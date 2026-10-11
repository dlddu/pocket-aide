// 검증 시나리오: test-approval-gate.md#시나리오 4
import XCTest

final class ApprovalOwnershipUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func openApprovals(in app: XCUIApplication, _ step: String) -> ApprovalsScreen {
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "\(step): the 승인 tab should be in the tab bar")
        tab.tap()
        let screen = ApprovalsScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "\(step): the 승인 screen should open")
        return screen
    }

    private func pendingIDs(_ api: BackendAPI, _ who: String) -> [String] {
        let reply = api.call("GET", "/api/approvals")
        XCTAssertEqual(reply.status, 200, "\(who)'s pending list should load: \(reply.text)")
        let items = reply.json?["items"] as? [[String: Any]] ?? []
        return items.compactMap { $0["id"] as? String }
    }

    func testRequestIsVisibleOnlyToTheKeyOwner() throws {
        let token = TodoUI.uniqueToken()
        let otherSubject = "e2e-approval-b-\(token.lowercased())"
        let owner = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get user A's token")
        let other = try XCTUnwrap(BackendAPI.signIn(subject: otherSubject), "The test runner should get user B's token")
        let infra = try XCTUnwrap(ApprovalSeed.issueKey(owner, name: "infra-ci-\(token)"), "User A should issue infra-ci")
        let foreign = try XCTUnwrap(ApprovalSeed.issueKey(other, name: "b-ci-\(token)"), "User B should issue b-ci")
        let externalID = "r-\(token)"
        let requestID = try XCTUnwrap(infra.create(externalID, context: "요청 \(externalID)", timeout: 1800), "infra-ci should create R")
        addTeardownBlock {
            _ = owner.call("POST", "/api/approvals/\(requestID)/reject")
            _ = owner.call("POST", "/api/approval-keys/\(infra.id)/revoke")
            _ = other.call("POST", "/api/approval-keys/\(foreign.id)/revoke")
        }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let ownerScreen = openApprovals(in: app, "Step 1 as user A")
        XCTAssertTrue(ownerScreen.row(externalID).waitForExistence(timeout: 15), "Step 1: R should be in user A's pending list")

        addTeardownBlock { AppAccount.restoreDefaultUser() }
        AppAccount.signOut(in: app)
        BackendAPI.setNextAppLoginSubject(otherSubject)
        AppAccount.signIn(in: app)

        let otherScreen = openApprovals(in: app, "Step 1 as user B")
        let empty = app.descendants(matching: .any).matching(identifier: "approvals.empty.state").firstMatch
        XCTAssertTrue(empty.waitForExistence(timeout: 20), "Step 1: user B's 승인 tab should have no pending request")
        XCTAssertFalse(otherScreen.row(externalID).exists, "Step 1: R should not be in user B's pending list")
        XCTAssertEqual(pendingIDs(other, "User B"), [], "Step 1: user B should own no pending request")
        XCTAssertTrue(pendingIDs(owner, "User A").contains(requestID), "Step 1: R should stay in user A's pending list")

        for (method, path) in [
            ("GET", "/api/approvals/\(requestID)"),
            ("POST", "/api/approvals/\(requestID)/approve"),
            ("POST", "/api/approvals/\(requestID)/reject")
        ] {
            let reply = other.call(method, path)
            XCTAssertEqual(reply.status, 404, "Step 2: user B's \(method) \(path) should not find R: \(reply.text)")
        }

        let anonymous = BackendAPI.anonymous("GET", "/api/approvals")
        XCTAssertEqual(anonymous.status, 401, "Step 3: the pending list should need sign-in: \(anonymous.text)")

        let foreignPoll = BackendAPI(token: foreign.key).call("GET", "/api/external/approvals/\(requestID)")
        XCTAssertEqual(foreignPoll.status, 404, "Step 4: b-ci should not learn that R exists: \(foreignPoll.text)")

        let ownerPoll = BackendAPI(token: infra.key).call("GET", "/api/external/approvals/\(requestID)")
        XCTAssertEqual(ownerPoll.status, 200, "infra-ci should still read R: \(ownerPoll.text)")
        XCTAssertEqual(ownerPoll.json?["status"] as? String, "PENDING", "Step 2: R should stay PENDING after user B's attempts")
        app.terminate()
    }
}
