import DesignSystem
import SwiftUI

struct ScratchpadAddSheet: View {
    let onSave: (String) -> Void
    let onCancel: () -> Void

    @State private var text = ""

    var body: some View {
        Sheet(area: .scratchpad, onClose: onCancel) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                Text("새 메모")
                    .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
                    .padding(.top, 8)
                    .accessibilityIdentifier("scratchpad.sheet.title")
                Card(area: .scratchpad, padding: .medium) {
                    TextField("분류는 나중에 — 일단 적어두세요", text: $text, axis: .vertical)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                        .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
                        .lineLimit(2...6)
                        .accessibilityIdentifier("scratchpad.sheet.text.field")
                }
                actions
                    .padding(.top, 8)
                    .padding(.bottom, 32)
            }
            .padding(.horizontal, 24)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var actions: some View {
        VStack(spacing: 4) {
            Button(action: handleSave) {
                Text("저장")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(Capsule().fill(DesignTokens.Color.ink(.scratchpad)))
                    .foregroundStyle(DesignTokens.Color.surface(.scratchpad))
            }
            .buttonStyle(.plain)
            .disabled(trimmedText.isEmpty)
            .opacity(trimmedText.isEmpty ? 0.5 : 1)
            .accessibilityIdentifier("scratchpad.sheet.save.button")

            Button("취소", action: onCancel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))
                .padding(.vertical, 8)
                .accessibilityIdentifier("scratchpad.sheet.cancel.button")
        }
    }

    private var trimmedText: String {
        text.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func handleSave() {
        guard !trimmedText.isEmpty else { return }
        onSave(trimmedText)
    }
}
