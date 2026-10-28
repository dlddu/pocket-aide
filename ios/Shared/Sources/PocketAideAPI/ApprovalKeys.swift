import Foundation

public struct ApprovalCallerKey: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let name: String
    public let keyPrefix: String
    public let serviceAccount: String?
    public let createdAt: Int64
    public let lastUsedAt: Int64?
    public let revokedAt: Int64?

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case keyPrefix = "key_prefix"
        case serviceAccount = "service_account"
        case createdAt = "created_at"
        case lastUsedAt = "last_used_at"
        case revokedAt = "revoked_at"
    }

    public init(
        id: Int64,
        name: String,
        keyPrefix: String,
        createdAt: Int64,
        lastUsedAt: Int64? = nil,
        revokedAt: Int64? = nil,
        serviceAccount: String? = nil
    ) {
        self.id = id
        self.name = name
        self.keyPrefix = keyPrefix
        self.serviceAccount = serviceAccount
        self.createdAt = createdAt
        self.lastUsedAt = lastUsedAt
        self.revokedAt = revokedAt
    }

    public var isRevoked: Bool { revokedAt != nil }

    public var isServiceAccount: Bool { serviceAccount != nil }
}

public struct IssuedApprovalKey: Decodable, Equatable, Sendable {
    public let key: ApprovalCallerKey
    public let secret: String

    enum CodingKeys: String, CodingKey {
        case secret = "key"
    }

    public init(key: ApprovalCallerKey, secret: String) {
        self.key = key
        self.secret = secret
    }

    public init(from decoder: Decoder) throws {
        key = try ApprovalCallerKey(from: decoder)
        secret = try decoder.container(keyedBy: CodingKeys.self).decode(String.self, forKey: .secret)
    }
}

public enum ApprovalKeyIssueOutcome: Equatable, Sendable {
    case issued(IssuedApprovalKey)
    case duplicateName
}

public enum ApprovalServiceAccountOutcome: Equatable, Sendable {
    case registered(ApprovalCallerKey)
    case duplicateName
    case duplicateServiceAccount
    case invalidServiceAccount
}

public enum ApprovalKeyFormatting {
    public static let duplicateNameNotice = "이미 쓰이고 있는 호출자 이름입니다. 다른 이름으로 발급하세요"
    public static let issuedTitle = "호출자 키를 발급했습니다"
    public static let secretWarning = "이 화면을 닫으면 원문을 다시 볼 수 없습니다. 지금 복사해 호출자의 시크릿에 넣으세요."
    public static let serviceAccountDuplicateNameNotice = "이미 쓰이고 있는 호출자 이름입니다. 다른 이름으로 등록하세요"
    public static let duplicateServiceAccountNotice = "이미 등록된 서비스 어카운트입니다"
    public static let invalidServiceAccountNotice = "네임스페이스와 서비스 어카운트는 소문자·숫자·하이픈으로 63자 이내입니다"

    private static let dnsLabelPattern = #"^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$"#

