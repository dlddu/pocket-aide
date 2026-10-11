// 검증 시나리오: test-approval-gate.md#시나리오 8
import XCTest

final class ApprovalCountdownUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testDecisionClosesWhenTheCountdownReachesZero() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "countdown-\(token)"), "A caller key should be issued")
        addTeardownBlock {
            _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let externalID = "countdown-\(token)"
        let id = try XCTUnwrap(caller.create(externalID, context: "카운트다운 \(token)", timeout: 45))
        let detail = ApprovalDetailScreen.open(externalID, in: app)

        XCTAssertTrue(detail.countdown.waitForExistence(timeout: 5), "A pending request with a deadline should count down")
        let first = detail.countdown.label
        Thread.sleep(forTimeInterval: 2.2)
        let second = detail.countdown.label
        XCTAssertLessThan(second, first, "The time left should go down second by second (\(first) → \(second))")
        XCTAssertTrue(detail.approve.isEnabled, "Before the deadline the 승인 button should be enabled")

        XCTAssertTrue(detail.waitForStatus("만료", timeout: 60), "At zero the status should read 만료 but read \(detail.status.label)")
        XCTAssertFalse(detail.approve.isEnabled, "At zero the 승인 button should be disabled")
        XCTAssertFalse(detail.reject.isEnabled, "At zero the 거절 button should be disabled")
        XCTAssertFalse(detail.countdown.exists, "The countdown should be gone once the request expired")
        XCTAssertTrue(
            RoutinesScreen.eventually(timeout: 10) { ApprovalLookup.status(api, id) == "EXPIRED" },
            "The server should also treat the request as expired"
        )
    }
}
