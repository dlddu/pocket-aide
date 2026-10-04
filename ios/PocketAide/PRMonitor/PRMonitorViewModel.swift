import Foundation
import PocketAideAPI

@MainActor
final class PRMonitorViewModel: ObservableObject {
    @Published private(set) var items: [NotificationHistoryItem] = []
    @Published private(set) var excludedRepos: [ExcludedRepo] = []
    @Published private(set) var isLoading = false
    @Published private(set) var isLoadingExcluded = false
    @Published var errorMessage: String?
    @Published var excludedRepoError: String?
    @Published private(set) var notificationSettings = NotificationSettings()
    @Published private(set) var isLoadingSettings = false
    @Published var settingsError: String?

    var groups: [HistoryGroup] {
        HistoryGrouping.group(items)
    }

    var totalUnacknowledgedCount: Int {
        items.reduce(0) { $0 + ($1.acknowledgedAt == nil ? 1 : 0) }
    }

    var unacknowledgedGroups: [HistoryGroup] {
        groups.filter { !$0.allAcknowledged }
    }

    var acknowledgedGroups: [HistoryGroup] {
        groups.filter { $0.allAcknowledged }
    }

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

    func loadNotificationSettings() async {
        guard let api else { return }
        isLoadingSettings = true
        settingsError = nil
        defer { isLoadingSettings = false }
        do {
            notificationSettings = try await api.notificationSettings()
        } catch {
            settingsError = String(describing: error)
        }
    }

    func setNotificationsEnabled(_ enabled: Bool) async {
        await updateNotificationSettings(NotificationSettingsPatch(enabled: enabled)) {
            $0.enabled = enabled
        }
    }

    func setNotificationOutcomes(_ outcomes: NotificationOutcomes) async {
        await updateNotificationSettings(NotificationSettingsPatch(outcomes: outcomes)) {
            $0.outcomes = outcomes
        }
    }

    private func updateNotificationSettings(
        _ patch: NotificationSettingsPatch,
        apply: (inout NotificationSettings) -> Void
    ) async {
        guard let api else { return }
        settingsError = nil
        let prior = notificationSettings
        apply(&notificationSettings)
        do {
            notificationSettings = try await api.updateNotificationSettings(patch)
        } catch {
            notificationSettings = prior
            settingsError = String(describing: error)
        }
    }

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
            headSHA: original.headSHA,
            conclusion: original.conclusion,
            acknowledgedAt: Int64(Date().timeIntervalSince1970),
            createdAt: original.createdAt
        )
        items[idx] = stamped
        do {
            try await api.acknowledgeNotification(id: id)
            WidgetRefresher.reloadAll()
        } catch {
            if let revertIdx = items.firstIndex(where: { $0.id == id }) {
                items[revertIdx] = original
            }
            errorMessage = String(describing: error)
        }
    }

    func acknowledgeGroup(_ group: HistoryGroup) async {
        let unackedIDs = group.items.compactMap { $0.acknowledgedAt == nil ? $0.id : nil }
        for id in unackedIDs {
            await acknowledge(id: id)
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
