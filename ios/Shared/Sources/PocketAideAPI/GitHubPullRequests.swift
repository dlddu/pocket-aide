import Foundation

public enum PullRequestRole: String, Equatable, Sendable {
    case author
    case reviewer

    public var label: String {
        switch self {
        case .author: return "작성자"
        case .reviewer: return "리뷰어"
        }
    }
}

public enum PullRequestCIStatus: String, Equatable, Sendable {
    case inProgress
    case success
    case failure
    case noChecks

    public init(rollupState: String?) {
        switch rollupState {
        case "SUCCESS": self = .success
        case "FAILURE", "ERROR": self = .failure
        case "PENDING", "EXPECTED": self = .inProgress
        default: self = .noChecks
        }
    }

    public var label: String {
        switch self {
        case .inProgress: return "진행 중"
        case .success: return "성공"
        case .failure: return "실패"
        case .noChecks: return "상태 없음"
        }
    }
}

public struct OpenPullRequest: Identifiable, Equatable, Sendable {
    public let repository: String
    public let number: Int
    public let title: String
    public let author: String
    public let url: URL?
    public let updatedAt: Date
    public let role: PullRequestRole
    public let ciStatus: PullRequestCIStatus

    public var id: String { "\(repository)#\(number)" }
}

public enum OpenPullRequestFilter: String, CaseIterable, Equatable, Sendable {
    case all
    case mine
    case reviewRequested
    case ciFailing

    public var label: String {
        switch self {
        case .all: return "전체"
        case .mine: return "내 PR만"
        case .reviewRequested: return "리뷰 요청만"
        case .ciFailing: return "CI 실패만"
        }
    }

    public func matches(_ pullRequest: OpenPullRequest) -> Bool {
        switch self {
        case .all: return true
        case .mine: return pullRequest.role == .author
        case .reviewRequested: return pullRequest.role == .reviewer
        case .ciFailing: return pullRequest.ciStatus == .failure
        }
    }

    public func apply(to pullRequests: [OpenPullRequest]) -> [OpenPullRequest] {
        pullRequests.filter(matches)
    }
}

public struct OpenPullRequestsResult: Equatable, Sendable {
    public let pullRequests: [OpenPullRequest]
    public let inaccessibleCount: Int
}

public enum GitHubError: Error, Equatable, CustomStringConvertible {
    case unauthorized
    case forbidden
    case rateLimited(resetAt: Date?)
    case badStatus(Int)
    case graphQL(String)
    case decoding(String)
    case transport(String)

    public var description: String {
        switch self {
        case .unauthorized:
            return "GitHub가 토큰을 거부했습니다(만료·폐기·오타). 토큰을 바꿔 다시 연결하세요."
        case .forbidden:
            return "GitHub가 이 토큰의 요청을 거부했습니다(권한 부족)."
        case .rateLimited:
            return "GitHub 요청 한도에 도달했습니다."
        case .badStatus(let status):
            return "GitHub 응답 오류 (HTTP \(status))"
        case .graphQL(let message):
            return "GitHub 조회 오류: \(message)"
        case .decoding(let detail):
            return "GitHub 응답을 해석하지 못했습니다: \(detail)"
        case .transport(let detail):
            return "GitHub에 연결하지 못했습니다: \(detail)"
        }
    }

    public static func classify(statusCode: Int, remaining: String? = nil, reset: String? = nil, retryAfter: String? = nil, now: Date = Date()) -> GitHubError? {
        if (200..<300).contains(statusCode) {
            return nil
        }
        if statusCode == 401 {
            return .unauthorized
        }
        if statusCode == 403 || statusCode == 429 {
            if let seconds = retryAfter.flatMap({ TimeInterval($0) }) {
                return .rateLimited(resetAt: now.addingTimeInterval(seconds))
            }
            if remaining == "0" || statusCode == 429 {
                let resetAt = reset.flatMap { TimeInterval($0) }.map { Date(timeIntervalSince1970: $0) }
                return .rateLimited(resetAt: resetAt)
            }
            return .forbidden
        }
        return .badStatus(statusCode)
    }
}

