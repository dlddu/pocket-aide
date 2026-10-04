import EventKit

enum CalendarAccess {
    static func requestIfNeeded() async {
        guard EKEventStore.authorizationStatus(for: .event) == .notDetermined else { return }
        let granted = (try? await EKEventStore().requestFullAccessToEvents()) ?? false
        if granted {
            WidgetRefresher.reloadAll()
        }
    }
}
