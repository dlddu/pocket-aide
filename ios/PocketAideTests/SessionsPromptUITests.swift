// 검증 시나리오: test-session-platform.md#시나리오 7 (1·2단계), #시나리오 8 (3단계)
import XCTest

final class SessionsPromptUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testPromptIsSentAndOutputSurvivesLeavingTheSession() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "프롬프트 세션"), "The seeded session should get an id")
        let app = TodoUI.launch()
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30))
        row.tap()

        XCTAssertTrue(app.staticTexts["sessions.detail.output.empty"].waitForExistence(timeout: 30), "A fresh session has no output yet")
        XCTAssertTrue(app.staticTexts["sessions.detail.output.empty"].label.contains("프롬프트를 적어 에이전트에게 일을 맡겨 보세요"))

        let prompt = "이 레포 테스트 실패 원인 찾아줘"
        let field = app.textFields["sessions.detail.prompt.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 15))
        field.tap()
        field.typeText(prompt)
        app.buttons["sessions.detail.send.button"].tap()

        let output = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(output.waitForExistence(timeout: 30), "The sent prompt should reach the platform and come back through the real read path")
        XCTAssertTrue(output.label.contains(prompt), "The output should carry the prompt the platform received")
        XCTAssertEqual(field.value as? String, "에이전트에게 맡길 일", "A sent prompt clears the input (placeholder is back)")

        // 시나리오 8 3단계: 세션을 나갔다 다시 들어와도 전체 출력이 같은 내용으로, 중복 없이 보인다.
        app.navigationBars.buttons.firstMatch.tap()
        XCTAssertTrue(app.staticTexts["sessions.title"].waitForExistence(timeout: 15))
        row.tap()
        let again = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(again.waitForExistence(timeout: 30))
        XCTAssertEqual(again.label.components(separatedBy: prompt).count - 1, 1, "Re-entering must not duplicate the same line")
    }
}
