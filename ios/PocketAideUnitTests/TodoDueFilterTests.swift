import XCTest
@testable import PocketAideAPI

final class TodoDueFilterTests: XCTestCase {
    private var calendar: Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.firstWeekday = 2
        calendar.timeZone = .current
        return calendar
    }

    private var now: Date {
        TodoDueDate.date(from: "2026-10-14")!
    }

    private let items = [
        TodoItem(id: 1, title: "지난 마감", dueDate: "2026-10-13"),
        TodoItem(id: 2, title: "오늘 마감", dueDate: "2026-10-14"),
        TodoItem(id: 3, title: "이번 주 일요일", dueDate: "2026-10-18"),
        TodoItem(id: 4, title: "다음 주", dueDate: "2026-10-19"),
        TodoItem(id: 5, title: "날짜 없음"),
        TodoItem(id: 6, title: "완료된 오늘 마감", dueDate: "2026-10-14", completedAt: 10),
    ]

    private func ids(_ filter: TodoDueFilter) -> [Int64] {
        items.filter { filter.includes($0, now: now, calendar: calendar) }.map(\.id)
    }

    func testAllKeepsEverything() {
        XCTAssertEqual(ids(.all), [1, 2, 3, 4, 5, 6])
    }

    func testTodayKeepsItemsDueTodayOrEarlier() {
        XCTAssertEqual(ids(.today), [1, 2, 6])
    }

    func testThisWeekKeepsItemsDueByTheLastDayOfTheWeek() {
        XCTAssertEqual(ids(.thisWeek), [1, 2, 3, 6])
    }

    func testNoDateKeepsOnlyUndatedItems() {
        XCTAssertEqual(ids(.noDate), [5])
    }

    func testDeadlineCountIsOpenItemsWithDueDate() {
        XCTAssertEqual(TodoDeadline.openCount(items), 4)
        XCTAssertEqual(TodoDeadline.openCount([]), 0)
    }
}
