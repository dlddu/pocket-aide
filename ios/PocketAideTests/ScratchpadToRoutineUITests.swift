// 검증 시나리오: test-scratchpad.md#시나리오 5
import XCTest

final class ScratchpadToRoutineUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testMemoMovedToRoutineBecomesADailyRoutineWithoutSteps() {
        let token = TodoUI.uniqueToken()
        let name = "저녁 정리 \(token)"
        let keptMemo = "나중에 볼 메모 \(token)"
        let app = TodoUI.launch()

        var scratchpad = ScratchpadScreen.open(in: app)
        scratchpad.add(keptMemo)
        scratchpad.add(name)

        scratchpad.move(name, to: "routine")
        XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.memo(name)), "The memo moved to 루틴 should leave 임시 공간")
        XCTAssertTrue(scratchpad.memo(keptMemo).exists, "The memo that was not moved should stay in 임시 공간")
        assertRoutine(name, notListing: keptMemo, in: app)

        TodoUI.relaunch(app)
        scratchpad = ScratchpadScreen.open(in: app)
        XCTAssertTrue(scratchpad.memo(keptMemo).waitForExistence(timeout: 15), "The kept memo should still be in 임시 공간")
        XCTAssertFalse(scratchpad.memo(name).exists, "The moved memo must not come back to 임시 공간")
        let routines = assertRoutine(name, notListing: keptMemo, in: app)

        routines.swipeDelete(routines.card(name).title)
        ScratchpadScreen.open(in: app).remove([keptMemo])
    }

    @discardableResult
    private func assertRoutine(_ name: String, notListing other: String, in app: XCUIApplication) -> RoutinesScreen {
        let routines = RoutinesScreen.open(in: app)
        let card = routines.card(name)
        card.assertState(summary: "매일 · 0 / 0 완료", progress: "대기", steps: [])
        XCTAssertNotNil(card.stepsButton, "The new routine should offer the 단계 button to attach steps")
        XCTAssertFalse(routines.card(other).title.exists, "The memo that was not moved must not become a routine")
        return routines
    }
}
