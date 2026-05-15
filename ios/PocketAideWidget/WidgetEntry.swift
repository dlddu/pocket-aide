import Foundation
import PocketAideAPI
import WidgetKit

enum WidgetAffirmationState: Equatable {
    case loaded(Affirmation)
    case empty
    case needsLogin
    case error
}

struct PocketAideWidgetEntry: TimelineEntry {
    let date: Date
    let state: WidgetAffirmationState
}
