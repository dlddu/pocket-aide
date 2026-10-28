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
        enableAutoApprove(api)

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let screen = ApprovalKeysScreen.open(in: app)
        let revoke = screen.revoke(revokedName)
        XCTAssertTrue(revoke.waitForExistence(timeout: 15), "\(revokedName) should offer 폐기")

        confirm(revoke, with: "취소", in: app)
        XCTAssertTrue(revoke.exists, "Cancelling should leave the key active")
        XCTAssertNil(ApprovalKeysAPI.key(api, named: revokedName)?["revoked_at"] as? Int, "Cancelling should not revoke")

        confirm(revoke, with: "폐기", in: app)
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
        let kept = try createApproved("kept-\(token)", by: keptCaller)
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
        assertDetailWarningThenHistory(pendingExternalID, id: pending, keyName: revokedName, api: api, in: app)
        app.terminate()
    }

    private func enableAutoApprove(_ api: BackendAPI) {
        addTeardownBlock { _ = api.call("PATCH", "/api/approval-settings", json: ["auto_response": "NONE"]) }
        let reply = api.call("PATCH", "/api/approval-settings", json: ["auto_response": "AUTO_APPROVE"])
        XCTAssertEqual(reply.status, 200, "Precondition: 자동 승인 should be stored: \(reply.text)")
    }

    private func createApproved(_ externalID: String, by caller: ApprovalCaller) throws -> String {
        let reply = BackendAPI(token: caller.key).call(
            "POST", "/api/external/approvals",
            json: ["externalId": externalID, "context": "요청 \(externalID)", "timeoutSeconds": 1800]
        )
        XCTAssertEqual(reply.status, 201, "The other key should still create requests: \(reply.text)")
        XCTAssertEqual(reply.json?["status"] as? String, "APPROVED", "In 자동 승인 the other key's request should be APPROVED at creation")
        return try XCTUnwrap(reply.json?["id"] as? String, "The create reply should carry an id")
    }

    private func assertDetailWarningThenHistory(_ externalID: String, id: String, keyName: String, api: BackendAPI, in app: XCUIApplication) {
        let detail = ApprovalDetailScreen.open(externalID, in: app)
        XCTAssertTrue(app.staticTexts["approval.detail.revoked"].waitForExistence(timeout: 15), "The detail should carry the revoked-key warning")
        XCTAssertTrue(detail.waitForStatus("대기 중"), "R should still read 대기 중 in 자동 승인: \(detail.status.label)")
        detail.choose(detail.reject, thenTap: "거절")
        XCTAssertTrue(detail.waitForStatus("거절됨"), "Rejecting by hand should reject R: \(detail.status.label)")
        XCTAssertEqual(ApprovalLookup.status(api, id), "REJECTED", "The server should store R as REJECTED")
        detail.goBack()
        let row = ApprovalHistoryScreen.open(in: app).row(externalID)
        XCTAssertTrue(row.waitForExistence(timeout: 15), "R should be in 처리 이력")
        XCTAssertTrue(row.label.contains(keyName), "R's history entry should keep the key name \(keyName): \(row.label)")
        XCTAssertTrue(row.label.contains("폐기된 키로 생성됨"), "R's history entry should note the revoked key: \(row.label)")
    }

    private func confirm(_ revoke: XCUIElement, with label: String, in app: XCUIApplication) {
        revoke.tap()
        let alert = app.alerts.firstMatch
        XCTAssertTrue(alert.waitForExistence(timeout: 5), "폐기 should ask for confirmation")
        alert.buttons[label].tap()
        XCTAssertTrue(TodoUI.waitToDisappear(alert, timeout: 5), "\(label) should close the confirmation")
    }
}
