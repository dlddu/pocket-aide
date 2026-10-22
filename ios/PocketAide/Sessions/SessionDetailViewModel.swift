import Foundation
import PocketAideAPI

@MainActor
final class SessionDetailViewModel: ObservableObject {
    @Published private(set) var session: AgentSession
    @Published private(set) var output = AgentSessionOutput()
    @Published private(set) var phase: AgentSessionResponsePhase = .none
    @Published private(set) var hasLoadedOutput = false
    @Published private(set) var isSending = false
    @Published private(set) var isRestoring = false
    @Published var error: AgentSessionError?

    static let pollInterval: Duration = .seconds(2)
    static let stateRecheckAfter: TimeInterval = 30

    private let api: APIClient?
    private var lastOutputAt = Date()
    private var lastPollAt: Date?

    init(session: AgentSession, api: APIClient?) {
        self.session = session
        self.api = api
    }

    var canSend: Bool {
        session.state != .snapshot && !isSending && !isRestoring
    }

    func pollWhileVisible() async {
        while !Task.isCancelled {
            await tick()
            try? await Task.sleep(for: Self.pollInterval)
        }
    }

    func tick() async {
        let stale = lastPollAt.map { Date().timeIntervalSince($0) > Self.stateRecheckAfter } ?? true
        lastPollAt = Date()
        if stale {
            await refreshState()
        }
        guard session.state != .snapshot, !isRestoring else { return }
        await readOutput()
    }

    func refreshState() async {
        guard let api else { return }
        do {
            session = try await api.agentSession(id: session.id)
        } catch {
            self.error = AgentSessionError(error)
        }
    }

    func readOutput() async {
        guard let api else { return }
        let requested = output.nextOffset
        do {
            var result = try await api.readAgentSession(id: session.id, offset: requested)
            var outcome = output.apply(result, requestedOffset: requested)
            if outcome == .needsFullReread {
                result = try await api.readAgentSession(id: session.id, offset: 0)
                outcome = output.apply(result, requestedOffset: 0)
            }
            if let latest = result.session {
                session = latest
            }
            let grew = outcome == .grew && hasLoadedOutput
            if grew {
                lastOutputAt = Date()
            }
            phase = phase.next(outputGrew: grew, secondsSinceLastOutput: Date().timeIntervalSince(lastOutputAt))
            hasLoadedOutput = true
            error = nil
        } catch {
            let category = AgentSessionError(error)
            self.error = category
            if category == .snapshotted || category == .notFound {
                await refreshState()
            }
        }
    }

    func send(_ prompt: String) async -> Bool {
        guard let api, canSend else { return false }
        if AgentSessionPrompt.exceedsLimit(prompt) {
            error = .tooLarge
            return false
        }
        isSending = true
        defer { isSending = false }
        do {
            let result = try await api.writeAgentSession(id: session.id, prompt: prompt)
            if let latest = result.session {
                session = latest
            }
            phase = .queued
            lastOutputAt = Date()
            error = nil
            return true
        } catch {
            self.error = AgentSessionError(error)
            return false
        }
    }

    func restore() async {
        guard let api, session.state == .snapshot, !isRestoring else { return }
        isRestoring = true
        defer { isRestoring = false }
        do {
            session = try await api.restoreAgentSession(id: session.id)
            error = nil
        } catch {
            self.error = AgentSessionError(error)
        }
    }
}
