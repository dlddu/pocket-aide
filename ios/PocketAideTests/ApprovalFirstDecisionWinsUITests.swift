// 검증 시나리오: test-approval-gate.md#시나리오 10
import XCTest

final class ApprovalFirstDecisionWinsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testLateDecisionIsRefusedAndExplained() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "race-\(token)"), "A caller key should be issued")
        let externalID = "race-\(token)"
        let id = try XCTUnwrap(caller.create(externalID, context: "동시 결정 \(token)", timeout: nil))
        addTeardownBlock {
            _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let detail = ApprovalDetailScreen.open(externalID, in: app)
        XCTAssertTrue(detail.waitForStatus("대기 중"), "The request should be pending on this device")

        let first = api.call("POST", "/api/approvals/\(id)/approve")
        XCTAssertEqual(first.status, 200, "The other device's approval should land first: \(first.text)")

        detail.choose(detail.reject, thenTap: "거절")
        XCTAssertTrue(
            RoutinesScreen.eventually(timeout: 15) { detail.notice.exists && detail.notice.label.contains("이미 처리된 요청입니다") },
            "The late rejection should be explained as already processed"
        )
        XCTAssertTrue(detail.waitForStatus("승인됨"), "The detail should show the approval that won but read \(detail.status.label)")
        XCTAssertFalse(detail.reject.exists, "A processed request should have no decision buttons")

        let again = api.call("POST", "/api/approvals/\(id)/reject")
        XCTAssertEqual(again.status, 409, "Deciding a processed request again should be refused: \(again.text)")
        XCTAssertEqual(ApprovalLookup.status(api, id), "APPROVED", "The first decision should stand")
    }
}
