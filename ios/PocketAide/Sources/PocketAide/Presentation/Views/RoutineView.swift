#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Routine feature tab.
public struct RoutineView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.routine.title,
                systemImage: AppTab.routine.symbolName,
                description: Text("Routine feature coming soon.")
            )
            .navigationTitle(AppTab.routine.title)
        }
    }
}
#endif
