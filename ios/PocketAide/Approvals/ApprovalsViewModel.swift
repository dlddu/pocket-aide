import Foundation
import PocketAideAPI

@MainActor
final class ApprovalsViewModel: ObservableObject {
    @Published private(set) var items: [ApprovalRequest] = []
    @Published private(set) var isLoading = false
    @Published private(set) var hasLoaded = false
    @Published var errorMessage: String?

    private(set) var api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    var pendingCount: Int { items.count }

    func load() async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            items = try await api.listPendingApprovals()
            errorMessage = nil
            hasLoaded = true
        } catch {
            errorMessage = String(describing: error)
        }
    }
}
