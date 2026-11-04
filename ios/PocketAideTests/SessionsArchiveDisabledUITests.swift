// 검증 시나리오: test-session-platform.md#시나리오 9 (나)
import XCTest

final class SessionsArchiveDisabledUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        SessionPlatformStub.reset()
    }

    func testArchiveDisabledKeepsTheSessionRunningAndSaysSo() throws {
        let id = try XCTUnwrap(SessionPlatformStub.seedSession(name: "아카이브 꺼진 세션"))
        SessionPlatformStub.write(id, "남아야 하는 출력")
        SessionPlatformStub.disableArchive()
        let app = TodoUI.launch()
        _ = SessionsScreen.open(in: app)
        let row = app.descendants(matching: .any)["sessions.row.\(id)"]
        XCTAssertTrue(row.waitForExistence(timeout: 30))
        row.tap()

        let menu = app.buttons["sessions.detail.actions.menu"]
        XCTAssertTrue(menu.waitForExistence(timeout: 15))
        menu.tap()
        app.buttons["sessions.detail.archive.button"].tap()

        let notice = app.staticTexts.matching(NSPredicate(format: "label CONTAINS '이 환경에서는 아카이브가 꺼져 있어요'")).firstMatch
        XCTAssertTrue(notice.waitForExistence(timeout: 30), "A 503 from the archive request reads as 「꺼져 있음」")
        XCTAssertEqual(SessionPlatformStub.sessionState(id), "active", "The session stays running")
        let output = app.descendants(matching: .any)["sessions.detail.output"]
        XCTAssertTrue(output.waitForExistence(timeout: 15))
        XCTAssertTrue(output.label.contains("남아야 하는 출력"), "The output is untouched")
        XCTAssertFalse(app.staticTexts["아카이브된 세션이에요"].exists)
    }
}
