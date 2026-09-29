import DesignSystem
import SwiftUI

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
