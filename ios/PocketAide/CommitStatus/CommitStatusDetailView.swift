import DesignSystem
import PocketAideAPI
import SwiftUI

struct CommitStatusDetailView: View {
    @StateObject private var viewModel: CommitStatusDetailViewModel
    @State private var confirming: ApprovalDecision?

    private static let area: DesignTokens.Area = .prMonitor

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.locale = Locale(identifier: "ko_KR")
        f.dateFormat = "MM/dd HH:mm:ss"
        return f
    }()

    init(item: CommitStatusRequest, api: APIClient?) {
        _viewModel = StateObject(wrappedValue: CommitStatusDetailViewModel(item: item, api: api))
    }

    var body: some View {
        let item = viewModel.item
        ZStack {
            DesignTokens.Color.surface(Self.area).ignoresSafeArea()
            ScrollView {
                VStack(alignment: .leading, spacing: 10) {
                    Text(CommitStatusFormatting.title(item))
                        .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area))
                        .accessibilityIdentifier("commitStatus.detail.title")
                    Text(CommitStatusFormatting.target(item))
                        .font(.system(size: DesignTokens.Typography.captionSm, design: .monospaced))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.7))
                    if let notice = viewModel.notice {
                        banner(notice, id: "notice")
                    } else if item.status == .pending, !item.lastError.isEmpty {
                        banner(CommitStatusFormatting.writeFailedNotice(item.lastError), id: "lastError")
                    } else if viewModel.loadError != nil {
                        reloadBanner
                    }
                    if !item.reason.isEmpty {
                        agentReason(item.reason)
                    }
                    facts(item)
                    if item.status == .pending {
                        decisionButtons(enabled: viewModel.canDecide)
                    }
                }
                .padding(.horizontal, DesignTokens.Spacing.xl)
                .padding(.top, DesignTokens.Spacing.md)
                .padding(.bottom, DesignTokens.Spacing.xxl)
            }
        }
        .navigationBarTitleDisplayMode(.inline)
        .task { await viewModel.reload() }
        .alert(
            confirming.map(CommitStatusFormatting.confirmTitle) ?? "",
            isPresented: Binding(get: { confirming != nil }, set: { if !$0 { confirming = nil } }),
            presenting: confirming
        ) { decision in
            Button("취소", role: .cancel) {}
            Button(decision == .approve ? "승인" : "거절") {
                Task { await viewModel.decide(decision) }
            }
        } message: { decision in
            Text(CommitStatusFormatting.confirmMessage(decision, item: viewModel.item))
        }
    }

    private func agentReason(_ reason: String) -> some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
            HStack(alignment: .firstTextBaseline) {
                Text("에이전트 설명")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.accent(Self.area))
                Spacer(minLength: DesignTokens.Spacing.sm)
                Text("에이전트가 작성 · 검증되지 않음")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            }
            Text(reason)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 10)
                .background(DesignTokens.Color.card(Self.area))
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
                )
                .accessibilityIdentifier("commitStatus.detail.reason")
        }
    }

    private func facts(_ item: CommitStatusRequest) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            fact("저장소", "\(item.owner)/\(item.repo)", id: "repo")
            fact("커밋", item.sha, id: "sha")
            fact("context", item.context, id: "context")
            fact("state", item.state, id: "state")
            if !item.description.isEmpty {
                fact("설명", item.description, id: "description")
            }
            if !item.targetUrl.isEmpty {
                VStack(alignment: .leading, spacing: 2) {
                    label("링크")
                    if let url = URL(string: item.targetUrl) {
                        Link(item.targetUrl, destination: url)
                            .font(.system(size: DesignTokens.Typography.captionXs, design: .monospaced))
                            .tint(DesignTokens.Color.accent(Self.area))
                            .accessibilityIdentifier("commitStatus.detail.targetUrl")
                    }
                }
            }
            fact("보낸 키", item.keyName, id: "key")
            fact("요청자", item.requester, id: "requester")
            if !item.sessionName.isEmpty || !item.sessionId.isEmpty {
                fact("세션", item.sessionName.isEmpty ? "제목 없음 · \(item.sessionId.prefix(8))" : item.sessionName, id: "session")
            }
            fact("요청 시각", time(item.createdAt), id: "createdAt")
            fact("상태", CommitStatusFormatting.statusLabel(item.status), id: "status")
            if item.status != .pending {
                fact("처리 시각", time(item.processedAt), id: "processedAt")
            }
            if let statusID = item.githubStatusId {
                fact("GitHub status id", String(statusID), id: "githubStatusId")
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.card(Self.area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
        )
    }

    private func label(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
    }

    private func fact(_ title: String, _ value: String, id: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            label(title)
            Text(value)
                .font(.system(size: DesignTokens.Typography.captionXs, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
                .textSelection(.enabled)
                .accessibilityIdentifier("commitStatus.detail.\(id)")
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
            .accessibilityIdentifier("commitStatus.detail.reject")
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
            .accessibilityIdentifier("commitStatus.detail.approve")
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
        .opacity(enabled ? 1 : 0.35)
        .padding(.top, DesignTokens.Spacing.xs)
    }

    private func banner(_ text: String, id: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
            .foregroundStyle(DesignTokens.Color.ink(Self.area))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, DesignTokens.Spacing.md)
            .padding(.vertical, 10)
            .background(DesignTokens.Color.soft(Self.area))
            .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
            .accessibilityIdentifier("commitStatus.detail.\(id)")
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
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(Self.area))
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}
