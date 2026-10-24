// 검증 시나리오: test-session-platform.md#시나리오 1
import XCTest

final class SessionsTabUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    // Step 2 (empty state) needs a session platform with no sessions; CI has none (BF-5).
    func testSessionsTabSitsInBottomBar() throws {
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        XCTAssertFalse(app.tabBars.firstMatch.buttons["채팅"].exists, "채팅 should have moved under More once 세션 took the fourth slot")
        XCTAssertTrue(screen.createButton.exists, "The 새 세션 button should be in the header")
    }
}
