import Foundation
import PocketAideAPI

/// Drives the PR-monitor tab (PRD-10 AC11/AC12). Loads `notification_history`
/// for the signed-in user, exposes an acknowledge action (optimistic on the
/// local row), and manages the "제외 레포" sheet via the `user_excluded_repos`
/// API. Push-arrival highlight is driven externally — the parent screen
/// passes a `highlightedEventID` that auto-clears after 5 seconds.
@MainActor
final class PRMonitorViewModel: ObservableObject {
    @Published private(set) var items: [NotificationHistoryItem] = []
    @Published private(set) var excludedRepos: [ExcludedRepo] = []
    @Published private(set) var isLoading = false
    @Published private(set) var isLoadingExcluded = false
    @Published var errorMessage: String?
    @Published var excludedRepoError: String?

    private(set) var api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    func load() async {
        guard let api else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            items = try await api.listNotificationHistory()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func loadExcludedRepos() async {
        guard let api else { return }
        isLoadingExcluded = true
        excludedRepoError = nil
        defer { isLoadingExcluded = false }
        do {
            excludedRepos = try await api.listExcludedRepos()
        } catch {
            excludedRepoError = String(describing: error)
        }
    }

    /// AC12: trigger by the explicit "확인" button only. Optimistic UI — set
    /// the timestamp immediately so the card flips visual state without a
    /// network round-trip; revert on failure.
    func acknowledge(id: Int64) async {
        guard let api else { return }
        guard let idx = items.firstIndex(where: { $0.id == id }) else { return }
        guard items[idx].acknowledgedAt == nil else { return }

        let original = items[idx]
        let stamped = NotificationHistoryItem(
            id: original.id,
            repoFullName: original.repoFullName,
            prNumber: original.prNumber,
            prTitle: original.prTitle,
            prURL: original.prURL,
            commitURL: original.commitURL,
            runURL: original.runURL,
            workflowName: original.workflowName,
            headBranch: original.headBranch,
            conclusion: original.conclusion,
            acknowledgedAt: Int64(Date().timeIntervalSince1970),
            createdAt: original.createdAt
        )
        items[idx] = stamped
        do {
            try await api.acknowledgeNotification(id: id)
        } catch {
            // Revert on failure so the unacked card returns and the user can
            // retry. The error is surfaced via errorMessage; we don't replace
            // the existing items load failure if any.
            if let revertIdx = items.firstIndex(where: { $0.id == id }) {
                items[revertIdx] = original
            }
            errorMessage = String(describing: error)
        }
    }

    func excludeRepo(_ repoFullName: String) async {
        guard let api else { return }
        excludedRepoError = nil
        do {
            let created = try await api.addExcludedRepo(repoFullName)
            excludedRepos.insert(created, at: 0)
        } catch {
            excludedRepoError = String(describing: error)
        }
    }

    func removeExcludedRepo(id: Int64) async {
        guard let api else { return }
        excludedRepoError = nil
        let prior = excludedRepos
        excludedRepos.removeAll { $0.id == id }
        do {
            try await api.removeExcludedRepo(id: id)
        } catch {
            excludedRepos = prior
            excludedRepoError = String(describing: error)
        }
    }
}
