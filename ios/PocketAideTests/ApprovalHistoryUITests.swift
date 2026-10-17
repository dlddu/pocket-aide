// 검증 시나리오: test-approval-gate.md#시나리오 12
import XCTest

final class ApprovalHistoryUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHistoryIsNewestFirstWithItsModeAndScrollsWithoutGapsOrDuplicates() throws {
        let token = TodoUI.uniqueToken()
        let subject = "e2e-approval-history-\(token.lowercased())"
        let user = try XCTUnwrap(BackendAPI.signIn(subject: subject), "The test runner should get the history user's token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(user, name: "infra-ci-\(token)"), "A caller key should be issued")
        addTeardownBlock { _ = user.call("POST", "/api/approval-keys/\(caller.id)/revoke") }

        let manual = "m-\(token)"
        let manualID = try XCTUnwrap(caller.create(manual, context: "수동 결정 요청 \(manual)", timeout: 600))
        XCTAssertEqual(user.call("POST", "/api/approvals/\(manualID)/approve").status, 200, "Step 1: M should be approved")
        Thread.sleep(forTimeInterval: 1.2)
        let autoRejected = "a-\(token)"
        seedAutoRejected(autoRejected, user: user, caller: caller)
        let expired = "e-\(token)"
        _ = try XCTUnwrap(caller.create(expired, context: "만료 요청 \(expired)", timeout: 3))
        Thread.sleep(forTimeInterval: 4)

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        addTeardownBlock { AppAccount.restoreDefaultUser() }
        AppAccount.signOut(in: app)
        BackendAPI.setNextAppLoginSubject(subject)
        AppAccount.signIn(in: app)

        var history = ApprovalHistoryScreen.open(in: app)
        let expiredRow = history.row(expired)
        let manualRow = history.row(manual)
        XCTAssertTrue(expiredRow.waitForExistence(timeout: 15), "Step 2: E should be in 처리 이력")
        XCTAssertTrue(manualRow.waitForExistence(timeout: 15), "Step 2: M should be in 처리 이력")
        XCTAssertLessThan(expiredRow.frame.minY, manualRow.frame.minY, "Step 2: the later-processed E should be above M")
        XCTAssertTrue(expiredRow.label.contains("만료 — 결정 없음"), "Step 2: E should read 만료: \(expiredRow.label)")
        XCTAssertTrue(manualRow.label.contains("수동 결정"), "Step 2: M should read 수동 결정: \(manualRow.label)")
        assertAutoRejected(history.row(autoRejected), between: expiredRow, and: manualRow)
        XCTAssertEqual(history.visibleRows().map(\.id), [expired, autoRejected, manual], "Step 2: the new user's history should hold only E, A and M")

        for index in 0..<26 {
            let externalID = "p\(index)-\(token)"
            let id = try XCTUnwrap(caller.create(externalID, context: "처리 요청 \(externalID)", timeout: 600))
            let verb = index.isMultiple(of: 2) ? "approve" : "reject"
            XCTAssertEqual(user.call("POST", "/api/approvals/\(id)/\(verb)").status, 200, "Step 3: \(externalID) should be processed")
        }
        let expected = ApprovalHistoryAPI.externalIDs(user)
        XCTAssertEqual(expected.count, 29, "Step 3: the history API should list 29 processed requests")
        XCTAssertEqual(Set(expected).count, 29, "Step 3: the history API should not repeat a request")

        history.goBack()
        history = ApprovalHistoryScreen.open(in: app)
        XCTAssertTrue(history.row(expected[0]).waitForExistence(timeout: 15), "Step 3: the newest request should open the history")
        let scrolled = history.scrollThrough(expecting: expected.count)
        XCTAssertEqual(scrolled.duplicates, [], "Step 3: no request should appear twice while scrolling")
        XCTAssertEqual(scrolled.order, expected, "Step 3: scrolling should load every request once, newest first")
        app.terminate()
    }

    private func seedAutoRejected(_ externalID: String, user: BackendAPI, caller: ApprovalCaller) {
        addTeardownBlock { _ = user.call("PATCH", "/api/approval-settings", json: ["auto_response": "NONE"]) }
        let on = user.call("PATCH", "/api/approval-settings", json: ["auto_response": "AUTO_REJECT"])
        XCTAssertEqual(on.status, 200, "Step 1: 자동 거절 should be stored: \(on.text)")
        let reply = BackendAPI(token: caller.key).call(
            "POST", "/api/external/approvals",
            json: ["externalId": externalID, "context": "자동 거절 요청 \(externalID)", "timeoutSeconds": 600]
        )
        XCTAssertEqual(reply.status, 201, "Step 1: creating \(externalID) should succeed: \(reply.text)")
        XCTAssertEqual(reply.json?["status"] as? String, "REJECTED", "Step 1: \(externalID) should be rejected at creation")
        let off = user.call("PATCH", "/api/approval-settings", json: ["auto_response": "NONE"])
        XCTAssertEqual(off.status, 200, "Step 1: 없음 should be restored: \(off.text)")
    }

    private func assertAutoRejected(_ row: XCUIElement, between above: XCUIElement, and below: XCUIElement) {
        XCTAssertTrue(row.waitForExistence(timeout: 15), "Step 2: A should be in 처리 이력")
        XCTAssertLessThan(above.frame.minY, row.frame.minY, "Step 2: the later-processed E should be above A")
        XCTAssertLessThan(row.frame.minY, below.frame.minY, "Step 2: A should be above the earlier-processed M")
        XCTAssertTrue(row.label.contains("자동 거절"), "Step 2: A should read 자동 거절: \(row.label)")
    }
}
