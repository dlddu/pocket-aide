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
    @StateObject private var scratchpad: ScratchpadViewModel

    init(selectedTab: Binding<RootTab>, highlightedEventID: Binding<Int64?> = .constant(nil)) {
        _selectedTab = selectedTab
        _highlightedEventID = highlightedEventID
        _scratchpad = StateObject(wrappedValue: ScratchpadViewModel(api: nil))
    }

    var body: some View {
        Group {
            if auth.signedIn {
                signedInContent
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
        // PR 모니터 탭은 More 로 밀리지 않는 앞자리에 둔다 — 푸시 탭 deep link 가
        // selection 을 이 탭으로 바꿔 바로 연다.
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
                ScratchpadTab(viewModel: scratchpad)
            }
            .badge(scratchpad.unclassifiedCount)
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
        .task {
            if scratchpad.api == nil, let api = auth.api {
                scratchpad.replaceAPI(api)
            }
            await scratchpad.load()
        }
    }
}
