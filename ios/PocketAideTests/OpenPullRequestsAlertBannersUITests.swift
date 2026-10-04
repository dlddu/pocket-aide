// 검증 시나리오: test-github-monitor.md#시나리오 10
import XCTest

final class OpenPullRequestsAlertBannersUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func banner(_ kind: String, in app: XCUIApplication) -> XCUIElement {
        OpenPullRequestsUI.element("openprs.banner.\(kind)", in: app)
    }

    func testBannersNameTheCauseAndClearOnceResolved() {
        var app = OpenPullRequestsUI.launchDisconnected()
        OpenPullRequestsUI.submit(OpenPullRequestsUI.Token.expires, in: app)
        XCTAssertTrue(banner("token", in: app).waitForExistence(timeout: 15), "A token GitHub rejects should raise the token banner")

        app.terminate()
        app = OpenPullRequestsUI.launch()
        OpenPullRequestsUI.openSheet(in: app)
        XCTAssertTrue(banner("token", in: app).waitForExistence(timeout: 15), "Entering with the expired token should show the token banner")
        XCTAssertEqual(OpenPullRequestsUI.rows(in: app).count, 0, "No PR should be listed while the token is rejected")
        let reconnect = app.buttons["openprs.banner.action"]
        XCTAssertEqual(reconnect.label, "토큰 다시 연결")
        reconnect.tap()

        OpenPullRequestsUI.submit(OpenPullRequestsUI.Token.valid, in: app)
        XCTAssertTrue(OpenPullRequestsUI.row(11, in: app).waitForExistence(timeout: 15), "The list should load once a valid token is connected")
        XCTAssertTrue(OpenPullRequestsUI.row(14, in: app).exists, "The whole list should be shown after resolving")
        XCTAssertFalse(banner("token", in: app).exists, "The token banner should disappear once resolved")
        XCTAssertFalse(banner("ratelimit", in: app).exists, "No rate-limit banner should be shown")

        XCTAssertTrue(banner("access", in: app).waitForExistence(timeout: 5), "A PR the token cannot see should raise the access banner")
        XCTAssertEqual(app.buttons["openprs.banner.action"].label, "토큰 바꾸기")

        app.buttons["openprs.token.replace"].tap()
        OpenPullRequestsUI.submit(OpenPullRequestsUI.Token.rateLimited, in: app)
        XCTAssertTrue(banner("ratelimit", in: app).waitForExistence(timeout: 15), "An exhausted rate limit should raise the rate-limit banner")
        XCTAssertFalse(banner("access", in: app).exists, "Only the rate-limit cause should be shown")
        let retry = app.buttons["openprs.banner.action"]
        XCTAssertEqual(retry.label, "다시 시도")
        retry.tap()
        XCTAssertTrue(banner("ratelimit", in: app).waitForExistence(timeout: 15), "Retrying while still limited should keep the banner")

        app.buttons["openprs.disconnect.button"].tap()
        XCTAssertTrue(app.secureTextFields["openprs.token.field"].waitForExistence(timeout: 10), "Disconnect should return to the token form")
    }
}
