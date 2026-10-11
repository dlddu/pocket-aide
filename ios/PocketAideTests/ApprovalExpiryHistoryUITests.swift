// 검증 시나리오: test-approval-gate.md#시나리오 11
import XCTest

final class ApprovalExpiryHistoryUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testExpiredRequestLeavesPendingAndStaysInHistoryAsExpired() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "infra-ci-\(token)"), "A caller key should be issued")
        let expiring = "e-\(token)"
        let waiting = "n-\(token)"
        let expiringID = try XCTUnwrap(caller.create(expiring, context: "만료 요청 \(expiring)", timeout: 5))
        let waitingID = try XCTUnwrap(caller.create(waiting, context: "대기 요청 \(waiting)", timeout: nil))
        addTeardownBlock {
            _ = api.call("POST", "/api/approvals/\(waitingID)/reject")
            _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        Thread.sleep(forTimeInterval: 6)

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "The 승인 tab should be in the tab bar")
        tab.tap()
        let pending = ApprovalsScreen(app: app)
        if !pending.title.waitForExistence(timeout: 15) {
            tab.tap()
        }
        XCTAssertTrue(pending.title.waitForExistence(timeout: 30), "The 승인 screen should open")
        if !pending.row(waiting).waitForExistence(timeout: 5) {
            pending.pullToRefresh()
        }
        XCTAssertTrue(pending.row(waiting).waitForExistence(timeout: 15), "Step 2: N should stay in the pending list")
        XCTAssertFalse(pending.row(expiring).exists, "Step 2: E should be gone from the pending list")

        let history = ApprovalHistoryScreen.open(in: app)
        let row = history.row(expiring)
        XCTAssertTrue(row.waitForExistence(timeout: 15), "Step 2: E should be in 처리 이력")
        XCTAssertTrue(row.label.contains("만료"), "Step 2: E should read as 만료: \(row.label)")
        XCTAssertTrue(row.label.contains("만료 — 결정 없음"), "Step 2: E's processing mode should be 만료: \(row.label)")
        XCTAssertFalse(history.row(waiting).exists, "Step 2: N should not be in 처리 이력")

        XCTAssertEqual(ApprovalLookup.status(api, expiringID), "EXPIRED", "E should read EXPIRED")
        let stored = try XCTUnwrap(ApprovalHistoryAPI.item(api, externalID: expiring), "The history API should list E")
        XCTAssertEqual(stored["decision_mode"] as? String, "expired")

        history.goBack()
        XCTAssertTrue(pending.title.waitForExistence(timeout: 10), "Going back should land on 승인")
        XCTAssertTrue(pending.row(waiting).waitForExistence(timeout: 15), "N should still be pending after visiting 처리 이력")
        app.terminate()
    }
}
