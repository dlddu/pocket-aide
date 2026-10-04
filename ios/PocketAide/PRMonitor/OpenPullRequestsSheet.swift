import DesignSystem
import PocketAideAPI
import SwiftUI

struct OpenPullRequestsSheet: View {
    @ObservedObject var viewModel: OpenPullRequestsViewModel
    @Binding var isPresented: Bool

    @State private var tokenInput = ""
    @State private var replacingToken = false

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isConnected && !replacingToken {
                    connectedContent
                } else {
                    connectForm
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .background(DesignTokens.Color.surface(.prMonitor))
            .navigationTitle("열린 PR")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("완료") { isPresented = false }
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                }
            }
        }
        .task {
            if viewModel.isConnected && !viewModel.hasLoaded {
                await viewModel.refresh()
            }
        }
    }

    private var connectForm: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            Text("GitHub 계정을 연결하면 내가 작성했거나 리뷰어로 지정된 열린 PR과 CI 상태를 볼 수 있습니다.")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                .accessibilityIdentifier("openprs.connect.prompt")
            Text("Personal Access Token(classic은 repo 범위)을 붙여 넣으세요. 토큰은 이 기기의 키체인에만 저장되고 GitHub에만 전송됩니다.")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
            SecureField("ghp_…", text: $tokenInput)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .padding(DesignTokens.Spacing.md)
                .background(DesignTokens.Color.card(.prMonitor))
                .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                .accessibilityIdentifier("openprs.token.field")
            if let message = viewModel.connectError {
                Text(message)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .accessibilityIdentifier("openprs.connect.error")
            }
            HStack(spacing: DesignTokens.Spacing.md) {
                Button(connectButtonTitle) {
                    Task { await submitToken() }
                }
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                .disabled(viewModel.isConnecting || tokenInput.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                .accessibilityIdentifier("openprs.connect.button")
                if replacingToken {
                    Button("취소") {
                        replacingToken = false
                        tokenInput = ""
                    }
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.7))
                    .accessibilityIdentifier("openprs.connect.cancel")
                }
            }
            Spacer()
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.lg)
    }

    private var connectButtonTitle: String {
        viewModel.isConnecting ? "확인 중…" : "연결"
    }

    private func submitToken() async {
        if await viewModel.connect(token: tokenInput) {
            tokenInput = ""
            replacingToken = false
        }
    }

    private var connectedContent: some View {
        VStack(alignment: .leading, spacing: 0) {
            if let alert = viewModel.alert {
                alertBanner(alert)
            }
            accountBar
            statusLine
            filterBar
            OpenPullRequestsList(viewModel: viewModel)
        }
    }

    private func alertBanner(_ alert: GitHubAlert) -> some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Image(systemName: "exclamationmark.triangle.fill")
                Text(alert.title)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                    .accessibilityIdentifier("openprs.banner.\(alert.kind)")
            }
            .foregroundStyle(DesignTokens.StatusColor.failure)
            Text(alert.message { $0.formatted(date: .omitted, time: .shortened) })
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.7))
                .fixedSize(horizontal: false, vertical: true)
            Button(alert.actionTitle) {
                if case .rateLimited = alert {
                    Task { await viewModel.refresh() }
                } else {
                    tokenInput = ""
                    replacingToken = true
                }
            }
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
            .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
            .disabled(viewModel.isLoading)
            .accessibilityIdentifier("openprs.banner.action")
        }
        .padding(DesignTokens.Spacing.md)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(.prMonitor))
        .overlay(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .stroke(DesignTokens.StatusColor.failure.opacity(0.4), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
    }

    private var filterBar: some View {
        FilterPills(
            area: .prMonitor,
            options: OpenPullRequestFilter.allCases,
            selection: $viewModel.filter
        ) { option in
            Text(option.label)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                .lineLimit(1)
                .minimumScaleFactor(0.8)
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.bottom, DesignTokens.Spacing.sm)
    }

    private var accountBar: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "person.crop.circle")
                .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
            Text("@\(viewModel.login ?? "")")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                .accessibilityIdentifier("openprs.account.login")
            Spacer()
            Button("토큰 바꾸기") {
                tokenInput = ""
                replacingToken = true
            }
            .accessibilityIdentifier("openprs.token.replace")
            Button("연결 해제", role: .destructive) {
                viewModel.disconnect()
            }
            .accessibilityIdentifier("openprs.disconnect.button")
        }
        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.vertical, DesignTokens.Spacing.md)
    }

    private var statusLine: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
            if let refreshed = viewModel.lastRefreshedAt {
                Text("마지막 갱신 \(refreshed.formatted(date: .omitted, time: .shortened))")
                    .accessibilityIdentifier("openprs.lastRefreshed")
            }
            if let error = viewModel.errorMessage, viewModel.alert == nil, !viewModel.pullRequests.isEmpty {
                Text(error)
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .accessibilityIdentifier("openprs.refresh.error")
            }
        }
        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
        .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.bottom, DesignTokens.Spacing.sm)
    }
}

