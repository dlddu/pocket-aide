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

    /// Rows keep the unread state they were fetched with, so 「새 알림」 stays for this viewing after the server marks them seen.
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
            await markShownAsRead(page, api: api)
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

    private func markShownAsRead(_ page: ExternalNotificationsPage, api: APIClient) async {
        guard page.unreadCount > 0, let newest = page.items.first else { return }
        do {
            try await api.markExternalNotificationsRead(upTo: newest.id)
            onMarkedRead()
        } catch {
            // Left unread on the server; the next load marks them again.
        }
    }
}
