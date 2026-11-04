// 검증 시나리오: test-session-platform.md#시나리오 13
import XCTest

final class SessionsTerminalUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testTerminalSessionTakesCommandsAndShowsOutput() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "셸 세션", workloadType: "shell"))
        let app = TodoUI.launch()
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30))
        row.tap()

        let field = app.textFields["sessions.detail.prompt.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 30))
        XCTAssertEqual(field.placeholderValue, "명령 입력", "A terminal session asks for a command, not a prompt")
        XCTAssertFalse(app.staticTexts["sessions.detail.phase"].exists, "A terminal session has no response phase")

        field.tap()
        field.typeText("ls")
        app.buttons["sessions.detail.send.button"].tap()
        let output = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(output.waitForExistence(timeout: 30), "The command should reach the platform and its output should show")
        XCTAssertTrue(output.label.contains("ls"))
        XCTAssertFalse(output.label.contains("^[["), "Control sequences must not show")
    }
}