private struct OpenPullRequestsList: View {
    @ObservedObject var viewModel: OpenPullRequestsViewModel
    @Environment(\.openURL) private var openURL

    var body: some View {
        if viewModel.isLoading && !viewModel.hasLoaded {
            VStack {
                Spacer()
                ProgressView()
                Spacer()
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .accessibilityIdentifier("openprs.loading")
        } else if let error = viewModel.errorMessage, viewModel.pullRequests.isEmpty {
            centered(title: "열린 PR을 불러오지 못했습니다", detail: error)
                .accessibilityIdentifier("openprs.error.state")
        } else if viewModel.pullRequests.isEmpty {
            centered(title: "열려 있는 PR이 없습니다", detail: "작성했거나 리뷰어로 지정된 PR이 열리면 여기에 표시됩니다.")
                .accessibilityIdentifier("openprs.empty.state")
        } else if viewModel.visiblePullRequests.isEmpty {
            centered(
                title: "조건에 맞는 PR이 없습니다",
                detail: "「\(viewModel.filter.label)」 필터를 끄면 열린 PR \(viewModel.pullRequests.count)개를 모두 볼 수 있습니다."
            )
            .accessibilityIdentifier("openprs.filter.empty")
        } else {
            list
        }
    }

    private func centered(title: String, detail: String) -> some View {
        ScrollView {
            VStack(spacing: DesignTokens.Spacing.sm) {
                Text(title)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                Text(detail)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
                    .multilineTextAlignment(.center)
            }
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.top, 120)
            .frame(maxWidth: .infinity)
        }
        .refreshable { await viewModel.refresh() }
    }

    private var list: some View {
        List(viewModel.visiblePullRequests) { pullRequest in
            Button {
                if let url = pullRequest.url { openURL(url) }
            } label: {
                OpenPullRequestRow(pullRequest: pullRequest)
            }
            .buttonStyle(.plain)
            .listRowInsets(EdgeInsets(
                top: DesignTokens.Spacing.xs,
                leading: DesignTokens.Spacing.xl,
                bottom: DesignTokens.Spacing.xs,
                trailing: DesignTokens.Spacing.xl
            ))
            .listRowSeparator(.hidden)
            .listRowBackground(Color.clear)
            .accessibilityIdentifier("openprs.row.\(pullRequest.id)")
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.refresh() }
    }
}

private struct OpenPullRequestRow: View {
    let pullRequest: OpenPullRequest

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Text(pullRequest.repository)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.7))
                Spacer()
                ciBadge
            }
            Text("#\(pullRequest.number) \(pullRequest.title)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                .lineLimit(2)
                .multilineTextAlignment(.leading)
            HStack(spacing: DesignTokens.Spacing.xs) {
                Text(pullRequest.author)
                Text("·")
                Text(pullRequest.role.label)
                Spacer()
                Text(pullRequest.updatedAt.formatted(.relative(presentation: .named)))
            }
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
        }
        .padding(DesignTokens.Spacing.md)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(.prMonitor))
        .overlay(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
    }

    private var ciBadge: some View {
        HStack(spacing: DesignTokens.Spacing.xs) {
            Circle()
                .fill(ciColor)
                .frame(width: 8, height: 8)
            Text(pullRequest.ciStatus.label)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                .foregroundStyle(ciColor)
        }
        .accessibilityElement(children: .combine)
    }

    private var ciColor: Color {
        switch pullRequest.ciStatus {
        case .success: return DesignTokens.StatusColor.success
        case .failure: return DesignTokens.StatusColor.failure
        case .inProgress: return DesignTokens.StatusColor.inProgress
        case .noChecks: return DesignTokens.Color.ink(.prMonitor).opacity(0.4)
        }
    }
}
