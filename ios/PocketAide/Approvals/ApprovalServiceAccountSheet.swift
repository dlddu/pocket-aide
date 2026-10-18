import DesignSystem
import PocketAideAPI
import SwiftUI

struct ApprovalServiceAccountRegisterButton: View {
    let action: () -> Void

    private static let area: DesignTokens.Area = .prMonitor

    var body: some View {
        Button(action: action) {
            Text("+ k8s 서비스 어카운트 등록")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                .frame(maxWidth: .infinity)
                .padding(.vertical, DesignTokens.Spacing.md)
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .stroke(DesignTokens.Color.accent(Self.area), style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
                )
        }
        .buttonStyle(.plain)
        .padding(.top, DesignTokens.Spacing.sm)
        .accessibilityIdentifier("approval.keys.register-sa")
    }
}

struct ApprovalServiceAccountSheet: View {
    @ObservedObject var viewModel: ApprovalKeysViewModel
    @Binding var isPresented: Bool
    @State private var name = ""
    @State private var namespace = ""
    @State private var account = ""

    private static let area: DesignTokens.Area = .prMonitor

    private var canSubmit: Bool {
        ApprovalKeyFormatting.trimmedName(name) != nil
            && ApprovalKeyFormatting.trimmedName(namespace) != nil
            && ApprovalKeyFormatting.trimmedName(account) != nil
            && !viewModel.isRegistering
    }

    private var accountRefused: Bool {
        viewModel.registerOutcome == .duplicateServiceAccount
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("k8s 서비스 어카운트 등록")
                .font(DesignTokens.Typography.font(size: 17, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
            label("호출자 이름")
            field(text: $name, id: "approval.keys.sa.name", refused: false)
            HStack(alignment: .top, spacing: 10) {
                VStack(alignment: .leading, spacing: 0) {
                    label("네임스페이스")
                    field(text: $namespace, id: "approval.keys.sa.namespace", refused: false)
                }
                VStack(alignment: .leading, spacing: 0) {
                    label("서비스 어카운트")
                    field(text: $account, id: "approval.keys.sa.account", refused: accountRefused)
                }
            }
            if let notice = viewModel.registerNotice {
                Text(notice)
                    .font(DesignTokens.Typography.font(size: 11.5))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .padding(.top, DesignTokens.Spacing.sm)
                    .accessibilityIdentifier("approval.keys.sa.notice")
            }
            Text("이 서비스 어카운트로 도는 파드가 audience pocket-aide 토큰으로 요청합니다. 발급할 키 원문은 없습니다.")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .lineHeight(.multiple(factor: 1.625))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                .padding(.top, DesignTokens.Spacing.sm)
            HStack(spacing: 10) {
                sheetButton("취소", primary: false, id: "approval.keys.sa.cancel") { isPresented = false }
                sheetButton("등록", primary: true, id: "approval.keys.sa.submit") {
                    Task {
                        if await viewModel.registerServiceAccount(name: name, namespace: namespace, account: account) {
                            isPresented = false
                        }
                    }
                }
                .disabled(!canSubmit)
            }
            .padding(.top, DesignTokens.Spacing.lg)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.xl)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(DesignTokens.Color.card(Self.area))
    }

    private func label(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            .padding(.top, DesignTokens.Spacing.md)
            .padding(.bottom, DesignTokens.Spacing.xs)
    }

    private func field(text: Binding<String>, id: String, refused: Bool) -> some View {
        TextField("homelab-k3s-mcp", text: text)
            .textInputAutocapitalization(.never)
            .autocorrectionDisabled()
            .font(.system(size: DesignTokens.Typography.body, design: .monospaced))
            .padding(.horizontal, DesignTokens.Spacing.md)
            .padding(.vertical, 10)
            .background(DesignTokens.Color.surface(Self.area))
            .overlay(
                RoundedRectangle(cornerRadius: 6, style: .continuous)
                    .stroke(refused ? DesignTokens.StatusColor.failure : DesignTokens.Color.rule(Self.area), lineWidth: 1)
            )
            .accessibilityIdentifier(id)
    }

    private func sheetButton(_ title: String, primary: Bool, id: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Text(title)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: primary ? .bold : .regular))
                .frame(maxWidth: .infinity)
                .padding(.vertical, DesignTokens.Spacing.md)
                .foregroundStyle(primary ? Color.white : DesignTokens.Color.ink(Self.area))
                .background(primary ? DesignTokens.Color.accent(Self.area) : DesignTokens.Color.card(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .stroke(primary ? DesignTokens.Color.accent(Self.area) : DesignTokens.Color.rule(Self.area), lineWidth: 1)
                )
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier(id)
    }
}
