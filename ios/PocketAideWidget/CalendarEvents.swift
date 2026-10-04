import EventKit
import Foundation
import PocketAideAPI

enum WidgetCalendarState: Equatable {
    case loaded(UpcomingEventsSummary)
    case needsPermission
}

struct CalendarSnapshot {
    let authorized: Bool
    let events: [UpcomingEvent]

    static func load(from start: Date, through end: Date) -> CalendarSnapshot {
        guard EKEventStore.authorizationStatus(for: .event) == .fullAccess else {
            return CalendarSnapshot(authorized: false, events: [])
        }
        let store = EKEventStore()
        let predicate = store.predicateForEvents(
            withStart: start,
            end: end.addingTimeInterval(UpcomingEvents.lookahead),
            calendars: nil
        )
        let events = store.events(matching: predicate).map { item in
            UpcomingEvent(
                title: item.title ?? "",
                start: item.startDate,
                end: item.endDate,
                isAllDay: item.isAllDay
            )
        }
        return CalendarSnapshot(authorized: true, events: events)
    }

    func state(at date: Date) -> WidgetCalendarState {
        guard authorized else { return .needsPermission }
        return .loaded(UpcomingEvents.summarize(events, now: date))
    }
}
