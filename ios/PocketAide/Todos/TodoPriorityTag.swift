import DesignSystem
import PocketAideAPI
import SwiftUI

struct TodoPriorityTag: View {
    let area: TodoArea
    let item: TodoItem

    private var tone: DesignTokens.Area { area.designArea }

    var body: some View {
        HStack(spacing: 6) {
            Text(TodoPriority.high.displayName)
                .font(DesignTokens.Typography.font(size: 10, weight: .bold))
                .tracking(0.5)
                .textCase(.uppercase)
                .foregroundStyle(DesignTokens.Color.accent(tone))
                .accessibilityIdentifier("todos.\(area.rawValue).row.\(item.id).priority")
        }
        .padding(.bottom, 2)
    }
}
