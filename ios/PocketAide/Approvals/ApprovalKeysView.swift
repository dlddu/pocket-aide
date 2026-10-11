import DesignSystem
import PocketAideAPI
import SwiftUI
import UIKit

struct ApprovalKeysView: View {
    @StateObject private var viewModel: ApprovalKeysViewModel
    @State private var showsIssueSheet = false
    @State private var revoking: ApprovalCallerKey?

    private static let area: DesignTokens.Area = .prMonitor

    init(api: APIClient?) {
        _viewModel = StateObject(wrappedValue: ApprovalKeysViewModel(api: api))
    }

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(Self.area).ignoresSafeArea()
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    Text("승인 게이트 설정")
                        .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area))
                        .padding(.bottom, DesignTokens.Spacing.md)
                        .accessibilityIdentifier("approval.keys.title")
                    sectionHeader
                    if let message = viewModel.errorMessage {
                        errorBanner(message)
                    }
                    keyList
                    issueButton
                }
                .padding(.horizontal, DesignTokens.Spacing.xl)
                .padding(.top, DesignTokens.Spacing.md)
                .padding(.bottom, DesignTokens.Spacing.xxl)
            }
            .refreshable { await viewModel.load() }
        }
        .navigationBarTitleDisplayMode(.inline)
        .task { await viewModel.load() }
        .sheet(isPresented: $showsIssueSheet, onDismiss: { viewModel.finishIssuing() }, content: {
            ApprovalKeyIssueSheet(viewModel: viewModel, isPresented: $showsIssueSheet)
                .presentationDetents([.medium, .large])
        })
        .alert(
            revoking.map(ApprovalKeyFormatting.revokeTitle) ?? "",
            isPresented: Binding(get: { revoking != nil }, set: { if !$0 { revoking = nil } }),
            presenting: revoking
        ) { key in
            Button("취소", role: .cancel) {}
            Button("폐기", role: .destructive) {
                Task { await viewModel.revoke(key) }
            }
        } message: { _ in
            Text(ApprovalKeyFormatting.revokeMessage)
        }
    }

    private var sectionHeader: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Rectangle()
                .fill(DesignTokens.Color.accent(Self.area))
                .frame(width: 4, height: 12)
            Text("호출자 키")
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                .tracking(0.3)
                .textCase(.uppercase)
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
            Rectangle()
                .fill(DesignTokens.Color.rule(Self.area))
                .frame(height: 1)
        }
        .padding(.top, DesignTokens.Spacing.xs)
        .padding(.bottom, DesignTokens.Spacing.sm)
    }

    @ViewBuilder
    private var keyList: some View {
        if viewModel.keys.isEmpty {
            Group {
                if viewModel.isLoading || !viewModel.hasLoaded {
                    ProgressView().frame(maxWidth: .infinity)
                } else {
                    Text("발급한 호출자 키가 없습니다")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                        .frame(maxWidth: .infinity)
                        .accessibilityIdentifier("approval.keys.empty")
                }
            }
            .padding(.vertical, DesignTokens.Spacing.lg)
        } else {
            VStack(spacing: 0) {
                ForEach(Array(viewModel.keys.enumerated()), id: \.element.id) { index, key in
                    if index > 0 {
                        Rectangle()
                            .fill(DesignTokens.Color.rule(Self.area))
                            .frame(height: 1)
                    }
                    row(for: key)
                }
            }
            .background(DesignTokens.Color.card(Self.area))
            .overlay(
                RoundedRectangle(cornerRadius: 6, style: .continuous)
                    .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        }
    }

    private func row(for key: ApprovalCallerKey) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                keyChip(key)
                Spacer()
                if let revoked = ApprovalKeyFormatting.revokedLabel(key) {
                    Text(revoked)
                        .font(.system(size: 10.5, design: .monospaced))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                        .accessibilityIdentifier("approval.keys.revoked.\(key.name)")
                } else {
                    Button("폐기") { revoking = key }
                        .buttonStyle(.borderless)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                        .foregroundStyle(DesignTokens.StatusColor.failure)
                        .accessibilityIdentifier("approval.keys.revoke.\(key.name)")
                }
            }
            Text(ApprovalKeyFormatting.meta(key))
                .font(.system(size: 10.5, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                .accessibilityIdentifier("approval.keys.meta.\(key.name)")
        }
        .padding(.horizontal, 14)
        .padding(.vertical, DesignTokens.Spacing.md)
        .opacity(key.isRevoked ? 0.6 : 1)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("approval.keys.row.\(key.name)")
    }

    private func keyChip(_ key: ApprovalCallerKey) -> some View {
        Text(key.name)
            .font(.system(size: 10.5, weight: .bold, design: .monospaced))
            .strikethrough(key.isRevoked)
            .padding(.horizontal, DesignTokens.Spacing.sm)
            .padding(.vertical, 2)
            .background(key.isRevoked ? Color.clear : DesignTokens.Color.soft(Self.area))
            .foregroundStyle(key.isRevoked ? DesignTokens.Color.ink(Self.area).opacity(0.55) : DesignTokens.StatusColor.arrivalGlow)
            .overlay(
                RoundedRectangle(cornerRadius: 2, style: .continuous)
                    .stroke(key.isRevoked ? DesignTokens.Color.rule(Self.area) : Color.clear, lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
    }

    private var issueButton: some View {
        Button {
            showsIssueSheet = true
        } label: {
            Text("+ 호출자 키 발급")
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
        .padding(.top, DesignTokens.Spacing.md)
        .accessibilityIdentifier("approval.keys.issue")
    }

    private func errorBanner(_ message: String) -> some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
            Text("호출자 키를 불러오지 못했습니다")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
            Spacer(minLength: 0)
            Button("다시 시도") {
                Task { await viewModel.load() }
            }
            .buttonStyle(.borderless)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
            .foregroundStyle(DesignTokens.Color.accent(Self.area))
            .accessibilityIdentifier("approval.keys.retry")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(Self.area))
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .padding(.bottom, DesignTokens.Spacing.sm)
    }
}

