import Foundation

public enum NotificationOutcomes: String, Codable, CaseIterable, Identifiable, Sendable {
    case both
    case success
    case failure

    public var id: String { rawValue }

    public var label: String {
        switch self {
        case .both: return "둘 다"
        case .success: return "성공만"
        case .failure: return "실패만"
        }
    }
}

public struct NotificationSettings: Codable, Equatable, Sendable {
    public var enabled: Bool
    public var outcomes: NotificationOutcomes

    public init(enabled: Bool = true, outcomes: NotificationOutcomes = .both) {
        self.enabled = enabled
        self.outcomes = outcomes
    }
}

public struct NotificationSettingsPatch: Encodable, Equatable, Sendable {
    public var enabled: Bool?
    public var outcomes: NotificationOutcomes?

    public init(enabled: Bool? = nil, outcomes: NotificationOutcomes? = nil) {
        self.enabled = enabled
        self.outcomes = outcomes
    }
}

public extension APIClient {
    func notificationSettings() async throws -> NotificationSettings {
        try await get("/api/notification-settings", authenticated: true)
    }

    func updateNotificationSettings(_ patch: NotificationSettingsPatch) async throws -> NotificationSettings {
        try await self.patch("/api/notification-settings", body: patch, authenticated: true)
    }
}
