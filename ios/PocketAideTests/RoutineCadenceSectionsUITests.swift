// 검증 시나리오: test-routines.md#시나리오 3
import XCTest

final class RoutineCadenceSectionsUITests: XCTestCase {
    private static let symbols = ["일", "월", "화", "수", "목", "금", "토"]
    private static let displayOrder = [1, 2, 3, 4, 5, 6, 0]

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testCadencesSplitRoutinesIntoTodayAndResting() {
        Self.waitPastMidnightIfClose()
        let now = Date()
        let weekday = Calendar.current.component(.weekday, from: now) - 1
        let monthDay = Calendar.current.component(.day, from: now)
        let restingDays = [(weekday + 2) % 7, (weekday + 4) % 7]
        let billDay = monthDay == 2 ? 3 : 2
        let token = TodoUI.uniqueToken()
        let evening = "저녁 정리 \(token)"
        let eveningStep = "책상 정리 \(token)"
        let review = "주간 회고 \(token)"
        let bills = "관리비 확인 \(token)"
        let water = "물 마시기 \(token)"

        let app = TodoUI.launch()
        let screen = RoutinesScreen.open(in: app)
        screen.removeAll()
        screen.removeResting()
        XCTAssertTrue(
            screen.text(Self.headerDate(now)).waitForExistence(timeout: 10),
            "The header should show today's date as '\(Self.headerDate(now))'"
        )

        screen.saveWeekdays(evening, step: eveningStep, days: restingDays)
        screen.saveWeekly(review, tryFirst: restingDays[0], then: weekday)
        screen.saveMonthly(bills, day: billDay)
        screen.create(name: water, steps: [])

        let weeklySummary = "매주 \(Self.symbols[weekday])요일"
        screen.card(review).assertState(summary: "\(weeklySummary) · 0 / 0 완료", progress: "대기", steps: [])
        screen.card(water).assertState(summary: "매일 · 0 / 0 완료", progress: "대기", steps: [])
        XCTAssertTrue(screen.text("오늘 · 2").waitForExistence(timeout: 10), "The 오늘 section should count the two active routines")

        screen.assertSections(today: [review, water], resting: [evening, bills])
        let restingSummary = Self.displayOrder.filter { restingDays.contains($0) }.map { Self.symbols[$0] }.joined(separator: "·")
        screen.assertResting(evening, summary: "\(restingSummary) · 단계 1개")
        screen.assertResting(bills, summary: "매월 \(billDay)일 · 단계 0개")
        XCTAssertEqual(
            screen.labels(containing: eveningStep).count,
            0,
            "A resting routine should not offer its steps to check"
        )

        screen.swipeDelete(screen.card(review).title)
        screen.swipeDelete(screen.card(water).title)
        screen.removeResting()
    }

    private static func headerDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.dateFormat = "M월 d일 · EEEE"
        return formatter.string(from: date)
    }

    private static func waitPastMidnightIfClose() {
        let now = Date()
        guard let midnight = Calendar.current.nextDate(
            after: now,
            matching: DateComponents(hour: 0, minute: 0, second: 0),
            matchingPolicy: .nextTime
        ) else { return }
        let remaining = midnight.timeIntervalSince(now)
        if remaining < 300 {
            Thread.sleep(forTimeInterval: remaining + 5)
        }
    }
}

