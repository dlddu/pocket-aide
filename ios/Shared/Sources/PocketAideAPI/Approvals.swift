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
    public let decisionMode: String?

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
        case decisionMode = "decision_mode"
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
        processedAt: Int64? = nil,
        decisionMode: String? = nil
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
        self.decisionMode = decisionMode
    }

    public func isExpired(at now: Date) -> Bool {
        guard status == .pending, let expiresAt else { return status == .expired }
        return expiresAt <= Int64(now.timeIntervalSince1970.rounded(.down))
    }
}

public enum ApprovalDecision: String, Sendable {
    case approve
    case reject
}

public enum ApprovalDecisionOutcome: Equatable, Sendable {
    case decided(ApprovalRequest)
    case conflict(ApprovalRequest)
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

    public static func countdown(expiresAt: Int64, now: Date = Date()) -> String {
        let left = max(0, expiresAt - Int64(now.timeIntervalSince1970.rounded(.down)))
        let hours = left / 3600
        let minutes = (left % 3600) / 60
        let seconds = left % 60
        if hours > 0 {
            return String(format: "%ld:%02ld:%02ld", hours, minutes, seconds)
        }
        return String(format: "%02ld:%02ld", minutes, seconds)
    }

    public static func progress(createdAt: Int64, expiresAt: Int64, now: Date = Date()) -> Double {
        let total = Double(expiresAt - createdAt)
        guard total > 0 else { return 0 }
        let left = Double(expiresAt) - now.timeIntervalSince1970
        return min(1, max(0, left / total))
    }

    public static func statusLabel(_ status: ApprovalStatus) -> String {
        switch status {
        case .pending: return "대기 중"
        case .approved: return "승인됨"
        case .rejected: return "거절됨"
        case .expired: return "만료"
        }
    }

    public static func decisionModeLabel(status: ApprovalStatus, mode: String?) -> String {
        let mode = mode ?? ""
        if status == .expired || mode == "expired" {
            return "만료 — 결정 없음"
        }
        if mode.hasPrefix("auto") {
            return status == .rejected ? "자동 거절" : "자동 승인"
        }
        return "수동 결정"
    }

    public static func confirmTitle(_ decision: ApprovalDecision) -> String {
        decision == .approve ? "승인할까요?" : "거절할까요?"
    }

    public static func confirmMessage(_ decision: ApprovalDecision, context: String) -> String {
        let verb = decision == .approve ? "허락합니다" : "거절합니다"
        return "\(summary(of: context))\(objectParticle(for: summary(of: context))) \(verb). 확정하면 되돌릴 수 없습니다."
    }

    public static func conflictNotice(_ item: ApprovalRequest) -> String {
        item.status == .expired ? "만료된 요청이라 결정되지 않았습니다" : "이미 처리된 요청입니다 — \(statusLabel(item.status))"
    }

    public static let failedDecisionNotice = "결정되지 않았습니다. 연결을 확인한 뒤 다시 시도하세요"

    static func objectParticle(for text: String) -> String {
        guard let last = text.unicodeScalars.last, (0xAC00...0xD7A3).contains(last.value) else { return "을(를)" }
        return (last.value - 0xAC00) % 28 == 0 ? "를" : "을"
    }
}

struct ApprovalConflictResponse: Decodable {
    let approval: ApprovalRequest
}

struct ApprovalDecisionPayload: Encodable {}

struct ApprovalListResponse: Decodable {
    let items: [ApprovalRequest]
}

public extension APIClient {
    func listPendingApprovals() async throws -> [ApprovalRequest] {
        let response: ApprovalListResponse = try await get("/api/approvals", authenticated: true)
        return ApprovalFormatting.newestFirst(response.items)
    }

    func approval(id: String) async throws -> ApprovalRequest {
        try await get("/api/approvals/\(id)", authenticated: true)
    }

    func decideApproval(id: String, _ decision: ApprovalDecision) async throws -> ApprovalDecisionOutcome {
        do {
            let decided: ApprovalRequest = try await post(
                "/api/approvals/\(id)/\(decision.rawValue)",
                body: ApprovalDecisionPayload(),
                authenticated: true
            )
            return .decided(decided)
        } catch APIError.badStatus(409, let body) {
            let conflict = try APIClient.decoder.decode(ApprovalConflictResponse.self, from: Data(body.utf8))
            return .conflict(conflict.approval)
        }
    }
}
