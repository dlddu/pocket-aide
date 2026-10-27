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
    public let sessionId: String?
    public let sessionName: String?
    public let reason: String?
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
        case sessionId = "session_id"
        case sessionName = "session_name"
        case reason
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
        sessionId: String? = nil,
        sessionName: String? = nil,
        reason: String? = nil,
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
        self.sessionId = sessionId
        self.sessionName = sessionName
        self.reason = reason
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

public enum ApprovalAutoResponse: String, Codable, CaseIterable, Identifiable, Sendable {
    case off = "NONE"
    case autoApprove = "AUTO_APPROVE"
    case autoReject = "AUTO_REJECT"

    public var id: String { rawValue }

    public var label: String {
        switch self {
        case .off: return "없음"
        case .autoApprove: return "자동 승인"
        case .autoReject: return "자동 거절"
        }
    }

    public var optionLabel: String {
        self == .off ? "없음 — 알림으로 받기" : label
    }

    public var isAutomatic: Bool { self != .off }

    public var bannerTitle: String? {
        isAutomatic ? "\(label) 모드가 켜져 있습니다." : nil
    }

    public var outcomeNotice: String {
        switch self {
        case .off: return "새 요청은 알림으로 받습니다."
        case .autoApprove: return "새 요청은 알림 없이 바로 승인됩니다."
        case .autoReject: return "새 요청은 알림 없이 바로 거절됩니다."
        }
    }
}

public struct ApprovalSettings: Codable, Equatable, Sendable {
    public let autoResponse: ApprovalAutoResponse

    enum CodingKeys: String, CodingKey {
        case autoResponse = "auto_response"
    }

    public init(autoResponse: ApprovalAutoResponse) {
        self.autoResponse = autoResponse
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

    public static func sessionLabel(_ item: ApprovalRequest) -> String? {
        let name = (item.sessionName ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        if !name.isEmpty { return name }
        let id = (item.sessionId ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        guard !id.isEmpty else { return nil }
        return "제목 없음 · \(String(id.prefix(8)))"
    }

    public static func agentReason(_ item: ApprovalRequest) -> String? {
        let reason = (item.reason ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        return reason.isEmpty ? nil : reason
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

    public static func historyMeta(_ item: ApprovalRequest) -> String {
        let mode = decisionModeLabel(status: item.status, mode: item.decisionMode)
        return item.keyRevoked ? "\(mode) · 폐기된 키로 생성됨" : mode
    }

    public static func historyTime(_ processedAt: Int64, now: Date = Date(), calendar: Calendar = .current) -> String {
        dayRelativeTime(processedAt, todayFormat: "HH:mm:ss", now: now, calendar: calendar)
    }

    public static func requestTime(_ createdAt: Int64, now: Date = Date(), calendar: Calendar = .current) -> String {
        dayRelativeTime(createdAt, todayFormat: "HH:mm", now: now, calendar: calendar)
    }

    private static func dayRelativeTime(_ seconds: Int64, todayFormat: String, now: Date, calendar: Calendar) -> String {
        let date = Date(timeIntervalSince1970: TimeInterval(seconds))
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.timeZone = calendar.timeZone
        if calendar.isDate(date, inSameDayAs: now) {
            formatter.dateFormat = todayFormat
            return formatter.string(from: date)
        }
        if let yesterday = calendar.date(byAdding: .day, value: -1, to: now), calendar.isDate(date, inSameDayAs: yesterday) {
            formatter.dateFormat = "HH:mm"
            return "어제 " + formatter.string(from: date)
        }
        formatter.dateFormat = "MM/dd HH:mm"
        return formatter.string(from: date)
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

public struct ApprovalHistoryPage: Decodable, Equatable, Sendable {
    public let items: [ApprovalRequest]
    public let nextCursor: String?

    enum CodingKeys: String, CodingKey {
        case items
        case nextCursor = "next_cursor"
    }

    public init(items: [ApprovalRequest], nextCursor: String?) {
        self.items = items
        self.nextCursor = nextCursor
    }
}

public extension APIClient {
    func listPendingApprovals() async throws -> [ApprovalRequest] {
        let response: ApprovalListResponse = try await get("/api/approvals", authenticated: true)
        return ApprovalFormatting.newestFirst(response.items)
    }

    func approvalHistory(cursor: String? = nil) async throws -> ApprovalHistoryPage {
        var path = "/api/approvals/history"
        if let cursor, let encoded = cursor.addingPercentEncoding(withAllowedCharacters: .alphanumerics) {
            path += "?cursor=" + encoded
        }
        return try await get(path, authenticated: true)
    }

    func approvalSettings() async throws -> ApprovalSettings {
        try await get("/api/approval-settings", authenticated: true)
    }

    func updateApprovalAutoResponse(_ mode: ApprovalAutoResponse) async throws -> ApprovalSettings {
        try await patch("/api/approval-settings", body: ApprovalSettings(autoResponse: mode), authenticated: true)
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

public enum ApprovalPushPayload {
    public static func approvalID(from info: [AnyHashable: Any]) -> String? {
        guard let raw = info["approval_id"] as? String, !raw.isEmpty else { return nil }
        return raw
    }

    public static func deepLinkURL(fromUserInfo info: [AnyHashable: Any]) -> URL? {
        guard let id = approvalID(from: info) else { return nil }
        var components = URLComponents()
        components.scheme = "pocketaide"
        components.host = "approvals"
        components.queryItems = [URLQueryItem(name: "id", value: id)]
        return components.url
    }

    public static func approvalID(fromDeepLink url: URL) -> String? {
        guard url.scheme == "pocketaide", url.host == "approvals",
              let id = URLComponents(url: url, resolvingAgainstBaseURL: false)?
                .queryItems?.first(where: { $0.name == "id" })?.value,
              !id.isEmpty else { return nil }
        return id
    }
}

public enum ExternalNotificationPushPayload {
    public static func matches(_ info: [AnyHashable: Any]) -> Bool {
        (info["kind"] as? String) == "external_notification"
    }

    public static func notificationID(from info: [AnyHashable: Any]) -> Int64? {
        let raw = info["notification_id"]
        if let n = raw as? Int64 { return n > 0 ? n : nil }
        if let n = raw as? Int { return n > 0 ? Int64(n) : nil }
        if let n = raw as? NSNumber { return n.int64Value > 0 ? n.int64Value : nil }
        if let s = raw as? String, let n = Int64(s), n > 0 { return n }
        return nil
    }

    public static func deepLinkURL(fromUserInfo info: [AnyHashable: Any]) -> URL? {
        var components = URLComponents()
        components.scheme = "pocketaide"
        components.host = "notifications"
        if let id = notificationID(from: info) {
            components.queryItems = [URLQueryItem(name: "id", value: String(id))]
        }
        return components.url
    }

    public static func notificationID(fromDeepLink url: URL) -> Int64? {
        guard url.scheme == "pocketaide", url.host == "notifications",
              let raw = URLComponents(url: url, resolvingAgainstBaseURL: false)?
                .queryItems?.first(where: { $0.name == "id" })?.value,
              let id = Int64(raw), id > 0 else { return nil }
        return id
    }
}

public enum PushDeepLink {
    public static func url(fromUserInfo info: [AnyHashable: Any]) -> URL? {
        if ExternalNotificationPushPayload.matches(info) { return ExternalNotificationPushPayload.deepLinkURL(fromUserInfo: info) }
        return ApprovalPushPayload.deepLinkURL(fromUserInfo: info) ?? PRMonitorPushPayload.deepLinkURL(fromUserInfo: info)
    }
}