public enum GitHubAlert: Equatable, Sendable {
    case tokenRejected
    case insufficientAccess(hiddenCount: Int)
    case rateLimited(resetAt: Date?)

    public static func make(error: GitHubError?, inaccessibleCount: Int) -> GitHubAlert? {
        switch error {
        case .unauthorized:
            return .tokenRejected
        case .forbidden:
            return .insufficientAccess(hiddenCount: inaccessibleCount)
        case .rateLimited(let resetAt):
            return .rateLimited(resetAt: resetAt)
        default:
            return inaccessibleCount > 0 ? .insufficientAccess(hiddenCount: inaccessibleCount) : nil
        }
    }

    public var kind: String {
        switch self {
        case .tokenRejected: return "token"
        case .insufficientAccess: return "access"
        case .rateLimited: return "ratelimit"
        }
    }

    public var title: String {
        switch self {
        case .tokenRejected: return "GitHub 토큰이 거부되었습니다"
        case .insufficientAccess: return "접근 권한이 부족합니다"
        case .rateLimited: return "GitHub 요청 한도에 도달했습니다"
        }
    }

    public func message(formatTime: (Date) -> String) -> String {
        switch self {
        case .tokenRejected:
            return "토큰이 만료·폐기되었거나 잘못 입력되었습니다. 새 토큰으로 다시 연결하세요."
        case .insufficientAccess(let hiddenCount):
            let lead = hiddenCount > 0 ? "PR \(hiddenCount)개는 이 토큰으로 볼 수 없어 표시하지 않았습니다." : "이 토큰으로는 GitHub 조회가 허용되지 않습니다."
            return "\(lead) 조직 SSO 승인이나 토큰 범위(classic은 repo)를 확인한 뒤 토큰을 바꾸세요."
        case .rateLimited(let resetAt):
            return resetAt.map { "\(formatTime($0)) 이후에 다시 시도하세요." } ?? "잠시 후 다시 시도하세요."
        }
    }

    public var actionTitle: String {
        switch self {
        case .tokenRejected: return "토큰 다시 연결"
        case .insufficientAccess: return "토큰 바꾸기"
        case .rateLimited: return "다시 시도"
        }
    }
}

public enum OpenPullRequests {
    public static let query = """
    query($authored: String!, $requested: String!, $reviewed: String!) {
      authored: search(query: $authored, type: ISSUE, first: 50) { nodes { ...OpenPR } }
      requested: search(query: $requested, type: ISSUE, first: 50) { nodes { ...OpenPR } }
      reviewed: search(query: $reviewed, type: ISSUE, first: 50) { nodes { ...OpenPR } }
    }
    fragment OpenPR on PullRequest {
      number
      title
      url
      updatedAt
      repository { nameWithOwner }
      author { login }
      commits(last: 1) { nodes { commit { statusCheckRollup { state } } } }
    }
    """

    public static func variables(login: String) -> [String: String] {
        let base = "is:pr is:open archived:false"
        return [
            "authored": "\(base) author:\(login)",
            "requested": "\(base) review-requested:\(login)",
            "reviewed": "\(base) reviewed-by:\(login) -author:\(login)",
        ]
    }

    public static func decode(_ data: Data) throws -> OpenPullRequestsResult {
        let response: SearchResponse
        do {
            response = try JSONDecoder().decode(SearchResponse.self, from: data)
        } catch {
            throw GitHubError.decoding(String(describing: error))
        }
        guard let payload = response.data else {
            if response.errors?.contains(where: { $0.type == "RATE_LIMITED" }) == true {
                throw GitHubError.rateLimited(resetAt: nil)
            }
            throw GitHubError.graphQL(response.errors?.first?.message ?? "data 없음")
        }
        let sources: [(SearchResponse.Connection?, PullRequestRole)] = [
            (payload.authored, .author),
            (payload.requested, .reviewer),
            (payload.reviewed, .reviewer),
        ]
        var seen = Set<String>()
        var merged: [OpenPullRequest] = []
        var inaccessible = 0
        for (connection, role) in sources {
            for node in connection?.nodes ?? [] {
                guard let pullRequest = node.flatMap({ $0.pullRequest(role: role) }) else {
                    inaccessible += 1
                    continue
                }
                if seen.insert(pullRequest.id).inserted {
                    merged.append(pullRequest)
                }
            }
        }
        merged.sort { $0.updatedAt > $1.updatedAt }
        return OpenPullRequestsResult(pullRequests: merged, inaccessibleCount: inaccessible)
    }
}

