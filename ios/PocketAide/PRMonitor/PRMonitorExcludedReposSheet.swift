import DesignSystem
import PocketAideAPI
import SwiftUI

/// "제외 레포" 관리 시트. PRD-10 AC6 — 사용자가 자신의 blacklist에 owner/repo
/// 항목을 추가/삭제한다.
///
/// `PRMonitorView`에서 추출 — PRMonitorView의 type_body가 SwiftLint 250라인
/// 한도를 넘어 시각적으로 독립적인 sheet를 별도 파일로 분리.
struct PRMonitorExcludedReposSheet: View {
    @ObservedObject var viewModel: PRMonitorViewModel
    @Binding var isPresented: Bool
    @State private var newRepoText = ""

    var body: some View {
        NavigationStack {
            VStack(spacing: DesignTokens.Spacing.md) {
                inputRow
                if let message = viewModel.excludedRepoError {
                    Text(message)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.StatusColor.failure)
                        .padding(.horizontal, DesignTokens.Spacing.xl)
                }
                reposList
            }
            .padding(.top, DesignTokens.Spacing.lg)
            .background(DesignTokens.Color.surface(.prMonitor))
            .navigationTitle("제외한 레포")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("완료") { isPresented = false }
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                }
            }
        }
    }

    private var inputRow: some View {
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
    }

    private var reposList: some View {
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
}
