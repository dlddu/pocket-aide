import Foundation

public enum AgentSessionState: String, Codable, Sendable, Hashable {
    case active
    case idle
    case snapshot

    public var label: String {
        switch self {
        case .active: return "실행 중"
        case .idle: return "대기"
        case .snapshot: return "스냅샷"
        }
    }
}

public struct AgentSessionCheckpoint: Codable, Equatable, Sendable, Hashable {
    public let sizeBytes: Int64?
    public let createdAt: String?

    public init(sizeBytes: Int64?, createdAt: String?) {
        self.sizeBytes = sizeBytes
        self.createdAt = createdAt
    }
}

public struct AgentSession: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: String
    public let name: String
    public let state: AgentSessionState
    public let model: String?
    public let createdAt: String
    public let lastAccess: String
    public let checkpoint: AgentSessionCheckpoint?

    public init(
        id: String,
        name: String,
        state: AgentSessionState,
        model: String? = nil,
        createdAt: String,
        lastAccess: String,
        checkpoint: AgentSessionCheckpoint? = nil
    ) {
        self.id = id
        self.name = name
        self.state = state
        self.model = model
        self.createdAt = createdAt
        self.lastAccess = lastAccess
        self.checkpoint = checkpoint
    }

    public var modelLabel: String { model.flatMap { $0.isEmpty ? nil : $0 } ?? "platform-default" }
}

public struct AgentSessionConfig: Decodable, Equatable, Sendable {
    public let defaultModel: String
    public let models: [String]

    public init(defaultModel: String, models: [String]) {
        self.defaultModel = defaultModel
        self.models = models
    }

    enum RootKeys: String, CodingKey { case claudeCode }
    enum CodingKeys: String, CodingKey { case defaultModel, models }

    public init(from decoder: Decoder) throws {
        let root = try decoder.container(keyedBy: RootKeys.self)
        let claude = try root.nestedContainer(keyedBy: CodingKeys.self, forKey: .claudeCode)
        defaultModel = try claude.decodeIfPresent(String.self, forKey: .defaultModel) ?? ""
        models = try claude.decodeIfPresent([String].self, forKey: .models) ?? []
    }
}

public struct AgentSessionReadResult: Decodable, Equatable, Sendable {
    public let session: AgentSession?
    public let payload: String
    public let nextOffset: Int64

    public init(session: AgentSession?, payload: String, nextOffset: Int64) {
        self.session = session
        self.payload = payload
        self.nextOffset = nextOffset
    }
}

public struct AgentSessionWriteResult: Decodable, Equatable, Sendable {
    public let session: AgentSession?
}

public struct AgentSessionOutput: Equatable, Sendable {
    public static let displayLimit = 120_000

    public private(set) var text = ""
    public private(set) var nextOffset: Int64 = 0
    public private(set) var isTrimmed = false

    public init() {}

    public enum Outcome: Equatable, Sendable {
        case grew
        case unchanged
        case needsFullReread
    }

    public mutating func apply(_ result: AgentSessionReadResult, requestedOffset: Int64) -> Outcome {
        if result.nextOffset < requestedOffset {
            return .needsFullReread
        }
        nextOffset = result.nextOffset
        if requestedOffset == 0 {
            let previous = text
            text = ""
            isTrimmed = false
            append(result.payload)
            return text == previous ? .unchanged : .grew
        }
        guard !result.payload.isEmpty else { return .unchanged }
        append(result.payload)
        return .grew
    }

    private mutating func append(_ payload: String) {
        text += payload
        if text.count > Self.displayLimit {
            text = String(text.suffix(Self.displayLimit))
            isTrimmed = true
        }
    }
}

public enum AgentSessionResponsePhase: Equatable, Sendable {
    case none
    case queued
    case streaming
    case quiet

    public static let quietAfter: TimeInterval = 10

    public var label: String? {
        switch self {
        case .none: return nil
        case .queued: return "전송됨 · 응답 대기 중"
        case .streaming: return "응답 진행 중"
        case .quiet: return "새 출력이 멈췄어요 · 응답이 끝났을 수 있어요"
        }
    }

    public func next(outputGrew: Bool, secondsSinceLastOutput: TimeInterval) -> AgentSessionResponsePhase {
        switch self {
        case .none:
            return .none
        case .queued:
            return outputGrew ? .streaming : .queued
        case .streaming, .quiet:
            if outputGrew { return .streaming }
            return secondsSinceLastOutput >= Self.quietAfter ? .quiet : self
        }
    }
}

public enum AgentSessionPrompt {
    public static let maxBytes = 1 << 20

    public static func exceedsLimit(_ prompt: String) -> Bool {
        prompt.utf8.count > maxBytes
    }
}

public enum AgentSessionError: Hashable, Sendable {
    case unauthorized
    case notFound
    case busy
    case tooLarge
    case snapshotted
    case queueFull
    case disabled
    case quotaFull
    case unreachable
    case unknown

