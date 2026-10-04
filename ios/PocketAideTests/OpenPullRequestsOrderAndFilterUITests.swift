// 검증 시나리오: test-github-monitor.md#시나리오 4
import XCTest

final class OpenPullRequestsOrderAndFilterUITests: XCTestCase {
    private let newestFirst = [11, 12, 13, 14]

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func assertListed(_ listed: [Int], in app: XCUIApplication, _ moment: String) {
        for number in newestFirst {
            let row = OpenPullRequestsUI.row(number, in: app)
            if listed.contains(number) {
                XCTAssertTrue(row.waitForExistence(timeout: 10), "\(moment): PR #\(number) should be listed")
            } else {
                XCTAssertFalse(row.exists, "\(moment): PR #\(number) should be filtered out")
            }
        }
        let tops = listed.map { OpenPullRequestsUI.row($0, in: app).frame.minY }
        XCTAssertEqual(tops, tops.sorted(), "\(moment): rows should stay newest-updated first \(listed)")
        XCTAssertEqual(Set(tops).count, tops.count, "\(moment): each PR should have its own row")
    }

    func testListIsNewestFirstAndFilterSurvivesRelaunch() {
        var app = OpenPullRequestsUI.launchDisconnected()
        OpenPullRequestsUI.submit(OpenPullRequestsUI.Token.valid, in: app)
        XCTAssertTrue(
            OpenPullRequestsUI.element("openprs.account.login", in: app).waitForExistence(timeout: 15),
            "A valid PAT should connect"
        )
        OpenPullRequestsUI.selectFilter("all", in: app)
        assertListed(newestFirst, in: app, "전체")

        OpenPullRequestsUI.selectFilter("mine", in: app)
        assertListed([11, 14], in: app, "내 PR만")
        OpenPullRequestsUI.selectFilter("all", in: app)
        assertListed(newestFirst, in: app, "전체 after 내 PR만")

        OpenPullRequestsUI.selectFilter("reviewRequested", in: app)
        assertListed([12, 13], in: app, "리뷰 요청만")
        OpenPullRequestsUI.selectFilter("all", in: app)
        assertListed(newestFirst, in: app, "전체 after 리뷰 요청만")

        OpenPullRequestsUI.selectFilter("ciFailing", in: app)
        assertListed([12], in: app, "CI 실패만")
        OpenPullRequestsUI.selectFilter("all", in: app)
        assertListed(newestFirst, in: app, "전체 after CI 실패만")

        OpenPullRequestsUI.selectFilter("reviewRequested", in: app)
        assertListed([12, 13], in: app, "리뷰 요청만 before relaunch")
        app.terminate()
        app = OpenPullRequestsUI.launch()
        OpenPullRequestsUI.openSheet(in: app)
        XCTAssertTrue(
            OpenPullRequestsUI.element("openprs.account.login", in: app).waitForExistence(timeout: 15),
            "The connection should survive the relaunch"
        )
        let kept = app.buttons["filter.pill.reviewRequested"]
        XCTAssertTrue(kept.waitForExistence(timeout: 10), "리뷰 요청만 should be offered after relaunch")
        XCTAssertTrue(kept.isSelected, "The last filter should still be selected after relaunch")
        XCTAssertFalse(app.buttons["filter.pill.all"].isSelected, "전체 should not be selected after relaunch")
        assertListed([12, 13], in: app, "리뷰 요청만 after relaunch")

        OpenPullRequestsUI.selectFilter("all", in: app)
        app.buttons["openprs.disconnect.button"].tap()
        XCTAssertTrue(app.secureTextFields["openprs.token.field"].waitForExistence(timeout: 10), "Disconnect should return to the token form")
    }
}
