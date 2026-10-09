import XCTest
@testable import PocketAideAPI

final class ScratchpadTests: XCTestCase {
    private var calendar: Calendar {
        var cal = Calendar(identifier: .gregorian)
        cal.timeZone = TimeZone(identifier: "Asia/Seoul") ?? .current
        return cal
    }

    private func epoch(_ day: Int, _ hour: Int, _ minute: Int = 0) -> Int64 {
        let date = calendar.date(from: DateComponents(year: 2026, month: 9, day: day, hour: hour, minute: minute)) ?? Date()
        return Int64(date.timeIntervalSince1970)
    }

    private var now: Date { Date(timeIntervalSince1970: TimeInterval(epoch(29, 15))) }

    func testGroupsByDayNewestFirst() {
        let items = [
            ScratchpadItem(id: 1, text: "어제 밤", capturedAt: epoch(28, 22, 51)),
            ScratchpadItem(id: 2, text: "오늘 아침", source: .shortcut, capturedAt: epoch(29, 9, 13)),
            ScratchpadItem(id: 3, text: "오늘 오후", capturedAt: epoch(29, 14, 8)),
            ScratchpadItem(id: 4, text: "지난주", capturedAt: epoch(22, 10)),
        ]
        let sections = ScratchpadSections.group(items, now: now, calendar: calendar)
        XCTAssertEqual(sections.map(\.title), ["오늘", "어제", "9월 22일"])
        XCTAssertEqual(sections.map { $0.items.map(\.id) }, [[3, 2], [1], [4]])
    }

    func testEmptyListHasNoSections() {
        XCTAssertTrue(ScratchpadSections.group([], now: now, calendar: calendar).isEmpty)
    }

    func testDecodesServerItemMetadata() throws {
        let json = #"{"id":7,"text":"retention fact-check","source":"shortcut","captured_at":1700000000,"created_at":1700000001}"#
        let item = try JSONDecoder().decode(ScratchpadItem.self, from: Data(json.utf8))
        XCTAssertEqual(item.source, .shortcut)
        XCTAssertEqual(item.capturedAt, 1_700_000_000)
        XCTAssertEqual(item.source.displayName, "숏컷 · 음성")
    }

    func testDecodesMoveResultForEachTarget() throws {
        let todoJSON = #"{"target":"work","todo":{"id":3,"title":"메일","memo":"","due_date":null,"priority":null,"completed_at":null,"created_at":1,"updated_at":1}}"#
        let todo = try JSONDecoder().decode(ScratchpadMoveResult.self, from: Data(todoJSON.utf8))
        XCTAssertEqual(todo.target, .work)
        XCTAssertEqual(todo.todo?.title, "메일")
        XCTAssertNil(todo.affirmation)

        let affJSON = #"{"target":"affirmation","affirmation":{"id":9,"text":"매일 1%","priority":"normal","created_at":1,"updated_at":1}}"#
        let aff = try JSONDecoder().decode(ScratchpadMoveResult.self, from: Data(affJSON.utf8))
        XCTAssertEqual(aff.affirmation?.priority, .normal)
        XCTAssertNil(aff.todo)
    }

    func testWidgetCountTextShowsUnclassifiedCount() {
        XCTAssertEqual(WidgetScratchpad.countText(3), "미분류 3개")
        XCTAssertEqual(WidgetScratchpad.countText(1), "미분류 1개")
    }

    func testWidgetCountTextMarksEmptyScratchpad() {
        XCTAssertEqual(WidgetScratchpad.countText(0), "정리할 항목 없음")
    }
}
