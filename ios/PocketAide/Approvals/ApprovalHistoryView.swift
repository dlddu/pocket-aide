import DesignSystem
import PocketAideAPI
import SwiftUI

struct ApprovalHistoryView: View {
    @StateObject private var viewModel: ApprovalHistoryViewModel

    private static let area: DesignTokens.Area = .prMonitor

    init(api: APIClient?) {
        _viewModel = StateObject(wrappedValue: ApprovalHistoryViewModel(api: api))
    }

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(Self.area).ignoresSafeArea()
            List {
                sectionHeader
                if let message = viewModel.errorMessage {
                    errorBanner(message)
                }
                if viewModel.items.isEmpty {
                    emptyState
                } else {
                    ForEach(viewModel.items) { item in
                        row(for: item)
                    }
                    if viewModel.hasMore {
                        loadingMore
                    }
                }
            }
            .listStyle(.plain)
            .scrollContentBackground(.hidden)
            .refreshable { await viewModel.load() }
        }
        .navigationBarTitleDisplayMode(.inline)
        .task { await viewModel.load() }
    }

    private var sectionHeader: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Rectangle()
                .fill(DesignTokens.Color.ink(Self.area).opacity(0.45))
                .frame(width: 4, height: 12)
            Text("처리 이력")
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.45))
                .accessibilityIdentifier("approvals.history.title")
            Rectangle()
                .fill(DesignTokens.Color.rule(Self.area))
                .frame(height: 1)
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private func errorBanner(_ message: String) -> some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "exclamationmark.circle")
                .foregroundStyle(DesignTokens.StatusColor.failure)
            Text(message)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                .lineLimit(2)
            Spacer(minLength: 0)
            Button("다시 시도") {
                Task { await viewModel.load() }
            }
            .buttonStyle(.borderless)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
            .foregroundStyle(DesignTokens.Color.accent(Self.area))
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(Self.area))
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("approvals.history.error")
    }

    @ViewBuilder
    private var emptyState: some View {
        Group {
            if viewModel.isLoading || (!viewModel.hasLoaded && viewModel.errorMessage == nil) {
                ProgressView().frame(maxWidth: .infinity)
            }
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private var loadingMore: some View {
        Text("이전 이력 불러오는 중…")
            .font(.system(size: 11, design: .monospaced))
            .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.45))
            .frame(maxWidth: .infinity)
            .padding(.vertical, DesignTokens.Spacing.sm)
            .listRowSeparator(.hidden)
            .listRowBackground(Color.clear)
            .accessibilityIdentifier("approvals.history.loadingMore")
            .task(id: viewModel.nextCursor) { await viewModel.loadMore() }
    }

    private func row(for item: ApprovalRequest) -> some View {
        ApprovalHistoryRowCard(item: item, area: Self.area)
            .listRowSeparator(.hidden)
            .listRowBackground(Color.clear)
            .listRowInsets(EdgeInsets(top: 5, leading: DesignTokens.Spacing.xl, bottom: 5, trailing: DesignTokens.Spacing.xl))
            .accessibilityElement(children: .combine)
            .accessibilityIdentifier("approvals.history.row.\(item.externalId)")
    }
}

private struct ApprovalHistoryRowCard: View {
    let item: ApprovalRequest
    let area: DesignTokens.Area

    private var tint: Color {
        switch item.status {
        case .approved: return DesignTokens.StatusColor.success
        case .rejected: return DesignTokens.StatusColor.failure
        case .pending, .expired: return DesignTokens.Color.ink(area).opacity(0.45)
        }
    }

    var body: some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
            Circle()
                .fill(tint)
                .frame(width: 10, height: 10)
                .padding(.top, 6)
            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Text(ApprovalFormatting.statusLabel(item.status))
                        .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                        .foregroundStyle(tint)
                    Spacer()
                    Text(item.processedAt.map { ApprovalFormatting.historyTime($0) } ?? "")
                        .font(.system(size: 10.5, design: .monospaced))
                        .monospacedDigit()
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.45))
                }
                Text(ApprovalFormatting.summary(of: item.context))
                    .font(DesignTokens.Typography.font(size: 13, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                    .lineLimit(2)
                HStack(spacing: DesignTokens.Spacing.sm) {
                    Text(item.keyName)
                        .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                        .padding(.horizontal, DesignTokens.Spacing.sm)
                        .padding(.vertical, 2)
                        .background(DesignTokens.Color.soft(area))
                        .foregroundStyle(DesignTokens.StatusColor.arrivalGlow)
                        .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
                    Text(ApprovalFormatting.historyMeta(item))
                        .font(.system(size: 10.5, design: .monospaced))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                }
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, DesignTokens.Spacing.md)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}
