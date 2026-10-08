import XCTest

struct SeededRoutine {
    let id: Int64
    let name: String
    let stepIDs: [Int64]
}

enum RoutineDays {
    static let symbols = ["일", "월", "화", "수", "목", "금", "토"]
    static let displayOrder = [1, 2, 3, 4, 5, 6, 0]

    static func today() -> Date {
        Calendar.current.startOfDay(for: Date())
    }

    static func day(_ offset: Int, from base: Date = today()) -> Date {
        Calendar.current.date(byAdding: .day, value: offset, to: base) ?? base
    }

    static func key(_ date: Date) -> String {
        let parts = Calendar.current.dateComponents([.year, .month, .day], from: date)
        return String(format: "%04d-%02d-%02d", parts.year ?? 0, parts.month ?? 0, parts.day ?? 0)
    }

    static func title(_ date: Date) -> String {
        let parts = Calendar.current.dateComponents([.month, .day], from: date)
        return "\(parts.month ?? 0)월 \(parts.day ?? 0)일"
    }

    static func weekday(_ date: Date) -> Int {
        Calendar.current.component(.weekday, from: date) - 1
    }

    static func mask(_ weekdays: [Int]) -> Int {
        weekdays.reduce(0) { $0 | (1 << $1) }
    }

    static func summary(_ weekdays: [Int]) -> String {
        displayOrder.filter { weekdays.contains($0) }.map { symbols[$0] }.joined(separator: "·")
    }

    static func waitPastMidnight() {
        let left = day(1).timeIntervalSinceNow
        if left < 600 {
            Thread.sleep(forTimeInterval: left + 5)
        }
    }
}

extension BackendAPI {
    func createRoutine(
        name: String,
        cadence: String,
        weekdays: Int = 0,
        startDay: Date,
        steps: [String],
        file: StaticString = #filePath,
        line: UInt = #line
    ) -> SeededRoutine? {
        let reply = call("POST", "/api/routines", json: [
            "name": name,
            "cadence": cadence,
            "weekdays": weekdays,
            "month_day": 0,
            "start_day": RoutineDays.key(startDay),
            "steps": steps
        ])
        guard reply.status == 201, let body = reply.json, let id = (body["id"] as? NSNumber)?.int64Value else {
            XCTFail("POST /api/routines should create '\(name)': \(reply.status) \(reply.text)", file: file, line: line)
            return nil
        }
        let stepIDs = ((body["steps"] as? [[String: Any]]) ?? [])
            .sorted { (($0["position"] as? NSNumber)?.intValue ?? 0) < (($1["position"] as? NSNumber)?.intValue ?? 0) }
            .compactMap { ($0["id"] as? NSNumber)?.int64Value }
        XCTAssertEqual(stepIDs.count, steps.count, "'\(name)' should come back with its steps", file: file, line: line)
        return SeededRoutine(id: id, name: name, stepIDs: stepIDs)
    }

    func check(_ routine: SeededRoutine, step index: Int, on day: Date, file: StaticString = #filePath, line: UInt = #line) {
        let path = "/api/routines/\(routine.id)/days/\(RoutineDays.key(day))/steps/\(routine.stepIDs[index])"
        let reply = call("PATCH", path, json: ["checked": true])
        XCTAssertEqual(reply.status, 200, "Checking step \(index) of '\(routine.name)' on \(RoutineDays.key(day)) should be stored: \(reply.text)", file: file, line: line)
    }

    func deleteRoutine(_ routine: SeededRoutine) {
        _ = call("DELETE", "/api/routines/\(routine.id)")
    }
}

struct RoutineHistoryPanel {
    let app: XCUIApplication

    var title: XCUIElement { element("routines.history.title", fallback: NSPredicate(format: "label ENDSWITH %@", " · 30일")) }
    var summary: XCUIElement { element("routines.history.summary", fallback: NSPredicate(format: "label CONTAINS %@", " · 예정 ")) }

    var closeButton: XCUIElement {
        let identified = app.buttons["routines.history.close.button"]
        return identified.exists ? identified : app.buttons.matching(NSPredicate(format: "label == %@", "닫기")).firstMatch
    }

    private func element(_ identifier: String, fallback: NSPredicate) -> XCUIElement {
        let identified = app.staticTexts[identifier]
        return identified.exists ? identified : app.staticTexts.matching(fallback).firstMatch
    }

    func row(_ day: Date) -> XCUIElement {
        let identified = app.descendants(matching: .any).matching(identifier: "routines.history.row.\(RoutineDays.key(day))").firstMatch
        if identified.exists {
            return identified
        }
        let dayTitle = RoutineDays.title(day)
        return app.descendants(matching: .any)
            .matching(NSPredicate(format: "label BEGINSWITH %@ OR label BEGINSWITH %@", dayTitle + ",", dayTitle + " "))
            .firstMatch
    }

    func status(_ day: Date) -> String? {
        let element = row(day)
        guard element.exists else { return nil }
        for candidate in ["쉬는 날", "미완료 0/2", "미완료 1/2", "미완료 2/2", "완료"] where element.label.hasSuffix(candidate) {
            return candidate
        }
        return element.label
    }

    func assertOpen(name: String, file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(title.waitForExistence(timeout: 10), "The history sheet should open", file: file, line: line)
        XCTAssertEqual(title.label, "\(name) · 30일", "The history sheet header should name the routine and its 30-day window", file: file, line: line)
    }

    func assertStatus(_ expected: String, on day: Date, file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(
            RoutinesScreen.eventually { status(day) == expected },
            "\(RoutineDays.key(day)) should read '\(expected)' but reads '\(status(day) ?? "-")'",
            file: file,
            line: line
        )
    }

    func close(file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(closeButton.waitForExistence(timeout: 5), "The history sheet should offer 닫기", file: file, line: line)
        closeButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(title), "닫기 should close the history sheet", file: file, line: line)
    }
}

extension RoutinesScreen {
    func restingRow(_ routine: SeededRoutine) -> XCUIElement {
        let identified = app.buttons["routines.resting.\(routine.id)"]
        if identified.exists {
            return identified
        }
        return app.descendants(matching: .any)
            .matching(NSPredicate(format: "label BEGINSWITH %@ AND label CONTAINS %@", routine.name, " · 단계 "))
            .firstMatch
    }

    func revealResting(_ routine: SeededRoutine, attempts: Int = 6) -> Bool {
        if Self.eventually(timeout: 10, { restingRow(routine).exists }) {
            return true
        }
        for _ in 0..<attempts where !restingRow(routine).exists {
            app.swipeUp()
        }
        return restingRow(routine).exists
    }

    func historyButton(of card: RoutineCard) -> XCUIElement? {
        let top = card.title.frame.minY - 1
        return app.buttons.matching(NSPredicate(format: "label ENDSWITH %@ OR identifier ENDSWITH %@", "이력", ".history.button"))
            .allElementsBoundByIndex
            .filter { $0.frame.minY >= top }
            .min { $0.frame.minY < $1.frame.minY }
    }
}
