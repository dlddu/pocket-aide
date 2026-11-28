import Foundation

// PRD-1 AC12: whether the codex session answering this conversation is on, and
// turning it back on when it is not. The backend asks the platform about that
// one session — the one the user's own conversation points at — and never
// reads its transcript, so this can be polled without restoring an archived
// session (PRD-11 AC5).

/// What the chat tab says about the session answering the conversation.
///
/// `unknown` is a real answer and not a parsing accident: the backend sends it
/// for a platform state it cannot place, and for a turn-on the platform took
/// without a state coming back. An answer this app does not recognise lands
/// here too, which is the same honest "we do not know".
public enum ChatSessionPowerState: String, Codable, Sendable, Hashable {
    /// No session yet — the first message opens one (AC4).
    case absent = "none"
    /// The workload is up: answering, or idling with its environment still there.
    case on
    /// Archived — the environment was handed back and a restore brings it back.
    case off
    /// The platform no longer has the session. The conversation is untouched.
    case gone
    case unknown

    public init(from decoder: Decoder) throws {
        let raw = try decoder.singleValueContainer().decode(String.self)
        self = ChatSessionPowerState(rawValue: raw) ?? .unknown
    }

    public var label: String {
        switch self {
        case .absent: return "세션 없음"
        case .on: return "세션 켜짐"
        case .off: return "세션 꺼짐"
        case .gone: return "세션 없음"
        case .unknown: return "세션 상태 알 수 없음"
        }
    }

    /// Whether the screen offers to turn the session on. Only an archived
    /// session is offered: there is nothing to restore in the others, and a
    /// button that would have to invent a session is not this feature.
    public var canPowerOn: Bool { self == .off }
}

public struct ChatSessionPower: Codable, Sendable, Hashable {
    public let power: ChatSessionPowerState
    /// The platform's own word for the state, when it had one — kept so an
    /// unknown state is shown as what it is instead of being guessed at.
    public let state: String?
    public let lastAccess: String?
    public let checkpoint: AgentSessionCheckpoint?
    /// True only when this answer came from a turn-on that actually restored
    /// the session, so "켰어요" is never said about one that was already on.
    public let restored: Bool

    enum CodingKeys: String, CodingKey {
        case power, state, checkpoint, restored
        case lastAccess = "last_access"
    }

    public init(
        power: ChatSessionPowerState,
        state: String? = nil,
        lastAccess: String? = nil,
        checkpoint: AgentSessionCheckpoint? = nil,
        restored: Bool = false
    ) {
        self.power = power
        self.state = state
        self.lastAccess = lastAccess
        self.checkpoint = checkpoint
        self.restored = restored
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        power = try c.decode(ChatSessionPowerState.self, forKey: .power)
        state = try c.decodeIfPresent(String.self, forKey: .state)
        lastAccess = try c.decodeIfPresent(String.self, forKey: .lastAccess)
        checkpoint = try c.decodeIfPresent(AgentSessionCheckpoint.self, forKey: .checkpoint)
        restored = try c.decodeIfPresent(Bool.self, forKey: .restored) ?? false
    }

    /// The line under the state, in the words the sessions tab uses for the
    /// same platform states (PRD-11 AC5) — and, where the platform gave us
    /// nothing to show, what happens next instead of a blank.
    public func detail(now: Date = Date()) -> String {
        switch power {
        case .absent:
            return "첫 메시지를 보내면 세션이 열려요"
        case .on:
            guard let lastAccess, !lastAccess.isEmpty else { return "환경이 떠 있어요" }
            return "환경이 떠 있어요 · 마지막 사용 \(AgentSessionFormatting.relative(lastAccess, now: now))"
        case .off:
            var parts: [String] = []
            if let created = checkpoint?.createdAt, !created.isEmpty {
                parts.append("아카이브 \(AgentSessionFormatting.relative(created, now: now))")
            }
            if let size = AgentSessionFormatting.size(checkpoint?.sizeBytes) {
                parts.append(size)
            }
            parts.append("환경 반납됨 · 켜면 이 대화가 그대로 이어져요")
            return parts.joined(separator: " · ")
        case .gone:
            return "세션이 사라졌어요 · 대화 기록은 그대로이고 다음 메시지가 새 세션을 열어요"
        case .unknown:
            guard let state, !state.isEmpty else { return "상태를 확인하지 못했어요" }
            return "플랫폼이 「\(state)」라고 알려 왔어요"
        }
    }

    /// What to say after a turn-on, which is not always "turned it on": the
    /// session may have been on already, or the platform may have taken the
    /// restore without a state coming back.
    public var powerOnNotice: String {
        switch (restored, power) {
        case (false, _):
            return "세션이 이미 켜져 있었어요."
        case (true, .on):
            return "세션을 켰어요. 이어서 보내면 같은 대화로 이어져요."
        default:
            return "켜기를 접수했어요. 켜졌는지는 아직 확인하지 못했어요 — 상태를 다시 읽어 주세요."
        }
    }
}

/// The refusals the chat tab explains differently. The backend's `code` is
/// what tells 409s apart, so a session held by the platform is never shown as
/// "send a message first".
public enum ChatPowerError: Hashable, Sendable {
    case noSession
    case gone
    case busy
    case slotsBusy
    case unreachable
    case failed

    public init(_ error: Error) {
        guard let apiError = error as? APIError else {
            self = .unreachable
            return
        }
        switch apiError {
        case .transport, .missingBaseURL:
            self = .unreachable
        case .decoding:
            self = .failed
        case .badStatus(let status, let body):
            self = Self.fromStatus(status, body: body, isAuthPool: apiError.isAuthPoolExhausted)
        }
    }

    private static func fromStatus(_ status: Int, body: String, isAuthPool: Bool) -> ChatPowerError {
        if isAuthPool { return .slotsBusy }
        switch status {
        case 404: return .gone
        case 409: return body.contains("chat_power_no_session") ? .noSession : .busy
        case 502, 503, 504: return .unreachable
        default: return .failed
        }
    }

    public var message: String {
        switch self {
        case .noSession:
            return "아직 세션이 없어요. 먼저 메시지를 한 번 보내 주세요."
        case .gone:
            return "그 세션은 이미 사라졌어요. 대화 기록은 그대로이고, 다음 메시지가 새 세션을 열어요."
        case .busy:
            return "플랫폼이 이 세션을 다루는 중이에요. 잠시 뒤 다시 시도해 주세요."
        case .slotsBusy:
            return "로그인 슬롯이 모두 사용 중이에요. 다른 세션이 끝난 뒤 다시 시도해 주세요."
        case .unreachable:
            return "세션 플랫폼에 연결하지 못했어요. 잠시 뒤 다시 시도해 주세요."
        case .failed:
            return "세션 상태를 처리하지 못했어요."
        }
    }
}

public extension APIClient {
    /// Reads whether the session answering this conversation is on. Safe to
    /// call on a schedule: the backend reads the session's record and not its
    /// transcript, so an archived session stays archived (PRD-11 AC5).
    func chatSessionPower() async throws -> ChatSessionPower {
        try await get("/api/chat/session", authenticated: true)
    }

    /// Turns that same session back on — the explicit restore, which is the
    /// only thing that may restore it (PRD-11 「복원 트리거」). It carries no
    /// session id: the backend restores the one answering the signed in user's
    /// conversation, so no tap can open a second session or reach somebody
    /// else's.
    func powerOnChatSession() async throws -> ChatSessionPower {
        try await post("/api/chat/session/power", body: ChatEmptyPayload(), authenticated: true)
    }
}
