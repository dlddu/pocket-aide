// Main entry point for the tabbed UI.
// SwiftUI code is guarded so that the SPM package can compile on Linux CI
// (which lacks UIKit / SwiftUI) while still being importable on Apple platforms.

#if canImport(SwiftUI)
import SwiftUI

// MARK: - MainTabView

/// Root tab container that hosts all seven top-level sections of PocketAide.
///
/// Each tab is backed by a placeholder view while the full feature is under
/// development. The `selectedTab` binding allows programmatic navigation.
public struct MainTabView: View {

    @State private var selectedTab: AppTab = .chat

    public init() {}

    public var body: some View {
        TabView(selection: $selectedTab) {
            ForEach(AppTab.allCases, id: \.self) { tab in
                tabView(for: tab)
                    .tabItem {
                        Label(tab.title, systemImage: tab.symbolName)
                    }
                    .tag(tab)
            }
        }
    }

    @ViewBuilder
    private func tabView(for tab: AppTab) -> some View {
        switch tab {
        case .chat:
            ChatView()
        case .routine:
            RoutineView()
        case .personalTodo:
            PersonalTodoView()
        case .workTodo:
            WorkTodoView()
        case .scratchPad:
            ScratchPadView()
        case .notifications:
            NotificationsView()
        case .quotes:
            QuotesView()
        }
    }
}

// MARK: - Preview

#Preview {
    MainTabView()
}
#endif
