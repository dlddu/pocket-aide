import DesignSystem
import PocketAideAPI
import SwiftUI

struct SessionsView: View {
    static let area: DesignTokens.Area = .prMonitor

    @EnvironmentObject private var auth: AppAuthCoordinator
    @Environment(\.scenePhase) private var scenePhase
    @StateObject private var viewModel = SessionsViewModel(api: nil)
    @State private var path: [AgentSession] = []
    @State private var showsCreate = false

    var body: some View {
        NavigationStack(path: $path) {
            ZStack {
                DesignTokens.Color.surface(Self.area).ignoresSafeArea()
                VStack(spacing: 0) {
                    header
                    list
                }
            }
            .toolbar(.hidden, for: .navigationBar)
            .navigationDestination(for: AgentSession.self) { session in
                SessionDetailView(session: session, type: viewModel.workloadType(of: session), api: viewModel.api ?? auth.api)
            }
            .overlay { createSheet }
            .animation(.easeInOut(duration: 0.18), value: showsCreate)
        }
        .accessibilityIdentifier("sessions.screen")
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .onChange(of: path) { _, newPath in
            if newPath.isEmpty {
                Task { await viewModel.load() }
            }
        }
        .onChange(of: scenePhase) { _, phase in
            if phase == .active, path.isEmpty {
                Task { await viewModel.load() }
            }
        }
    }

    private var header: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 6) {
                AreaLabel(area: Self.area, text: "AGENT · SESSIONS", showsDot: true)
                Text("에이전트 세션")
                    .font(DesignTokens.Typography.font(size: 26, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .accessibilityIdentifier("sessions.title")
            }
            Spacer()
            Button {
                showsCreate = true
                Task { await viewModel.loadConfig() }
            } label: {
                HStack(spacing: DesignTokens.Spacing.xs) {
                    Image(systemName: "plus")
                        .font(.system(size: 13, weight: .bold))
                    Text("새 세션")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                }
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 6)
                .background(DesignTokens.Color.ink(Self.area))
                .clipShape(Capsule())
                .foregroundStyle(DesignTokens.Color.surface(Self.area))
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("sessions.create.button")
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
        .padding(.bottom, DesignTokens.Spacing.md)
    }

    private var list: some View {
        List {
            if let error = viewModel.error {
                SessionErrorBanner(error: error, area: Self.area) {
                    Task { await viewModel.load() }
                }
                .listRowSeparator(.hidden)
                .listRowBackground(Color.clear)
            }
            if viewModel.sessions.isEmpty {
                emptyState
            } else {
                ForEach(viewModel.sessions) { session in
                    NavigationLink(value: session) {
                        SessionRow(session: session, type: viewModel.workloadType(of: session), area: Self.area)
                    }
                    .listRowInsets(EdgeInsets(top: 5, leading: DesignTokens.Spacing.xl, bottom: 5, trailing: DesignTokens.Spacing.xl))
                    .listRowSeparator(.hidden)
                    .listRowBackground(Color.clear)
                    .accessibilityIdentifier("sessions.row.\(session.id)")
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.load() }
    }

    @ViewBuilder
    private var emptyState: some View {
        Group {
            if !viewModel.hasLoaded || (viewModel.isLoading && viewModel.error == nil) {
                ProgressView().frame(maxWidth: .infinity)
            } else if viewModel.error == nil {
                Text("아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                    .accessibilityIdentifier("sessions.empty.state")
            }
        }
        .padding(.vertical, DesignTokens.Spacing.xl)
        .padding(.horizontal, DesignTokens.Spacing.xs)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    @ViewBuilder
    private var createSheet: some View {
        if showsCreate {
            SessionCreateSheet(
                config: viewModel.config,
                isCreating: viewModel.isCreating,
                error: viewModel.createError,
                onCreate: { name, workload, model in
                    Task {
                        if let created = await viewModel.create(name: name, workload: workload, model: model) {
                            showsCreate = false
                            path.append(created)
                        }
                    }
                },
                onCancel: {
                    viewModel.createError = nil
                    showsCreate = false
                }
            )
            .transition(.move(edge: .bottom).combined(with: .opacity))
        }
    }
}

struct SessionRow: View {
    let session: AgentSession
    let type: AgentSessionWorkloadType?
    let area: DesignTokens.Area

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                SessionStateBadge(state: session.state, area: area)
                Spacer(minLength: 0)
                Text("생성 \(AgentSessionFormatting.relative(session.createdAt))")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
            }
            Text(session.name)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .lineLimit(2)
            Text(session.kindLabel(type))
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .medium))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.7))
                .lineLimit(1)
            Text(AgentSessionFormatting.stateDetail(session))
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .lineLimit(1)
                .accessibilityIdentifier("sessions.row.\(session.id).state")
        }
        .padding(DesignTokens.Spacing.lg)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(area))
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
    }

}

struct SessionStateBadge: View {
    let state: AgentSessionState
    let area: DesignTokens.Area

    var body: some View {
        HStack(spacing: 5) {
            Circle()
                .fill(color)
                .frame(width: 6, height: 6)
            Text(state.label)
                .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold))
                .tracking(0.5)
                .foregroundStyle(color)
        }
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("sessions.state.\(state.rawValue)")
    }

    private var color: Color {
        switch state {
        case .active: return DesignTokens.StatusColor.success
        case .idle: return DesignTokens.StatusColor.inProgress
        case .snapshot: return DesignTokens.Color.accent(area)
        }
    }
}

struct SessionErrorBanner: View {
    let error: AgentSessionError
    let area: DesignTokens.Area
    let onRetry: () -> Void

    var body: some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "exclamationmark.circle")
                .foregroundStyle(DesignTokens.StatusColor.failure)
            Text(error.message)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .fixedSize(horizontal: false, vertical: true)
            Spacer(minLength: 0)
            Button("다시 시도", action: onRetry)
                .buttonStyle(.borderless)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.accent(area))
                .accessibilityIdentifier("sessions.error.retry")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.accent(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .accessibilityIdentifier("sessions.error")
    }
}
