import DesignSystem
import PocketAideAPI
import SwiftUI

struct PriorityEditSheet: View {
    let mode: Mode
    let onSave: (String, AffirmationPriority) -> Void
    let onCancel: () -> Void
    let onDelete: (() -> Void)?

    @State private var text: String
    @State private var priority: AffirmationPriority

    init(
        mode: Mode,
        onSave: @escaping (String, AffirmationPriority) -> Void,
        onCancel: @escaping () -> Void,
        onDelete: (() -> Void)? = nil
    ) {
        self.mode = mode
        self.onSave = onSave
        self.onCancel = onCancel
        self.onDelete = onDelete
        switch mode {
        case .create:
            _text = State(initialValue: "")
            _priority = State(initialValue: .normal)
        case .edit(let existing):
            _text = State(initialValue: existing.text)
            _priority = State(initialValue: existing.priority)
        }
    }

    enum Mode: Equatable {
        case create
        case edit(Affirmation)

        var title: String {
            switch self {
            case .create: return "새 다짐"
            case .edit: return "우선순위 설정"
            }
        }
    }

    var body: some View {
        Sheet(area: .affirmations, onClose: onCancel) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
                Text(mode.title)
                    .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                    .tracking(-0.5)
                    .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    .padding(.horizontal, 24)
                    .padding(.top, 8)
                    .accessibilityIdentifier("sheet.title")

                VStack(alignment: .leading, spacing: DesignTokens.Spacing.xl) {
                    editorCard
                    prioritySection
                }
                .padding(.horizontal, 24)

                VStack(spacing: 0) {
                    Rectangle()
                        .fill(DesignTokens.Color.rule(.affirmations).opacity(0.6))
                        .frame(height: 1)
                    actions
                        .padding(.horizontal, 24)
                        .padding(.top, 8)
                        .padding(.bottom, 28)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    @ViewBuilder
    private var editorCard: some View {
        if isEditing {
            Card(area: .affirmations, padding: .medium) {
                textField
            }
            .overlay {
                Text("\"")
                    .font(DesignTokens.Typography.font(size: 64, family: .serif))
                    .lineHeight(.multiple(factor: 1))
                    .foregroundStyle(DesignTokens.Color.accent(.affirmations).opacity(0.1))
                    .fixedSize()
                    .padding(.leading, -4)
                    .padding(.top, -8)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
            }
        } else {
            textField
                .padding(.horizontal, DesignTokens.Spacing.lg)
                .padding(.vertical, 14)
                .background(DesignTokens.Color.card(.affirmations))
                .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
                .overlay(
                    RoundedRectangle(cornerRadius: 24, style: .continuous)
                        .stroke(DesignTokens.Color.accent(.affirmations), lineWidth: 1)
                )
        }
    }

    private var textField: some View {
        TextField(
            "다짐 문장을 입력하세요",
            text: $text,
            axis: .vertical
        )
        .font(DesignTokens.Typography.font(size: 16.5, family: .serif))
        .lineHeight(.multiple(factor: 1.5))
        .foregroundStyle(DesignTokens.Color.ink(.affirmations))
        .lineLimit(2...6)
        .accessibilityIdentifier("sheet.text.field")
    }

    private var prioritySection: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                AreaLabel(area: .affirmations, text: "노출 빈도")
                FilterPills(
                    area: .affirmations,
                    options: AffirmationPriority.allCases,
                    selection: $priority
                ) { option in
                    VStack(spacing: 4) {
                        priorityDots(for: option)
                        Text(option.displayName)
                            .font(DesignTokens.Typography.font(
                                size: DesignTokens.Typography.body,
                                weight: option == priority ? .bold : .medium
                            ))
                    }
                }
            }
            Text("위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다. 나중에 카드를 길게 눌러 변경할 수 있습니다.")
                .font(DesignTokens.Typography.font(
                    size: 11.5,
                    weight: .regular
                ))
                .lineHeight(.multiple(factor: 1.625))
                .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    private var actions: some View {
        VStack(spacing: 4) {
            Button(action: handleSave) {
                Text("저장")
                    .font(DesignTokens.Typography.font(
                        size: DesignTokens.Typography.bodyLg,
                        weight: .bold
                    ))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(
                        Capsule().fill(DesignTokens.Color.ink(.affirmations))
                    )
                    .foregroundStyle(DesignTokens.Color.surface(.affirmations))
            }
            .buttonStyle(.plain)
            .disabled(trimmedText.isEmpty)
            .opacity(trimmedText.isEmpty ? 0.5 : 1)
            .accessibilityIdentifier("sheet.save.button")

            if isEditing, let onDelete {
                Button(action: onDelete) {
                    Text("삭제")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.bodyLg,
                            weight: .semibold
                        ))
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .overlay(
                            Capsule()
                                .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)
                        )
                        .foregroundStyle(DesignTokens.Color.destructive(.affirmations))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("sheet.delete.button")
            }

            Button("취소", action: onCancel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))
                .padding(.vertical, 10)
                .accessibilityIdentifier("sheet.cancel.button")
        }
    }

    private var isEditing: Bool {
        if case .edit = mode { return true }
        return false
    }

    private func priorityDots(for option: AffirmationPriority) -> some View {
        let filled: Int = {
            switch option {
            case .high: return 3
            case .normal: return 2
            case .low: return 1
            }
        }()
        return HStack(spacing: 2) {
            ForEach(0..<3, id: \.self) { idx in
                Circle()
                    .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled ? 1 : 0.3))
                    .frame(width: 6, height: 6)
            }
        }
    }

    private var trimmedText: String {
        text.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func handleSave() {
        guard !trimmedText.isEmpty else { return }
        onSave(trimmedText, priority)
    }
}
