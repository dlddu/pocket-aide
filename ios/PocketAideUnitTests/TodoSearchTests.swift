import XCTest
@testable import PocketAideAPI

final class TodoSearchTests: XCTestCase {
    private let items = [
        TodoItem(id: 1, title: "치과 예약 잡기"),
        TodoItem(id: 2, title: "Retention fact-check", memo: "임시공간에서 이동됨"),
        TodoItem(id: 3, title: "책장 정리", completedAt: 10),
    ]

    func testEmptyQueryReturnsEverything() {
        XCTAssertEqual(TodoSearch.filter(items, query: "  ").map(\.id), [1, 2, 3])
    }

    func testMatchesTitleCaseInsensitively() {
        XCTAssertEqual(TodoSearch.filter(items, query: "retention").map(\.id), [2])
    }

    func testMatchesMemo() {
        XCTAssertEqual(TodoSearch.filter(items, query: "임시공간").map(\.id), [2])
    }

    func testNoMatchReturnsEmpty() {
        XCTAssertTrue(TodoSearch.filter(items, query: "분기 리뷰").isEmpty)
    }

    func testDueDateRoundTrip() {
        let date = TodoDueDate.date(from: "2026-10-02")
        XCTAssertNotNil(date)
        XCTAssertEqual(date.map(TodoDueDate.string(from:)), "2026-10-02")
        XCTAssertNil(TodoDueDate.date(from: "10/02"))
    }

    func testDraftEncodingOmitsUnsetOptionals() throws {
        let data = try JSONEncoder().encode(TodoDraft(title: "x"))
        let object = try JSONSerialization.jsonObject(with: data) as? [String: Any]
        XCTAssertEqual(object?["title"] as? String, "x")
        XCTAssertEqual(object?["done"] as? Bool, false)
        XCTAssertNil(object?["due_date"])
        XCTAssertNil(object?["priority"])
    }
}