private extension RoutinesScreen {
    func text(_ label: String) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@", label)).firstMatch
    }

    func labels(containing fragment: String) -> [String] {
        app.descendants(matching: .any)
            .matching(NSPredicate(format: "label CONTAINS %@", fragment))
            .allElementsBoundByIndex
            .map(\.label)
    }

    func chooseCadence(_ label: String) {
        let segment = app.segmentedControls.buttons[label]
        let target = segment.exists ? segment : app.buttons.matching(NSPredicate(format: "label == %@", label)).firstMatch
        XCTAssertTrue(target.waitForExistence(timeout: 5), "The 반복 주기 segment '\(label)' should exist")
        target.tap()
    }

    func tapWeekday(_ index: Int) {
        let button = app.buttons["routines.sheet.weekday.\(index)"]
        XCTAssertTrue(button.waitForExistence(timeout: 5), "Weekday button \(index) should be in the sheet")
        button.tap()
    }

    func saveWeekdays(_ name: String, step: String, days: [Int]) {
        openAddSheet()
        type(name, into: Self.nameField)
        chooseCadence("특정 요일")
        XCTAssertFalse(sheetSaveButton.isEnabled, "저장 must be disabled while 특정 요일 has no weekday")
        chooseCadence("매주")
        XCTAssertFalse(sheetSaveButton.isEnabled, "저장 must be disabled while 매주 has no weekday")
        chooseCadence("특정 요일")
        for day in days {
            tapWeekday(day)
        }
        fillSteps([step])
        saveAddSheet()
    }

    func saveWeekly(_ name: String, tryFirst first: Int, then chosen: Int) {
        openAddSheet()
        type(name, into: Self.nameField)
        chooseCadence("매주")
        tapWeekday(first)
        XCTAssertTrue(sheetSaveButton.isEnabled, "저장 should be enabled once 매주 has a weekday")
        tapWeekday(chosen)
        XCTAssertTrue(sheetSaveButton.isEnabled, "Tapping another weekday under 매주 should move the single selection")
        saveAddSheet()
    }

    func saveMonthly(_ name: String, day: Int) {
        openAddSheet()
        type(name, into: Self.nameField)
        chooseCadence("매월")
        let stepper = monthDayStepper()
        XCTAssertTrue(stepper.waitForExistence(timeout: 5), "The 매월 stepper should appear")
        for _ in 1..<day {
            incrementMonthDay(stepper)
        }
        XCTAssertTrue(
            Self.eventually(timeout: 5) { !labels(containing: "매월 \(day)일").isEmpty },
            "The stepper should read 매월 \(day)일"
        )
        saveAddSheet()
    }

    func named(_ name: String) -> XCUIElement {
        let title = text(name)
        if title.exists {
            return title
        }
        return app.descendants(matching: .any).matching(NSPredicate(format: "label BEGINSWITH %@", name)).firstMatch
    }

    func reveal(_ name: String, attempts: Int = 6) -> Bool {
        if Self.eventually(timeout: 10, { named(name).exists }) {
            return true
        }
        for _ in 0..<attempts where !named(name).exists {
            app.swipeUp()
        }
        return named(name).exists
    }

    func monthDayStepper() -> XCUIElement {
        let identified = app.descendants(matching: .any).matching(identifier: "routines.sheet.monthday.stepper").firstMatch
        return identified.waitForExistence(timeout: 5) ? identified : app.steppers.firstMatch
    }

    func incrementMonthDay(_ stepper: XCUIElement) {
        let named = stepper.buttons.matching(NSPredicate(format: "label IN %@", ["Increment", "증가", "+"])).firstMatch
        if named.exists {
            named.tap()
        } else {
            stepper.coordinate(withNormalizedOffset: CGVector(dx: 0.75, dy: 0.5)).tap()
        }
    }

    func restingSummary(_ name: String) -> String? {
        let merged = app.descendants(matching: .any)
            .matching(NSPredicate(format: "label BEGINSWITH %@ AND label CONTAINS %@", name, " · 단계 "))
            .firstMatch
        if merged.exists {
            return merged.label
        }
        let title = text(name)
        guard title.exists else { return nil }
        return app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", " · 단계 "))
            .allElementsBoundByIndex
            .filter { $0.frame.minY >= title.frame.minY - 1 }
            .min { $0.frame.minY < $1.frame.minY }?.label
    }

    func assertResting(_ name: String, summary: String, file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(reveal(name), "'\(name)' should be listed on the 루틴 screen", file: file, line: line)
        XCTAssertTrue(
            Self.eventually { restingSummary(name)?.contains(summary) == true },
            "'\(name)' should rest today with '\(summary)' but reads '\(restingSummary(name) ?? "-")'",
            file: file,
            line: line
        )
    }

    func assertSections(today: [String], resting: [String], file: StaticString = #filePath, line: UInt = #line) {
        let header = text("오늘 쉬는 루틴")
        XCTAssertTrue(header.waitForExistence(timeout: 10), "The 오늘 쉬는 루틴 section should appear", file: file, line: line)
        for name in today {
            XCTAssertTrue(named(name).exists, "'\(name)' should be on screen with the section headers", file: file, line: line)
            XCTAssertLessThan(named(name).frame.minY, header.frame.minY, "'\(name)' should sit in the 오늘 section", file: file, line: line)
        }
        for name in resting {
            XCTAssertTrue(named(name).exists, "'\(name)' should be on screen with the section headers", file: file, line: line)
            XCTAssertGreaterThan(named(name).frame.minY, header.frame.minY, "'\(name)' should sit under 오늘 쉬는 루틴", file: file, line: line)
        }
    }

    func removeResting(limit: Int = 8) {
        let rows = app.descendants(matching: .any).matching(NSPredicate(format: "label CONTAINS %@", " · 단계 "))
        for _ in 0..<limit {
            let row = rows.firstMatch
            guard row.waitForExistence(timeout: 2), swipeDelete(row) else { return }
        }
    }
}
