// 검증 시나리오: test-routines.md#시나리오 6
import XCTest

final class RoutineCheckDayResetUITests: XCTestCase {
    private struct Seeded {
        let carried: SeededRoutine
        let bare: SeededRoutine
        let resting: SeededRoutine
    }

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testChecksHoldOnlyForTheirDayAndStepLessRoutinesNeverComplete() throws {
        RoutineDays.waitPastMidnight()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let token = TodoUI.uniqueToken()
        let today = RoutineDays.today()
        let yesterday = RoutineDays.day(-1)
        let steps = ["양치 \(token)", "세수 \(token)"]
        let restingStep = "걷기 \(token)"
        let seeded = try seed(api, token: token, steps: steps, restingStep: restingStep)
        let carried = seeded.carried
        let bare = seeded.bare
        let resting = seeded.resting

        let app = TodoUI.launch()
        var screen = RoutinesScreen.open(in: app)
        screen.card(bare.name).assertState(summary: "매일 · 0 / 0 완료", progress: "대기", steps: [])

        XCTAssertTrue(screen.revealResting(resting), "'\(resting.name)' should be listed under 오늘 쉬는 루틴")
        screen.restingRow(resting).tap()
        let sheet = RoutineHistoryPanel(app: app)
        sheet.assertOpen(name: resting.name)
        XCTAssertFalse(screen.stepsSheetTitle.exists, "Tapping a resting routine should not open the steps sheet")
        XCTAssertEqual(
            app.buttons.matching(NSPredicate(format: "label ENDSWITH %@", restingStep)).count,
            0,
            "A resting day should offer no step to check"
        )
        sheet.assertStatus("쉬는 날", on: today)
        sheet.close()

        screen = screen.reenter()
        let card = screen.card(carried.name)
        card.assertState(summary: "매일 · 0 / 2 완료", progress: "대기", steps: steps)
        for step in steps {
            XCTAssertFalse(card.isChecked(step), "'\(step)' was checked yesterday and should start unchecked today")
        }
        let history = try XCTUnwrap(screen.historyButton(of: card), "'\(carried.name)' should offer 이력")
        history.tap()
        sheet.assertOpen(name: carried.name)
        sheet.assertStatus("완료", on: yesterday)
        sheet.assertStatus("미완료 0/2", on: today)
        sheet.close()

        for step in steps {
            card.toggle(step)
        }
        card.assertState(summary: "매일 · 2 / 2 완료", progress: "완료", steps: steps)
        screen = screen.reenter()
        screen.card(bare.name).assertState(summary: "매일 · 0 / 0 완료", progress: "대기", steps: [])

        for routine in [carried, bare, resting] {
            api.deleteRoutine(routine)
        }
    }

    private func seed(
        _ api: BackendAPI,
        token: String,
        steps: [String],
        restingStep: String
    ) throws -> Seeded {
        let today = RoutineDays.today()
        let yesterday = RoutineDays.day(-1)
        let carried = try XCTUnwrap(api.createRoutine(
            name: "아침 준비 \(token)", cadence: "daily", startDay: yesterday, steps: steps
        ))
        api.check(carried, step: 0, on: yesterday)
        api.check(carried, step: 1, on: yesterday)
        let bare = try XCTUnwrap(api.createRoutine(
            name: "단계 없는 루틴 \(token)", cadence: "daily", startDay: today, steps: []
        ))
        let resting = try XCTUnwrap(api.createRoutine(
            name: "쉬는 루틴 \(token)",
            cadence: "weekly",
            weekdays: RoutineDays.mask([(RoutineDays.weekday(today) + 3) % 7]),
            startDay: today,
            steps: [restingStep]
        ))
        return Seeded(carried: carried, bare: bare, resting: resting)
    }
}
