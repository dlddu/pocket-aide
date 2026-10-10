// 검증 시나리오: test-github-monitor.md#시나리오 3
import XCTest

final class PullRequestsCIStatusUITests: XCTestCase {
    private let labels = ["성공", "실패", "진행 중", "상태 없음"]

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func reveal(_ number: Int, in app: XCUIApplication) -> XCUIElement {
        let row = OpenPullRequestsUI.row(number, in: app)
        for _ in 0..<4 where !row.waitForExistence(timeout: 3) {
            let anchor = OpenPullRequestsUI.rows(in: app).firstMatch
            guard anchor.exists else { break }
            let start = anchor.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
            start.press(forDuration: 0.1, thenDragTo: start.withOffset(CGVector(dx: 0, dy: -260)))
        }
        return row
    }

    private func scrollBackToTop(in app: XCUIApplication) {
        let disconnect = app.buttons["openprs.disconnect.button"]
        for _ in 0..<4 where !(disconnect.exists && disconnect.isHittable) {
            let anchor = OpenPullRequestsUI.rows(in: app).firstMatch
            guard anchor.exists else { break }
            let start = anchor.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
            start.press(forDuration: 0.1, thenDragTo: start.withOffset(CGVector(dx: 0, dy: 260)))
        }
    }

    private func assertStatus(_ number: Int, reads expected: String, in app: XCUIApplication, _ moment: String) {
        let row = reveal(number, in: app)
        XCTAssertTrue(row.exists, "\(moment): PR #\(number) should be listed")
        XCTAssertTrue(row.label.contains(expected), "\(moment): PR #\(number) should read \(expected): \(row.label)")
        for other in labels where other != expected {
            XCTAssertFalse(row.label.contains(other), "\(moment): PR #\(number) should not read \(other): \(row.label)")
        }
    }

    func testEachPullRequestShowsTheRollupOfItsHeadCommitChecks() {
        let app = OpenPullRequestsUI.launchDisconnected()
        OpenPullRequestsUI.connect(OpenPullRequestsUI.Token.rollup, in: app)
        OpenPullRequestsUI.selectFilter("all", in: app)

        assertStatus(41, reads: "성공", in: app, "(a) every check passed")
        assertStatus(42, reads: "실패", in: app, "(b) one check failed")
        assertStatus(43, reads: "진행 중", in: app, "(c) checks running")
        assertStatus(44, reads: "상태 없음", in: app, "(d) no checks")
        assertStatus(45, reads: "진행 중", in: app, "passed and running mixed")
        assertStatus(46, reads: "실패", in: app, "running with one failure")

        scrollBackToTop(in: app)
        OpenPullRequestsUI.disconnect(in: app)
    }
}
