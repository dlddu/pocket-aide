import Foundation

public struct SignOutReason: Codable, Equatable, Sendable {
    public let detail: String
    public let at: Date

    public init(detail: String, at: Date) {
        self.detail = detail
        self.at = at
    }

    public func message(timeZone: TimeZone = .current) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.timeZone = timeZone
        formatter.dateFormat = "M월 d일 a h:mm"
        return "\(formatter.string(from: at))에 자동으로 로그아웃됐어요.\n\n\(detail)"
    }

    public static func refreshHTTPFailure(status: Int, body: String, at: Date) -> SignOutReason {
        let verdict = (400..<500).contains(status) ? "토큰 갱신 거절" : "토큰 갱신 실패"
        return SignOutReason(detail: "\(verdict) — HTTP \(status)\(suffix(oauthError(body)))", at: at)
    }

    public static func refreshFailure(_ message: String, at: Date) -> SignOutReason {
        SignOutReason(detail: "토큰 갱신 실패 — \(oneLine(message))", at: at)
    }

    public static func afterUnauthorized(_ stored: SignOutReason?, apiBody: String, at: Date) -> SignOutReason {
        let api = "API 응답 — HTTP 401\(suffix(oneLine(apiBody)))"
        let detail = [stored?.detail, api].compactMap { $0 }.joined(separator: "\n")
        return SignOutReason(detail: detail, at: stored?.at ?? at)
    }

    private static let maxDetailLength = 200

    private static func oauthError(_ body: String) -> String {
        guard let object = try? JSONSerialization.jsonObject(with: Data(body.utf8)) as? [String: Any],
              let error = object["error"] as? String else {
            return oneLine(body)
        }
        guard let description = object["error_description"] as? String, !description.isEmpty else {
            return oneLine(error)
        }
        return oneLine("\(error): \(description)")
    }

    private static func suffix(_ text: String) -> String {
        text.isEmpty ? "" : ": \(text)"
    }

    private static func oneLine(_ text: String) -> String {
        let flat = text.split(whereSeparator: \.isWhitespace).joined(separator: " ")
        return flat.count <= maxDetailLength ? flat : String(flat.prefix(maxDetailLength)) + "…"
    }
}

public struct SignOutReasonStore {
    private static let key = "auth.signOutReason"
    private let defaults: UserDefaults

    public init(defaults: UserDefaults? = UserDefaults(suiteName: Weather.appGroup)) {
        self.defaults = defaults ?? .standard
    }

    public func load() -> SignOutReason? {
        guard let data = defaults.data(forKey: Self.key) else { return nil }
        return try? JSONDecoder().decode(SignOutReason.self, from: data)
    }

    public func save(_ reason: SignOutReason) {
        guard let data = try? JSONEncoder().encode(reason) else { return }
        defaults.set(data, forKey: Self.key)
    }

    public func clear() {
        defaults.removeObject(forKey: Self.key)
    }
}
