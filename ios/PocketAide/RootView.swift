import DesignSystem
import SwiftUI
import UIKit

enum RootTab: Hashable {
    case chat, scratchpad, personal, work, routines, affirmations, prMonitor
}

struct RootView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Binding var selectedTab: RootTab
    @Binding var highlightedEventID: Int64?

    init(selectedTab: Binding<RootTab>, highlightedEventID: Binding<Int64?> = .constant(nil)) {
        _selectedTab = selectedTab
        _highlightedEventID = highlightedEventID
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

    private var activeTint: Color {
        // SwiftUI applies `.tint` globally to the TabView; we still want each
        // tab's selected state to match its area accent. Picking the active
        // tab's accent at render time keeps the visual identity consistent
        // for the PR-monitor and affirmations tabs (the two with the loudest
        // custom palettes).
        switch selectedTab {
        case .prMonitor: return DesignTokens.Color.accent(.prMonitor)
        case .affirmations: return DesignTokens.Color.accent(.affirmations)
        case .personal: return DesignTokens.Color.accent(.personal)
        case .work: return DesignTokens.Color.accent(.work)
        case .routines: return DesignTokens.Color.accent(.routines)
        case .chat: return DesignTokens.Color.accent(.aiChat)
        case .scratchpad: return DesignTokens.Color.accent(.scratchpad)
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

            PRMonitorView(highlightedEventID: $highlightedEventID)
                .tabItem { Label("PR 모니터", systemImage: "checkmark.seal") }
                .tag(RootTab.prMonitor)
        }
        .tint(activeTint)
    }
}
