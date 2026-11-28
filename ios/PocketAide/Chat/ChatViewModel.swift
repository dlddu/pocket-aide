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

    static let replyPollNanos: UInt64 = 1_500_000_000
    static let idlePollNanos: UInt64 = 15_000_000_000

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

    func pollWhileVisible() async {
        while !Task.isCancelled {
            await load()
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
