import DesignSystem
import PocketAideAPI
import SwiftUI

struct ScratchpadCountSection: View {
    let state: WidgetScratchpadState

    var body: some View {
        Link(destination: URL(string: "pocketaide://scratchpad")!) {
            HStack(spacing: DesignTokens.Spacing.xs) {
                AreaLabel(area: .scratchpad, text: "임시 공간", showsDot: false)
                Spacer(minLength: DesignTokens.Spacing.sm)
                content
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(.plain)
    }

    @ViewBuilder
    private var content: some View {
        switch state {
        case .loaded(let unclassified) where unclassified > 0:
            Text(WidgetScratchpad.countText(unclassified))
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                .monospacedDigit()
                .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
                .lineLimit(1)
        case .loaded:
            notice(WidgetScratchpad.countText(0))
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
            .lineLimit(1)
    }
}
