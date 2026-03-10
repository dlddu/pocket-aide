#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Personal Todo feature tab.
public struct PersonalTodoView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.personalTodo.title,
                systemImage: AppTab.personalTodo.symbolName,
                description: Text("Personal todo feature coming soon.")
            )
            .navigationTitle(AppTab.personalTodo.title)
        }
    }
}
#endif
