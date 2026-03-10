#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Chat feature tab.
public struct ChatView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.chat.title,
                systemImage: AppTab.chat.symbolName,
                description: Text("Chat feature coming soon.")
            )
            .navigationTitle(AppTab.chat.title)
        }
    }
}
#endif
