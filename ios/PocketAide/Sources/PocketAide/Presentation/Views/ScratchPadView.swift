#if canImport(SwiftUI)
import SwiftUI

/// Placeholder for the Scratch Pad feature tab.
public struct ScratchPadView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            ContentUnavailableView(
                AppTab.scratchPad.title,
                systemImage: AppTab.scratchPad.symbolName,
                description: Text("Scratch pad feature coming soon.")
            )
            .navigationTitle(AppTab.scratchPad.title)
        }
    }
}
#endif
