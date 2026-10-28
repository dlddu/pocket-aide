import Foundation
import PocketAideAPI

@MainActor
final class ExternalNotificationsViewModel: ObservableObject {
    @Published private(set) var items: [ExternalNotification] = []
    @Published private(set) var unreadCount = 0
    @Published private(set) var isLoading = false
    @Published private(set) var hasLoaded = false
    @Published private(set) var hasMore = false
    @Published var errorMessage: String?

    private(set) var api: APIClient?
    private let onMarkedRead: () -> Void

    init(api: APIClient?, onMarkedRead: @escaping () -> Void = {}) {
        self.api = api
        self.onMarkedRead = onMarkedRead
    }

    func replaceAPI(_ api: APIClient) {
        self.api = api
    }

    func load() async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let page = try await api.externalNotifications(limit: ExternalNotificationFormatting.pageSize)
            items = page.items
            unreadCount = page.unreadCount
            hasMore = page.items.count >= ExternalNotificationFormatting.pageSize
            errorMessage = nil
            hasLoaded = true
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func loadMore() async {
        guard let api, !isLoading, hasMore, let oldest = items.last else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let page = try await api.externalNotifications(limit: ExternalNotificationFormatting.pageSize, before: oldest.id)
            let known = Set(items.map(\.id))
            items += page.items.filter { !known.contains($0.id) }
            hasMore = page.items.count >= ExternalNotificationFormatting.pageSize
            errorMessage = nil
        } catch {
            errorMessage = String(describing: error)
        }
    }

    /// The notification a push tap names, which may not be in the loaded page.
    func notification(id: Int64) async -> ExternalNotification? {
        if let loaded = items.first(where: { $0.id == id }) {
            return loaded
        }
        guard let api else { return nil }
        do {
            return try await api.externalNotification(id: id)
        } catch {
            errorMessage = String(describing: error)
            return nil
        }
    }

    /// Opening the detail is what counts as seeing a notification (PRD-12 AC17).
    func markRead(_ item: ExternalNotification) async {
        guard item.isUnread, let api else { return }
        do {
            try await api.markExternalNotificationRead(id: item.id)
        } catch {
            return
        }
        let now = Int64(Date().timeIntervalSince1970)
        if let index = items.firstIndex(where: { $0.id == item.id }), items[index].isUnread {
            items[index] = items[index].markedRead(at: now)
            unreadCount = max(0, unreadCount - 1)
        }
        onMarkedRead()
    }
}
