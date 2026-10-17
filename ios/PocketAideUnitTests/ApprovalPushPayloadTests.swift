import XCTest
@testable import PocketAideAPI

final class ApprovalPushPayloadTests: XCTestCase {
    func testApprovalPushOpensTheRequestDetailLink() {
        let url = PushDeepLink.url(fromUserInfo: ["aps": [:], "approval_id": "0f8c-1d2e"])
        XCTAssertEqual(url?.absoluteString, "pocketaide://approvals?id=0f8c-1d2e")
        XCTAssertEqual(url.flatMap(ApprovalPushPayload.approvalID(fromDeepLink:)), "0f8c-1d2e")
    }

    func testPushWithoutApprovalIDStillGoesToThePRMonitor() {
        XCTAssertEqual(
            PushDeepLink.url(fromUserInfo: ["event_id": 42])?.absoluteString,
            "pocketaide://pr-monitor?eventId=42"
        )
        XCTAssertEqual(PushDeepLink.url(fromUserInfo: [:])?.absoluteString, "pocketaide://pr-monitor")
    }

    func testApprovalIDMustBeANonEmptyString() {
        XCTAssertNil(ApprovalPushPayload.approvalID(from: ["approval_id": ""]))
        XCTAssertNil(ApprovalPushPayload.approvalID(from: ["approval_id": 7]))
        XCTAssertNil(ApprovalPushPayload.deepLinkURL(fromUserInfo: ["event_id": 42]))
    }

    func testDeepLinkParsingIgnoresOtherHostsAndMissingIDs() throws {
        XCTAssertNil(ApprovalPushPayload.approvalID(fromDeepLink: try XCTUnwrap(URL(string: "pocketaide://pr-monitor?id=1"))))
        XCTAssertNil(ApprovalPushPayload.approvalID(fromDeepLink: try XCTUnwrap(URL(string: "pocketaide://approvals"))))
        XCTAssertNil(ApprovalPushPayload.approvalID(fromDeepLink: try XCTUnwrap(URL(string: "pocketaide://approvals?id="))))
    }

    func testLockScreenActionsMapToDecisionsAndReportTheResult() {
        XCTAssertEqual(ApprovalNotificationAction.categoryIdentifier, "APPROVAL_REQUEST")
        XCTAssertEqual(ApprovalNotificationAction.decision(forActionIdentifier: ApprovalNotificationAction.approveIdentifier), .approve)
        XCTAssertEqual(ApprovalNotificationAction.decision(forActionIdentifier: ApprovalNotificationAction.rejectIdentifier), .reject)
        XCTAssertNil(ApprovalNotificationAction.decision(forActionIdentifier: "com.apple.UNNotificationDefaultActionIdentifier"))
        XCTAssertEqual(ApprovalNotificationAction.title(.approve), "승인")
        XCTAssertEqual(ApprovalNotificationAction.title(.reject), "거절")

        let approved = ApprovalRequest(id: "r1", status: .approved, context: "prod 배포 v1.42\n변경 3건", createdAt: 1)
        XCTAssertEqual(ApprovalNotificationAction.resultNotice(.success(.decided(approved))), "승인됨 — prod 배포 v1.42")
        let rejected = ApprovalRequest(id: "r2", status: .rejected, context: "캐시 비움", createdAt: 1)
        XCTAssertEqual(ApprovalNotificationAction.resultNotice(.success(.decided(rejected))), "거절됨 — 캐시 비움")
        let expired = ApprovalRequest(id: "r3", status: .expired, context: "c", createdAt: 1)
        XCTAssertEqual(ApprovalNotificationAction.resultNotice(.success(.conflict(expired))), "만료된 요청이라 결정되지 않았습니다")
        XCTAssertEqual(ApprovalNotificationAction.resultNotice(.success(.conflict(approved))), "이미 처리된 요청입니다 — 승인됨")
        XCTAssertEqual(
            ApprovalNotificationAction.resultNotice(.failure(URLError(.notConnectedToInternet))),
            "결정되지 않았습니다. 연결을 확인한 뒤 다시 시도하세요"
        )
    }
}
