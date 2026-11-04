// 검증 시나리오: test-session-platform.md#시나리오 7 (2·3단계), #시나리오 8 (1·2단계)
import XCTest

final class SessionsResponsePhaseUITests: XCTestCase {
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

    private func send(_ app: XCUIApplication, _ prompt: String) {
        let field = app.textFields["sessions.detail.prompt.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 15))
        field.tap()
        field.typeText(prompt)
        app.buttons["sessions.detail.send.button"].tap()
    }

    private func waitForPhase(_ app: XCUIApplication, containing text: String, timeout: TimeInterval) {
        let phase = app.descendants(matching: .any)["sessions.detail.phase"]
        expectation(for: NSPredicate(format: "label CONTAINS %@", text), evaluatedWith: phase)
        waitForExpectations(timeout: timeout)
    }

    func testPhaseGoesQueuedThenStreamingThenQuiet() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "진행 단계 세션"))
        SessionPlatformStub.streamReplies(["첫 조각", "둘째 조각"], delay: 6, interval: 2)
        let app = TodoUI.launch()
        openDetail(app, id: id)

        send(app, "응답이 천천히 오는 프롬프트")
        waitForPhase(app, containing: "전송됨 · 응답 대기 중", timeout: 15)
        waitForPhase(app, containing: "응답 진행 중", timeout: 30)
        waitForPhase(app, containing: "새 출력이 멈췄어요", timeout: 40)
        let output = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(output.label.contains("둘째 조각"), "The streamed reply should be on screen when the phase goes quiet")
    }

    func testOutputContinuesAcrossBackgroundWithoutDuplicates() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "백그라운드 세션"))
        SessionPlatformStub.streamReplies(["줄 하나", "줄 둘", "줄 셋"], delay: 2, interval: 6)
        let app = TodoUI.launch()
        openDetail(app, id: id)

        send(app, "긴 응답 프롬프트")
        let output = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(output.waitForExistence(timeout: 30))

        XCUIDevice.shared.press(.home)
        sleep(15)
        app.activate()

        expectation(for: NSPredicate(format: "label CONTAINS '줄 셋'"), evaluatedWith: output)
        waitForExpectations(timeout: 45)
        for line in ["줄 하나", "줄 둘", "줄 셋"] {
            XCTAssertEqual(output.label.components(separatedBy: line).count - 1, 1, "\(line) must appear exactly once after returning")
        }
    }
}
