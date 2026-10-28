// 검증 시나리오: test-session-platform.md#시나리오 4
import XCTest

final class SessionsCreateUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testCreateSessionFromPlatformCatalog() throws {
        let app = TodoUI.launch()
        let screen = SessionsScreen.open(in: app)
        XCTAssertTrue(screen.emptyState.waitForExistence(timeout: 30))
        screen.createButton.tap()

        XCTAssertTrue(app.staticTexts["sessions.create.title"].waitForExistence(timeout: 15))
        XCTAssertFalse(screen.saveButton.isEnabled, "An empty name cannot be created")
        let name = app.textFields["sessions.create.name.field"]
        XCTAssertTrue(name.waitForExistence(timeout: 15))
        name.tap()
        name.typeText("테스트 실패 원인 찾기")
        XCTAssertTrue(screen.saveButton.waitForExistence(timeout: 5))
        XCTAssertTrue(screen.saveButton.isEnabled)
        screen.saveButton.tap()

        XCTAssertTrue(app.otherElements["sessions.detail.screen"].waitForExistence(timeout: 30), "Creating should land on the new session's detail")
        XCTAssertTrue(app.staticTexts["sessions.detail.state"].waitForExistence(timeout: 15))
        XCTAssertTrue(app.staticTexts["sessions.detail.state"].label.contains("실행 중"))
    }
}
