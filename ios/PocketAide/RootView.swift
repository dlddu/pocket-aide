import DesignSystem
import SwiftUI
import UIKit

enum RootTab: Hashable {
    case chat, scratchpad, personal, work, routines, affirmations
}

struct RootView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Binding var selectedTab: RootTab

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
                    signedInContent
                }
            } else {
                LoginView()
            }
        }
    }

    private var signedInContent: some View {
        VStack(spacing: 0) {
            if auth.pushAuthorizationDenied {
                pushDeniedBanner
            }
            signedInTabs
        }
    }

    private var pushDeniedBanner: some View {
        Button {
            if let url = URL(string: UIApplication.openSettingsURLString) {
                UIApplication.shared.open(url)
            }
        } label: {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Image(systemName: "bell.slash.fill")
                Text("알림 권한이 꺼져 있어 PR 푸시가 도착하지 않습니다. 설정에서 켜기")
                    .font(.system(size: DesignTokens.Typography.captionXs))
                    .multilineTextAlignment(.leading)
                Spacer(minLength: 0)
            }
            .padding(.horizontal, DesignTokens.Spacing.md)
            .padding(.vertical, DesignTokens.Spacing.sm)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(DesignTokens.Color.accent(.work).opacity(0.18))
            .foregroundStyle(DesignTokens.Color.ink(.work))
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("PushDeniedBanner")
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
                .tag(RootTab.chat)

            ScratchpadTab()
                .tabItem { Label("임시공간", systemImage: "doc.text") }
                .tag(RootTab.scratchpad)

            PersonalTab()
                .tabItem { Label("개인", systemImage: "person") }
                .tag(RootTab.personal)

            WorkTab()
                .tabItem { Label("회사", systemImage: "briefcase") }
                .tag(RootTab.work)

            RoutinesTab()
                .tabItem { Label("루틴", systemImage: "arrow.triangle.2.circlepath") }
                .tag(RootTab.routines)

            AffirmationsView()
                .tabItem { Label("다짐", systemImage: "heart.fill") }
                .tag(RootTab.affirmations)
        }
        .tint(DesignTokens.Color.accent(.affirmations))
    }
}
