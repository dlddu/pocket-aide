// 검증 시나리오: test-session-platform.md#시나리오 14
import XCTest

final class SessionsTurnMarkersUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testEachWriteFoldsIntoATurnThatExpandsAndCollapses() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "turn 세션"))
        SessionPlatformStub.streamReplies(["응답 첫 줄", "응답 마지막 줄"], delay: 1, interval: 1, markers: true)
        let app = TodoUI.launch()
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30))
        row.tap()

        let first = "이 레포 테스트 실패 원인 찾아줘"
        let field = app.textFields["sessions.detail.prompt.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 15))
        field.tap()
        field.typeText(first)
        app.buttons["sessions.detail.send.button"].tap()

        let phase = app.descendants(matching: .any)["sessions.detail.phase"]
        expectation(for: NSPredicate(format: "label CONTAINS '응답 완료'"), evaluatedWith: phase)
        waitForExpectations(timeout: 45)
        XCTAssertTrue(app.staticTexts["#1"].waitForExistence(timeout: 15), "The first write becomes turn #1")
        XCTAssertFalse(app.staticTexts.matching(NSPredicate(format: "label CONTAINS 'session-platform input'")).firstMatch.exists, "Marker text must never show")

        field.tap()
        field.typeText("lastAccess 로 바꾸고 테스트도 추가해줘")
        app.buttons["sessions.detail.send.button"].tap()
        XCTAssertTrue(app.staticTexts["#2"].waitForExistence(timeout: 30), "The second write becomes turn #2 below #1")

        let input = app.buttons["sessions.detail.turn.1.input.toggle"]
        let output = app.buttons["sessions.detail.turn.1.output.toggle"]
        XCTAssertTrue(input.waitForExistence(timeout: 15))
        XCTAssertFalse(app.staticTexts["sessions.detail.turn.1.input.full"].exists, "Turns start folded")
        input.tap()
        XCTAssertTrue(app.staticTexts["sessions.detail.turn.1.input.full"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["sessions.detail.turn.1.input.full"].label.contains(first))
        output.tap()
        XCTAssertTrue(app.staticTexts["sessions.detail.turn.1.output.full"].waitForExistence(timeout: 10))
        input.tap()
        XCTAssertFalse(app.staticTexts["sessions.detail.turn.1.input.full"].exists, "Tapping the header again folds it")

        app.navigationBars.buttons.firstMatch.tap()
        XCTAssertTrue(app.staticTexts["sessions.title"].waitForExistence(timeout: 15))
        row.tap()
        XCTAssertTrue(app.staticTexts["#1"].waitForExistence(timeout: 30), "The turn list comes back from the start")
        XCTAssertFalse(app.staticTexts["sessions.detail.turn.1.output.full"].exists, "Re-entering folds every turn")
    }
}
