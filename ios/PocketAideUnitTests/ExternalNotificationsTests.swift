import XCTest
@testable import PocketAideAPI

final class ExternalNotificationsTests: XCTestCase {
    func testDecodesHistoryPage() throws {
        let json = """
        {"items":[
          {"id":9,"title":"배포 완료","body":"","caller_name":"homelab-k3s-mcp","caller_kind":"k8s_sa",
           "requester":"jwt:user-123","devices":2,"sent":1,"created_at":1800000000},
          {"id":8,"title":"ping","body":"x","caller_name":"infra-ci","caller_kind":"key",
           "requester":"","devices":0,"sent":0,"read_at":1800000100,"created_at":1799999000}
        ],"unread_count":1}
        """
        let page = try APIClient.decoder.decode(ExternalNotificationsPage.self, from: Data(json.utf8))
        XCTAssertEqual(page.unreadCount, 1)
        XCTAssertEqual(page.items.map(\.id), [9, 8])
        XCTAssertTrue(page.items[0].isUnread)
        XCTAssertEqual(page.items[1].readAt, 1_800_000_100)
    }

    func testCallerRequesterAndDeliveryLabels() {
        let sa = ExternalNotification(
            id: 1, title: "t", body: "", callerName: "homelab-k3s-mcp", callerKind: "k8s_sa",
            requester: "jwt:user-123", devices: 2, sent: 1, readAt: nil, createdAt: 0
        )
        XCTAssertEqual(ExternalNotificationFormatting.callerLabel(sa), "homelab-k3s-mcp · SA")
        XCTAssertEqual(ExternalNotificationFormatting.requesterLabel(sa), "요청 jwt:user-123")
        XCTAssertEqual(ExternalNotificationFormatting.deliveryLabel(sa), "1/2 기기")

        let key = ExternalNotification(
            id: 2, title: "t", body: "", callerName: "infra-ci", callerKind: "key",
            requester: "", devices: 0, sent: 0, readAt: nil, createdAt: 0
        )
        XCTAssertEqual(ExternalNotificationFormatting.callerLabel(key), "infra-ci · 키")
        XCTAssertEqual(ExternalNotificationFormatting.requesterLabel(key), "요청 주체 미상")
        XCTAssertEqual(ExternalNotificationFormatting.deliveryLabel(key), "등록된 기기 없음")
    }

    func testMarkReadPayloadUsesSnakeCase() throws {
        let data = try JSONEncoder().encode(MarkExternalNotificationsReadPayload(upToID: 42))
        XCTAssertEqual(String(decoding: data, as: UTF8.self), #"{"up_to_id":42}"#)
    }
}
