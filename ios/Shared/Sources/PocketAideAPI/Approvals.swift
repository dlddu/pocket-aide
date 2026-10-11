import Foundation

public enum ApprovalStatus: String, Codable, Sendable, Hashable {
    case pending = "PENDING"
    case approved = "APPROVED"
    case rejected = "REJECTED"
    case expired = "EXPIRED"
}

public struct ApprovalRequest: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: String
    public let status: ApprovalStatus
    public let externalId: String
    public let context: String
    public let requesterName: String
    public let keyName: String
    public let keyRevoked: Bool
    public let createdAt: Int64
    public let expiresAt: Int64?
    public let processedAt: Int64?

    enum CodingKeys: String, CodingKey {
        case id
        case status
        case externalId = "external_id"
        case context
        case requesterName = "requester_name"
        case keyName = "key_name"
        case keyRevoked = "key_revoked"
        case createdAt = "created_at"
        case expiresAt = "expires_at"
        case processedAt = "processed_at"
    }

    public init(
        id: String,
        status: ApprovalStatus = .pending,
        externalId: String = "",
        context: String,
        requesterName: String = "",
        keyName: String = "",
        keyRevoked: Bool = false,
        createdAt: Int64,
        expiresAt: Int64? = nil,
        processedAt: Int64? = nil
    ) {
        self.id = id
        self.status = status
        self.externalId = externalId
        self.context = context
        self.requesterName = requesterName
        self.keyName = keyName
        self.keyRevoked = keyRevoked
        self.createdAt = createdAt
        self.expiresAt = expiresAt
        self.processedAt = processedAt
    }
}

public enum ApprovalFormatting {
    public static let noDeadline = "기한 없음"

    public static func newestFirst(_ items: [ApprovalRequest]) -> [ApprovalRequest] {
        items.sorted { ($0.createdAt, $0.id) > ($1.createdAt, $1.id) }
    }

    public static func remaining(expiresAt: Int64?, now: Date = Date()) -> String {
        guard let expiresAt else { return noDeadline }
        let left = max(0, expiresAt - Int64(now.timeIntervalSince1970.rounded(.down)))
        let hours = left / 3600
        let minutes = (left % 3600) / 60
        let seconds = left % 60
        if hours > 0 {
            return String(format: "%ld:%02ld:%02ld 남음", hours, minutes, seconds)
        }
        return String(format: "%02ld:%02ld 남음", minutes, seconds)
    }

    public static func summary(of context: String) -> String {
        let firstLine = context
            .split(whereSeparator: \.isNewline)
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .first { !$0.isEmpty }
        return firstLine ?? context.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    public static func pendingHeader(_ count: Int) -> String {
        "대기 · \(count)건"
    }
}

struct ApprovalListResponse: Decodable {
    let items: [ApprovalRequest]
}

public extension APIClient {
    func listPendingApprovals() async throws -> [ApprovalRequest] {
        let response: ApprovalListResponse = try await get("/api/approvals", authenticated: true)
        return ApprovalFormatting.newestFirst(response.items)
    }
}
