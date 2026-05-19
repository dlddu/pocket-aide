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
    @State private var newRepoText = ""
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
        }
        .onAppear { scheduleHighlightClear(for: highlightedEventID) }
        .sheet(isPresented: $showingExcludedSheet) {
            excludedReposSheet
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
            ForEach(viewModel.items) { item in
                PRMonitorHistoryRow(
                    item: item,
                    isHighlighted: highlightedEventID == item.id,
                    onAcknowledge: {
                        Task { await viewModel.acknowledge(id: item.id) }
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
                .accessibilityIdentifier("prmonitor.row.\(item.id)")
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .background(DesignTokens.Color.surface(.prMonitor))
        .refreshable {
            await viewModel.load()
        }
    }

    private var excludedReposSheet: some View {
        NavigationStack {
            VStack(spacing: DesignTokens.Spacing.md) {
                HStack {
                    TextField("owner/repo", text: $newRepoText)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                        .padding(.vertical, DesignTokens.Spacing.sm)
                        .padding(.horizontal, DesignTokens.Spacing.md)
                        .background(DesignTokens.Color.card(.prMonitor))
                        .overlay(
                            RoundedRectangle(cornerRadius: 8, style: .continuous)
                                .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                        )
                        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                        .accessibilityIdentifier("prmonitor.excluded.input")
                    Button {
                        let value = newRepoText.trimmingCharacters(in: .whitespaces)
                        guard !value.isEmpty else { return }
                        let captured = value
                        newRepoText = ""
                        Task { await viewModel.excludeRepo(captured) }
                    } label: {
                        Text("추가")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                            .padding(.vertical, DesignTokens.Spacing.sm)
                            .padding(.horizontal, DesignTokens.Spacing.md)
                            .background(DesignTokens.Color.accent(.prMonitor))
                            .foregroundStyle(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                    }
                    .accessibilityIdentifier("prmonitor.excluded.add.button")
                }
                .padding(.horizontal, DesignTokens.Spacing.xl)

                if let message = viewModel.excludedRepoError {
                    Text(message)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.StatusColor.failure)
                        .padding(.horizontal, DesignTokens.Spacing.xl)
                }

                List {
                    if viewModel.excludedRepos.isEmpty && !viewModel.isLoadingExcluded {
                        Text("제외한 레포가 없습니다")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                            .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
                            .listRowBackground(Color.clear)
                    }
                    ForEach(viewModel.excludedRepos) { repo in
                        Text(repo.repoFullName)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                            .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                            .listRowBackground(DesignTokens.Color.surface(.prMonitor))
                            .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                                Button(role: .destructive) {
                                    Task { await viewModel.removeExcludedRepo(id: repo.id) }
                                } label: {
                                    Label("삭제", systemImage: "trash")
                                }
                            }
                            .accessibilityIdentifier("prmonitor.excluded.row.\(repo.id)")
                    }
                }
                .listStyle(.plain)
                .scrollContentBackground(.hidden)
                .background(DesignTokens.Color.surface(.prMonitor))
            }
            .padding(.top, DesignTokens.Spacing.lg)
            .background(DesignTokens.Color.surface(.prMonitor))
            .navigationTitle("제외한 레포")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("완료") { showingExcludedSheet = false }
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                }
            }
        }
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
