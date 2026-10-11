import XCTest
@testable import PocketAideAPI

final class ApprovalsTests: XCTestCase {
    func testDecodesServerApprovalResponse() throws {
        let json = #"""
        {"id":"0f1e2d3c4b5a69788796a5b4c3d2e1f0","status":"PENDING","external_id":"deploy-42",
         "context":"prod 배포 v1.42\n변경 3건","requester_name":"Deploy Bot","key_name":"infra-ci",
         "key_revoked":false,"created_at":1760000000,"expires_at":1760000600,"processed_at":null}
        """#
        let item = try JSONDecoder().decode(ApprovalRequest.self, from: Data(json.utf8))
        XCTAssertEqual(item.status, .pending)
        XCTAssertEqual(item.externalId, "deploy-42")
        XCTAssertEqual(item.requesterName, "Deploy Bot")
        XCTAssertEqual(item.keyName, "infra-ci")
        XCTAssertFalse(item.keyRevoked)
        XCTAssertEqual(item.expiresAt, 1_760_000_600)
        XCTAssertNil(item.processedAt)
    }

    func testDecodesListWithoutDeadline() throws {
        let json = #"{"items":[{"id":"a","status":"PENDING","external_id":"x","context":"c","requester_name":"","key_name":"k","key_revoked":true,"created_at":1,"expires_at":null,"processed_at":null}]}"#
        let response = try JSONDecoder().decode(ApprovalListResponse.self, from: Data(json.utf8))
        XCTAssertEqual(response.items.count, 1)
        XCTAssertNil(response.items[0].expiresAt)
        XCTAssertTrue(response.items[0].keyRevoked)
    }

    func testNewestFirstOrdersByCreationTime() {
        let items = [
            ApprovalRequest(id: "b", context: "둘째", createdAt: 200),
            ApprovalRequest(id: "a", context: "첫째", createdAt: 100),
            ApprovalRequest(id: "c", context: "셋째", createdAt: 300),
        ]
        XCTAssertEqual(ApprovalFormatting.newestFirst(items).map(\.id), ["c", "b", "a"])
    }

    func testRemainingTime() {
        let now = Date(timeIntervalSince1970: 1000)
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: nil, now: now), "기한 없음")
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: 1252, now: now), "04:12 남음")
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: 1000 + 3600 + 61, now: now), "1:01:01 남음")
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: 900, now: now), "00:00 남음")
    }

    func testSummaryUsesFirstNonEmptyLine() {
        XCTAssertEqual(ApprovalFormatting.summary(of: "\n  deployment/dear-baby-api 삭제 \n상세"), "deployment/dear-baby-api 삭제")
        XCTAssertEqual(ApprovalFormatting.summary(of: "한 줄"), "한 줄")
        XCTAssertEqual(ApprovalFormatting.pendingHeader(3), "대기 · 3건")
    }
}
