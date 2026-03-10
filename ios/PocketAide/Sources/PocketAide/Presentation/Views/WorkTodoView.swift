#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Work Todo feature tab.
public struct WorkTodoView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.workTodo.title,
                systemImage: AppTab.workTodo.symbolName,
                description: Text("Work todo feature coming soon.")
            )
            .navigationTitle(AppTab.workTodo.title)
        }
    }
}
#endif
