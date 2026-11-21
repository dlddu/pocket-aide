// 검증 시나리오: test-session-platform.md#시나리오 1
import XCTest

final class SessionsTabUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testSessionsTabSitsInBottomBar() throws {
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        XCTAssertTrue(app.tabBars.firstMatch.buttons["채팅"].exists, "채팅 should occupy the second bottom slot")
        XCTAssertTrue(screen.createButton.exists, "The 새 세션 button should be in the header")
    }

    func testEmptyPlatformShowsStartHint() throws {
        SessionPlatformStub.reset()
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        XCTAssertTrue(screen.emptyState.waitForExistence(timeout: 30), "A platform with no sessions should show the start hint")
        XCTAssertFalse(screen.listRetry.exists, "A reachable platform must not show the unreachable banner")
    }
}
