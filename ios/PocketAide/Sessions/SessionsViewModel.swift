import Foundation
import PocketAideAPI

@MainActor
final class SessionsViewModel: ObservableObject {
    @Published private(set) var sessions: [AgentSession] = []
    @Published private(set) var isLoading = false
    @Published private(set) var hasLoaded = false
    @Published var error: AgentSessionError?

    @Published private(set) var config: AgentSessionConfig?
    @Published private(set) var isCreating = false
    @Published var createError: AgentSessionError?

    private(set) var api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    func load() async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            sessions = try await api.listAgentSessions()
            error = nil
        } catch {
            self.error = AgentSessionError(error)
        }
        hasLoaded = true
    }

    func loadConfig() async {
        guard let api else { return }
        config = try? await api.agentSessionConfig()
    }

    func create(name: String, model: String?) async -> AgentSession? {
        guard let api, !isCreating else { return nil }
        isCreating = true
        createError = nil
        defer { isCreating = false }
        do {
            let created = try await api.createAgentSession(name: name, model: model)
            sessions.removeAll { $0.id == created.id }
            sessions.insert(created, at: 0)
            return created
        } catch {
            createError = AgentSessionError(error)
            return nil
        }
    }
}