struct SearchResponse: Decodable {
    struct Payload: Decodable {
        let authored: Connection?
        let requested: Connection?
        let reviewed: Connection?
    }

    struct Connection: Decodable {
        let nodes: [Node?]
    }

    struct Node: Decodable {
        let number: Int?
        let title: String?
        let url: String?
        let updatedAt: String?
        let repository: Repository?
        let author: Actor?
        let commits: Commits?

        func pullRequest(role: PullRequestRole) -> OpenPullRequest? {
            guard let number,
                  let title,
                  let repository,
                  let updatedAt,
                  let updated = try? Date(updatedAt, strategy: .iso8601) else {
                return nil
            }
            let rollup = commits?.nodes?.compactMap { $0 }.last?.commit?.statusCheckRollup?.state
            return OpenPullRequest(
                repository: repository.nameWithOwner,
                number: number,
                title: title,
                author: author?.login ?? "ghost",
                url: url.flatMap(URL.init(string:)),
                updatedAt: updated,
                role: role,
                ciStatus: PullRequestCIStatus(rollupState: rollup)
            )
        }
    }

    struct Repository: Decodable {
        let nameWithOwner: String
    }

    struct Actor: Decodable {
        let login: String
    }

    struct Commits: Decodable {
        let nodes: [CommitNode?]?
    }

    struct CommitNode: Decodable {
        let commit: Commit?
    }

    struct Commit: Decodable {
        let statusCheckRollup: Rollup?
    }

    struct Rollup: Decodable {
        let state: String?
    }

    struct GraphQLError: Decodable {
        let type: String?
        let message: String?
    }

    let data: Payload?
    let errors: [GraphQLError]?
}

public struct GitHubClient: Sendable {
    public static let defaultBaseURL = URL(string: "https://api.github.com")!

    public let baseURL: URL
    private let session: URLSession

    public init(baseURL: URL = GitHubClient.defaultBaseURL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    public func viewerLogin(token: String) async throws -> String {
        let data = try await send(request(path: "user", method: "GET", token: token))
        do {
            return try JSONDecoder().decode(Viewer.self, from: data).login
        } catch {
            throw GitHubError.decoding(String(describing: error))
        }
    }

    public func openPullRequests(token: String, login: String) async throws -> OpenPullRequestsResult {
        var req = request(path: "graphql", method: "POST", token: token)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try JSONEncoder().encode(GraphQLRequest(
            query: OpenPullRequests.query,
            variables: OpenPullRequests.variables(login: login)
        ))
        return try OpenPullRequests.decode(try await send(req))
    }

    private struct Viewer: Decodable {
        let login: String
    }

    private struct GraphQLRequest: Encodable {
        let query: String
        let variables: [String: String]
    }

    private func request(path: String, method: String, token: String) -> URLRequest {
        var req = URLRequest(url: baseURL.appendingPathComponent(path))
        req.httpMethod = method
        req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        req.setValue("application/vnd.github+json", forHTTPHeaderField: "Accept")
        req.setValue("2022-11-28", forHTTPHeaderField: "X-GitHub-Api-Version")
        return req
    }

    private func send(_ req: URLRequest) async throws -> Data {
        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: req)
        } catch {
            throw GitHubError.transport(String(describing: error))
        }
        guard let http = response as? HTTPURLResponse else {
            throw GitHubError.transport("HTTP 응답이 아님")
        }
        if let error = GitHubError.classify(
            statusCode: http.statusCode,
            remaining: http.value(forHTTPHeaderField: "X-RateLimit-Remaining"),
            reset: http.value(forHTTPHeaderField: "X-RateLimit-Reset"),
            retryAfter: http.value(forHTTPHeaderField: "Retry-After")
        ) {
            throw error
        }
        return data
    }
}
