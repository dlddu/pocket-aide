import DesignSystem
import PocketAideAPI
import SwiftUI

struct RoutineHistoryRange: View {
    let days: [RoutineHistoryDay]

    var body: some View {
        if let first = days.first, let today = days.last {
            HStack {
                Text(RoutineDayFormat.title(of: first.day))
                Spacer(minLength: 0)
                if today.scheduled {
                    Text("오늘 · \(today.progressPercent)%")
                        .accessibilityIdentifier("routines.history.today")
                }
            }
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
            .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
            .padding(.top, -DesignTokens.Spacing.sm)
        }
    }
}
