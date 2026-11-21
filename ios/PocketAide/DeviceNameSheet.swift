import DesignSystem
import SwiftUI

struct DeviceNameSheet: View {
    let onSave: (String) -> Void
    let onSkip: () -> Void

    @State private var text: String

    init(initial: String, onSave: @escaping (String) -> Void, onSkip: @escaping () -> Void) {
        self.onSave = onSave
        self.onSkip = onSkip
        _text = State(initialValue: initial)
    }

    var body: some View {
        Sheet(area: .prMonitor, onClose: onSkip) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                    Text("이 기기 이름")
                        .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                        .accessibilityIdentifier("device.name.sheet.title")
                    Text("푸시를 받는 기기를 구분하는 이름이에요.")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                        .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.6))
                }
                .padding(.top, 8)
                Card(area: .prMonitor, padding: .medium) {
                    TextField("예: 출근용 폰", text: $text)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                        .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                        .submitLabel(.done)
                        .onChange(of: text) { _, new in
                            if new.count > DeviceName.maxLength { text = String(new.prefix(DeviceName.maxLength)) }
                        }
                        .accessibilityIdentifier("device.name.sheet.field")
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
                    .background(Capsule().fill(DesignTokens.Color.ink(.prMonitor)))
                    .foregroundStyle(DesignTokens.Color.surface(.prMonitor))
            }
            .buttonStyle(.plain)
            .disabled(name.isEmpty)
            .opacity(name.isEmpty ? 0.5 : 1)
            .accessibilityIdentifier("device.name.sheet.save.button")

            Button("건너뛰기", action: onSkip)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
                .padding(.vertical, 8)
                .accessibilityIdentifier("device.name.sheet.skip.button")
        }
    }

    private var name: String { DeviceName.normalize(text) }

    private func handleSave() {
        guard !name.isEmpty else { return }
        onSave(name)
    }
}
