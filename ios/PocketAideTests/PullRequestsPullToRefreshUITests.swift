// 검증 시나리오: test-github-monitor.md#시나리오 5
import XCTest

final class PullRequestsPullToRefreshUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func waitUntilTheMinuteChanges() {
        let second = Calendar.current.component(.second, from: Date())
        Thread.sleep(forTimeInterval: TimeInterval(62 - second))
    }

    private func pullDown(_ row: XCUIElement) {
        let start = row.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
        start.press(forDuration: 0.1, thenDragTo: start.withOffset(CGVector(dx: 0, dy: 420)))
    }

    private func wait(for row: XCUIElement, toRead text: String, timeout: TimeInterval) -> Bool {
        let reads = XCTNSPredicateExpectation(predicate: NSPredicate(format: "label CONTAINS %@", text), object: row)
        return XCTWaiter().wait(for: [reads], timeout: timeout) == .completed
    }

    func testPullToRefreshShowsTheChangedCIStatusAndANewRefreshTime() {
        let app = OpenPullRequestsUI.launchDisconnected()
        OpenPullRequestsUI.connect(OpenPullRequestsUI.Token.rerun(), in: app)
        OpenPullRequestsUI.selectFilter("all", in: app)
        let row = OpenPullRequestsUI.row(31, in: app)
        XCTAssertTrue(row.waitForExistence(timeout: 15), "PR #31 should be listed")
        XCTAssertTrue(row.label.contains("진행 중"), "PR #31 should start with checks running: \(row.label)")
        let stamp = OpenPullRequestsUI.element("openprs.lastRefreshed", in: app)
        XCTAssertTrue(stamp.waitForExistence(timeout: 5), "The list should show when it was last refreshed")
        let before = stamp.label
        XCTAssertTrue(before.hasPrefix("마지막 갱신"), "The refresh time should be labelled: \(before)")

        waitUntilTheMinuteChanges()
        var refreshed = false
        for _ in 0..<3 where !refreshed {
            pullDown(row)
            refreshed = wait(for: row, toRead: "성공", timeout: 10)
        }
        XCTAssertTrue(app.navigationBars["열린 PR"].exists, "Pulling the list should not dismiss the sheet")
        XCTAssertTrue(refreshed, "Pull to refresh should show the finished checks: \(row.label)")
        XCTAssertFalse(row.label.contains("진행 중"), "The old CI status should be replaced: \(row.label)")
        XCTAssertTrue(stamp.label.hasPrefix("마지막 갱신"), "The refresh time should stay labelled: \(stamp.label)")
        XCTAssertNotEqual(stamp.label, before, "The refresh time should move to the pull-to-refresh time")

        OpenPullRequestsUI.disconnect(in: app)
    }
}
