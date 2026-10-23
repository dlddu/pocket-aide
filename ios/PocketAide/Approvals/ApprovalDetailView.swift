import DesignSystem
import PocketAideAPI
import SwiftUI

struct ApprovalDetailView: View {
    @StateObject private var viewModel: ApprovalDetailViewModel
    @State private var confirming: ApprovalDecision?

    private static let area: DesignTokens.Area = .prMonitor

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.locale = Locale(identifier: "ko_KR")
        f.dateFormat = "HH:mm:ss"
        return f
    }()

    init(item: ApprovalRequest, api: APIClient?) {
        _viewModel = StateObject(wrappedValue: ApprovalDetailViewModel(item: item, api: api))
    }

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(Self.area).ignoresSafeArea()
            ScrollView {
                VStack(alignment: .leading, spacing: 10) {
                    Text(ApprovalFormatting.summary(of: viewModel.item.context))
                        .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area))
                        .accessibilityIdentifier("approval.detail.title")
                    TimelineView(.periodic(from: .now, by: 1)) { context in
                        content(now: context.date)
                    }
                }
                .padding(.horizontal, DesignTokens.Spacing.xl)
                .padding(.top, DesignTokens.Spacing.md)
                .padding(.bottom, DesignTokens.Spacing.xxl)
            }
        }
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                keyChip
            }
        }
        .task { await viewModel.reload() }
        .alert(
            confirming.map(ApprovalFormatting.confirmTitle) ?? "",
            isPresented: Binding(get: { confirming != nil }, set: { if !$0 { confirming = nil } }),
            presenting: confirming
        ) { decision in
            Button("취소", role: .cancel) {}
            Button(decision == .approve ? "승인" : "거절") {
                Task { await viewModel.decide(decision) }
            }
        } message: { decision in
            Text(ApprovalFormatting.confirmMessage(decision, context: viewModel.item.context))
        }
    }

    private var keyChip: some View {
        Text(viewModel.item.keyName)
            .font(.system(size: 10.5, weight: .bold, design: .monospaced))
            .padding(.horizontal, DesignTokens.Spacing.sm)
            .padding(.vertical, DesignTokens.Spacing.xs)
            .background(DesignTokens.Color.soft(Self.area))
            .foregroundStyle(DesignTokens.StatusColor.arrivalGlow)
            .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
            .accessibilityIdentifier("approval.detail.key.chip")
    }

    @ViewBuilder
    private func content(now: Date) -> some View {
        let item = viewModel.item
        let expiredNow = item.isExpired(at: now)
        VStack(alignment: .leading, spacing: 10) {
            if item.status == .pending, !expiredNow, let expiresAt = item.expiresAt {
                countdown(item: item, expiresAt: expiresAt, now: now)
            }
            if let notice = viewModel.notice {
                noticeBanner(notice)
            } else if viewModel.loadError != nil {
                reloadBanner
            }
            if item.keyRevoked {
                Text("폐기된 키로 생성됨 — 자동 처리되지 않음")
                    .font(DesignTokens.Typography.font(size: 10.5, weight: .bold))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .accessibilityIdentifier("approval.detail.revoked")
            }
            Text(item.context)
                .font(.system(size: 10.5, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 10)
                .background(DesignTokens.Color.card(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
                )
                .accessibilityIdentifier("approval.detail.context")
            facts(item: item, status: expiredNow ? .expired : item.status)
            if item.status == .pending {
                decisionButtons(enabled: viewModel.canDecide(at: now))
            }
        }
    }

    private func countdown(item: ApprovalRequest, expiresAt: Int64, now: Date) -> some View {
        HStack(spacing: 10) {
            Text("남은 시간")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            Text(ApprovalFormatting.countdown(expiresAt: expiresAt, now: now))
                .font(.system(size: 18, weight: .bold, design: .monospaced))
                .monospacedDigit()
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
                .accessibilityIdentifier("approval.detail.countdown")
            ProgressView(value: ApprovalFormatting.progress(createdAt: item.createdAt, expiresAt: expiresAt, now: now))
                .tint(DesignTokens.Color.accent(Self.area))
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, DesignTokens.Spacing.sm)
        .background(DesignTokens.Color.card(Self.area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
        )
    }

    private func facts(item: ApprovalRequest, status: ApprovalStatus) -> some View {
        LazyVGrid(columns: [GridItem(.flexible(), alignment: .topLeading), GridItem(.flexible(), alignment: .topLeading)], spacing: 6) {
            fact("보낸 키", item.keyName, id: "key")
            fact("요청자", item.requesterName.isEmpty ? "—" : item.requesterName, id: "requester")
            fact("요청 ID", item.externalId, id: "externalId")
            fact("요청 시각", time(item.createdAt), id: "createdAt")
            fact("상태", ApprovalFormatting.statusLabel(status), id: "status")
            if status != .pending {
                fact("처리 방식", ApprovalFormatting.decisionModeLabel(status: status, mode: item.decisionMode), id: "mode")
                fact("처리 시각", time(item.processedAt ?? item.expiresAt), id: "processedAt")
            }
        }
    }

    private func fact(_ label: String, _ value: String, id: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            Text(value)
                .font(.system(size: DesignTokens.Typography.captionXs, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
                .accessibilityIdentifier("approval.detail.\(id)")
        }
    }

    private func time(_ seconds: Int64?) -> String {
        guard let seconds else { return "—" }
        return Self.timeFormatter.string(from: Date(timeIntervalSince1970: TimeInterval(seconds)))
    }

    private func decisionButtons(enabled: Bool) -> some View {
        HStack(spacing: 10) {
            Button { confirming = .reject } label: {
                Text("거절")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, DesignTokens.Spacing.md)
                    .foregroundStyle(DesignTokens.Color.accent(Self.area))
                    .background(DesignTokens.Color.card(Self.area))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10, style: .continuous)
                            .stroke(DesignTokens.Color.accent(Self.area), lineWidth: 1)
                    )
                    .contentShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
            .accessibilityIdentifier("approval.detail.reject")
            Button { confirming = .approve } label: {
                Text("승인")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, DesignTokens.Spacing.md)
                    .foregroundStyle(.white)
                    .background(DesignTokens.Color.accent(Self.area))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                    .contentShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
            .accessibilityIdentifier("approval.detail.approve")
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
        .opacity(enabled ? 1 : 0.35)
        .padding(.top, DesignTokens.Spacing.xs)
    }

    private func noticeBanner(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
            .foregroundStyle(DesignTokens.Color.ink(Self.area))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, DesignTokens.Spacing.md)
            .padding(.vertical, 10)
            .background(DesignTokens.Color.soft(Self.area))
            .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
            .accessibilityIdentifier("approval.detail.notice")
    }

    private var reloadBanner: some View {
        HStack {
            Text("최신 상태를 불러오지 못했습니다")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
            Spacer(minLength: 0)
            Button("다시 시도") {
                Task { await viewModel.reload() }
            }
            .buttonStyle(.borderless)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
            .foregroundStyle(DesignTokens.Color.accent(Self.area))
            .accessibilityIdentifier("approval.detail.retry")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(Self.area))
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}
