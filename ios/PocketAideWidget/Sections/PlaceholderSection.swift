import DesignSystem
import SwiftUI

/// Shared shape for the 4 not-yet-wired widget sections (weather, calendar,
/// mail, notifications). Real data ships per PRD-위젯 AC2/3/4/6 follow-ups —
/// for now we show the area-tinted label and an explicit "곧 추가" line so
/// the widget reads as intentional, not broken.
struct PlaceholderSection: View {
    let area: DesignTokens.Area
    let label: String

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
            AreaLabel(area: area, text: label, showsDot: false)
            Text("곧 추가")
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.captionXs,
                    family: .sans
                ))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.6))
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}
