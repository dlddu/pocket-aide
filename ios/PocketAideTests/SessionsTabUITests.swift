// 검증 시나리오: test-session-platform.md#시나리오 1
import XCTest

final class SessionsTabUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testSessionsTabSitsInBottomBarAndStartsEmpty() throws {
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        XCTAssertFalse(app.tabBars.firstMatch.buttons["채팅"].exists, "채팅 should have moved under More once 세션 took the fourth slot")
        XCTAssertTrue(screen.emptyState.waitForExistence(timeout: 30), "With no app-created sessions the empty state should invite a new session")
        XCTAssertEqual(screen.emptyState.label, "아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요.")
        XCTAssertTrue(screen.createButton.exists, "The 새 세션 button should be in the header")
    }
}
