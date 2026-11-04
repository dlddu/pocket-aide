// 검증 시나리오: test-session-platform.md#시나리오 9 (가), #시나리오 10
import XCTest

final class SessionsArchiveRestoreUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    private func openDetail(_ app: XCUIApplication, id: String) {
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30), "A platform session should show in the list")
        row.tap()
    }

    func testArchiveNowReturnsTheEnvironmentAndShowsRestore() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "아카이브할 세션"))
        let app = TodoUI.launch()
        openDetail(app, id: id)

        let menu = app.buttons["sessions.detail.actions.menu"]
        XCTAssertTrue(menu.waitForExistence(timeout: 15))
        menu.tap()
        app.buttons["sessions.detail.archive.button"].tap()

        XCTAssertTrue(app.staticTexts["아카이브된 세션이에요"].waitForExistence(timeout: 30), "Archiving should replace the output with the archived panel")
        XCTAssertTrue(app.buttons["sessions.detail.restore.button"].exists, "An archived session offers 복원")
        XCTAssertEqual(SessionPlatformStub.sessionState(id), "snapshot", "The platform should hold the session as a snapshot")
    }

    func testArchivedSessionComesBackOnlyOnExplicitRestore() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "복원할 세션"))
        SessionPlatformStub.write(id, "이전 출력 한 줄")
        SessionPlatformStub.snapshot(id)
        let app = TodoUI.launch()
        openDetail(app, id: id)

        let restore = app.buttons["sessions.detail.restore.button"]
        XCTAssertTrue(restore.waitForExistence(timeout: 30), "An archived session shows 복원")
        XCTAssertFalse(app.textFields["sessions.detail.prompt.field"].isEnabled, "The composer is unusable while archived")
        sleep(10)
        XCTAssertEqual(SessionPlatformStub.sessionState(id), "snapshot", "Merely opening the detail must not restore the session")

        restore.tap()
        let output = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(output.waitForExistence(timeout: 60), "Restoring should bring the previous output back")
        XCTAssertTrue(output.label.contains("이전 출력 한 줄"))
        XCTAssertEqual(SessionPlatformStub.sessionState(id), "active")

        let field = app.textFields["sessions.detail.prompt.field"]
        field.tap()
        field.typeText("후속 프롬프트")
        app.buttons["sessions.detail.send.button"].tap()
        let followUp = NSPredicate(format: "label CONTAINS '이전 출력 한 줄' AND label CONTAINS '후속 프롬프트'")
        expectation(for: followUp, evaluatedWith: output)
        waitForExpectations(timeout: 30)
    }
}
