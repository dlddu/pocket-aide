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
        XCTAssertFalse(app.tabBars.firstMatch.buttons["채팅"].exists, "채팅 should have moved under More once 세션 took the fourth slot")
        XCTAssertTrue(screen.createButton.exists, "The 새 세션 button should be in the header")
    }

    // 시나리오 1 의 2단계 — 플랫폼에 세션이 하나도 없으면 시작을 안내한다.
    func testEmptyPlatformShowsStartHint() throws {
        SessionPlatformStub.reset()
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        XCTAssertTrue(screen.emptyState.waitForExistence(timeout: 30), "A platform with no sessions should show the start hint")
        XCTAssertFalse(screen.listRetry.exists, "A reachable platform must not show the unreachable banner")
    }
}
