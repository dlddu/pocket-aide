// 검증 시나리오: test-github-monitor.md#시나리오 2
import XCTest

final class PullRequestsAppearAndCloseUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func assertRow(_ number: Int, reads parts: [String], in app: XCUIApplication) {
        let row = OpenPullRequestsUI.row(number, in: app)
        XCTAssertTrue(row.waitForExistence(timeout: 15), "PR #\(number) should be listed")
        let label = row.label
        for part in parts {
            XCTAssertTrue(label.contains(part), "PR #\(number) row should read \(part): \(label)")
        }
        XCTAssertNotNil(
            label.range(of: "[0-9]+ ?(분|min)", options: .regularExpression),
            "PR #\(number) row should read when it was last updated: \(label)"
        )
    }

    func testAuthoredAndReviewRequestedPullRequestsAreListedThenLeaveOnceMergedOrClosed() {
        var app = OpenPullRequestsUI.launchDisconnected()
        OpenPullRequestsUI.connect(OpenPullRequestsUI.Token.closing(), in: app)
        OpenPullRequestsUI.selectFilter("all", in: app)
        assertRow(
            21,
            reads: ["dlddu/pocket-aide-e2e", "#21", "e2e authored to be merged", "pocket-aide-e2e", "작성자"],
            in: app
        )
        assertRow(
            22,
            reads: ["dlddu/pocket-aide-e2e", "#22", "e2e review requested to be closed", "octocat", "리뷰어"],
            in: app
        )

        app.terminate()
        app = OpenPullRequestsUI.launch()
        OpenPullRequestsUI.openSheet(in: app)
        XCTAssertTrue(
            OpenPullRequestsUI.element("openprs.account.login", in: app).waitForExistence(timeout: 15),
            "The connection should survive the relaunch"
        )
        XCTAssertTrue(
            app.staticTexts["열려 있는 PR이 없습니다"].waitForExistence(timeout: 15),
            "Once both PRs are merged or closed the list should be empty"
        )
        XCTAssertFalse(OpenPullRequestsUI.row(21, in: app).exists, "The merged PR should leave the list")
        XCTAssertFalse(OpenPullRequestsUI.row(22, in: app).exists, "The closed PR should leave the list")
        XCTAssertEqual(OpenPullRequestsUI.rows(in: app).count, 0, "No PR should stay listed")

        OpenPullRequestsUI.disconnect(in: app)
    }
}
