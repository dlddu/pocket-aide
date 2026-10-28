// 검증 시나리오: test-session-platform.md#시나리오 12
import XCTest

final class SessionsPlatformUnreachableUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testCreateSheetExplainsUnreachablePlatformWithoutLeakingDetails() throws {
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        let banner = app.staticTexts["세션 플랫폼에 연결할 수 없어요. 네트워크를 확인하고 다시 시도해 주세요."]
        XCTAssertTrue(banner.waitForExistence(timeout: 45), "Without a reachable session platform the list should say it cannot connect")
        XCTAssertTrue(screen.listRetry.exists, "The list failure should offer 다시 시도")
        XCTAssertFalse(screen.emptyState.exists, "An unreachable platform must not look like an empty one")
        screen.createButton.tap()

        let notice = app.staticTexts["세션 플랫폼에 연결하지 못해 타입 목록을 불러오지 못했어요."]
        XCTAssertTrue(notice.waitForExistence(timeout: 45), "Without a reachable session platform the type list should explain it cannot load")
        XCTAssertTrue(screen.catalogRetry.exists, "The type list failure should offer 다시 시도")
        XCTAssertFalse(screen.saveButton.isEnabled, "Nothing can be created without a type catalog")
        let leaked = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] %@ OR label CONTAINS[c] %@", "svc.cluster.local", "unreachable"))
        XCTAssertEqual(leaked.count, 0, "Internal addresses and raw upstream errors must not reach the screen")
    }
}
