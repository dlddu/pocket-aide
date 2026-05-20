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
        // 탭 순서: 구현된 탭(다짐·PR 모니터)을 앞에 두고 placeholder를 뒤로.
        // iPhone compact는 첫 5개를 직접 노출하고 6번째부터 More로 보내므로,
        // PR 모니터가 직접 탭으로 노출되어 deep link selection이 즉시 작동한다.
        // iOS 18+ Tab(value:) API — selection이 customization/More 안 자식까지
        // 전파되어 deep link로 PR 모니터 탭을 외부에서 활성화할 수 있다.
        TabView(selection: $selectedTab) {
            Tab("다짐", systemImage: "heart.fill", value: RootTab.affirmations) {
                AffirmationsView()
            }
            Tab("PR 모니터", systemImage: "checkmark.seal", value: RootTab.prMonitor) {
                PRMonitorView(highlightedEventID: $highlightedEventID)
            }
            Tab("채팅", systemImage: "bubble.left.and.bubble.right", value: RootTab.chat) {
                ChatTab()
            }
            Tab("임시공간", systemImage: "doc.text", value: RootTab.scratchpad) {
                ScratchpadTab()
            }
            Tab("개인", systemImage: "person", value: RootTab.personal) {
                PersonalTab()
            }
            Tab("회사", systemImage: "briefcase", value: RootTab.work) {
                WorkTab()
            }
            Tab("루틴", systemImage: "arrow.triangle.2.circlepath", value: RootTab.routines) {
                RoutinesTab()
            }
        }
        .tint(activeTint)
    }
}
