// 검증 시나리오: test-approval-gate.md#시나리오 16
import XCTest

final class ApprovalKeyRevokeUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testRevokingOneKeyBlocksOnlyThatKeyAndKeepsItsPendingRequest() throws {
        let token = TodoUI.uniqueToken()
        let revokedName = "infra-ci-\(token)"
        let keptName = "staging-ci-\(token)"
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let revokedCaller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: revokedName))
        let keptCaller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: keptName))
        let created = ApprovalIDs()
        addTeardownBlock {
            created.ids.forEach { _ = api.call("POST", "/api/approvals/\($0)/reject") }
            _ = api.call("POST", "/api/approval-keys/\(revokedCaller.id)/revoke")
            _ = api.call("POST", "/api/approval-keys/\(keptCaller.id)/revoke")
        }
        let pendingExternalID = "pending-\(token)"
        let pending = try XCTUnwrap(revokedCaller.create(pendingExternalID, context: "요청 \(pendingExternalID)", timeout: 1800))
        created.ids.append(pending)

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let screen = ApprovalKeysScreen.open(in: app)
        let revoke = screen.revoke(revokedName)
        XCTAssertTrue(revoke.waitForExistence(timeout: 15), "\(revokedName) should offer 폐기")

        revoke.tap()
        let alert = app.alerts.firstMatch
        XCTAssertTrue(alert.waitForExistence(timeout: 5), "폐기 should ask for confirmation")
        alert.buttons["취소"].tap()
        XCTAssertTrue(TodoUI.waitToDisappear(alert, timeout: 5), "취소 should close the confirmation")
        XCTAssertTrue(revoke.exists, "Cancelling should leave the key active")
        XCTAssertNil(ApprovalKeysAPI.key(api, named: revokedName)?["revoked_at"] as? Int, "Cancelling should not revoke")

        revoke.tap()
        XCTAssertTrue(alert.waitForExistence(timeout: 5), "폐기 should ask for confirmation")
        alert.buttons["폐기"].tap()
        XCTAssertTrue(screen.revoked(revokedName).waitForExistence(timeout: 15), "The revoked key should stay listed with its revoke date")
        XCTAssertTrue(screen.revoked(revokedName).label.hasPrefix("폐기됨 "))
        XCTAssertFalse(screen.revoke(revokedName).exists, "A revoked key should offer no way to re-enable it")
        XCTAssertTrue(screen.revoke(keptName).exists, "The other key should stay active")

        let blocked = BackendAPI(token: revokedCaller.key).call(
            "POST", "/api/external/approvals", json: ["externalId": "after-\(token)", "context": "c"]
        )
        XCTAssertEqual(blocked.status, 401, "The revoked key should not create requests: \(blocked.text)")
        let blockedPoll = BackendAPI(token: revokedCaller.key).call("GET", "/api/external/approvals/\(pending)")
        XCTAssertEqual(blockedPoll.status, 401, "The revoked key should not read requests: \(blockedPoll.text)")
        let kept = try XCTUnwrap(keptCaller.create("kept-\(token)", context: "요청 kept-\(token)", timeout: 1800))
        created.ids.append(kept)

        screen.goBack()
        let list = ApprovalsScreen(app: app)
        XCTAssertTrue(list.title.waitForExistence(timeout: 10), "Going back should land on 승인")
        let row = list.row(pendingExternalID)
        if !row.waitForExistence(timeout: 5) {
            list.pullToRefresh()
        }
        XCTAssertTrue(row.waitForExistence(timeout: 15), "The revoked key's pending request should stay in the list")
        XCTAssertTrue(row.label.contains("폐기된 키로 생성됨"), "The pending request should carry the revoked-key warning: \(row.label)")
        XCTAssertEqual(ApprovalLookup.status(api, pending), "PENDING", "Revoking should not decide the pending request")
        app.terminate()
    }
}
