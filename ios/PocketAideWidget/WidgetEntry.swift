import Foundation
import PocketAideAPI
import WidgetKit

enum WidgetAffirmationState: Equatable {
    case loaded(Affirmation)
    case empty
    case needsLogin
    case error
}

enum WidgetNotificationState: Equatable {
    case loaded(WidgetNotificationsSummary)
    case needsLogin
    case error
}

struct PocketAideWidgetEntry: TimelineEntry {
    let date: Date
    let state: WidgetAffirmationState
    let calendar: WidgetCalendarState
    let notifications: WidgetNotificationState
}
