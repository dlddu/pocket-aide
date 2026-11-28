import Foundation

// Archiving the codex session answering the chat tab's one conversation
// (PRD-1 AC11). It is the platform's own snapshot — the same action the
// sessions tab takes (PRD-11 AC8) — and the conversation is not part of what
// is handed back: the next message restores this very session and carries on
// (PRD-1 AC4).

/// What became of the session, as the server read it back from the platform.
/// `archived` is false when the platform took the request without the record
/// reading archived yet — the request was accepted, and saying the environment
/// is already handed back would be more than we know.
public struct ChatArchiveResult: Codable, Hashable, Sendable {
    public let sessionID: String
    public let state: String
    public let archived: Bool
    public let checkpoint: ChatArchiveCheckpoint?

    enum CodingKeys: String, CodingKey {
        case sessionID = "session_id"
        case state, archived, checkpoint
    }

    public init(sessionID: String = "", state: String = "", archived: Bool = false, checkpoint: ChatArchiveCheckpoint? = nil) {
        self.sessionID = sessionID
        self.state = state
        self.archived = archived
        self.checkpoint = checkpoint
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        sessionID = try container.decodeIfPresent(String.self, forKey: .sessionID) ?? ""
        state = try container.decodeIfPresent(String.self, forKey: .state) ?? ""
        archived = try container.decodeIfPresent(Bool.self, forKey: .archived) ?? false
        checkpoint = try container.decodeIfPresent(ChatArchiveCheckpoint.self, forKey: .checkpoint)
    }

    /// What to tell the person, said from what the server read back — never
    /// from what the tap hoped for.
    public var notice: String {
        guard archived else {
            return "아카이브를 요청했어요. 환경 반납은 잠시 뒤 상태로 확인해 주세요. 대화는 그대로예요."
        }
        let detail = [checkpoint?.relativeCreatedAt, checkpoint?.formattedSize]
            .compactMap { $0 }
            .joined(separator: " · ")
        let suffix = detail.isEmpty ? "" : " (\(detail))"
        return "아카이브됐어요 — 환경을 반납했어요\(suffix). 대화는 그대로이고, 다음 메시지가 이 세션을 복원해요."
    }
}

/// When the archive was taken and how big it is — the same two fields the
/// sessions tab shows for an archived session (PRD-11 AC5).
public struct ChatArchiveCheckpoint: Codable, Hashable, Sendable {
    public let createdAt: String
    public let sizeBytes: Int64?

    public init(createdAt: String = "", sizeBytes: Int64? = nil) {
        self.createdAt = createdAt
        self.sizeBytes = sizeBytes
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        createdAt = try container.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
        sizeBytes = try container.decodeIfPresent(Int64.self, forKey: .sizeBytes)
    }

    var relativeCreatedAt: String? {
        let text = AgentSessionFormatting.relative(createdAt)
        return text == "—" ? nil : text
    }

    var formattedSize: String? { AgentSessionFormatting.size(sizeBytes) }
}

/// Why an archive did not happen. The server gives each refusal its own code
/// because each is a different thing to say — and `.disabled` is not a failure
/// at all: the session stays up and nothing is lost (PRD-11 AC8).
public enum ChatArchiveError: Hashable, Sendable {
    case noSession
    case sessionChanged
    case queued
    case busy
    case gone
    case disabled
    case slotsBusy
    case unreachable
    case failed

    // Read by code, not by status: two of these share 409 and two share 503,
    // and the app says something different for each.
    private static let byCode: [(code: String, refusal: ChatArchiveError)] = [
        ("chat_archive_no_session", .noSession),
        ("chat_archive_session_changed", .sessionChanged),
        ("chat_archive_queued", .queued),
        ("chat_archive_busy", .busy),
        ("chat_archive_session_gone", .gone),
        ("chat_archive_disabled", .disabled),
    ]

    public init(_ error: Error) {
        switch error as? APIError {
        case let .badStatus(status, body)?:
            self = Self.named(in: body) ?? Self.unnamed(status, body)
        case .transport?, .missingBaseURL?:
            self = .unreachable
        default:
            self = .failed
        }
    }

    private static func named(in body: String) -> ChatArchiveError? {
        byCode.first { body.contains("\"\($0.code)\"") }?.refusal
    }

    /// A refusal this route did not label: the platform's own login-slot 503,
    /// which is not about this feature (PRD-1 AC7), and everything else.
    private static func unnamed(_ status: Int, _ body: String) -> ChatArchiveError {
        if body.contains(APIError.authPoolExhaustedCode) { return .slotsBusy }
        return [502, 503, 504].contains(status) ? .unreachable : .failed
    }

    /// The refusals that mean the screen's idea of the session is stale, so it
    /// reads the conversation again.
    public var rereadsConversation: Bool { self == .sessionChanged || self == .gone }

    public var message: String {
        switch self {
        case .noSession: return "아직 이 대화를 맡은 세션이 없어요. 먼저 메시지를 한 번 보내 주세요."
        case .sessionChanged: return "그 사이 다른 세션이 이 대화를 맡았어요. 지금 담당 세션을 다시 읽었으니 확인하고 다시 해 주세요."
        case .queued: return "보낼 메시지가 기다리고 있어요. 그 메시지가 나간 뒤에 다시 해 주세요."
        case .busy: return "답을 쓰는 중이에요. 답이 끝난 뒤에 다시 해 주세요."
        case .gone: return "이 대화를 맡았던 세션은 이미 사라졌어요. 대화 기록은 그대로이고, 다음 메시지가 새 세션을 열어요."
        case .disabled: return "이 환경에서는 아카이브가 꺼져 있어요. 세션은 그대로 유지되고 데이터는 사라지지 않아요."
        case .slotsBusy: return "로그인 슬롯이 모두 사용 중이에요. 다른 세션이 끝난 뒤 다시 시도해 주세요."
        case .unreachable: return "세션 플랫폼에 연결하지 못했어요. 잠시 뒤 다시 시도해 주세요."
        case .failed: return "아카이브하지 못했어요."
        }
    }
}

struct ChatArchivePayload: Encodable {
    let sessionID: String

    enum CodingKeys: String, CodingKey {
        case sessionID = "session_id"
    }
}

public extension APIClient {
    /// Hands back the environment of `sessionID` — the session the person was
    /// shown and confirmed. The server refuses if that is no longer the session
    /// answering this conversation, so a confirmation drawn before a handover
    /// cannot archive the one that took over.
    func archiveChatSession(sessionID: String) async throws -> ChatArchiveResult {
        try await post("/api/chat/archive", body: ChatArchivePayload(sessionID: sessionID), authenticated: true)
    }
}
