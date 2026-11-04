// 검증 시나리오: test-session-platform.md#시나리오 7 (4단계)
import XCTest

final class SessionsOversizePromptUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testOversizePastedPromptIsRefusedBeforeSending() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "큰 프롬프트 세션"))
        let app = TodoUI.launch()
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30))
        row.tap()

        let field = app.textFields["sessions.detail.prompt.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 15))

        UIPasteboard.general.string = String(repeating: "a", count: 1_048_576 + 1)
        field.tap()
        field.press(forDuration: 1.0)
        let paste = app.menuItems["붙여넣기"].firstMatch
        let pasteEn = app.menuItems["Paste"].firstMatch
        if paste.waitForExistence(timeout: 5) {
            paste.tap()
        } else {
            XCTAssertTrue(pasteEn.waitForExistence(timeout: 5), "The paste menu should appear on the prompt field")
            pasteEn.tap()
        }
        let allow = app.alerts.buttons["붙여넣기 허용"].firstMatch
        if allow.waitForExistence(timeout: 3) { allow.tap() }

        XCTAssertTrue(app.staticTexts["sessions.detail.prompt.toolarge"].waitForExistence(timeout: 30), "Over 1 MiB shows the notice before any send")
        XCTAssertFalse(app.buttons["sessions.detail.send.button"].isEnabled, "The send button is disabled over 1 MiB")
        XCTAssertFalse(app.descendants(matching: .any)["sessions.detail.output"].exists, "Nothing reached the platform")
    }
}
