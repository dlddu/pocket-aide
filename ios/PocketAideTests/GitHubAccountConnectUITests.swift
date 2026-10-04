// 검증 시나리오: test-github-monitor.md#시나리오 1
import XCTest

final class GitHubAccountConnectUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func assertDisconnected(_ app: XCUIApplication, _ moment: String) {
        let prompt = OpenPullRequestsUI.element("openprs.connect.prompt", in: app)
        XCTAssertTrue(prompt.waitForExistence(timeout: 10), "\(moment): the sheet should invite connecting a GitHub account")
        XCTAssertTrue(app.secureTextFields["openprs.token.field"].exists, "\(moment): the token form should be shown")
        XCTAssertFalse(OpenPullRequestsUI.element("openprs.account.login", in: app).exists, "\(moment): no handle should remain")
        XCTAssertFalse(app.buttons["openprs.disconnect.button"].exists, "\(moment): there is nothing to disconnect")
        XCTAssertEqual(OpenPullRequestsUI.rows(in: app).count, 0, "\(moment): no GitHub data should be listed")
    }

    func testConnectShowsHandleAndDisconnectReturnsToConnectPrompt() {
        var app = OpenPullRequestsUI.launchDisconnected()
        assertDisconnected(app, "Before connecting")

        OpenPullRequestsUI.submit(OpenPullRequestsUI.Token.valid, in: app)
        let login = OpenPullRequestsUI.element("openprs.account.login", in: app)
        XCTAssertTrue(login.waitForExistence(timeout: 15), "A valid PAT should connect")
        XCTAssertEqual(login.label, OpenPullRequestsUI.handle)
        XCTAssertTrue(OpenPullRequestsUI.row(11, in: app).waitForExistence(timeout: 15), "A connected account should list its open PRs")
        XCTAssertFalse(OpenPullRequestsUI.element("openprs.connect.prompt", in: app).exists, "A connected sheet should not invite connecting")

        app.buttons["openprs.disconnect.button"].tap()
        assertDisconnected(app, "Right after disconnecting")

        OpenPullRequestsUI.closeSheet(in: app)
        OpenPullRequestsUI.openSheet(in: app)
        assertDisconnected(app, "Re-entering the sheet")

        app.terminate()
        app = OpenPullRequestsUI.launch()
        OpenPullRequestsUI.openSheet(in: app)
        assertDisconnected(app, "After relaunching")
    }
}
