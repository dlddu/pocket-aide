import Foundation
import PocketAideAPI

@MainActor
final class ChatViewModel: ObservableObject {
    @Published private(set) var conversation: ChatConversation?
    @Published private(set) var messages: [ChatMessage] = []
    @Published private(set) var pending: ChatPending?
    @Published private(set) var queued: ChatQueued?
    @Published private(set) var hasLoaded = false
    @Published private(set) var isSending = false
    @Published private(set) var isRefreshing = false
    @Published private(set) var isSyncing = false
    @Published var notice: String?
    @Published private(set) var error: ChatError?
    // Whether the session answering this conversation is on (PRD-1 AC12), and
    // the turn-on's own in-flight flag and failure — kept apart from the
    // conversation's so a failed status read never looks like a failed send,
    // and neither hides the other.
    @Published private(set) var sessionPower: ChatSessionPower?
    @Published private(set) var isPoweringOn = false
    @Published private(set) var powerError: ChatPowerError?

    static let replyPollNanos: UInt64 = 1_500_000_000
    static let idlePollNanos: UInt64 = 15_000_000_000
    /// How stale the session's state may get while the screen is open. The
    /// state changes on the platform's own schedule (a 60 minute idle freeze),
    /// not with the conversation, so it is read far less often than the
    /// messages — one read a minute per open screen rather than one per poll.
    static let powerRefreshSeconds: TimeInterval = 60

    private var api: APIClient?
    // The newest message id a read has delivered — not one appended by send(),
    // which could be newer than a reply the polls have not fetched yet.
    private var cursor: Int64?
    private var etag: String?
    // readSeq orders the conversation's reads. The poll loop and the sync
    // button both call load(), so two reads are in the air whenever somebody
    // taps while a poll is out; a reply that is no longer the newest read's is
    // dropped rather than applied, because its pending/queued would put the
    // screen back to a state it has already left.
    private var readSeq: Int64 = 0
    // When the session's state was last read, so the poll loop can refresh it
    // on its own slower schedule. Cleared by anything that changes the state —
    // a send restores an archived session (AC4) — so the next poll reads it
    // again rather than showing what was true before.
    private var powerReadAt: Date?

    init(api: APIClient?) {
        self.api = api
    }

    func replaceAPI(_ api: APIClient?) {
        self.api = api
    }

    var isReplying: Bool { pending != nil }
    // Sending while a reply is in progress queues instead of failing (server
    // decides), so only an already-queued message blocks the composer.
    var canSend: Bool { hasLoaded && !isSending && queued == nil && !isRefreshing }
    var canRefresh: Bool { hasLoaded && !isSending && !isReplying && !isRefreshing && !messages.isEmpty }
    // A plain re-read of the conversation, so it stays available even while a
    // reply is in progress, queued or sending — only guarded against itself
    // (rapid taps).
    var canSync: Bool { hasLoaded && !isSyncing }
    /// Only an archived session is turned on from here, and only one tap at a
    /// time (PRD-1 AC12).
    var canPowerOn: Bool { sessionPower?.power.canPowerOn == true && !isPoweringOn }

    func pollWhileVisible() async {
        while !Task.isCancelled {
            await load()
            await loadSessionPower()
            try? await Task.sleep(nanoseconds: isReplying ? Self.replyPollNanos : Self.idlePollNanos)
        }
    }

    /// Reads the conversation: only what came after the last read, and nothing
    /// at all when the server answers 304. The whole conversation is read when
    /// there is no read to continue from — before the first one, and whenever an
    /// answer turns out to be a different conversation than the one on screen,
    /// whose ids this cursor says nothing about.
    func load() async {
        guard let api else { return }
        readSeq += 1
        let mine = readSeq
        do {
            if let cursor, let known = conversation?.id {
                switch try await api.chat(after: cursor, ifNoneMatch: etag) {
                // Nothing changed, so nothing is applied — in particular the
                // messages on screen are not replaced by the empty body.
                case .notModified:
                    break
                case .fresh(let view, let tag) where view.conversation.id == known:
                    guard mine == readSeq else { return }
                    applyView(view, merging: true)
                    etag = tag
                case .fresh:
                    try await loadAll(api, mine)
                }
            } else {
                try await loadAll(api, mine)
            }
            guard mine == readSeq else { return }
            hasLoaded = true
            if error == .unreachable { error = nil }
        } catch is CancellationError {
        } catch {
            // A superseded read's failure is not the screen's: a newer read has
            // answered, and its outcome is the one the screen is showing.
            guard mine == readSeq else { return }
            hasLoaded = true
            self.error = ChatError(error)
        }
    }

