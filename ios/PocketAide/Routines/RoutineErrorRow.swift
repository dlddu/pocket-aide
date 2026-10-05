import DesignSystem
import SwiftUI

struct RoutineErrorRow: View {
    let message: String
    let canRetry: Bool
    let onRetry: () -> Void

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: DesignTokens.Spacing.sm) {
            Text(message)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.destructive(.routines))
                .accessibilityIdentifier("routines.error")
            Spacer(minLength: 0)
            if canRetry {
                Button("다시 시도", action: onRetry)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.accent(.routines))
                    .buttonStyle(.borderless)
                    .accessibilityIdentifier("routines.error.retry")
            }
        }
    }
}
