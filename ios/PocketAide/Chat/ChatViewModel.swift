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
    /// Whether the conversation goes on before what is held, and the id to ask
    /// for that page by — the newest page's word, never an increment's. An
    /// incremental read carries nothing about the far past, and taking its
    /// silence for "the conversation starts here" would take 「이전 대화 더
    /// 보기」 away from somebody who has not reached the start.
    @Published private(set) var hasOlder = false
    @Published private(set) var isLoadingOlder = false
    /// Kept apart from `error` because the two are about different things: one
    /// read failed to fetch the past, which leaves what is on screen exactly as
    /// it was and is retried by tapping again; the other is the conversation
    /// itself failing to read.
    @Published private(set) var olderError: ChatError?
    /// The message that was at the top before the last page was prepended, so
    /// the screen can put it back where it was. Non-nil for exactly one change
    /// of `messages`; the view takes it with `consumePrependAnchor()`.
    @Published private(set) var prependAnchor: Int64?
    @Published var notice: String?
    @Published private(set) var error: ChatError?

    static let replyPollNanos: UInt64 = 1_500_000_000
    static let idlePollNanos: UInt64 = 15_000_000_000
    /// One page of the conversation. Enough that opening the screen shows more
    /// than fits on it — so the first thing somebody does is read, not fetch —
    /// and small enough that it is a page rather than the history
    /// (PRD-AI채팅.AC1, 2026-11-28 개정). The backend caps it at 200.
    static let pageSize = 30

    private var api: APIClient?
    // The newest message id a read has delivered — not one appended by send(),
    // which could be newer than a reply the polls have not fetched yet.
    private var cursor: Int64?
    private var etag: String?
    // The cursor for the page before the oldest message held.
    private var olderCursor: Int64 = 0
    // readSeq orders the conversation's reads. The poll loop and the sync
    // button both call load(), so two reads are in the air whenever somebody
    // taps while a poll is out; a reply that is no longer the newest read's is
    // dropped rather than applied, because its pending/queued would put the
    // screen back to a state it has already left.
    private var readSeq: Int64 = 0
    // olderReadSeq does the same for the reads of the past. A first page moves
    // it too, so a page that was in the air when the conversation was replaced
    // is not prepended to the new one.
    private var olderReadSeq: Int64 = 0

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
    // Reading the past is guarded only against itself: a page before the oldest
    // message held does not change with a reply coming in, so there is no
    // reason to make somebody wait for one to finish before reading back.
    var canLoadOlder: Bool { hasOlder && olderCursor > 0 && !isLoadingOlder }

    func pollWhileVisible() async {
        while !Task.isCancelled {
            await load()
            try? await Task.sleep(nanoseconds: isReplying ? Self.replyPollNanos : Self.idlePollNanos)
        }
    }

    /// Reads the conversation: only what came after the last read, and nothing
    /// at all when the server answers 304. Its newest page is read when there is
    /// no read to continue from — before the first one, and whenever an answer
    /// turns out to be a different conversation than the one on screen, whose
    /// ids this cursor says nothing about.
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
                    try await loadFirstPage(api, mine)
                }
            } else {
                try await loadFirstPage(api, mine)
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

    /// The conversation's newest page, which is where the screen starts and
    /// where it starts over when the conversation turns out to be another one.
    ///
    /// Starting over drops the older pages that were read: they belong to the
    /// conversation that is no longer on screen, and the page just read says
    /// from scratch how far back this one goes.
    private func loadFirstPage(_ api: APIClient, _ mine: Int64) async throws {
        // Checked before the request, not only after it: a superseded read that
        // was answered with another conversation would otherwise fall back to
        // reading that conversation's page just to throw the answer away. The
        // newest read is already reading it.
        guard mine == readSeq else { return }
        let view = try await api.chatLatest(limit: Self.pageSize)
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
        guard !merging else { return }
        // A first page also ends whatever read of the past was in the air and
        // clears its failure: that read was about the conversation being
        // replaced.
        olderReadSeq += 1
        hasOlder = view.hasMore ?? false
        olderCursor = view.prevCursor ?? 0
        isLoadingOlder = false
        olderError = nil
        prependAnchor = nil
    }

    /// Reads the page before the oldest message held and puts it in front of
    /// them. True when a page was read.
    ///
    /// The page is bounded by a cursor, which is a message id, so it is the
    /// same page no matter what arrives in the conversation meanwhile — a reply
    /// landing mid-read gets an id above everything held and comes down the
    /// poll's `?after` read instead. Merging by id on top of that leaves no
    /// room for a gap, a repeat, or a reordering: the retry of a page whose
    /// answer never arrived is the same page, and applying it twice is applying
    /// it once.
    @discardableResult
    func loadOlder() async -> Bool {
        guard let api, canLoadOlder else { return false }
        olderReadSeq += 1
        let mine = olderReadSeq
        let from = olderCursor
        // What is on screen stays on screen while the past is fetched: this
        // read is somebody asking for *more*, not for what they have to be
        // taken away and given back.
        isLoadingOlder = true
        olderError = nil
        do {
            let page = try await api.chatMessages(before: from, limit: Self.pageSize)
            // A page that lost its race belongs to a conversation the screen
            // has left, or to a cursor it has already read past.
            guard mine == olderReadSeq else { return false }
            // Taken before the merge: it is the message the screen is anchored
            // on, and after the merge it is no longer the first. Only armed
            // when the page actually brought something — the view takes the
            // anchor on the next change of the list, and a page that added
            // nothing produces no such change, which would leave the anchor
            // waiting to pull some later arrival back up here.
            if !page.messages.isEmpty {
                prependAnchor = messages.first?.id
            }
            messages = messages.merging(page.messages)
            hasOlder = page.hasMore ?? false
            // Keep nothing to retry once the start of the conversation is
            // reached: an empty page before a cursor that still promised more
            // would otherwise be offered again.
            olderCursor = (page.hasMore ?? false) ? (page.prevCursor ?? 0) : 0
            isLoadingOlder = false
            return true
        } catch is CancellationError {
            guard mine == olderReadSeq else { return false }
            isLoadingOlder = false
            return false
        } catch {
            guard mine == olderReadSeq else { return false }
            // The failure is the read's, not the conversation's: the messages,
            // the reply in progress and the cursor all stay as they are, so the
            // same tap retries the same page.
            isLoadingOlder = false
            olderError = ChatError(error)
            return false
        }
    }

    /// The anchor to put back on screen, once. Returning it clears it, so the
    /// next change of `messages` follows the end again (AC9) instead of
    /// jumping back to where the last page landed.
    func consumePrependAnchor() -> Int64? {
        defer { prependAnchor = nil }
        return prependAnchor
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
