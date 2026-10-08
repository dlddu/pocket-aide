import XCTest
@testable import PocketAideAPI

final class AffirmationSortTests: XCTestCase {
    private func aff(_ id: Int64, _ p: AffirmationPriority, _ createdAt: Int64) -> Affirmation {
        Affirmation(id: id, text: "t\(id)", priority: p, createdAt: createdAt, updatedAt: createdAt)
    }

    private var pool: [Affirmation] {
        [aff(1, .low, 300), aff(2, .high, 100), aff(3, .normal, 400), aff(4, .high, 200), aff(5, .normal, 400)]
    }

    func testPriorityPutsHighFirstThenNewest() {
        XCTAssertEqual(AffirmationSortOrder.priority.sorted(pool).map(\.id), [4, 2, 5, 3, 1])
    }

    func testNewestIgnoresPriority() {
        XCTAssertEqual(AffirmationSortOrder.newest.sorted(pool).map(\.id), [5, 3, 1, 4, 2])
    }

    func testEmptyStaysEmpty() {
        XCTAssertTrue(AffirmationSortOrder.priority.sorted([]).isEmpty)
        XCTAssertTrue(AffirmationSortOrder.newest.sorted([]).isEmpty)
    }
}
