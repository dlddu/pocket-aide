// 검증 시나리오: test-approval-gate.md#시나리오 1
import XCTest

final class ApprovalKeyIssueUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testIssuedKeyShowsItsSecretOnlyOnceAndStaysWithItsOwner() throws {
        let token = TodoUI.uniqueToken()
        let name = "infra-ci-\(token)"
        let owner = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        addTeardownBlock { ApprovalKeysAPI.revokeAll(owner, named: name) }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        var screen = ApprovalKeysScreen.open(in: app)

        screen.issue(name)
        XCTAssertTrue(screen.issuedTitle.waitForExistence(timeout: 15), "Issuing should show the key once")
        let secret = screen.issuedSecret.label
        XCTAssertTrue(secret.hasPrefix("pak_"), "The secret should be shown right after issuing: \(secret)")
        XCTAssertTrue(screen.copy.isHittable, "The secret should be copyable")
        screen.copy.tap()
        XCTAssertTrue(RoutinesScreen.eventually(timeout: 5) { screen.copy.label == "복사됨" }, "복사 should confirm the copy")
        screen.done.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.issuedTitle, timeout: 10), "완료 should close the sheet")

        XCTAssertTrue(screen.row(name).waitForExistence(timeout: 15), "\(name) should be listed")
        let meta = screen.meta(name).label
        XCTAssertTrue(meta.hasPrefix(String(secret.prefix(12)) + "…"), "The list should show only the start of the key: \(meta)")
        XCTAssertTrue(meta.hasSuffix("마지막 사용 없음"), "A fresh key should have no last use yet: \(meta)")
        XCTAssertTrue(screen.secretIsNowhere(secret), "The secret should not be on screen after the sheet closed")

        screen.goBack()
        XCTAssertTrue(app.staticTexts["approvals.title"].waitForExistence(timeout: 10), "Going back should land on 승인")
        screen = ApprovalKeysScreen.open(in: app)
        XCTAssertTrue(screen.row(name).waitForExistence(timeout: 15), "\(name) should still be listed on return")
        XCTAssertTrue(screen.secretIsNowhere(secret), "The secret should not be shown again on return")

        screen.issue(name)
        XCTAssertTrue(screen.issueNotice.waitForExistence(timeout: 15), "A second key with the same name should be refused")
        XCTAssertFalse(screen.issuedTitle.exists, "The refused issue should not show a secret")
        screen.cancel.tap()

        let stored = try XCTUnwrap(ApprovalKeysAPI.key(owner, named: name), "The server should list \(name)")
        XCTAssertNil(stored["key"], "The key list should never carry the secret")
        XCTAssertEqual(ApprovalKeysAPI.keys(owner).filter { $0["name"] as? String == name }.count, 1, "Only one \(name) should exist")

        addTeardownBlock { ApprovalUserSwitch.restoreDefaultUser() }
        let other = "e2e-approval-keys-other-\(token)"
        ApprovalUserSwitch.switchUser(to: other, in: app)
        screen = ApprovalKeysScreen.open(in: app)
        XCTAssertTrue(screen.empty.waitForExistence(timeout: 20), "The other user should start with no caller keys")
        XCTAssertFalse(screen.row(name).exists, "\(name) should not be in the other user's list")
        let second = try XCTUnwrap(BackendAPI.signIn(subject: other), "The test runner should get the other user's token")
        XCTAssertNil(ApprovalKeysAPI.key(second, named: name), "The other user should own no \(name)")
        app.terminate()
    }
}
