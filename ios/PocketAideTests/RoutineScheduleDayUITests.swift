// 검증 시나리오: test-routines.md#시나리오 4
import XCTest

final class RoutineScheduleDayUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testRoutinesComeUpOnTheDaysTheirCadenceReturns() throws {
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        api.removeAllRoutines()
        let token = TodoUI.uniqueToken()
        let start = try XCTUnwrap(Self.date("2026-01-01"))
        let evening = try XCTUnwrap(api.createRoutine(
            name: "저녁 정리 \(token)",
            cadence: "weekdays",
            weekdays: RoutineDays.mask([1, 4]),
            startDay: start,
            steps: ["책상 정리 \(token)"]
        ))
        let review = try XCTUnwrap(api.createRoutine(
            name: "주간 회고 \(token)",
            cadence: "weekly",
            weekdays: RoutineDays.mask([0]),
            startDay: start,
            steps: ["한 주 돌아보기 \(token)"]
        ))
        let bills = try XCTUnwrap(api.createRoutine(
            name: "관리비 확인 \(token)",
            cadence: "monthly",
            monthDay: 31,
            startDay: start,
            steps: ["고지서 확인 \(token)"]
        ))
        let water = try XCTUnwrap(api.createRoutine(
            name: "물 마시기 \(token)",
            cadence: "daily",
            startDay: start,
            steps: ["물 한 잔 \(token)"]
        ))

        let app = XCUIApplication()
        assertDay("2026-09-28", header: "9월 28일 · 월요일", in: app, today: [evening, water], resting: [review, bills])
        assertDay("2026-09-30", header: "9월 30일 · 수요일", in: app, today: [bills, water], resting: [evening, review])
        assertDay("2027-02-28", header: "2월 28일 · 일요일", in: app, today: [review, bills, water], resting: [evening])

        app.terminate()
        for routine in [evening, review, bills, water] {
            api.deleteRoutine(routine)
        }
    }

    private func assertDay(
        _ day: String,
        header: String,
        in app: XCUIApplication,
        today: [SeededRoutine],
        resting: [SeededRoutine],
        file: StaticString = #filePath,
        line: UInt = #line
    ) {
        app.terminate()
        // mock-exception: DET — 「오늘」은 기기 달력이 정하고 XCUITest 는 시뮬레이터 시계를 바꿀 수 없다; 앱이 「오늘」로 쓸 날짜 값만 주입한다 (docs/e2e-mocking-policy.md)
        app.launchEnvironment["ROUTINES_TODAY"] = day
        app.launch()
        let screen = RoutinesScreen.open(in: app)

        let date = app.staticTexts["routines.today"]
        XCTAssertTrue(date.waitForExistence(timeout: 10), "The 루틴 header should show the day it treats as today", file: file, line: line)
        XCTAssertEqual(date.label, header, "The 루틴 header should read the launched day \(day)", file: file, line: line)
        XCTAssertTrue(
            app.staticTexts.matching(NSPredicate(format: "label == %@", "오늘 · \(today.count)")).firstMatch.waitForExistence(timeout: 10),
            "On \(day) the 오늘 section should count \(today.count) routines",
            file: file,
            line: line
        )

        let restingHeader = app.staticTexts.matching(NSPredicate(format: "label == %@", "오늘 쉬는 루틴")).firstMatch
        XCTAssertTrue(restingHeader.waitForExistence(timeout: 10), "On \(day) the 오늘 쉬는 루틴 section should appear", file: file, line: line)
        for routine in today {
            let card = screen.card(routine.name)
            XCTAssertTrue(card.reveal(), "On \(day) '\(routine.name)' should be listed", file: file, line: line)
            XCTAssertFalse(screen.restingRow(routine).exists, "On \(day) '\(routine.name)' should not rest", file: file, line: line)
            XCTAssertLessThan(card.title.frame.minY, restingHeader.frame.minY, "On \(day) '\(routine.name)' should sit in the 오늘 section", file: file, line: line)
        }
        for routine in resting {
            XCTAssertTrue(screen.revealResting(routine), "On \(day) '\(routine.name)' should rest under 오늘 쉬는 루틴", file: file, line: line)
            XCTAssertGreaterThan(screen.restingRow(routine).frame.minY, restingHeader.frame.minY, "On \(day) '\(routine.name)' should sit under 오늘 쉬는 루틴", file: file, line: line)
        }
    }

    private static func date(_ key: String) -> Date? {
        let parts = key.split(separator: "-").compactMap { Int($0) }
        guard parts.count == 3 else { return nil }
        return Calendar.current.date(from: DateComponents(year: parts[0], month: parts[1], day: parts[2], hour: 12))
    }
}
