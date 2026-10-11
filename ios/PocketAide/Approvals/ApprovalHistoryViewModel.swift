import Foundation
import PocketAideAPI

@MainActor
final class ApprovalHistoryViewModel: ObservableObject {
    @Published private(set) var items: [ApprovalRequest] = []
    @Published private(set) var isLoading = false
    @Published private(set) var hasLoaded = false
    @Published private(set) var nextCursor: String?
    @Published var errorMessage: String?

    private let api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    var hasMore: Bool { nextCursor != nil }

    func load() async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let page = try await api.approvalHistory()
            items = page.items
            nextCursor = page.nextCursor
            errorMessage = nil
            hasLoaded = true
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func loadMore() async {
        guard let api, !isLoading, let cursor = nextCursor else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let page = try await api.approvalHistory(cursor: cursor)
            items += page.items
            nextCursor = page.nextCursor
            errorMessage = nil
        } catch {
            errorMessage = String(describing: error)
        }
    }
}
