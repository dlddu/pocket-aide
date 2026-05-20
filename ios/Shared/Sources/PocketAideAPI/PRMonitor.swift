import Foundation

/// One row of `notification_history` belonging to the authenticated user
/// (PRD-10 AC11). Optional fields are null when the underlying workflow_run
/// did not have a PR linked (e.g. push to main) — the iOS card uses the
/// fallback `repo — conclusion · workflow_name` text in that case.
public struct NotificationHistoryItem: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let repoFullName: String
    public let prNumber: Int?
    public let prTitle: String?
    public let prURL: String?
    public let commitURL: String?
    public let runURL: String?
    public let workflowName: String
    public let headBranch: String
    public let headSHA: String
    public let conclusion: String
    public let acknowledgedAt: Int64?
    public let createdAt: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case repoFullName = "repo_full_name"
        case prNumber = "pr_number"
        case prTitle = "pr_title"
        case prURL = "pr_url"
        case commitURL = "commit_url"
        case runURL = "run_url"
        case workflowName = "workflow_name"
        case headBranch = "head_branch"
        case headSHA = "head_sha"
        case conclusion
        case acknowledgedAt = "acknowledged_at"
        case createdAt = "created_at"
    }

    public init(
        id: Int64,
        repoFullName: String,
        prNumber: Int?,
        prTitle: String?,
        prURL: String?,
        commitURL: String?,
        runURL: String?,
        workflowName: String,
        headBranch: String,
        headSHA: String,
        conclusion: String,
        acknowledgedAt: Int64?,
        createdAt: Int64
    ) {
        self.id = id
        self.repoFullName = repoFullName
        self.prNumber = prNumber
        self.prTitle = prTitle
        self.prURL = prURL
        self.commitURL = commitURL
        self.runURL = runURL
        self.workflowName = workflowName
        self.headBranch = headBranch
        self.headSHA = headSHA
        self.conclusion = conclusion
        self.acknowledgedAt = acknowledgedAt
        self.createdAt = createdAt
    }
}

public struct ExcludedRepo: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let repoFullName: String
    public let createdAt: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case repoFullName = "repo_full_name"
        case createdAt = "created_at"
    }

    public init(id: Int64, repoFullName: String, createdAt: Int64) {
        self.id = id
        self.repoFullName = repoFullName
        self.createdAt = createdAt
    }
}

/// `{}` body — backend ignores the body for the ack endpoint but post()
/// requires *something* Encodable.
struct EmptyPayload: Encodable {}

/// Pure helper exposed to both the app delegate (push tap → deep link) and
/// to unit tests. Parses the `event_id` field from a push payload (which
/// arrives as a heterogeneous `[AnyHashable: Any]`) and synthesizes the
/// `pocketaide://pr-monitor?eventId=<id>` URL the app uses to route the
/// notification into the PR monitor tab.
public enum PRMonitorPushPayload {
    public static func deepLinkURL(fromUserInfo info: [AnyHashable: Any]) -> URL? {
        guard let id = eventID(from: info) else { return nil }
        return URL(string: "pocketaide://pr-monitor?eventId=\(id)")
    }

    public static func eventID(from info: [AnyHashable: Any]) -> Int64? {
        let raw = info["event_id"]
        if let n = raw as? Int64 { return n }
        if let n = raw as? Int { return Int64(n) }
        if let n = raw as? NSNumber { return n.int64Value }
        if let s = raw as? String, let n = Int64(s) { return n }
        return nil
    }
}

struct NotificationHistoryListResponse: Decodable {
    let items: [NotificationHistoryItem]
}

struct ExcludedRepoListResponse: Decodable {
    let items: [ExcludedRepo]
}

struct ExcludedRepoPayload: Encodable {
    let repoFullName: String

    enum CodingKeys: String, CodingKey {
        case repoFullName = "repo_full_name"
    }
}

public extension APIClient {
    func listNotificationHistory(limit: Int? = nil, before: Int64? = nil) async throws -> [NotificationHistoryItem] {
        var path = "/api/notification-history"
        var query: [String] = []
        if let limit { query.append("limit=\(limit)") }
        if let before { query.append("before=\(before)") }
        if !query.isEmpty { path += "?" + query.joined(separator: "&") }
        let response: NotificationHistoryListResponse = try await get(path, authenticated: true)
        return response.items
    }

    /// AC12: the only path that should call this is the explicit "확인" button.
    /// Push taps (AC7) and external-link taps must NOT call it.
    func acknowledgeNotification(id: Int64) async throws {
        _ = try await post(
            "/api/notification-history/\(id)/ack",
            body: EmptyPayload(),
            authenticated: true,
            decodeAs: APIClient.EmptyResponse.self
        )
    }

    func listExcludedRepos() async throws -> [ExcludedRepo] {
        let response: ExcludedRepoListResponse = try await get(
            "/api/excluded-repos",
            authenticated: true
        )
        return response.items
    }

    func addExcludedRepo(_ repoFullName: String) async throws -> ExcludedRepo {
        try await post(
            "/api/excluded-repos",
            body: ExcludedRepoPayload(repoFullName: repoFullName),
            authenticated: true
        )
    }

    func removeExcludedRepo(id: Int64) async throws {
        try await delete("/api/excluded-repos/\(id)", authenticated: true)
    }
}
