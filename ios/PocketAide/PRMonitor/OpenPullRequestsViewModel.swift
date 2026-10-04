import Foundation
import PocketAideAPI
import PocketAideStorage

@MainActor
final class OpenPullRequestsViewModel: ObservableObject {
    @Published private(set) var login: String?
    @Published private(set) var pullRequests: [OpenPullRequest] = []
    @Published private(set) var inaccessibleCount = 0
    @Published private(set) var lastRefreshedAt: Date?
    @Published private(set) var isLoading = false
    @Published private(set) var hasLoaded = false
    @Published private(set) var isConnecting = false
    @Published private(set) var lastError: GitHubError?
    @Published var errorMessage: String?
    @Published var connectError: String?
    @Published var filter: OpenPullRequestFilter {
        didSet { defaults.set(filter.rawValue, forKey: Self.filterKey) }
    }

    static let filterKey = "openprs.filter"

    private let store: GitHubCredentialStoring
    private let client: GitHubClient
    private let defaults: UserDefaults

    init(
        store: GitHubCredentialStoring = KeychainGitHubCredentialStore(),
        client: GitHubClient = GitHubClient(),
        defaults: UserDefaults = .standard
    ) {
        self.store = store
        self.client = client
        self.defaults = defaults
        self.login = (try? store.load())?.login
        self.filter = defaults.string(forKey: Self.filterKey).flatMap(OpenPullRequestFilter.init(rawValue:)) ?? .all
    }

    var isConnected: Bool { login != nil }

    var visiblePullRequests: [OpenPullRequest] { filter.apply(to: pullRequests) }

    var alert: GitHubAlert? { GitHubAlert.make(error: lastError, inaccessibleCount: inaccessibleCount) }

    func connect(token raw: String) async -> Bool {
        let token = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !token.isEmpty else { return false }
        isConnecting = true
        connectError = nil
        defer { isConnecting = false }
        do {
            let viewer = try await client.viewerLogin(token: token)
            try store.save(GitHubCredential(token: token, login: viewer))
            resetList()
            login = viewer
        } catch {
            connectError = Self.message(for: error)
            return false
        }
        await refresh()
        return true
    }

    func disconnect() {
        try? store.clear()
        login = nil
        resetList()
    }

    func refresh() async {
        guard let credential = try? store.load() else {
            login = nil
            resetList()
            return
        }
        isLoading = true
        errorMessage = nil
        lastError = nil
        defer { isLoading = false }
        do {
            let result = try await client.openPullRequests(token: credential.token, login: credential.login)
            pullRequests = result.pullRequests
            inaccessibleCount = result.inaccessibleCount
            lastRefreshedAt = Date()
            hasLoaded = true
        } catch {
            lastError = error as? GitHubError
            errorMessage = Self.message(for: error)
        }
    }

    private func resetList() {
        pullRequests = []
        inaccessibleCount = 0
        lastRefreshedAt = nil
        hasLoaded = false
        errorMessage = nil
        lastError = nil
    }

    private static func message(for error: Error) -> String {
        (error as? GitHubError)?.description ?? String(describing: error)
    }
}
