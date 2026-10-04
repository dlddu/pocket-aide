// 검증 시나리오: test-github-monitor.md#시나리오 11
import XCTest

final class PullRequestsEmptyAndLoadingUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testEmptyAccountShowsTheEmptyStateAndASlowFirstLoadShowsASpinner() {
        var app = OpenPullRequestsUI.launchDisconnected()
        OpenPullRequestsUI.connect(OpenPullRequestsUI.Token.empty, in: app)
        XCTAssertTrue(
            OpenPullRequestsUI.element("openprs.empty.state", in: app).waitForExistence(timeout: 15),
            "An account without open PRs should show the empty state"
        )
        XCTAssertTrue(app.staticTexts["열려 있는 PR이 없습니다"].exists, "The empty state should say there is no open PR")
        XCTAssertEqual(OpenPullRequestsUI.rows(in: app).count, 0, "No PR should be listed for the empty account")
        XCTAssertFalse(OpenPullRequestsUI.element("openprs.loading", in: app).exists, "The spinner should be gone once the empty result arrived")

        OpenPullRequestsUI.disconnect(in: app)
        OpenPullRequestsUI.connect(OpenPullRequestsUI.Token.slow, in: app)
        app.terminate()
        app = OpenPullRequestsUI.launch()
        OpenPullRequestsUI.openSheet(in: app)
        let loading = OpenPullRequestsUI.element("openprs.loading", in: app)
        XCTAssertTrue(loading.waitForExistence(timeout: 12), "The first entry should show a spinner while GitHub has not answered")
        XCTAssertFalse(OpenPullRequestsUI.row(11, in: app).exists, "No PR should be listed before the answer arrives")
        XCTAssertFalse(app.staticTexts["열려 있는 PR이 없습니다"].exists, "The empty state should not stand in for loading")

        XCTAssertTrue(OpenPullRequestsUI.row(11, in: app).waitForExistence(timeout: 40), "The list should replace the spinner once the answer arrives")
        XCTAssertFalse(loading.exists, "The spinner should disappear once the list is shown")

        OpenPullRequestsUI.disconnect(in: app)
    }
}
