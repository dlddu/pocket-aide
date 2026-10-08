// 검증 시나리오: test-routines.md#시나리오 7
import XCTest

final class RoutineHistoryDaysUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHistorySheetListsThirtyDaysWithEachDayStatus() throws {
        RoutineDays.waitPastMidnight()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let token = TodoUI.uniqueToken()
        let full = RoutineDays.day(-5)
        let partial = RoutineDays.day(-3)
        let missed = RoutineDays.day(-1)
        let weekdays = [full, partial, missed].map(RoutineDays.weekday)
        let name = "이력 확인 \(token)"
        let routine = try XCTUnwrap(api.createRoutine(
            name: name,
            cadence: "weekdays",
            weekdays: RoutineDays.mask(weekdays),
            startDay: full,
            steps: ["스트레칭 \(token)", "물 한 잔 \(token)"]
        ))
        api.check(routine, step: 0, on: full)
        api.check(routine, step: 1, on: full)
        api.check(routine, step: 0, on: partial)

        let app = TodoUI.launch()
        let screen = RoutinesScreen.open(in: app)
        XCTAssertTrue(screen.revealResting(routine), "'\(name)' should rest today and be listed under 오늘 쉬는 루틴")
        screen.restingRow(routine).tap()

        let sheet = RoutineHistoryPanel(app: app)
        sheet.assertOpen(name: name)
        XCTAssertTrue(sheet.summary.waitForExistence(timeout: 5), "The history sheet should show its summary line")
        XCTAssertEqual(
            sheet.summary.label,
            "\(RoutineDays.summary(weekdays)) · 예정 3일 중 1일 완료",
            "The summary should count the three scheduled days since the start and the one completed"
        )

        sheet.assertStatus("완료", on: full)
        sheet.assertStatus("미완료 1/2", on: partial)
        sheet.assertStatus("미완료 0/2", on: missed)
        for offset in [0, -2, -4] {
            sheet.assertStatus("쉬는 날", on: RoutineDays.day(offset))
        }
        for offset in [-8, -10, -12, -29] {
            sheet.assertStatus("쉬는 날", on: RoutineDays.day(offset))
        }
        XCTAssertFalse(sheet.row(RoutineDays.day(-30)).exists, "The list should stop at 30 days")

        let newestFirst = [0, -1, -3, -5, -29].map { sheet.row(RoutineDays.day($0)) }
        let tops = newestFirst.map(\.frame.minY)
        XCTAssertEqual(tops, tops.sorted(), "The daily list should run from the most recent day down")
        XCTAssertEqual(Set(tops).count, tops.count, "Each day should have its own row")

        sheet.close()
        api.deleteRoutine(routine)
    }
}
