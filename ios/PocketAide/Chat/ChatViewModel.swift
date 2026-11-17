import Foundation
import PocketAideAPI

@MainActor
final class ChatViewModel: ObservableObject {
    @Published private(set) var conversation: ChatConversation?
    @Published private(set) var messages: [ChatMessage] = []
    @Published private(set) var pending: ChatPending?
    @Published private(set) var hasLoaded = false
    @Published private(set) var isSending = false
    @Published private(set) var isRefreshing = false
    @Published private(set) var isSyncing = false
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
    var canSend: Bool { hasLoaded && !isSending && !isReplying && !isRefreshing }
    var canRefresh: Bool { canSend && !messages.isEmpty }
    // A plain re-read of the conversation, so it stays available even while a
    // reply is in progress or sending — only guarded against itself (rapid taps).
    var canSync: Bool { hasLoaded && !isSyncing }

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
            let sent = try await api.sendChatMessage(text)
            messages.append(sent)
            if pending == nil { pending = ChatPending() }
            return true
        } catch {
            self.error = ChatError(error)
            return false
        }
    }

    // Pulls the conversation right away instead of waiting for the next poll
    // tick — for history the server added while the tab wasn't visible. The
    // background poll loop keeps calling load() on its own schedule the whole
    // time; both calls are plain reads that each replace state with the
    // server's current truth, so running at the same time is harmless.
    @discardableResult
    func sync() async -> Bool {
        guard canSync else { return false }
        isSyncing = true
        defer { isSyncing = false }
        await load()
        return true
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
