import Foundation

public struct UpcomingEvent: Equatable, Sendable {
    public let title: String
    public let start: Date
    public let end: Date
    public let isAllDay: Bool

    public init(title: String, start: Date, end: Date, isAllDay: Bool) {
        self.title = title
        self.start = start
        self.end = end
        self.isAllDay = isAllDay
    }
}

public struct UpcomingEventsSummary: Equatable, Sendable {
    public let events: [UpcomingEvent]
    public let moreCount: Int

    public init(events: [UpcomingEvent], moreCount: Int) {
        self.events = events
        self.moreCount = moreCount
    }
}

public enum UpcomingEvents {
    public static let limit = 3
    public static let lookahead: TimeInterval = 7 * 24 * 60 * 60

    public static func summarize(_ events: [UpcomingEvent], now: Date) -> UpcomingEventsSummary {
        let upcoming = events
            .filter { $0.end > now && $0.start < now.addingTimeInterval(lookahead) }
            .sorted { lhs, rhs in
                if lhs.start != rhs.start { return lhs.start < rhs.start }
                return lhs.title < rhs.title
            }
        return UpcomingEventsSummary(
            events: Array(upcoming.prefix(limit)),
            moreCount: max(0, upcoming.count - limit)
        )
    }

    public static func startLabel(_ event: UpcomingEvent, now: Date, calendar: Calendar) -> String {
        let time = event.isAllDay ? "종일" : clock(event.start, calendar: calendar)
        if event.start <= now || calendar.isDate(event.start, inSameDayAs: now) {
            return time
        }
        if let tomorrow = calendar.date(byAdding: .day, value: 1, to: now),
           calendar.isDate(event.start, inSameDayAs: tomorrow) {
            return "내일 \(time)"
        }
        let parts = calendar.dateComponents([.month, .day], from: event.start)
        return "\(parts.month ?? 0)/\(parts.day ?? 0) \(time)"
    }

    public static func rangeLabel(_ event: UpcomingEvent, now: Date, calendar: Calendar) -> String {
        let start = startLabel(event, now: now, calendar: calendar)
        guard !event.isAllDay else { return start }
        return "\(start) — \(clock(event.end, calendar: calendar))"
    }

    private static func clock(_ date: Date, calendar: Calendar) -> String {
        let parts = calendar.dateComponents([.hour, .minute], from: date)
        return String(format: "%02d:%02d", parts.hour ?? 0, parts.minute ?? 0)
    }
}
