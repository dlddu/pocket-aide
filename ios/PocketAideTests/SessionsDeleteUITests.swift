// 검증 시나리오: test-session-platform.md#시나리오 11
import XCTest

final class SessionsDeleteUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testDeleteAsksForConfirmationAndRemovesSession() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "지울 세션"), "The seeded session should get an id")
        let app = TodoUI.launch()
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30), "A platform session should show in the list")
        row.tap()

        let menu = app.buttons["sessions.detail.actions.menu"]
        XCTAssertTrue(menu.waitForExistence(timeout: 15))
        menu.tap()
        app.buttons["sessions.detail.delete.button"].tap()
        let alert = app.alerts["세션을 삭제할까요?"]
        XCTAssertTrue(alert.waitForExistence(timeout: 10))
        alert.buttons["취소"].tap()
        XCTAssertTrue(SessionPlatformStub.sessionExists(id), "Cancelling must delete nothing")

        menu.tap()
        app.buttons["sessions.detail.delete.button"].tap()
        XCTAssertTrue(alert.waitForExistence(timeout: 10))
        alert.buttons["삭제"].tap()
        XCTAssertTrue(app.staticTexts["sessions.title"].waitForExistence(timeout: 30), "Deleting should return to the list")
        XCTAssertFalse(row.waitForExistence(timeout: 5), "The deleted session should be gone from the list")
        XCTAssertFalse(SessionPlatformStub.sessionExists(id), "The platform should no longer have the session")
    }
}
