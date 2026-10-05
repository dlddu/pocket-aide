// 검증 시나리오: test-routines.md#시나리오 1
import XCTest

final class RoutineCreateUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testRoutineKeepsItsNameCadenceAndOrderedSteps() {
        let token = TodoUI.uniqueToken()
        let name = "아침 루틴 \(token)"
        let steps = ["물 한 잔", "스트레칭", "일정 확인"].map { "\($0) \(token)" }
        let app = TodoUI.launch()
        var screen = RoutinesScreen.open(in: app)
        screen.removeAll()

        screen.openAddSheet()
        XCTAssertFalse(screen.sheetSaveButton.isEnabled, "저장 must be disabled while the name is empty")
        screen.type(" ", into: RoutinesScreen.nameField)
        XCTAssertFalse(screen.sheetSaveButton.isEnabled, "저장 must stay disabled while the name is only whitespace")
        screen.type(name, into: RoutinesScreen.nameField)
        screen.fillSteps(steps, trailingBlank: true)
        screen.saveAddSheet()

        let summary = "매일 · 0 / 3 완료"
        screen.card(name).assertState(summary: summary, progress: "대기", steps: steps)

        screen = screen.reenter()
        screen.card(name).assertState(summary: summary, progress: "대기", steps: steps)

        TodoUI.relaunch(app)
        screen = RoutinesScreen.open(in: app)
        let card = screen.card(name)
        card.assertState(summary: summary, progress: "대기", steps: steps)
        screen.swipeDelete(card.title)
    }
}
