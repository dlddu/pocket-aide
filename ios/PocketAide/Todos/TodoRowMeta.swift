import DesignSystem
import PocketAideAPI
import SwiftUI

struct TodoRowMeta: View {
    let area: TodoArea
    let item: TodoItem

    private var tone: DesignTokens.Area { area.designArea }

    var body: some View {
        if !parts.isEmpty {
            HStack(spacing: DesignTokens.Spacing.md) {
                ForEach(Array(parts.enumerated()), id: \.offset) { index, part in
                    if area == .work && index > 0 {
                        Text("·")
                    }
                    if area == .personal && index == 0 && item.dueDate != nil {
                        TodoDueItem(text: part)
                    } else {
                        Text(part)
                    }
                }
            }
            .font(metaFont)
            .textCase(area == .work ? .uppercase : nil)
            .foregroundStyle(DesignTokens.Color.ink(tone).opacity(0.55))
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(parts.joined(separator: " · "))
            .accessibilityAddTraits(.isStaticText)
            .accessibilityIdentifier("todos.\(area.rawValue).row.\(item.id).meta")
        }
    }

    private var parts: [String] {
        var parts: [String] = []
        if let due = item.dueDate { parts.append(due) }
        if let priority = item.priority {
            parts.append(area == .work ? "P\(workRank(priority))" : priority.displayName)
        }
        if !item.memo.isEmpty { parts.append("메모: \(item.memo)") }
        return parts
    }

    private func workRank(_ priority: TodoPriority) -> Int {
        switch priority {
        case .high: return 1
        case .normal: return 2
        case .low: return 3
        }
    }

    private var metaFont: Font {
        let base = DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)
        return area == .work ? base.monospaced() : base
    }
}

private struct TodoDueItem: View {
    let text: String

    var body: some View {
        HStack(spacing: DesignTokens.Spacing.xs) {
            Image(systemName: "calendar")
                .font(.system(size: 11))
            Text(text)
        }
    }
}
