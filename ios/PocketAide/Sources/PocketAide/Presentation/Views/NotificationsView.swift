#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Notifications feature tab.
public struct NotificationsView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.notifications.title,
                systemImage: AppTab.notifications.symbolName,
                description: Text("Notifications feature coming soon.")
            )
            .navigationTitle(AppTab.notifications.title)
        }
    }
}
#endif
