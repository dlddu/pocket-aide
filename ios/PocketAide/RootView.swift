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
        TabView(selection: $selectedTab) {
            ChatTab()
                .tabItem { Label("채팅", systemImage: "bubble.left.and.bubble.right") }
                .tag(Tab.chat)
                .accessibilityIdentifier("tab.chat")

            ScratchpadTab()
                .tabItem { Label("임시공간", systemImage: "doc.text") }
                .tag(Tab.scratchpad)
                .accessibilityIdentifier("tab.scratchpad")

            PersonalTab()
                .tabItem { Label("개인", systemImage: "person") }
                .tag(Tab.personal)
                .accessibilityIdentifier("tab.personal")

            WorkTab()
                .tabItem { Label("회사", systemImage: "briefcase") }
                .tag(Tab.work)
                .accessibilityIdentifier("tab.work")

            RoutinesTab()
                .tabItem { Label("루틴", systemImage: "arrow.triangle.2.circlepath") }
                .tag(Tab.routines)
                .accessibilityIdentifier("tab.routines")

            AffirmationsView()
                .tabItem { Label("다짐", systemImage: "heart.fill") }
                .tag(Tab.affirmations)
                .accessibilityIdentifier("tab.affirmations")
        }
        .tint(DesignTokens.Color.accent(.affirmations))
    }
}
