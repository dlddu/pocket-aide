// 검증 시나리오: test-todo.md#시나리오 10
import XCTest

final class TodoDueFilterUITests: XCTestCase {
    private struct DueTitles {
        let overdue: String
        let today: String
        let doneToday: String
        let weekEnd: String
        let nextWeek: String
        let noDate: String
        let work: String

        init(token: String) {
            overdue = "지난 마감 \(token)"
            today = "오늘 마감 \(token)"
            doneToday = "완료된 오늘 마감 \(token)"
            weekEnd = "이번 주 마감 \(token)"
            nextWeek = "다음 주 마감 \(token)"
            noDate = "날짜 없는 일 \(token)"
            work = "회사 오늘 마감 \(token)"
        }

        var personal: [String] { [overdue, today, weekEnd, nextWeek, noDate, doneToday] }
    }

    private struct DueWeek {
        let locale: String
        let yesterday: String
        let today: String
        let lastDay: String
        let nextFirstDay: String

        init(now: Date = Date()) {
            let weekday = Calendar(identifier: .gregorian).component(.weekday, from: now)
            locale = weekday == 7 ? "en_GB" : "en_US"
            var calendar = Calendar(identifier: .gregorian)
            calendar.timeZone = .current
            calendar.firstWeekday = weekday == 7 ? 2 : 1
            let week = calendar.dateInterval(of: .weekOfYear, for: now) ?? DateInterval(start: now, duration: 0)
            let formatter = DateFormatter()
            formatter.calendar = Calendar(identifier: .gregorian)
            formatter.locale = Locale(identifier: "en_US_POSIX")
            formatter.timeZone = .current
            formatter.dateFormat = "yyyy-MM-dd"
            yesterday = formatter.string(from: calendar.date(byAdding: .day, value: -1, to: now) ?? now)
            today = formatter.string(from: now)
            lastDay = formatter.string(from: calendar.date(byAdding: .day, value: -1, to: week.end) ?? now)
            nextFirstDay = formatter.string(from: week.end)
        }
    }

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testDueFilterChipsNarrowThePersonalList() throws {
        let titles = DueTitles(token: TodoUI.uniqueToken())
        let week = DueWeek()
        XCTAssertNotEqual(week.lastDay, week.today, "Today must not be the last day of the \(week.locale) week")
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        seed(titles, week: week, with: api)

        let app = XCUIApplication()
        app.launchArguments += ["-AppleLocale", week.locale]
        app.launch()
        let screen = TodoScreen.open(.personal, in: app)

        XCTAssertTrue(chip(.all, in: app).waitForExistence(timeout: 10), "The due filter chips should sit under the search field")
        for filter in ["all", "today", "thisWeek", "noDate"] {
            XCTAssertTrue(app.buttons["todos.personal.filter.\(filter)"].exists, "The '\(filter)' chip should be shown")
        }
        assertSelected(.all, in: app)
        assertListed(titles.personal, absent: [titles.work], on: screen)
        screen.list.swipeDown()

        select(.today, in: app, screen: screen)
        assertListed([titles.overdue, titles.today, titles.doneToday], absent: [titles.weekEnd, titles.nextWeek, titles.noDate, titles.work], on: screen)
        assertSections(open: [titles.overdue, titles.today], done: titles.doneToday, on: screen)

        select(.thisWeek, in: app, screen: screen)
        assertListed([titles.overdue, titles.today, titles.weekEnd, titles.doneToday], absent: [titles.nextWeek, titles.noDate, titles.work], on: screen)

        select(.noDate, in: app, screen: screen)
        assertListed([titles.noDate], absent: [titles.overdue, titles.today, titles.doneToday, titles.weekEnd, titles.nextWeek, titles.work], on: screen)

        select(.today, in: app, screen: screen)
        screen.search("완료된")
        assertListed([titles.doneToday], absent: [titles.overdue, titles.today, titles.weekEnd, titles.nextWeek, titles.noDate, titles.work], on: screen)
        assertSelected(.today, in: app)
        screen.clearSearch()
        screen.dismissKeyboard()
        assertListed([titles.overdue, titles.today, titles.doneToday], absent: [titles.weekEnd, titles.nextWeek, titles.noDate, titles.work], on: screen)
        assertUnchangedSummary(on: screen, in: app)

        select(.all, in: app, screen: screen)
        assertListed(titles.personal, absent: [titles.work], on: screen)

        let work = TodoScreen.open(.work, in: app)
        XCTAssertTrue(work.row(titles.work).waitForExistence(timeout: 10), "The work item due today should be listed in the work tab")
        let workChips = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "todos.work.filter."))
        XCTAssertEqual(workChips.count, 0, "The work tab must not show the due filter chips")
        XCTAssertFalse(app.buttons["todos.personal.filter.all"].exists, "The personal chips must not leak into the work tab")
    }

    private func seed(_ titles: DueTitles, week: DueWeek, with api: BackendAPI) {
        let listed = api.call("GET", "/api/todos/personal")
        XCTAssertEqual(listed.status, 200, "GET /api/todos/personal should list the items: \(listed.text)")
        for item in (listed.json?["items"] as? [[String: Any]]) ?? [] {
            guard let id = (item["id"] as? NSNumber)?.int64Value else { continue }
            _ = api.call("DELETE", "/api/todos/personal/\(id)")
        }
        let emptied = api.call("GET", "/api/todos/personal").json?["items"] as? [[String: Any]]
        XCTAssertEqual(emptied?.count, 0, "The personal list should be empty before seeding")

        api.addTodo(titles.overdue, in: .personal, dueDate: week.yesterday)
        api.addTodo(titles.today, in: .personal, dueDate: week.today)
        api.addTodo(titles.weekEnd, in: .personal, dueDate: week.lastDay)
        api.addTodo(titles.nextWeek, in: .personal, dueDate: week.nextFirstDay)
        api.addTodo(titles.noDate, in: .personal)
        api.addTodo(titles.work, in: .work, dueDate: week.today)
        guard let done = api.addTodo(titles.doneToday, in: .personal, dueDate: week.today) else { return }
        let completed = api.call(
            "PATCH",
            "/api/todos/personal/\(done.id)",
            json: ["title": titles.doneToday, "due_date": week.today, "done": true]
        )
        XCTAssertEqual(completed.status, 200, "PATCH should complete '\(titles.doneToday)': \(completed.text)")
    }

    private func chip(_ filter: DueChip, in app: XCUIApplication) -> XCUIElement {
        app.buttons["todos.personal.filter.\(filter.rawValue)"]
    }

    private func select(_ filter: DueChip, in app: XCUIApplication, screen: TodoScreen, file: StaticString = #filePath, line: UInt = #line) {
        let button = chip(filter, in: app)
        XCTAssertTrue(button.waitForExistence(timeout: 5), "The '\(filter.rawValue)' chip should exist", file: file, line: line)
        button.tap()
        assertSelected(filter, in: app, file: file, line: line)
        assertUnchangedSummary(on: screen, in: app, file: file, line: line)
    }

    private func assertSelected(_ filter: DueChip, in app: XCUIApplication, file: StaticString = #filePath, line: UInt = #line) {
        for other in DueChip.allCases {
            let button = chip(other, in: app)
            if other == filter {
                XCTAssertTrue(waitForSelected(button), "The '\(other.rawValue)' chip should be on", file: file, line: line)
            } else {
                XCTAssertFalse(button.isSelected, "The '\(other.rawValue)' chip should be off", file: file, line: line)
            }
        }
    }

    private func waitForSelected(_ element: XCUIElement) -> Bool {
        let expectation = XCTNSPredicateExpectation(predicate: NSPredicate(format: "selected == true"), object: element)
        return XCTWaiter.wait(for: [expectation], timeout: 5) == .completed
    }

    private func assertUnchangedSummary(on screen: TodoScreen, in app: XCUIApplication, file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(screen.waitForCounts(TodoCounts(open: 5, done: 1)), "The header summary should stay at 5 open · 1 done", file: file, line: line)
        XCTAssertEqual(chip(.all, in: app).label, "전체 · 6", "The 'all' chip should keep counting the whole area", file: file, line: line)
    }

    private func assertListed(_ present: [String], absent: [String], on screen: TodoScreen, file: StaticString = #filePath, line: UInt = #line) {
        for title in absent {
            let element = screen.app.staticTexts[title]
            XCTAssertTrue(TodoUI.waitToDisappear(element, timeout: 5), "'\(title)' must not be listed under the current chip", file: file, line: line)
        }
        for title in present {
            let element = screen.row(title)
            _ = element.waitForExistence(timeout: 5)
            XCTAssertTrue(screen.reveal(element), "'\(title)' should be listed", file: file, line: line)
        }
    }

    private func assertSections(open: [String], done: String, on screen: TodoScreen, file: StaticString = #filePath, line: UInt = #line) {
        let doneHeader = screen.sectionHeader(done: true)
        XCTAssertTrue(doneHeader.waitForExistence(timeout: 5), "The done section should be shown", file: file, line: line)
        XCTAssertGreaterThan(screen.row(done).frame.minY, doneHeader.frame.minY, "'\(done)' should sit in the done section", file: file, line: line)
        for title in open {
            XCTAssertLessThan(screen.row(title).frame.minY, doneHeader.frame.minY, "'\(title)' should sit in the open section", file: file, line: line)
        }
    }
}

private enum DueChip: String, CaseIterable {
    case all
    case today
    case thisWeek
    case noDate
}
