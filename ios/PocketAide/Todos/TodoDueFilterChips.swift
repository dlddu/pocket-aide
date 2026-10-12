import DesignSystem
import PocketAideAPI
import SwiftUI

struct TodoDueFilterChips: View {
    let area: TodoArea
    let total: Int
    @Binding var selection: TodoDueFilter

    private var tone: DesignTokens.Area { area.designArea }

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 6) {
                ForEach(TodoDueFilter.allCases, id: \.self) { filter in
                    let isSelected = filter == selection
                    Button {
                        selection = filter
                    } label: {
                        Text(title(filter))
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: isSelected ? .semibold : .medium))
                            .foregroundStyle(isSelected ? DesignTokens.Color.ink(tone) : DesignTokens.Color.ink(tone).opacity(0.7))
                            .lineLimit(1)
                            .padding(.horizontal, DesignTokens.Spacing.md)
                            .padding(.vertical, DesignTokens.Spacing.xs)
                            .background(Capsule().fill(isSelected ? DesignTokens.Color.soft(tone) : Color.clear))
                            .overlay(Capsule().stroke(isSelected ? Color.clear : DesignTokens.Color.rule(tone), lineWidth: 1))
                    }
                    .buttonStyle(.plain)
                    .accessibilityIdentifier("todos.\(area.rawValue).filter.\(filter.rawValue)")
                    .accessibilityAddTraits(isSelected ? .isSelected : [])
                }
            }
            .padding(.horizontal, DesignTokens.Spacing.xl)
        }
        .padding(.bottom, DesignTokens.Spacing.md)
    }

    private func title(_ filter: TodoDueFilter) -> String {
        switch filter {
        case .all: return "전체 · \(total)"
        case .today: return "오늘"
        case .thisWeek: return "이번 주"
        case .noDate: return "날짜 없음"
        }
    }
}
