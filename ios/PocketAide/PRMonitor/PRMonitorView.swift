import DesignSystem
import PocketAideAPI
import SwiftUI

struct PRMonitorView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @StateObject private var viewModel: PRMonitorViewModel

    /// Set by the parent when a push tap opens this tab via deep-link.
    /// Cleared after `arrivalHighlightDuration` seconds so the glow doesn't
    /// linger forever (AC7 — routing only, no acknowledgement).
    @Binding var highlightedEventID: Int64?

    @State private var showingExcludedSheet = false
    @State private var arrivalClearTask: Task<Void, Never>?

    private let arrivalHighlightDuration: TimeInterval = 5

    init(highlightedEventID: Binding<Int64?> = .constant(nil)) {
        _highlightedEventID = highlightedEventID
        _viewModel = StateObject(wrappedValue: PRMonitorViewModel(api: nil))
    }

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(.prMonitor).ignoresSafeArea()
            VStack(spacing: 0) {
                ScreenHeader(area: .prMonitor, title: "PR 모니터") {
                    HStack(spacing: DesignTokens.Spacing.sm) {
                        if viewModel.totalUnacknowledgedCount > 0 {
                            unreadBadge
                        }
                        Button {
                            showingExcludedSheet = true
                            Task { await viewModel.loadExcludedRepos() }
                        } label: {
                            Image(systemName: "line.3.horizontal.decrease")
                                .font(.system(size: 14, weight: .bold))
                                .frame(width: 36, height: 36)
                                .background(DesignTokens.Color.card(.prMonitor))
                                .overlay(
                                    RoundedRectangle(cornerRadius: 8, style: .continuous)
                                        .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                                )
                                .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                                .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                        }
                        .accessibilityIdentifier("prmonitor.excluded.button")
                    }
                }

                content
            }
        }
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .onChange(of: highlightedEventID) { _, newValue in
            scheduleHighlightClear(for: newValue)
            // Foreground push tap: highlight is set but viewModel.items may be
            // stale and not yet contain the newly-arrived event. Re-fetch so
            // the matching card actually exists in the list and can light up.
            if newValue != nil {
                Task { await viewModel.load() }
            }
        }
        .onAppear { scheduleHighlightClear(for: highlightedEventID) }
        .sheet(isPresented: $showingExcludedSheet) {
            PRMonitorExcludedReposSheet(
                viewModel: viewModel,
                isPresented: $showingExcludedSheet
            )
        }
    }

    @ViewBuilder
    private var content: some View {
        if viewModel.isLoading && viewModel.items.isEmpty {
            VStack {
                Spacer()
                ProgressView()
                Spacer()
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else if let error = viewModel.errorMessage, viewModel.items.isEmpty {
            VStack(spacing: DesignTokens.Spacing.sm) {
                Spacer()
                Text("이력을 불러오지 못했습니다")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                Text(error)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, DesignTokens.Spacing.xl)
                Button("다시 시도") {
                    Task { await viewModel.load() }
                }
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                Spacer()
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .accessibilityIdentifier("prmonitor.error.state")
        } else if viewModel.items.isEmpty {
            VStack(spacing: DesignTokens.Spacing.sm) {
                Spacer()
                Text("아직 도착한 알림이 없습니다")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                Text("CI가 완료되면 여기에 표시됩니다.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
                Spacer()
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .accessibilityIdentifier("prmonitor.empty.state")
        } else {
            historyList
        }
    }

    private var historyList: some View {
        List {
            let unread = viewModel.unacknowledgedGroups
            let read = viewModel.acknowledgedGroups
            if !unread.isEmpty {
                sectionHeader(
                    title: "미확인 · \(unread.count)개 그룹",
                    accentColor: DesignTokens.Color.accent(.prMonitor),
                    titleColor: DesignTokens.Color.accent(.prMonitor)
                )
                ForEach(unread) { group in
                    groupCard(group)
                }
            }
            if !read.isEmpty {
                sectionHeader(
                    title: "확인 완료",
                    accentColor: DesignTokens.Color.rule(.prMonitor),
                    titleColor: DesignTokens.Color.ink(.prMonitor).opacity(0.55)
                )
                ForEach(read) { group in
                    groupCard(group)
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .background(DesignTokens.Color.surface(.prMonitor))
        .refreshable {
            await viewModel.load()
        }
    }

    @ViewBuilder
    private func groupCard(_ group: HistoryGroup) -> some View {
        PRMonitorGroupCard(
            group: group,
            highlightedEventID: highlightedEventID,
            onAcknowledge: { id in
                Task { await viewModel.acknowledge(id: id) }
            },
            onAcknowledgeGroup: {
                Task { await viewModel.acknowledgeGroup(group) }
            }
        )
        .listRowInsets(EdgeInsets(
            top: DesignTokens.Spacing.xs,
            leading: DesignTokens.Spacing.xl,
            bottom: DesignTokens.Spacing.xs,
            trailing: DesignTokens.Spacing.xl
        ))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .accessibilityIdentifier("prmonitor.group.\(group.id)")
    }

    @ViewBuilder
    private func sectionHeader(title: String, accentColor: Color, titleColor: Color) -> some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Rectangle()
                .fill(accentColor)
                .frame(width: 3, height: 12)
            Text(title)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                .foregroundStyle(titleColor)
                .textCase(.uppercase)
            Rectangle()
                .fill(DesignTokens.Color.rule(.prMonitor))
                .frame(height: 1)
        }
        .listRowInsets(EdgeInsets(
            top: DesignTokens.Spacing.sm,
            leading: DesignTokens.Spacing.xl,
            bottom: DesignTokens.Spacing.xs,
            trailing: DesignTokens.Spacing.xl
        ))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    @ViewBuilder
    private var unreadBadge: some View {
        VStack(spacing: 2) {
            Text("\(viewModel.totalUnacknowledgedCount)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                .foregroundStyle(.white)
            Text("미확인")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))
                .foregroundStyle(.white)
                .textCase(.uppercase)
        }
        .padding(.vertical, 4)
        .padding(.horizontal, DesignTokens.Spacing.sm)
        .background(DesignTokens.Color.accent(.prMonitor))
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .accessibilityIdentifier("prmonitor.unread.badge")
    }

    /// Schedule the auto-clear of the push-arrival highlight. Calls cancel
    /// on any previous task so a quick second push doesn't fight an
    /// in-flight clear.
    private func scheduleHighlightClear(for id: Int64?) {
        arrivalClearTask?.cancel()
        guard id != nil else { return }
        let duration = arrivalHighlightDuration
        arrivalClearTask = Task { @MainActor in
            try? await Task.sleep(nanoseconds: UInt64(duration * 1_000_000_000))
            if Task.isCancelled { return }
            if highlightedEventID == id {
                highlightedEventID = nil
            }
        }
    }
}
