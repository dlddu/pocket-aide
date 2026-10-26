import Foundation

/// One notification an external caller sent through PocketAide, with who sent it (PRD-12 AC17).
public struct ExternalNotification: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let title: String
    public let body: String
    public let callerName: String
    public let callerKind: String
    public let requester: String
    public let devices: Int
    public let sent: Int
    public let readAt: Int64?
    public let createdAt: Int64

    enum CodingKeys: String, CodingKey {
        case id, title, body, requester, devices, sent
        case callerName = "caller_name"
        case callerKind = "caller_kind"
        case readAt = "read_at"
        case createdAt = "created_at"
    }

    public init(
        id: Int64,
        title: String,
        body: String,
        callerName: String,
        callerKind: String,
        requester: String,
        devices: Int,
        sent: Int,
        readAt: Int64?,
        createdAt: Int64
    ) {
        self.id = id
        self.title = title
        self.body = body
        self.callerName = callerName
        self.callerKind = callerKind
        self.requester = requester
        self.devices = devices
        self.sent = sent
        self.readAt = readAt
        self.createdAt = createdAt
    }

    public var isUnread: Bool { readAt == nil }
}

/// A page of the notification history plus the user's total unread count.
public struct ExternalNotificationsPage: Decodable, Equatable, Sendable {
    public let items: [ExternalNotification]
    public let unreadCount: Int

    enum CodingKeys: String, CodingKey {
        case items
        case unreadCount = "unread_count"
    }

    public init(items: [ExternalNotification], unreadCount: Int) {
        self.items = items
        self.unreadCount = unreadCount
    }
}

struct MarkExternalNotificationsReadPayload: Encodable {
    let upToID: Int64

    enum CodingKeys: String, CodingKey {
        case upToID = "up_to_id"
    }
}

struct MarkExternalNotificationsReadResponse: Decodable {
    let updated: Int
}

public extension APIClient {
    func externalNotifications(limit: Int? = nil, before: Int64? = nil) async throws -> ExternalNotificationsPage {
        var path = "/api/external-notifications"
        var query: [String] = []
        if let limit { query.append("limit=\(limit)") }
        if let before { query.append("before=\(before)") }
        if !query.isEmpty { path += "?" + query.joined(separator: "&") }
        return try await get(path, authenticated: true)
    }

    /// Marks the user's unread notifications up to and including `id` as seen.
    @discardableResult
    func markExternalNotificationsRead(upTo id: Int64) async throws -> Int {
        let response: MarkExternalNotificationsReadResponse = try await post(
            "/api/external-notifications/read",
            body: MarkExternalNotificationsReadPayload(upToID: id),
            authenticated: true
        )
        return response.updated
    }
}

/// Display strings for the notification history screen (PRD-12 AC17).
public enum ExternalNotificationFormatting {
    public static let pageSize = 50

    public static func callerLabel(_ item: ExternalNotification) -> String {
        "\(item.callerName) · \(item.callerKind == "k8s_sa" ? "SA" : "키")"
    }

    public static func requesterLabel(_ item: ExternalNotification) -> String {
        item.requester.isEmpty ? "요청 주체 미상" : "요청 \(item.requester)"
    }

    public static func deliveryLabel(_ item: ExternalNotification) -> String {
        item.devices == 0 ? "등록된 기기 없음" : "\(item.sent)/\(item.devices) 기기"
    }

    public static func time(_ item: ExternalNotification, now: Date = Date(), calendar: Calendar = .current) -> String {
        ApprovalFormatting.requestTime(item.createdAt, now: now, calendar: calendar)
    }
}

public struct WidgetNotificationsSummary: Equatable, Sendable {
    public let latest: ExternalNotification?
    public let moreCount: Int

    public init(latest: ExternalNotification?, moreCount: Int) {
        self.latest = latest
        self.moreCount = moreCount
    }
}

/// What the widget's 「PocketAide 알림」 slice shows: unread caller notifications (PRD-8 AC6, 2026-10-27).
public enum WidgetNotifications {
    public static func summarize(_ page: ExternalNotificationsPage) -> WidgetNotificationsSummary {
        let latest = page.items.filter(\.isUnread).max { $0.id < $1.id }
        guard let latest else { return WidgetNotificationsSummary(latest: nil, moreCount: 0) }
        return WidgetNotificationsSummary(latest: latest, moreCount: max(0, page.unreadCount - 1))
    }
}
