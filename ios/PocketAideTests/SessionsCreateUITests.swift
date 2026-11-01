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
        XCTAssertTrue(screen.emptyState.waitForExistence(timeout: 30), "The platform should show the empty state before any session exists")
        screen.createButton.tap()

        XCTAssertTrue(app.staticTexts["sessions.create.title"].waitForExistence(timeout: 15), "The create sheet should open")
        XCTAssertFalse(screen.saveButton.isEnabled, "An empty name cannot be created")
        let name = app.textFields["sessions.create.name.field"]
        XCTAssertTrue(name.waitForExistence(timeout: 15), "The name field should be in the create sheet")
        name.tap()
        name.typeText("테스트 실패 원인 찾기")
        XCTAssertTrue(screen.saveButton.waitForExistence(timeout: 5), "The create button should be in the create sheet")
        XCTAssertTrue(screen.saveButton.isEnabled, "A named session with the catalog default type should be creatable")
        screen.saveButton.tap()

        XCTAssertTrue(app.buttons["sessions.detail.actions.menu"].waitForExistence(timeout: 30), "Creating should land on the new session's detail")
        XCTAssertTrue(app.staticTexts["sessions.detail.state"].waitForExistence(timeout: 15), "The detail screen should show the session state line")
        XCTAssertTrue(app.staticTexts["sessions.detail.state"].label.contains("환경이 떠 있어요"), "An active session should report an environment that is up")
    }
}
