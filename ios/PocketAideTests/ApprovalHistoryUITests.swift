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
        XCTAssertEqual(history.visibleRows().map(\.id), [expired, manual], "Step 2: the new user's history should hold only E and M")

        for index in 0..<26 {
            let externalID = "p\(index)-\(token)"
            let id = try XCTUnwrap(caller.create(externalID, context: "처리 요청 \(externalID)", timeout: 600))
            let verb = index.isMultiple(of: 2) ? "approve" : "reject"
            XCTAssertEqual(user.call("POST", "/api/approvals/\(id)/\(verb)").status, 200, "Step 3: \(externalID) should be processed")
        }
        let expected = ApprovalHistoryAPI.externalIDs(user)
        XCTAssertEqual(expected.count, 28, "Step 3: the history API should list 28 processed requests")
        XCTAssertEqual(Set(expected).count, 28, "Step 3: the history API should not repeat a request")

        history.goBack()
        history = ApprovalHistoryScreen.open(in: app)
        XCTAssertTrue(history.row(expected[0]).waitForExistence(timeout: 15), "Step 3: the newest request should open the history")
        let scrolled = history.scrollThrough(expecting: expected.count)
        XCTAssertEqual(scrolled.duplicates, [], "Step 3: no request should appear twice while scrolling")
        XCTAssertEqual(scrolled.order, expected, "Step 3: scrolling should load every request once, newest first")
        app.terminate()
    }
}
