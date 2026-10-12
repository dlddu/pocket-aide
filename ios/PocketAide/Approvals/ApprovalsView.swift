import DesignSystem
import PocketAideAPI
import SwiftUI

struct ApprovalsView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Environment(\.scenePhase) private var scenePhase
    @ObservedObject var viewModel: ApprovalsViewModel
    @Binding var openedApprovalID: String?
    @State private var path: [ApprovalRequest] = []
    @State private var showsSettings = false
    @State private var showsHistory = false

    private static let area: DesignTokens.Area = .prMonitor

    var body: some View {
        NavigationStack(path: $path) {
            ZStack {
                DesignTokens.Color.surface(Self.area).ignoresSafeArea()
                VStack(spacing: 0) {
                    header
                    list
                }
            }
            .navigationTitle("승인")
            .toolbar(.hidden, for: .navigationBar)
            .navigationDestination(for: ApprovalRequest.self) { item in
                ApprovalDetailView(item: item, api: viewModel.api ?? auth.api)
            }
            .navigationDestination(isPresented: $showsSettings) {
                ApprovalKeysView(api: viewModel.api ?? auth.api)
            }
            .navigationDestination(isPresented: $showsHistory) {
                ApprovalHistoryView(api: viewModel.api ?? auth.api)
            }
        }
        .onChange(of: path) { _, newPath in
            if newPath.isEmpty {
                Task { await viewModel.load() }
            }
        }
        .onChange(of: showsSettings) { _, shown in
            if !shown {
                Task { await viewModel.load() }
            }
        }
        .onChange(of: showsHistory) { _, shown in
            if !shown {
                Task { await viewModel.load() }
            }
        }
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .onChange(of: scenePhase) { _, phase in
            if phase == .active {
                Task { await viewModel.load() }
            }
        }
        .modifier(ApprovalDeepLinkOpener(openedApprovalID: $openedApprovalID, path: $path, showsSettings: $showsSettings, showsHistory: $showsHistory, api: viewModel.api ?? auth.api))
    }

    private var header: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(DesignTokens.Color.accent(Self.area))
                        .frame(width: 8, height: 8)
                    Text("APPROVAL · GATE")
                        .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                        .tracking(2.4)
                        .foregroundStyle(DesignTokens.Color.accent(Self.area))
                }
                Text("승인")
                    .font(DesignTokens.Typography.font(size: 26, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .accessibilityIdentifier("approvals.title")
            }
            Spacer()
            VStack(spacing: 2) {
                Text("\(viewModel.pendingCount)")
                    .font(.system(size: 15, weight: .bold, design: .monospaced))
                    .accessibilityIdentifier("approvals.pending.count")
                Text("대기")
                    .font(.system(size: 8, weight: .bold, design: .monospaced))
            }
            .foregroundStyle(.white)
            .padding(.horizontal, 10)
            .padding(.vertical, DesignTokens.Spacing.xs)
            .background(DesignTokens.Color.accent(Self.area))
            .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
            Button {
                showsHistory = true
            } label: {
                Image(systemName: "clock")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .frame(width: 36, height: 36)
                    .background(DesignTokens.Color.card(Self.area))
                    .overlay(
                        RoundedRectangle(cornerRadius: 6, style: .continuous)
                            .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
                    )
            }
            .buttonStyle(.plain)
            .accessibilityLabel("처리 이력")
            .accessibilityIdentifier("approvals.history.button")
            Button {
                showsSettings = true
            } label: {
                Image(systemName: "gearshape")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .frame(width: 36, height: 36)
                    .background(DesignTokens.Color.card(Self.area))
                    .overlay(
                        RoundedRectangle(cornerRadius: 6, style: .continuous)
                            .stroke(DesignTokens.Color.rule(Self.area), lineWidth: 1)
                    )
            }
            .buttonStyle(.plain)
            .accessibilityLabel("승인 게이트 설정")
            .accessibilityIdentifier("approvals.settings.button")
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
        .padding(.bottom, DesignTokens.Spacing.md)
    }

    private var list: some View {
        List {
            if let message = viewModel.errorMessage {
                errorBanner(message)
            }
            if viewModel.items.isEmpty {
                emptyState
            } else {
                sectionHeader
                ForEach(viewModel.items) { item in
                    row(for: item)
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.load() }
        .overlay(alignment: .top) {
            if viewModel.isLoading && !viewModel.items.isEmpty {
                ProgressView()
                    .padding(.top, DesignTokens.Spacing.xs)
                    .accessibilityIdentifier("approvals.loading")
            }
        }
    }

    private func errorBanner(_ message: String) -> some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "exclamationmark.circle")
                .foregroundStyle(DesignTokens.StatusColor.failure)
            VStack(alignment: .leading, spacing: 2) {
                Text("승인 요청을 불러오지 못했습니다")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                Text(message)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                    .lineLimit(2)
            }
            Spacer(minLength: 0)
            Button("다시 시도") {
                Task { await viewModel.load() }
            }
            .buttonStyle(.borderless)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
            .foregroundStyle(DesignTokens.Color.accent(Self.area))
            .accessibilityIdentifier("approvals.error.retry")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(Self.area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.accent(Self.area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("approvals.error.banner")
    }

    @ViewBuilder
    private var emptyState: some View {
        Group {
            if viewModel.isLoading || (!viewModel.hasLoaded && viewModel.errorMessage == nil) {
                ProgressView().frame(maxWidth: .infinity)
            } else if viewModel.errorMessage == nil {
                Text("대기 중인 승인 요청이 없습니다")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                    .frame(maxWidth: .infinity)
                    .padding(.top, DesignTokens.Spacing.xxl)
                    .accessibilityIdentifier("approvals.empty.state")
            }
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private var sectionHeader: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Rectangle()
                .fill(DesignTokens.Color.accent(Self.area))
                .frame(width: 4, height: 12)
            Text(ApprovalFormatting.pendingHeader(viewModel.pendingCount))
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
                .accessibilityIdentifier("approvals.section.pending")
            Rectangle()
                .fill(DesignTokens.Color.rule(Self.area))
                .frame(height: 1)
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private func row(for item: ApprovalRequest) -> some View {
        Button {
            path.append(item)
        } label: {
            if item.expiresAt == nil {
                card(for: item, now: Date())
            } else {
                TimelineView(.periodic(from: .now, by: 1)) { context in
                    card(for: item, now: context.date)
                }
            }
        }
        .buttonStyle(.plain)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .listRowInsets(EdgeInsets(top: 5, leading: DesignTokens.Spacing.xl, bottom: 5, trailing: DesignTokens.Spacing.xl))
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("approvals.row.\(item.externalId)")
    }

    private func card(for item: ApprovalRequest, now: Date) -> ApprovalRowCard {
        ApprovalRowCard(
            item: item,
            remaining: ApprovalFormatting.remaining(expiresAt: item.expiresAt, now: now),
            createdText: ApprovalFormatting.requestTime(item.createdAt, now: now),
            area: Self.area
        )
    }
}

private struct ApprovalRowCard: View {
    let item: ApprovalRequest
    let remaining: String
    let createdText: String
    let area: DesignTokens.Area

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text(item.keyName)
                    .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                    .padding(.horizontal, DesignTokens.Spacing.sm)
                    .padding(.vertical, 2)
                    .background(DesignTokens.Color.soft(area))
                    .foregroundStyle(DesignTokens.StatusColor.arrivalGlow)
                    .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
                Spacer()
                Text(remaining)
                    .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                    .monospacedDigit()
                    .foregroundStyle(DesignTokens.Color.accent(area))
            }
            Text(ApprovalFormatting.summary(of: item.context))
                .font(DesignTokens.Typography.font(size: 13.5, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .lineLimit(2)
            Text(item.requesterName.isEmpty ? createdText : "\(item.requesterName) · \(createdText)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
            if item.keyRevoked {
                Text("폐기된 키로 생성됨 — 자동 처리되지 않음")
                    .font(DesignTokens.Typography.font(size: 10.5, weight: .bold))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, DesignTokens.Spacing.md)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(area))
        .overlay(alignment: .leading) {
            Rectangle()
                .fill(DesignTokens.Color.accent(area))
                .frame(width: 3)
        }
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}
