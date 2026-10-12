import XCTest
@testable import PocketAideAPI

final class ApprovalHistoryTests: XCTestCase {
    func testDecodesHistoryPageWithAndWithoutCursor() throws {
        let json = #"""
        {"items":[
         {"id":"e1","status":"EXPIRED","external_id":"e","context":"prod 배포","requester_name":"","key_name":"infra-ci","key_revoked":false,"decision_mode":"expired","created_at":1760000000,"expires_at":1760000005,"processed_at":1760000005},
         {"id":"m1","status":"APPROVED","external_id":"m","context":"삭제","requester_name":"bot","key_name":"k3s-mcp","key_revoked":true,"decision_mode":"manual","created_at":1759990000,"expires_at":null,"processed_at":1759990100}
        ],"next_cursor":"1759990100.m1"}
        """#
        let page = try JSONDecoder().decode(ApprovalHistoryPage.self, from: Data(json.utf8))
        XCTAssertEqual(page.items.map(\.externalId), ["e", "m"])
        XCTAssertEqual(page.items[0].status, .expired)
        XCTAssertEqual(page.items[1].processedAt, 1_759_990_100)
        XCTAssertEqual(page.nextCursor, "1759990100.m1")

        let last = try JSONDecoder().decode(ApprovalHistoryPage.self, from: Data(#"{"items":[],"next_cursor":null}"#.utf8))
        XCTAssertTrue(last.items.isEmpty)
        XCTAssertNil(last.nextCursor)
    }

    func testHistoryMetaNamesTheProcessingMode() {
        func item(_ status: ApprovalStatus, _ mode: String?, revoked: Bool = false) -> ApprovalRequest {
            ApprovalRequest(id: "x", status: status, context: "c", keyRevoked: revoked, createdAt: 0, processedAt: 1, decisionMode: mode)
        }
        XCTAssertEqual(ApprovalFormatting.historyMeta(item(.approved, "manual")), "수동 결정")
        XCTAssertEqual(ApprovalFormatting.historyMeta(item(.rejected, "manual")), "수동 결정")
        XCTAssertEqual(ApprovalFormatting.historyMeta(item(.approved, "auto")), "자동 승인")
        XCTAssertEqual(ApprovalFormatting.historyMeta(item(.rejected, "auto")), "자동 거절")
        XCTAssertEqual(ApprovalFormatting.historyMeta(item(.expired, "expired")), "만료 — 결정 없음")
        XCTAssertEqual(ApprovalFormatting.historyMeta(item(.rejected, "manual", revoked: true)), "수동 결정 · 폐기된 키로 생성됨")
    }

    func testHistoryTimeIsTodayYesterdayOrDate() throws {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = try XCTUnwrap(TimeZone(identifier: "Asia/Seoul"))
        let now = Date(timeIntervalSince1970: 1_760_000_000)
        XCTAssertEqual(ApprovalFormatting.historyTime(1_760_000_000 - 60, now: now, calendar: calendar), "17:52:20")
        XCTAssertEqual(ApprovalFormatting.historyTime(1_760_000_000 - 86_400, now: now, calendar: calendar), "어제 17:53")
        XCTAssertEqual(ApprovalFormatting.historyTime(1_760_000_000 - 8 * 86_400, now: now, calendar: calendar), "10/01 17:53")
    }

    func testRequestTimeIsTodayYesterdayOrDate() throws {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = try XCTUnwrap(TimeZone(identifier: "Asia/Seoul"))
        let now = Date(timeIntervalSince1970: 1_760_000_000)
        XCTAssertEqual(ApprovalFormatting.requestTime(1_760_000_000 - 60, now: now, calendar: calendar), "17:52")
        XCTAssertEqual(ApprovalFormatting.requestTime(1_760_000_000 - 86_400, now: now, calendar: calendar), "어제 17:53")
        XCTAssertEqual(ApprovalFormatting.requestTime(1_760_000_000 - 8 * 86_400, now: now, calendar: calendar), "10/01 17:53")
    }

    func testRequestURLKeepsTheQueryOutOfThePath() throws {
        let base = try XCTUnwrap(URL(string: "https://aide.example.com"))
        XCTAssertEqual(
            APIClient.url(base, "/api/approvals/history?cursor=1760000000%2Eab").absoluteString,
            "https://aide.example.com/api/approvals/history?cursor=1760000000%2Eab"
        )
        XCTAssertEqual(APIClient.url(base, "/api/approvals").absoluteString, "https://aide.example.com/api/approvals")
    }
}
