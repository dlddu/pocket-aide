import WidgetKit

enum WidgetRefresher {
    static func reloadAll() {
        WidgetCenter.shared.reloadAllTimelines()
    }
}
