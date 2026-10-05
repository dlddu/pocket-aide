// 검증 시나리오: test-routines.md#시나리오 5
import XCTest

final class RoutineStepCheckUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testCheckingEveryStepCompletesTheRoutineForTheDay() {
        let token = TodoUI.uniqueToken()
        let name = "저녁 루틴 \(token)"
        let steps = ["설거지", "내일 가방 싸기", "불 끄기"].map { "\($0) \(token)" }
        let app = TodoUI.launch()
        var screen = RoutinesScreen.open(in: app)
        screen.removeAll()
        var card = screen.create(name: name, steps: steps)
        card.assertState(summary: "매일 · 0 / 3 완료", progress: "대기", steps: steps)

        card.toggle(steps[0])
        card.assertState(summary: "매일 · 1 / 3 완료", progress: "33%", steps: steps)
        XCTAssertTrue(card.isChecked(steps[0]), "The tapped step should be checked")
        XCTAssertFalse(card.isChecked(steps[1]), "A step that was not tapped must stay unchecked")

        card.toggle(steps[1])
        card.assertState(summary: "매일 · 2 / 3 완료", progress: "67%", steps: steps)
        card.toggle(steps[2])
        card.assertState(summary: "매일 · 3 / 3 완료", progress: "완료", steps: steps)

        TodoUI.relaunch(app)
        screen = RoutinesScreen.open(in: app)
        card = screen.card(name)
        card.assertState(summary: "매일 · 3 / 3 완료", progress: "완료", steps: steps)
        for step in steps {
            XCTAssertTrue(card.isChecked(step), "'\(step)' should still be checked after a relaunch")
        }

        card.toggle(steps[2])
        card.assertState(summary: "매일 · 2 / 3 완료", progress: "67%", steps: steps)
        XCTAssertFalse(card.isChecked(steps[2]), "Tapping a checked step again should uncheck it")
        screen.swipeDelete(card.title)
    }
}
