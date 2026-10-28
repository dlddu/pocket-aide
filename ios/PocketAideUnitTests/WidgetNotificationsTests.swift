import XCTest
@testable import PocketAideAPI

final class WidgetNotificationsTests: XCTestCase {
    private func item(_ id: Int64, read: Bool = false, title: String = "백업 완료", body: String = "") -> ExternalNotification {
        ExternalNotification(
            id: id,
            title: title,
            body: body,
            callerName: "infra-ci",
            callerKind: "key",
            requester: "",
            devices: 1,
            sent: 1,
            readAt: read ? 2_000 : nil,
            createdAt: 1_000 + id
        )
    }

    func testLatestUnreadFirstAndRestCountedFromServerTotal() {
        let page = ExternalNotificationsPage(items: [item(5), item(4), item(3, read: true)], unreadCount: 7)
        let summary = WidgetNotifications.summarize(page)
        XCTAssertEqual(summary.latest?.id, 5)
        XCTAssertEqual(summary.moreCount, 6)
    }

    func testReadNotificationsAreLeftOut() {
        let page = ExternalNotificationsPage(items: [item(5, read: true), item(4)], unreadCount: 1)
        let summary = WidgetNotifications.summarize(page)
        XCTAssertEqual(summary.latest?.id, 4)
        XCTAssertEqual(summary.moreCount, 0)
    }

    func testNothingUnreadShowsNothing() {
        XCTAssertEqual(
            WidgetNotifications.summarize(ExternalNotificationsPage(items: [item(1, read: true)], unreadCount: 0)),
            WidgetNotificationsSummary(latest: nil, moreCount: 0)
        )
        XCTAssertNil(WidgetNotifications.summarize(ExternalNotificationsPage(items: [], unreadCount: 0)).latest)
    }
}
