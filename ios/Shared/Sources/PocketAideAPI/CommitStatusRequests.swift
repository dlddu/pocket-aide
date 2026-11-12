import Foundation

public enum CommitStatusRequestStatus: String, Codable, Sendable, Hashable {
    case pending = "PENDING"
    case approved = "APPROVED"
    case rejected = "REJECTED"
}

public struct CommitStatusRequest: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: String
    public let status: CommitStatusRequestStatus
    public let owner: String
    public let repo: String
    public let sha: String
    public let state: String
    public let context: String
    public let description: String
    public let targetUrl: String
    public let requester: String
    public let reason: String
    public let sessionId: String
    public let sessionName: String
    public let keyName: String
    public let keyRevoked: Bool
    public let githubStatusId: Int64?
    public let lastError: String
    public let createdAt: Int64
    public let processedAt: Int64?

    enum CodingKeys: String, CodingKey {
        case id, status, owner, repo, sha, state, context, description, requester, reason
        case targetUrl = "target_url"
        case sessionId = "session_id"
        case sessionName = "session_name"
        case keyName = "key_name"
        case keyRevoked = "key_revoked"
        case githubStatusId = "github_status_id"
        case lastError = "last_error"
        case createdAt = "created_at"
        case processedAt = "processed_at"
    }

    public init(
        id: String,
        status: CommitStatusRequestStatus = .pending,
        owner: String = "dlddu",
        repo: String,
        sha: String,
        state: String = "success",
        context: String,
        description: String = "",
        targetUrl: String = "",
        requester: String = "",
        reason: String = "",
        sessionId: String = "",
        sessionName: String = "",
        keyName: String = "",
        keyRevoked: Bool = false,
        githubStatusId: Int64? = nil,
        lastError: String = "",
        createdAt: Int64,
        processedAt: Int64? = nil
    ) {
        self.id = id
        self.status = status
        self.owner = owner
        self.repo = repo
        self.sha = sha
        self.state = state
        self.context = context
        self.description = description
        self.targetUrl = targetUrl
        self.requester = requester
        self.reason = reason
        self.sessionId = sessionId
        self.sessionName = sessionName
        self.keyName = keyName
        self.keyRevoked = keyRevoked
        self.githubStatusId = githubStatusId
        self.lastError = lastError
        self.createdAt = createdAt
        self.processedAt = processedAt
    }
}

public struct CommitStatusRequestList: Decodable, Equatable, Sendable {
    public let items: [CommitStatusRequest]
    public let history: [CommitStatusRequest]

    public init(items: [CommitStatusRequest], history: [CommitStatusRequest]) {
        self.items = items
        self.history = history
    }
}

public enum CommitStatusDecisionOutcome: Equatable, Sendable {
    case decided(CommitStatusRequest)
    case conflict(CommitStatusRequest)
    case writeFailed(CommitStatusRequest, String)
    case unavailable(String)
}

public enum CommitStatusFormatting {
    public static let failedDecisionNotice = "결정되지 않았습니다. 연결을 확인한 뒤 다시 시도하세요"
    public static let unavailableNotice = "커밋 상태를 쓸 GitHub App 이 설정되지 않아 승인할 수 없습니다"

    public static func title(_ item: CommitStatusRequest) -> String {
        "\(item.context) → \(item.state)"
    }

    public static func shortSHA(_ sha: String) -> String {
        String(sha.prefix(7))
    }

    public static func target(_ item: CommitStatusRequest) -> String {
        "\(item.owner)/\(item.repo)@\(shortSHA(item.sha))"
    }

    public static func statusLabel(_ status: CommitStatusRequestStatus) -> String {
        switch status {
        case .pending: return "대기 중"
        case .approved: return "승인됨 — 기록함"
        case .rejected: return "거절됨"
        }
    }

    public static func pendingHeader(_ count: Int) -> String {
        "대기 · \(count)건"
    }

    public static let historyHeader = "최근 처리"

    public static func confirmTitle(_ decision: ApprovalDecision) -> String {
        decision == .approve ? "커밋 상태를 기록할까요?" : "거절할까요?"
    }

    public static func confirmMessage(_ decision: ApprovalDecision, item: CommitStatusRequest) -> String {
        if decision == .approve {
            return "\(target(item)) 에 \(title(item)) 을(를) GitHub 에 기록합니다. 확정하면 되돌릴 수 없습니다."
        }
        return "\(title(item)) 요청을 거절합니다. GitHub 에는 아무것도 기록하지 않습니다."
    }

