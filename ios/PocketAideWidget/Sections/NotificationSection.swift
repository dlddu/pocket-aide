import DesignSystem
import PocketAideAPI
import SwiftUI

struct NotificationSection: View {
    let state: WidgetNotificationState

    var body: some View {
        Link(destination: URL(string: "pocketaide://pr-monitor")!) {
            Group {
                if case .loaded(let summary) = state, let latest = summary.latest {
                    VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                        label
                        detail(latest, moreCount: summary.moreCount)
                    }
                } else {
                    HStack(spacing: DesignTokens.Spacing.xs) {
                        label
                        Spacer(minLength: DesignTokens.Spacing.sm)
                        notice
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(.plain)
    }

    private var label: some View {
        AreaLabel(area: .scratchpad, text: "PocketAide 알림", showsDot: false)
    }

    private func detail(_ latest: NotificationHistoryItem, moreCount: Int) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(PushText.title(for: latest))
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
                .lineLimit(1)
            Text(PushText.body(for: latest))
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.75))
                .lineLimit(1)
            if moreCount > 0 {
                Text("+ \(moreCount) 더")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))
            }
        }
    }

    private var notice: some View {
        Text(noticeText)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.6))
            .lineLimit(1)
    }

    private var noticeText: String {
        switch state {
        case .loaded: return "확인할 알림이 없어요."
        case .needsLogin: return "앱에서 로그인이 필요해요."
        case .error: return "잠시 후 다시 시도할게요."
        }
    }
}
