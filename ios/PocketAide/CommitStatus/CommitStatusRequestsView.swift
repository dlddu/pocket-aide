import DesignSystem
import PocketAideAPI
import SwiftUI

struct CommitStatusRequestsView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Environment(\.scenePhase) private var scenePhase
    @ObservedObject var viewModel: CommitStatusRequestsViewModel
    @Binding var openedRequestID: String?
    @State private var path: [CommitStatusRequest] = []

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
            .navigationTitle("커밋 상태")
            .toolbar(.hidden, for: .navigationBar)
            .navigationDestination(for: CommitStatusRequest.self) { item in
                CommitStatusDetailView(item: item, api: viewModel.api ?? auth.api)
                    .id(item.id)
            }
        }
        .onChange(of: path) { _, newPath in
            if newPath.isEmpty {
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
        .onChange(of: openedRequestID, initial: true) { _, id in
            guard let id else { return }
            openedRequestID = nil
            Task { await open(id) }
        }
    }

    private func open(_ id: String) async {
        guard let api = viewModel.api ?? auth.api, let item = try? await api.commitStatusRequest(id: id) else { return }
        path = [item]
    }

    private var header: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(DesignTokens.Color.accent(Self.area))
                        .frame(width: 8, height: 8)
                    Text("COMMIT · STATUS")
                        .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                        .tracking(2.4)
                        .foregroundStyle(DesignTokens.Color.accent(Self.area))
                }
                Text("커밋 상태")
                    .font(DesignTokens.Typography.font(size: 26, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .accessibilityIdentifier("commitStatus.title")
            }
            Spacer()
            VStack(spacing: 2) {
                Text("\(viewModel.pendingCount)")
                    .font(.system(size: 15, weight: .bold, design: .monospaced))
                    .accessibilityIdentifier("commitStatus.pending.count")
                Text("대기")
                    .font(.system(size: 8, weight: .bold, design: .monospaced))
            }
            .foregroundStyle(.white)
            .padding(.horizontal, 10)
            .padding(.vertical, DesignTokens.Spacing.xs)
            .background(DesignTokens.Color.accent(Self.area))
            .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
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
                sectionHeader(CommitStatusFormatting.pendingHeader(viewModel.pendingCount), id: "pending")
                ForEach(viewModel.items) { item in
                    row(for: item)
                }
            }
            if !viewModel.history.isEmpty {
                sectionHeader(CommitStatusFormatting.historyHeader, id: "history")
                ForEach(viewModel.history) { item in
                    row(for: item)
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.load() }
    }

    private var emptyState: some View {
        VStack(spacing: DesignTokens.Spacing.sm) {
            if viewModel.isLoading || (!viewModel.hasLoaded && viewModel.errorMessage == nil) {
                ProgressView()
            } else if viewModel.errorMessage == nil {
                Image(systemName: "checkmark.circle")
                    .font(.system(size: 28))
                    .foregroundStyle(DesignTokens.Color.accent(Self.area))
                Text("기록을 기다리는 커밋 상태가 없습니다")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                Text("에이전트가 커밋 상태를 요청하면 여기에 쌓이고 알림이 옵니다.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                    .multilineTextAlignment(.center)
            }
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, DesignTokens.Spacing.xxl)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .accessibilityIdentifier("commitStatus.empty")
    }

    private func errorBanner(_ message: String) -> some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "exclamationmark.circle")
                .foregroundStyle(DesignTokens.StatusColor.failure)
            VStack(alignment: .leading, spacing: 2) {
                Text("커밋 상태 요청을 불러오지 못했습니다")
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
            .accessibilityIdentifier("commitStatus.error.retry")
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
    }

    private func sectionHeader(_ title: String, id: String) -> some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Rectangle()
                .fill(DesignTokens.Color.accent(Self.area))
                .frame(width: 4, height: 12)
            Text(title)
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
                .accessibilityIdentifier("commitStatus.section.\(id)")
            Rectangle()
                .fill(DesignTokens.Color.rule(Self.area))
                .frame(height: 1)
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private func row(for item: CommitStatusRequest) -> some View {
        Button {
            path.append(item)
        } label: {
            CommitStatusRowCard(item: item, area: Self.area)
        }
        .buttonStyle(.plain)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .listRowInsets(EdgeInsets(top: 5, leading: DesignTokens.Spacing.xl, bottom: 5, trailing: DesignTokens.Spacing.xl))
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("commitStatus.row.\(item.id)")
    }
}

private struct CommitStatusRowCard: View {
    let item: CommitStatusRequest
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
                Text(item.status == .pending
                     ? ApprovalFormatting.requestTime(item.createdAt)
                     : CommitStatusFormatting.statusLabel(item.status))
                    .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                    .foregroundStyle(DesignTokens.Color.accent(area))
            }
            Text(CommitStatusFormatting.title(item))
                .font(DesignTokens.Typography.font(size: 13.5, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .lineLimit(2)
            Text(CommitStatusFormatting.target(item))
                .font(.system(size: DesignTokens.Typography.captionXs, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(area))
            Text(item.requester)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .lineLimit(1)
            if item.status == .pending, !item.lastError.isEmpty {
                Text(CommitStatusFormatting.writeFailedNotice(item.lastError))
                    .font(DesignTokens.Typography.font(size: 10.5, weight: .bold))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .lineLimit(2)
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, DesignTokens.Spacing.md)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(area))
        .overlay(alignment: .leading) {
            Rectangle()
                .fill(item.status == .pending ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area))
                .frame(width: 3)
        }
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .opacity(item.status == .pending ? 1 : 0.7)
    }
}
