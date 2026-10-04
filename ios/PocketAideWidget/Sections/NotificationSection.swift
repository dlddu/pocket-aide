import DesignSystem
import PocketAideAPI
import SwiftUI

struct NotificationSection: View {
    let state: WidgetNotificationState

    var body: some View {
        Link(destination: URL(string: "pocketaide://pr-monitor")!) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                AreaLabel(area: .scratchpad, text: "PocketAide 알림", showsDot: false)
                content
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(.plain)
    }

    @ViewBuilder
    private var content: some View {
        switch state {
        case .loaded(let summary):
            if let latest = summary.latest {
                VStack(alignment: .leading, spacing: 2) {
                    Text(PushText.title(for: latest))
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
                        .lineLimit(1)
                    Text(PushText.body(for: latest))
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                        .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.75))
                        .lineLimit(1)
                    if summary.moreCount > 0 {
                        Text("+ \(summary.moreCount) 더")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                            .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))
                    }
                }
            } else {
                notice("확인할 알림이 없어요.")
            }
        case .needsLogin:
            notice("앱에서 로그인이 필요해요.")
        case .error:
            notice("잠시 후 다시 시도할게요.")
        }
    }

    private func notice(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.6))
            .lineLimit(2)
    }
}
