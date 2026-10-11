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

@MainActor
final class ApprovalDetailViewModel: ObservableObject {
    @Published private(set) var item: ApprovalRequest
    @Published private(set) var isDeciding = false
    @Published var notice: String?
    @Published var loadError: String?

    private let api: APIClient?

    init(item: ApprovalRequest, api: APIClient?) {
        self.item = item
        self.api = api
    }

    func canDecide(at now: Date) -> Bool {
        item.status == .pending && !item.isExpired(at: now) && !isDeciding
    }

    func reload() async {
        guard let api else { return }
        do {
            item = try await api.approval(id: item.id)
            loadError = nil
        } catch {
            loadError = String(describing: error)
        }
    }

    func decide(_ decision: ApprovalDecision) async {
        guard let api, !isDeciding else { return }
        isDeciding = true
        defer { isDeciding = false }
        do {
            switch try await api.decideApproval(id: item.id, decision) {
            case .decided(let updated):
                item = updated
                notice = nil
            case .conflict(let current):
                item = current
                notice = ApprovalFormatting.conflictNotice(current)
            }
        } catch {
            notice = ApprovalFormatting.failedDecisionNotice
        }
    }
}
