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
            .accessibilityIdentifier("prmonitor.empty.state")
        } else {
            historyList
        }
    }

    private var historyList: some View {
        List {
            ForEach(viewModel.items) { item in
                HistoryRow(
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

// MARK: - HistoryRow

private struct HistoryRow: View {
    let item: NotificationHistoryItem
    let isHighlighted: Bool
    let onAcknowledge: () -> Void

    var body: some View {
        let acked = item.acknowledgedAt != nil
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            header
            titleLine
            metaLine
            actionsRow
        }
        .padding(DesignTokens.Spacing.md)
        .background(rowBackground)
        .overlay(rowBorder)
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(highlightOverlay)
        .opacity(acked ? 0.7 : 1.0)
    }

    @ViewBuilder
    private var header: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            statusBadge
            statusLabel
            Spacer()
            Text(timestampString)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(.secondary)
        }
    }

    @ViewBuilder
    private var statusBadge: some View {
        Circle()
            .fill(statusColor)
            .frame(width: 18, height: 18)
            .overlay(
                Image(systemName: badgeIcon)
                    .font(.system(size: 9, weight: .bold))
                    .foregroundStyle(.white)
            )
    }

    @ViewBuilder
    private var statusLabel: some View {
        if item.acknowledgedAt != nil {
            Text("\(verdictLabel) · 확인됨")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                .foregroundStyle(.secondary)
                .accessibilityIdentifier("prmonitor.row.\(item.id).status")
        } else {
            Text(verdictLabel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                .foregroundStyle(statusColor)
                .accessibilityIdentifier("prmonitor.row.\(item.id).status")
        }
    }

    @ViewBuilder
    private var titleLine: some View {
        Text(titleString)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: item.acknowledgedAt == nil ? .semibold : .medium))
            .foregroundStyle(item.acknowledgedAt == nil
                ? DesignTokens.Color.ink(.prMonitor)
                : DesignTokens.Color.ink(.prMonitor).opacity(0.55))
            .strikethrough(item.acknowledgedAt != nil)
            .accessibilityIdentifier("prmonitor.row.\(item.id).title")
    }

    @ViewBuilder
    private var metaLine: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Text(item.workflowName.isEmpty ? "workflow" : item.workflowName)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(.secondary)
            if !item.headBranch.isEmpty {
                Text("·")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(.tertiary)
                Text(item.headBranch)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(.secondary)
            }
            if let ackedAt = item.acknowledgedAt {
                Spacer()
                Text("확인 \(relativeAck(ackedAt))")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(.secondary)
            }
        }
    }

    @ViewBuilder
    private var actionsRow: some View {
        if item.acknowledgedAt == nil {
            HStack(spacing: DesignTokens.Spacing.sm) {
                if let url = item.prURL {
                    linkChip(label: "PR", url: url, identifier: "prmonitor.row.\(item.id).link.pr")
                }
                if let url = item.commitURL {
                    linkChip(label: "커밋", url: url, identifier: "prmonitor.row.\(item.id).link.commit")
                }
                if let url = item.runURL {
                    linkChip(label: "런", url: url, identifier: "prmonitor.row.\(item.id).link.run")
                }
                Spacer()
                Button {
                    onAcknowledge()
                } label: {
                    Label("확인", systemImage: "checkmark")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                        .padding(.vertical, 6)
                        .padding(.horizontal, DesignTokens.Spacing.md)
                        .background(DesignTokens.Color.accent(.prMonitor))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .accessibilityIdentifier("prmonitor.row.\(item.id).ack.button")
            }
        }
    }

    private func linkChip(label: String, url: String, identifier: String) -> some View {
        let action = Button {
            if let u = URL(string: url) {
                #if canImport(UIKit)
                UIApplication.shared.open(u)
                #endif
            }
        } label: {
            Text("↗ \(label)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                .padding(.vertical, 4)
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                )
                .foregroundStyle(.secondary)
        }
        .accessibilityIdentifier(identifier)
        return action
    }

    private var rowBackground: some View {
        item.acknowledgedAt == nil
            ? DesignTokens.Color.card(.prMonitor)
            : DesignTokens.Color.surface(.prMonitor).opacity(0.6)
    }

    @ViewBuilder
    private var rowBorder: some View {
        if item.acknowledgedAt == nil {
            // Left accent stripe + thin rule on remaining sides for unacked.
            ZStack(alignment: .leading) {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                Rectangle()
                    .fill(DesignTokens.Color.accent(.prMonitor))
                    .frame(width: 3)
                    .clipShape(
                        UnevenRoundedRectangle(cornerRadii: .init(
                            topLeading: 12, bottomLeading: 12,
                            bottomTrailing: 0, topTrailing: 0
                        ), style: .continuous)
                    )
            }
        } else {
            // Dashed border for acked.
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(
                    DesignTokens.Color.rule(.prMonitor),
                    style: StrokeStyle(lineWidth: 1, dash: [4, 3])
                )
        }
    }

    @ViewBuilder
    private var highlightOverlay: some View {
        if isHighlighted {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .stroke(DesignTokens.StatusColor.arrivalGlow.opacity(0.55), lineWidth: 3)
                .blur(radius: 1.5)
        }
    }

    private var statusColor: Color {
        switch item.conclusion.lowercased() {
        case "success": return DesignTokens.StatusColor.success
        case "failure", "timed_out", "cancelled", "action_required": return DesignTokens.StatusColor.failure
        default: return DesignTokens.StatusColor.failure
        }
    }

    private var badgeIcon: String {
        switch item.conclusion.lowercased() {
        case "success": return "checkmark"
        default: return "xmark"
        }
    }

    private var verdictLabel: String {
        switch item.conclusion.lowercased() {
        case "success": return "CI 통과"
        case "failure": return "CI 실패"
        case "cancelled": return "CI 취소"
        case "timed_out": return "CI 타임아웃"
        default: return "CI \(item.conclusion)"
        }
    }

    private var titleString: String {
        if let number = item.prNumber, let title = item.prTitle, !title.isEmpty {
            return "\(item.repoFullName) · #\(number) \(title)"
        }
        if let number = item.prNumber {
            return "\(item.repoFullName) · #\(number)"
        }
        // No PR linked — fallback (AC6 PR-less case).
        if item.headBranch.isEmpty {
            return item.repoFullName
        }
        return "\(item.repoFullName) · \(item.headBranch)"
    }

    private var timestampString: String {
        Self.timeFormatter.string(from: Date(timeIntervalSince1970: TimeInterval(item.createdAt)))
    }

    private func relativeAck(_ epoch: Int64) -> String {
        let date = Date(timeIntervalSince1970: TimeInterval(epoch))
        return Self.relativeFormatter.localizedString(for: date, relativeTo: Date())
    }

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "HH:mm"
        return f
    }()

    private static let relativeFormatter = RelativeDateTimeFormatter()
}
