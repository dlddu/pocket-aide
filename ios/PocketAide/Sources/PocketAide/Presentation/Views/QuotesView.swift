#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Quotes feature tab.
public struct QuotesView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.quotes.title,
                systemImage: AppTab.quotes.symbolName,
                description: Text("Quotes feature coming soon.")
            )
            .navigationTitle(AppTab.quotes.title)
        }
    }
}
#endif
