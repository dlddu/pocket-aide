// 검증 시나리오: test-routines.md#시나리오 2
import XCTest

final class RoutineStepsEditUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testStepsCanBeAddedAndRemovedAfterTheRoutineExists() {
        let token = TodoUI.uniqueToken()
        let name = "아침 루틴 \(token)"
        let steps = ["물 한 잔", "스트레칭", "일정 확인"].map { "\($0) \(token)" }
        let added = "비타민 먹기 \(token)"
        let app = TodoUI.launch()
        var screen = RoutinesScreen.open(in: app)
        screen.removeAll()
        var card = screen.create(name: name, steps: steps)
        card.assertState(summary: "매일 · 0 / 3 완료", progress: "대기", steps: steps)

        card.openStepsSheet(in: screen)
        XCTAssertFalse(screen.stepsSheetSaveButton.isEnabled, "단계 추가 must be disabled while the field is empty")
        screen.type(added, into: RoutinesScreen.stepsSheetField)
        XCTAssertTrue(screen.stepsSheetSaveButton.isEnabled, "단계 추가 should be enabled once the field is not empty")
        screen.stepsSheetSaveButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.stepsSheetTitle), "The steps sheet should close after adding a step")
        card.assertState(summary: "매일 · 0 / 4 완료", progress: "대기", steps: steps + [added])

        card.openStepsSheet(in: screen)
        let deletes = screen.stepsSheetDeleteButtons
        XCTAssertEqual(deletes.count, 4, "The steps sheet should offer one minus button per step")
        deletes[1].tap()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.stepsSheetTitle), "The steps sheet should close after removing a step")
        let remaining = [steps[0], steps[2], added]
        card.assertState(summary: "매일 · 0 / 3 완료", progress: "대기", steps: remaining)
        XCTAssertFalse(card.step(steps[1]).exists, "The removed step must leave the card")

        TodoUI.relaunch(app)
        screen = RoutinesScreen.open(in: app)
        card = screen.card(name)
        card.assertState(summary: "매일 · 0 / 3 완료", progress: "대기", steps: remaining)
        XCTAssertFalse(card.step(steps[1]).exists, "The removed step must stay removed after a relaunch")
        screen.swipeDelete(card.title)
    }
}
