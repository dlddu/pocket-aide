import DesignSystem
import SwiftUI

struct RootView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @State private var selectedTab: Tab = .affirmations

    enum Tab: Hashable {
        case chat, scratchpad, personal, work, routines, affirmations
    }

    /// LoginUITests still target HelloWorldView's identifiers
    /// (SignedInLabel, SignOutButton). Tests opt in by setting
    /// `UI_TESTS_USE_LEGACY_HOME=1`; production never sees this branch.
    private var useLegacyHome: Bool {
        ProcessInfo.processInfo.environment["UI_TESTS_USE_LEGACY_HOME"] == "1"
    }

    var body: some View {
        Group {
            if auth.signedIn {
                if useLegacyHome {
                    HelloWorldView()
                } else {
                    signedInTabs
                }
            } else {
                LoginView()
            }
        }
    }

    private var signedInTabs: some View {
        // NOTE: `.accessibilityIdentifier` is intentionally NOT applied to tab
        // children. On iOS 26 SwiftUI cascades that identifier to every
        // descendant accessibility element, overriding the more specific
        // identifiers we set inside (e.g. `screen.header.title`,
        // `affirmations.add.button`). UI tests query inner identifiers
        // directly, so the cascade did harm without helping the tab bar
        // (which uses the system Tab type and ignores the modifier anyway).
        TabView(selection: $selectedTab) {
            ChatTab()
                .tabItem { Label("채팅", systemImage: "bubble.left.and.bubble.right") }
                .tag(Tab.chat)

            ScratchpadTab()
                .tabItem { Label("임시공간", systemImage: "doc.text") }
                .tag(Tab.scratchpad)

            PersonalTab()
                .tabItem { Label("개인", systemImage: "person") }
                .tag(Tab.personal)

            WorkTab()
                .tabItem { Label("회사", systemImage: "briefcase") }
                .tag(Tab.work)

            RoutinesTab()
                .tabItem { Label("루틴", systemImage: "arrow.triangle.2.circlepath") }
                .tag(Tab.routines)

            AffirmationsView()
                .tabItem { Label("다짐", systemImage: "heart.fill") }
                .tag(Tab.affirmations)
        }
        .tint(DesignTokens.Color.accent(.affirmations))
    }
}
