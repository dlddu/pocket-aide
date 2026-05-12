import DesignSystem
import PocketAideAPI
import SwiftUI

struct PriorityEditSheet: View {
    enum Mode {
        case create
        case edit(Affirmation)
    }

    let area: DesignTokens.Area = .affirmations
    let mode: Mode
    var onSave: (String, Priority) async -> Void
    var onCancel: () -> Void

    @State private var text: String
    @State private var priority: Priority

    init(mode: Mode, onSave: @escaping (String, Priority) async -> Void, onCancel: @escaping () -> Void) {
        self.mode = mode
        self.onSave = onSave
        self.onCancel = onCancel
        switch mode {
        case .create:
            _text = State(initialValue: "")
            _priority = State(initialValue: .normal)
        case .edit(let existing):
            _text = State(initialValue: existing.text)
            _priority = State(initialValue: existing.priority)
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("우선순위 설정")
                .font(DesignTokens.Typography.font(
                    size: 18,
                    weight: .bold
                ))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .padding(.horizontal, DesignTokens.Spacing.xxl)
                .padding(.top, DesignTokens.Spacing.sm)
                .padding(.bottom, DesignTokens.Spacing.md)

            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                DSCard(area: area, padding: .small) {
                    VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                        Text("다짐 문장")
                            .font(DesignTokens.Typography.font(
                                size: DesignTokens.Typography.captionXs,
                                weight: .bold
                            ))
                            .tracking(2.4)
                            .textCase(.uppercase)
                            .foregroundStyle(DesignTokens.Color.accent(area))
                        TextField(
                            "다짐 문장을 입력하세요",
                            text: $text,
                            axis: .vertical
                        )
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.bodyLg,
                            family: .serif
                        ))
                        .foregroundStyle(DesignTokens.Color.ink(area))
                        .lineLimit(3...6)
                        .accessibilityIdentifier("priority.sheet.text")
                    }
                }

                VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                    Text("노출 빈도")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.captionXs,
                            weight: .bold
                        ))
                        .tracking(2.4)
                        .textCase(.uppercase)
                        .foregroundStyle(DesignTokens.Color.accent(area))

                    DSFilterPills(
                        area: area,
                        options: Priority.allCases,
                        selection: $priority
                    ) { option, _ in
                        VStack(spacing: 4) {
                            PriorityDots(priority: option, layout: .horizontal)
                            Text(label(for: option))
                                .font(DesignTokens.Typography.font(
                                    size: DesignTokens.Typography.body,
                                    weight: .semibold
                                ))
                        }
                    }
                    .accessibilityIdentifier("priority.sheet.pills")
                }
            }
            .padding(.horizontal, DesignTokens.Spacing.xxl)
            .padding(.bottom, DesignTokens.Spacing.md)

            Divider()
                .overlay(DesignTokens.Color.rule(area))

            VStack(spacing: DesignTokens.Spacing.xs) {
                Button {
                    Task {
                        await onSave(text.trimmingCharacters(in: .whitespacesAndNewlines), priority)
                    }
                } label: {
                    Text("저장")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.bodyLg,
                            weight: .bold
                        ))
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, DesignTokens.Spacing.md)
                        .background(DesignTokens.Color.ink(area))
                        .foregroundStyle(DesignTokens.Color.surface(area))
                        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.chip))
                }
                .buttonStyle(.plain)
                .disabled(text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                .accessibilityIdentifier("priority.sheet.save")

                Button("취소", action: onCancel)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                    .padding(.vertical, DesignTokens.Spacing.sm)
                    .accessibilityIdentifier("priority.sheet.cancel")
            }
            .padding(.horizontal, DesignTokens.Spacing.xxl)
            .padding(.top, DesignTokens.Spacing.sm)
            .padding(.bottom, DesignTokens.Spacing.xl)
        }
    }

    private func label(for priority: Priority) -> String {
        switch priority {
        case .high: return "높음"
        case .normal: return "보통"
        case .low: return "가끔"
        }
    }
}
