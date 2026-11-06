import Foundation

// PRD-1: the chat tab is one long conversation with the user's codex session
// on session-platform. A send is answered at once (202) and the reply arrives
// through GET /api/chat, so the screen polls while one is pending.

public struct ChatConversation: Codable, Hashable, Sendable {
    public let id: Int64
    public let title: String

    public init(id: Int64, title: String) {
        self.id = id
        self.title = title
    }
}

public struct ChatMessage: Codable, Hashable, Identifiable, Sendable {
    public let id: Int64
    public let role: String
    public let content: String
    public let createdAt: Int64

    enum CodingKeys: String, CodingKey {
        case id, role, content
        case createdAt = "created_at"
    }

    public var isMine: Bool { role == "user" }
}

public struct ChatPending: Codable, Hashable, Sendable {
    public let content: String
    public let startedAt: Int64

    enum CodingKeys: String, CodingKey {
        case content
        case startedAt = "started_at"
    }

    public init(content: String = "", startedAt: Int64 = 0) {
        self.content = content
        self.startedAt = startedAt
    }
}

public struct ChatThread: Codable, Sendable {
    public let conversation: ChatConversation
    public let messages: [ChatMessage]
    public let pending: ChatPending?
}

public struct ChatPluginRefresh: Codable, Hashable, Sendable {
    public let state: String
    public let revision: String?
    public let added: [String]
    public let updated: [String]
    public let removed: [String]

    public var changedCount: Int { added.count + updated.count + removed.count }

    public var notice: String {
        if state == "unconfigured" { return "이 대화에는 새로 받을 스킬이 없어요." }
        if changedCount == 0 { return "스킬이 이미 최신이에요." }
        return "스킬을 새로 받았어요 — 추가 \(added.count) · 수정 \(updated.count) · 삭제 \(removed.count). 다음 답부터 반영돼요."
    }
}

public enum ChatError: Hashable, Sendable {
    case busy
    case noConversation
    case tooLarge
    case unreachable
    case failed

    public static let maxMessageBytes = 16 * 1024

    public init(_ error: Error, refreshing: Bool = false) {
        switch error as? APIError {
        case let .badStatus(status, body)?:
            switch status {
            case 409: self = refreshing && body.contains("send a message first") ? .noConversation : .busy
            case 413: self = .tooLarge
            case 502, 503, 504: self = .unreachable
            default: self = .failed
            }
        case .transport?, .missingBaseURL?:
            self = .unreachable
        default:
            self = .failed
        }
    }

    public var message: String {
        switch self {
        case .busy: return "이전 답이 끝난 뒤에 다시 해 주세요."
        case .noConversation: return "먼저 메시지를 한 번 보내 주세요."
        case .tooLarge: return "메시지가 너무 깁니다."
        case .unreachable: return "세션 플랫폼에 연결하지 못했어요. 잠시 뒤 다시 시도해 주세요."
        case .failed: return "요청을 처리하지 못했어요."
        }
    }
}

struct ChatSendPayload: Encodable {
    let content: String
}

struct ChatSendResponse: Decodable {
    let message: ChatMessage
}

struct ChatEmptyPayload: Encodable {}

public extension APIClient {
    func chat() async throws -> ChatThread {
        try await get("/api/chat", authenticated: true)
    }

    func sendChatMessage(_ content: String) async throws -> ChatMessage {
        let response: ChatSendResponse = try await post("/api/chat/messages", body: ChatSendPayload(content: content), authenticated: true)
        return response.message
    }

    func refreshChatPlugins() async throws -> ChatPluginRefresh {
        try await post("/api/chat/refresh-plugins", body: ChatEmptyPayload(), authenticated: true)
    }
}