    private func loadAll(_ api: APIClient, _ mine: Int64) async throws {
        // Checked before the request, not only after it: a superseded read that
        // was answered with another conversation would otherwise fall back to
        // reading that whole conversation — megabytes on a long one — just to
        // throw the answer away. The newest read is already reading it.
        guard mine == readSeq else { return }
        let view = try await api.chat()
        // Checked again: the newest read can change while this one is in the
        // air, and then this answer is not the screen's to apply.
        guard mine == readSeq else { return }
        applyView(view, merging: false)
        etag = nil
    }

    private func applyView(_ view: ChatThread, merging: Bool) {
        conversation = view.conversation
        messages = merging ? messages.merging(view.messages) : view.messages
        pending = view.pending
        queued = view.queued
        let newest = view.messages.last?.id ?? 0
        cursor = merging ? max(cursor ?? 0, newest) : newest
    }

    func send(_ content: String) async -> Bool {
        let text = content.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let api, !text.isEmpty, canSend else { return false }
        if text.utf8.count > ChatError.maxMessageBytes {
            error = .tooLarge
            return false
        }
        isSending = true
        error = nil
        notice = nil
        defer { isSending = false }
        do {
            switch try await api.sendChatMessage(text) {
            case .sent(let sent):
                messages.append(sent)
                if pending == nil { pending = ChatPending() }
            case .queued(let held):
                queued = held
            }
            // A send opens the session, or restores an archived one (AC4), so
            // whatever the chip was showing is no longer what is true.
            powerReadAt = nil
            return true
        } catch {
            self.error = ChatError(error)
            return false
        }
    }

    // Reads right away instead of waiting for the next poll tick, and reads the
    // same way a poll does: what came after the last read. A tap is therefore
    // the cost of one poll rather than of the whole conversation, which on a
    // long one was megabytes (PRD-AI채팅.AC1, 2026-11-28 개정). The background
    // poll loop keeps calling load() on its own schedule the whole time; the
    // two overlapping is what readSeq is for.
    @discardableResult
    func sync() async -> Bool {
        guard canSync else { return false }
        isSyncing = true
        defer { isSyncing = false }
        await load()
        return true
    }

    /// Drops the message waiting for after the reply in progress. A 404
    /// means it was already dispatched just as the user cancelled — not an
    /// error worth showing, since it is about to appear as a sent message.
    func cancelQueued() async {
        guard let api, let waiting = queued else { return }
        do {
            try await api.cancelQueuedChatMessage(id: waiting.id)
        } catch {
            if case .badStatus(404, _)? = error as? APIError {
            } else {
                self.error = ChatError(error)
                return
            }
        }
        queued = nil
    }

    /// Reads whether this conversation's session is on, at most once every
    /// `powerRefreshSeconds` unless `force` says somebody asked for it now —
    /// the chip's own retry, and right after a turn-on. A failure is kept in
    /// `powerError` and the last state is left on screen: a chip blanked by a
    /// network blip reads as "no session", which is a different fact.
    func loadSessionPower(force: Bool = false) async {
        guard let api else { return }
        if !force, let read = powerReadAt, Date().timeIntervalSince(read) < Self.powerRefreshSeconds {
            return
        }
        do {
            sessionPower = try await api.chatSessionPower()
            powerReadAt = Date()
            powerError = nil
        } catch is CancellationError {
        } catch {
            powerError = ChatPowerError(error)
        }
    }

    /// Turns this conversation's session back on. The same session comes back
    /// — the server restores the one its own row names — so the conversation
    /// carries on in it and no second session appears (PRD-1 AC12 · AC4).
    func powerOn() async {
        guard let api, canPowerOn else { return }
        isPoweringOn = true
        powerError = nil
        notice = nil
        defer { isPoweringOn = false }
        do {
            let result = try await api.powerOnChatSession()
            sessionPower = result
            // An accepted turn-on whose state never came back is not a state
            // to sit on: the next poll reads it again rather than leaving the
            // chip on "알 수 없음" for the whole refresh window.
            powerReadAt = result.power == .unknown ? nil : Date()
            notice = result.powerOnNotice
        } catch is CancellationError {
        } catch {
            powerError = ChatPowerError(error)
        }
    }

    func refreshPlugins() async {
        guard let api, canRefresh else { return }
        isRefreshing = true
        error = nil
        notice = nil
        defer { isRefreshing = false }
        do {
            notice = try await api.refreshChatPlugins().notice
        } catch {
            self.error = ChatError(error, refreshing: true)
        }
    }
}