    public static func conflictNotice(_ item: CommitStatusRequest) -> String {
        "이미 처리된 요청입니다 — \(statusLabel(item.status))"
    }

    public static func writeFailedNotice(_ message: String) -> String {
        "GitHub 에 기록하지 못했습니다 — \(message)"
    }

    public static func newestFirst(_ items: [CommitStatusRequest]) -> [CommitStatusRequest] {
        items.sorted { ($0.createdAt, $0.id) > ($1.createdAt, $1.id) }
    }
}

struct CommitStatusDecisionErrorResponse: Decodable {
    let error: String
    let request: CommitStatusRequest?
}

struct CommitStatusDecisionPayload: Encodable {}

public extension APIClient {
    func listCommitStatusRequests() async throws -> CommitStatusRequestList {
        let list: CommitStatusRequestList = try await get("/api/commit-status-requests", authenticated: true)
        return CommitStatusRequestList(items: CommitStatusFormatting.newestFirst(list.items), history: list.history)
    }

    func commitStatusRequest(id: String) async throws -> CommitStatusRequest {
        try await get("/api/commit-status-requests/\(id)", authenticated: true)
    }

    func decideCommitStatusRequest(id: String, _ decision: ApprovalDecision) async throws -> CommitStatusDecisionOutcome {
        do {
            let decided: CommitStatusRequest = try await post(
                "/api/commit-status-requests/\(id)/\(decision.rawValue)",
                body: CommitStatusDecisionPayload(),
                authenticated: true
            )
            return .decided(decided)
        } catch APIError.badStatus(let code, let body) where [409, 502, 503].contains(code) {
            let response = try APIClient.decoder.decode(CommitStatusDecisionErrorResponse.self, from: Data(body.utf8))
            switch (code, response.request) {
            case (409, let request?):
                return .conflict(request)
            case (502, let request?):
                return .writeFailed(request, response.error)
            case (503, _):
                return .unavailable(response.error)
            default:
                throw APIError.badStatus(code, body)
            }
        }
    }
}

public enum CommitStatusNotificationAction {
    public static let categoryIdentifier = "COMMIT_STATUS_REQUEST"
    public static let approveIdentifier = "COMMIT_STATUS_APPROVE"
    public static let rejectIdentifier = "COMMIT_STATUS_REJECT"

    public static func title(_ decision: ApprovalDecision) -> String {
        decision == .approve ? "승인" : "거절"
    }

    public static func decision(forActionIdentifier identifier: String) -> ApprovalDecision? {
        switch identifier {
        case approveIdentifier: return .approve
        case rejectIdentifier: return .reject
        default: return nil
        }
    }

    public static func resultNotice(_ result: Result<CommitStatusDecisionOutcome, Error>) -> String {
        switch result {
        case .success(.decided(let item)):
            return "\(CommitStatusFormatting.statusLabel(item.status)) — \(CommitStatusFormatting.title(item))"
        case .success(.conflict(let item)):
            return CommitStatusFormatting.conflictNotice(item)
        case .success(.writeFailed(_, let message)):
            return CommitStatusFormatting.writeFailedNotice(message)
        case .success(.unavailable):
            return CommitStatusFormatting.unavailableNotice
        case .failure:
            return CommitStatusFormatting.failedDecisionNotice
        }
    }
}

public enum CommitStatusPushPayload {
    public static let kind = "commit_status_request"

    public static func matches(_ info: [AnyHashable: Any]) -> Bool {
        (info["kind"] as? String) == kind
    }

    public static func requestID(from info: [AnyHashable: Any]) -> String? {
        guard matches(info), let raw = info["commit_status_request_id"] as? String, !raw.isEmpty else { return nil }
        return raw
    }

    public static func deepLinkURL(fromUserInfo info: [AnyHashable: Any]) -> URL? {
        var components = URLComponents()
        components.scheme = "pocketaide"
        components.host = "commit-status"
        if let id = requestID(from: info) {
            components.queryItems = [URLQueryItem(name: "id", value: id)]
        }
        return components.url
    }

    public static func requestID(fromDeepLink url: URL) -> String? {
        guard url.scheme == "pocketaide", url.host == "commit-status",
              let id = URLComponents(url: url, resolvingAgainstBaseURL: false)?
                .queryItems?.first(where: { $0.name == "id" })?.value,
              !id.isEmpty else { return nil }
        return id
    }
}