    public init(_ error: Error) {
        guard let apiError = error as? APIError else {
            self = .unreachable
            return
        }
        switch apiError {
        case .transport, .missingBaseURL:
            self = .unreachable
        case .decoding:
            self = .unknown
        case .badStatus(let status, _):
            self = Self.fromStatus(status)
        }
    }

    static let byStatus: [Int: AgentSessionError] = [
        401: .unauthorized,
        404: .notFound,
        409: .busy,
        413: .tooLarge,
        422: .snapshotted,
        429: .queueFull,
        503: .disabled,
        507: .quotaFull,
        502: .unreachable,
        504: .unreachable,
    ]

    static func fromStatus(_ status: Int) -> AgentSessionError {
        byStatus[status] ?? .unknown
    }

    private static let messages: [AgentSessionError: String] = [
        .unauthorized: "로그인이 만료됐어요. 다시 로그인해 주세요.",
        .notFound: "세션을 찾을 수 없어요. 목록을 새로고침해 주세요.",
        .busy: "세션이 다른 작업 중이에요. 잠시 후 다시 시도해 주세요.",
        .tooLarge: "프롬프트가 너무 커요. 1 MiB 이하로 줄여 주세요.",
        .snapshotted: "스냅샷 상태인 세션이에요. 복원한 뒤 이어서 쓸 수 있어요.",
        .queueFull: "앞선 프롬프트가 아직 처리 중이에요. 끝난 뒤 다시 보내 주세요.",
        .disabled: "이 환경에서는 꺼져 있는 기능이에요. 세션은 그대로 유지돼요.",
        .quotaFull: "세션 출력 한도가 찼어요. 기존 출력은 볼 수 있지만 새 프롬프트는 보낼 수 없어요.",
        .unreachable: "세션 플랫폼에 연결할 수 없어요. 네트워크를 확인하고 다시 시도해 주세요.",
    ]

    public var message: String {
        Self.messages[self] ?? "세션 플랫폼에서 오류가 났어요. 잠시 후 다시 시도해 주세요."
    }
}

public enum AgentSessionFormatting {
    public static func date(_ raw: String?) -> Date? {
        guard let raw, !raw.isEmpty else { return nil }
        let trimmed = raw.replacingOccurrences(of: #"\.\d+"#, with: "", options: .regularExpression)
        return ISO8601DateFormatter().date(from: trimmed)
    }

    public static func relative(_ raw: String?, now: Date = Date()) -> String {
        guard let date = date(raw) else { return "—" }
        if now.timeIntervalSince(date) < 60 { return "방금" }
        let formatter = RelativeDateTimeFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.unitsStyle = .short
        return formatter.localizedString(for: date, relativeTo: now)
    }

    public static func size(_ bytes: Int64?) -> String? {
        guard let bytes else { return nil }
        return ByteCountFormatter.string(fromByteCount: bytes, countStyle: .file)
    }
}

struct AgentSessionListResponse: Decodable {
    let sessions: [AgentSession]
}

struct AgentSessionCreatePayload: Encodable {
    let name: String
    let model: String?
}

struct AgentSessionReadPayload: Encodable {
    let offset: Int64
}

struct AgentSessionWritePayload: Encodable {
    let payload: String
}

struct AgentSessionEmptyPayload: Encodable {}

public extension APIClient {
    func agentSessionConfig() async throws -> AgentSessionConfig {
        try await get("/api/sessions/config", authenticated: true)
    }

    func listAgentSessions() async throws -> [AgentSession] {
        let response: AgentSessionListResponse = try await get("/api/sessions", authenticated: true)
        return response.sessions
    }

    func createAgentSession(name: String, model: String?) async throws -> AgentSession {
        let trimmedModel = model?.trimmingCharacters(in: .whitespacesAndNewlines)
        let payload = AgentSessionCreatePayload(
            name: name.trimmingCharacters(in: .whitespacesAndNewlines),
            model: (trimmedModel?.isEmpty ?? true) ? nil : trimmedModel
        )
        return try await post("/api/sessions", body: payload, authenticated: true)
    }

    func agentSession(id: String) async throws -> AgentSession {
        try await get("/api/sessions/\(id)", authenticated: true)
    }

    func readAgentSession(id: String, offset: Int64) async throws -> AgentSessionReadResult {
        try await post("/api/sessions/\(id)/read", body: AgentSessionReadPayload(offset: offset), authenticated: true)
    }

    func writeAgentSession(id: String, prompt: String) async throws -> AgentSessionWriteResult {
        try await post("/api/sessions/\(id)/write", body: AgentSessionWritePayload(payload: prompt), authenticated: true)
    }

    func restoreAgentSession(id: String) async throws -> AgentSession {
        try await post("/api/sessions/\(id)/switch", body: AgentSessionEmptyPayload(), authenticated: true)
    }
}
