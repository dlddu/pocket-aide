// 검증 시나리오: test-approval-gate.md#시나리오 9
import XCTest

final class ApprovalDecisionUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testApprovalNeedsConfirmationAndReachesTheCaller() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "decision-\(token)"), "A caller key should be issued")
        let externalID = "decision-\(token)"
        let id = try XCTUnwrap(caller.create(externalID, context: "prod 배포 \(token)\n변경 3건", timeout: 1800))
        addTeardownBlock {
            _ = api.call("POST", "/api/approvals/\(id)/reject")
            _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let detail = ApprovalDetailScreen.open(externalID, in: app)
        XCTAssertTrue(detail.waitForStatus("대기 중"), "A new request should be pending")

        detail.choose(detail.approve, thenTap: "취소")
        XCTAssertTrue(detail.waitForStatus("대기 중"), "Cancelling the confirmation should leave the request pending")
        XCTAssertEqual(ApprovalLookup.status(api, id), "PENDING", "Cancelling should not reach the server")

        detail.choose(detail.approve, thenTap: "승인")
        XCTAssertTrue(detail.waitForStatus("승인됨"), "Confirming should approve the request but read \(detail.status.label)")
        XCTAssertEqual(detail.mode.label, "수동 결정", "The detail should say the request was decided by hand")
        XCTAssertTrue(detail.processedAt.exists, "The detail should show when it was processed")
        XCTAssertFalse(detail.approve.exists, "A processed request should have no decision buttons")
        XCTAssertFalse(detail.reject.exists, "A processed request should have no decision buttons")

        let seen = try XCTUnwrap(ApprovalLookup.callerView(caller, id), "The caller should be able to poll the request")
        XCTAssertEqual(seen["status"] as? String, "APPROVED", "The caller should receive APPROVED")
        XCTAssertNotNil(seen["processedAt"] as? Int, "The caller should see processedAt: \(seen)")

        detail.goBack()
        let list = ApprovalsScreen(app: app)
        XCTAssertTrue(TodoUI.waitToDisappear(list.row(externalID), timeout: 15), "The approved request should leave the pending list")
    }
}
