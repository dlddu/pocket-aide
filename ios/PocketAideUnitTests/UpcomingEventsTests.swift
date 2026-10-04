import XCTest
@testable import PocketAideAPI

final class UpcomingEventsTests: XCTestCase {
    private var calendar: Calendar {
        var cal = Calendar(identifier: .gregorian)
        cal.timeZone = TimeZone(identifier: "Asia/Seoul")!
        return cal
    }

    private func date(_ month: Int, _ day: Int, _ hour: Int, _ minute: Int = 0) -> Date {
        calendar.date(from: DateComponents(year: 2026, month: month, day: day, hour: hour, minute: minute))!
    }

    private func event(_ title: String, _ start: Date, hours: Double = 1, allDay: Bool = false) -> UpcomingEvent {
        UpcomingEvent(title: title, start: start, end: start.addingTimeInterval(hours * 3600), isAllDay: allDay)
    }

    func testShowsNearestThreeInStartOrder() {
        let now = date(10, 4, 9)
        let events = [
            event("d", date(10, 4, 18)),
            event("a", date(10, 4, 10)),
            event("c", date(10, 4, 15)),
            event("b", date(10, 4, 12)),
        ]
        let summary = UpcomingEvents.summarize(events, now: now)
        XCTAssertEqual(summary.events.map(\.title), ["a", "b", "c"])
        XCTAssertEqual(summary.moreCount, 1)
    }

    func testDeletedNearestLetsNextMoveUp() {
        let now = date(10, 4, 9)
        let events = [
            event("b", date(10, 4, 12)),
            event("c", date(10, 4, 15)),
            event("d", date(10, 4, 18)),
        ]
        let summary = UpcomingEvents.summarize(events, now: now)
        XCTAssertEqual(summary.events.map(\.title), ["b", "c", "d"])
        XCTAssertEqual(summary.moreCount, 0)
    }

    func testEndedEventsAreDroppedAndOngoingKept() {
        let now = date(10, 4, 13)
        let events = [
            event("ended", date(10, 4, 10)),
            event("ongoing", date(10, 4, 12, 30)),
            event("later", date(10, 4, 16)),
        ]
        let summary = UpcomingEvents.summarize(events, now: now)
        XCTAssertEqual(summary.events.map(\.title), ["ongoing", "later"])
    }

    func testEventsBeyondLookaheadAreExcluded() {
        let now = date(10, 4, 9)
        let events = [event("far", date(10, 12, 9))]
        XCTAssertEqual(UpcomingEvents.summarize(events, now: now), UpcomingEventsSummary(events: [], moreCount: 0))
    }

    func testLabels() {
        let now = date(10, 4, 9)
        let today = event("today", date(10, 4, 14), hours: 1.5)
        let tomorrow = event("tomorrow", date(10, 5, 8, 5))
        let later = event("later", date(10, 7, 19))
        let allDay = event("allDay", date(10, 4, 0), hours: 24, allDay: true)
        XCTAssertEqual(UpcomingEvents.rangeLabel(today, now: now, calendar: calendar), "14:00 — 15:30")
        XCTAssertEqual(UpcomingEvents.startLabel(tomorrow, now: now, calendar: calendar), "내일 08:05")
        XCTAssertEqual(UpcomingEvents.startLabel(later, now: now, calendar: calendar), "10/7 19:00")
        XCTAssertEqual(UpcomingEvents.rangeLabel(allDay, now: now, calendar: calendar), "종일")
    }
}