    private static let dayFormatter: DateFormatter = {
        let f = DateFormatter()
        f.locale = Locale(identifier: "ko_KR")
        f.dateFormat = "MM/dd"
        return f
    }()

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.locale = Locale(identifier: "ko_KR")
        f.dateFormat = "HH:mm"
        return f
    }()

    public static func ordered(_ keys: [ApprovalCallerKey]) -> [ApprovalCallerKey] {
        keys.sorted { lhs, rhs in
            if lhs.isRevoked != rhs.isRevoked { return !lhs.isRevoked }
            return (lhs.createdAt, lhs.id) > (rhs.createdAt, rhs.id)
        }
    }

    public static func meta(_ key: ApprovalCallerKey, now: Date = Date(), calendar: Calendar = .current) -> String {
        let lastUsed = key.lastUsedAt.map { "마지막 사용 \(moment($0, now: now, calendar: calendar))" } ?? "마지막 사용 없음"
        if let account = key.serviceAccount {
            return "SA \(account) · 등록 \(day(key.createdAt)) · \(lastUsed)"
        }
        let issued = "발급 \(day(key.createdAt))"
        return "\(key.keyPrefix)… · \(issued) · \(lastUsed)"
    }

    public static func revokedLabel(_ key: ApprovalCallerKey) -> String? {
        key.revokedAt.map { "폐기됨 \(day($0))" }
    }

    public static func revokeTitle(_ key: ApprovalCallerKey) -> String {
        "\(key.name) 키를 폐기할까요?"
    }

    public static let revokeMessage = "폐기하면 이 키로 요청을 만들거나 조회할 수 없고 되돌릴 수 없습니다. 이미 만든 대기 요청은 목록에 남습니다."
    public static let serviceAccountRevokeMessage = "폐기하면 이 서비스 어카운트의 토큰으로 요청을 만들거나 조회할 수 없고 되돌릴 수 없습니다. 이미 만든 대기 요청은 목록에 남고, 같은 서비스 어카운트는 다시 등록할 수 있습니다."

    public static func revokeMessage(for key: ApprovalCallerKey) -> String {
        key.isServiceAccount ? serviceAccountRevokeMessage : revokeMessage
    }

    public static func trimmedName(_ raw: String) -> String? {
        let name = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        return name.isEmpty ? nil : name
    }

    public static func serviceAccount(namespace rawNamespace: String, name rawName: String) -> String? {
        let namespace = rawNamespace.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = rawName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard isDNSLabel(namespace), isDNSLabel(name) else { return nil }
        return "\(namespace)/\(name)"
    }

    static func isDNSLabel(_ value: String) -> Bool {
        value.range(of: dnsLabelPattern, options: .regularExpression) != nil
    }

    static func day(_ seconds: Int64) -> String {
        dayFormatter.string(from: Date(timeIntervalSince1970: TimeInterval(seconds)))
    }

    static func moment(_ seconds: Int64, now: Date, calendar: Calendar) -> String {
        let date = Date(timeIntervalSince1970: TimeInterval(seconds))
        return calendar.isDate(date, inSameDayAs: now) ? timeFormatter.string(from: date) : dayFormatter.string(from: date)
    }
}

struct ApprovalKeyListResponse: Decodable {
    let keys: [ApprovalCallerKey]
}

struct ApprovalKeyIssuePayload: Encodable {
    let name: String
}

struct ApprovalServiceAccountPayload: Encodable {
    let name: String
    let kind = "k8s_sa"
    let serviceAccount: String

    enum CodingKeys: String, CodingKey {
        case name
        case kind
        case serviceAccount = "service_account"
    }
}

public extension APIClient {
    func listApprovalKeys() async throws -> [ApprovalCallerKey] {
        let response: ApprovalKeyListResponse = try await get("/api/approval-keys", authenticated: true)
        return ApprovalKeyFormatting.ordered(response.keys)
    }

    func issueApprovalKey(name: String) async throws -> ApprovalKeyIssueOutcome {
        do {
            let issued: IssuedApprovalKey = try await post(
                "/api/approval-keys",
                body: ApprovalKeyIssuePayload(name: name),
                authenticated: true
            )
            return .issued(issued)
        } catch APIError.badStatus(409, _) {
            return .duplicateName
        }
    }

    func registerApprovalServiceAccount(name: String, serviceAccount: String) async throws -> ApprovalServiceAccountOutcome {
        do {
            let registered: ApprovalCallerKey = try await post(
                "/api/approval-keys",
                body: ApprovalServiceAccountPayload(name: name, serviceAccount: serviceAccount),
                authenticated: true
            )
            return .registered(registered)
        } catch APIError.badStatus(409, let body) where body.contains("service account already registered") {
            return .duplicateServiceAccount
        } catch APIError.badStatus(409, _) {
            return .duplicateName
        } catch APIError.badStatus(400, let body) where body.contains("service_account") {
            return .invalidServiceAccount
        }
    }

    func revokeApprovalKey(id: Int64) async throws -> ApprovalCallerKey {
        try await post("/api/approval-keys/\(id)/revoke", body: ApprovalDecisionPayload(), authenticated: true)
    }
}
