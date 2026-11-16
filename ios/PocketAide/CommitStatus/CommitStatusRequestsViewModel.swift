import Foundation
import PocketAideAPI

@MainActor
final class CommitStatusRequestsViewModel: ObservableObject {
    @Published private(set) var items: [CommitStatusRequest] = []
    @Published private(set) var history: [CommitStatusRequest] = []
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
            let list = try await api.listCommitStatusRequests()
            items = list.items
            history = list.history
            errorMessage = nil
            hasLoaded = true
        } catch {
            errorMessage = String(describing: error)
        }
    }
}

@MainActor
final class CommitStatusDetailViewModel: ObservableObject {
    @Published private(set) var item: CommitStatusRequest
    @Published private(set) var isDeciding = false
    @Published var notice: String?
    @Published var loadError: String?

    private let api: APIClient?

    init(item: CommitStatusRequest, api: APIClient?) {
        self.item = item
        self.api = api
    }

    var canDecide: Bool {
        item.status == .pending && !isDeciding
    }

    func reload() async {
        guard let api else { return }
        do {
            item = try await api.commitStatusRequest(id: item.id)
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
            switch try await api.decideCommitStatusRequest(id: item.id, decision) {
            case .decided(let updated):
                item = updated
                notice = nil
            case .conflict(let current):
                item = current
                notice = CommitStatusFormatting.conflictNotice(current)
            case .writeFailed(let current, let message):
                item = current
                notice = CommitStatusFormatting.writeFailedNotice(message)
            case .unavailable:
                notice = CommitStatusFormatting.unavailableNotice
            }
        } catch {
            notice = CommitStatusFormatting.failedDecisionNotice
        }
    }
}
