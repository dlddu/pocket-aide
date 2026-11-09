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
    @Published var notice: String?
    @Published private(set) var error: ChatError?

    static let replyPollNanos: UInt64 = 1_500_000_000
    static let idlePollNanos: UInt64 = 15_000_000_000

    private var api: APIClient?

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

    func pollWhileVisible() async {
        while !Task.isCancelled {
            await load()
            try? await Task.sleep(nanoseconds: isReplying ? Self.replyPollNanos : Self.idlePollNanos)
        }
    }

    func load() async {
        guard let api else { return }
        do {
            let view = try await api.chat()
            conversation = view.conversation
            messages = view.messages
            pending = view.pending
            queued = view.queued
            hasLoaded = true
            if error == .unreachable { error = nil }
        } catch is CancellationError {
        } catch {
            hasLoaded = true
            self.error = ChatError(error)
        }
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

    /// Drops the message waiting for after the reply in progress. A 404
    /// means it was already dispatched just as the user cancelled — not an
    /// error worth showing, since it is about to appear as a sent message.
    func cancelQueued() async {
        guard let api, queued != nil else { return }
        do {
            try await api.cancelQueuedChatMessage()
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