private struct ApprovalKeyIssueSheet: View {
    @ObservedObject var viewModel: ApprovalKeysViewModel
    @Binding var isPresented: Bool
    @State private var name = ""
    @State private var copied = false

    private static let area: DesignTokens.Area = .prMonitor

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if let issued = viewModel.issued {
                issuedContent(issued)
            } else {
                form
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.xl)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(DesignTokens.Color.card(Self.area))
    }

    private var form: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("호출자 키 발급")
                .font(DesignTokens.Typography.font(size: 17, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
            label("호출자 이름")
            TextField("infra-ci", text: $name)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .font(.system(size: DesignTokens.Typography.body, design: .monospaced))
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 10)
                .background(DesignTokens.Color.surface(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
                )
                .accessibilityIdentifier("approval.keys.issue.name")
            if let notice = viewModel.issueNotice {
                Text(notice)
                    .font(DesignTokens.Typography.font(size: 11.5))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .padding(.top, DesignTokens.Spacing.sm)
                    .accessibilityIdentifier("approval.keys.issue.notice")
            }
            HStack(spacing: 10) {
                sheetButton("취소", primary: false, id: "approval.keys.issue.cancel") { isPresented = false }
                sheetButton("발급", primary: true, id: "approval.keys.issue.submit") {
                    Task { await viewModel.issue(name: name) }
                }
                .disabled(ApprovalKeyFormatting.trimmedName(name) == nil || viewModel.isIssuing)
            }
            .padding(.top, DesignTokens.Spacing.lg)
        }
    }

    private func issuedContent(_ issued: IssuedApprovalKey) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            Text(ApprovalKeyFormatting.issuedTitle)
                .font(DesignTokens.Typography.font(size: 17, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
                .accessibilityIdentifier("approval.keys.issued.title")
            label("호출자 이름")
            Text(issued.key.name)
                .font(.system(size: DesignTokens.Typography.body, design: .monospaced))
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 10)
                .background(DesignTokens.Color.surface(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
                )
            label("키 원문")
            Text(issued.secret)
                .font(.system(size: DesignTokens.Typography.captionSm, design: .monospaced))
                .textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 10)
                .background(DesignTokens.Color.soft(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.accent(Self.area), lineWidth: 1)
                )
                .accessibilityIdentifier("approval.keys.issued.secret")
            Text(ApprovalKeyFormatting.secretWarning)
                .font(DesignTokens.Typography.font(size: 11.5))
                .lineHeight(.multiple(factor: 1.625))
                .foregroundStyle(DesignTokens.StatusColor.failure)
                .padding(.top, DesignTokens.Spacing.sm)
            HStack(spacing: 10) {
                sheetButton(copied ? "복사됨" : "복사", primary: true, id: "approval.keys.issued.copy") {
                    UIPasteboard.general.string = issued.secret
                    copied = true
                }
                sheetButton("완료", primary: false, id: "approval.keys.issued.done") { isPresented = false }
            }
            .padding(.top, DesignTokens.Spacing.lg)
        }
    }

    private func label(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            .padding(.top, DesignTokens.Spacing.md)
            .padding(.bottom, DesignTokens.Spacing.xs)
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
