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

    func testExternalNotificationOpensTheHistory() {
        XCTAssertEqual(
            PushDeepLink.url(fromUserInfo: ["aps": [:], "kind": "external_notification"])?.absoluteString,
            "pocketaide://notifications"
        )
        XCTAssertEqual(
            PushDeepLink.url(fromUserInfo: ["kind": "something-else"])?.absoluteString,
            "pocketaide://pr-monitor"
        )
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
}
