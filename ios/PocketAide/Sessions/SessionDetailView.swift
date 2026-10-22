import DesignSystem
import PocketAideAPI
import SwiftUI

struct SessionDetailView: View {
    @Environment(\.scenePhase) private var scenePhase
    @StateObject private var viewModel: SessionDetailViewModel
    @State private var prompt = ""

    private let area = SessionsView.area
    private static let bottomAnchor = "sessions.detail.output.bottom"

    init(session: AgentSession, api: APIClient?) {
        _viewModel = StateObject(wrappedValue: SessionDetailViewModel(session: session, api: api))
    }

    var body: some View {
        VStack(spacing: 0) {
            statusStrip
            if let error = viewModel.error {
                SessionErrorBanner(error: error, area: area) {
                    Task { await viewModel.tick() }
                }
                .padding(.horizontal, DesignTokens.Spacing.lg)
                .padding(.bottom, DesignTokens.Spacing.sm)
            }
            if viewModel.session.state == .snapshot {
                snapshotPanel
            } else {
                outputArea
            }
            composer
        }
        .background(DesignTokens.Color.surface(area).ignoresSafeArea())
        .navigationTitle(viewModel.session.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar(.visible, for: .navigationBar)
        .task(id: scenePhase) {
            guard scenePhase == .active else { return }
            await viewModel.pollWhileVisible()
        }
        .accessibilityIdentifier("sessions.detail.screen")
    }

    private var statusStrip: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            SessionStateBadge(state: viewModel.session.state, area: area)
            Text(viewModel.session.modelLabel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .lineLimit(1)
            Spacer(minLength: 0)
            if let label = viewModel.phase.label {
                HStack(spacing: 4) {
                    if viewModel.phase == .queued || viewModel.phase == .streaming {
                        ProgressView().controlSize(.mini)
                    }
                    Text(label)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                        .lineLimit(1)
                }
                .accessibilityIdentifier("sessions.detail.phase")
            }
        }
        .padding(.horizontal, DesignTokens.Spacing.lg)
        .padding(.vertical, DesignTokens.Spacing.sm)
    }

    private var outputArea: some View {
        ScrollViewReader { proxy in
            ScrollView {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                    if viewModel.output.isTrimmed {
                        Text("앞부분 출력은 생략했어요.")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.5))
                    }
                    outputBody
                    Color.clear.frame(height: 1).id(Self.bottomAnchor)
                }
                .padding(DesignTokens.Spacing.lg)
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .background(DesignTokens.Color.card(area))
            .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                    .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
            )
            .padding(.horizontal, DesignTokens.Spacing.lg)
            .onChange(of: viewModel.output.nextOffset) { _, _ in
                proxy.scrollTo(Self.bottomAnchor, anchor: .bottom)
            }
        }
    }

    @ViewBuilder
    private var outputBody: some View {
        if !viewModel.hasLoadedOutput {
            ProgressView().frame(maxWidth: .infinity)
        } else if viewModel.output.text.isEmpty {
            Text("아직 출력이 없어요. 아래에 프롬프트를 적어 에이전트에게 일을 맡겨 보세요.")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .accessibilityIdentifier("sessions.detail.output.empty")
        } else {
            Text(viewModel.output.text)
                .font(.system(size: DesignTokens.Typography.bodySm, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityIdentifier("sessions.detail.output")
        }
    }

    private var snapshotPanel: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            Text("스냅샷으로 보관된 세션이에요")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area))
            Text("오래 쓰지 않아 컴퓨트를 반납한 상태예요. 복원하면 새 환경이 뜨고 이전 출력이 그대로 이어져요. 수십 초 걸릴 수 있어요.")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.65))
                .fixedSize(horizontal: false, vertical: true)
            Button {
                Task { await viewModel.restore() }
            } label: {
                HStack(spacing: DesignTokens.Spacing.sm) {
                    if viewModel.isRestoring {
                        ProgressView().tint(DesignTokens.Color.surface(area))
                    }
                    Text(viewModel.isRestoring ? "복원하는 중…" : "복원")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 12)
                .background(Capsule().fill(DesignTokens.Color.ink(area)))
                .foregroundStyle(DesignTokens.Color.surface(area))
            }
            .buttonStyle(.plain)
            .disabled(viewModel.isRestoring)
            .accessibilityIdentifier("sessions.detail.restore.button")
            Spacer(minLength: 0)
        }
        .padding(DesignTokens.Spacing.xl)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }

    private var composer: some View {
        VStack(alignment: .leading, spacing: 4) {
            if promptTooLarge {
                Text("1 MiB를 넘는 프롬프트는 보낼 수 없어요.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.StatusColor.failure)
                    .accessibilityIdentifier("sessions.detail.prompt.toolarge")
            }
            HStack(alignment: .bottom, spacing: DesignTokens.Spacing.sm) {
                TextField("에이전트에게 맡길 일", text: $prompt, axis: .vertical)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                    .lineLimit(1...6)
                    .padding(.horizontal, DesignTokens.Spacing.md)
                    .padding(.vertical, 10)
                    .background(DesignTokens.Color.card(area))
                    .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
                    .overlay(
                        RoundedRectangle(cornerRadius: 18, style: .continuous)
                            .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
                    )
                    .disabled(viewModel.session.state == .snapshot)
                    .accessibilityIdentifier("sessions.detail.prompt.field")
                sendButton
            }
        }
        .padding(.horizontal, DesignTokens.Spacing.lg)
        .padding(.vertical, DesignTokens.Spacing.sm)
    }

    private var sendButton: some View {
        Button {
            let text = trimmedPrompt
            Task {
                if await viewModel.send(text) {
                    prompt = ""
                }
            }
        } label: {
            Group {
                if viewModel.isSending {
                    ProgressView().tint(DesignTokens.Color.surface(area))
                } else {
                    Image(systemName: "arrow.up")
                        .font(.system(size: 16, weight: .bold))
                }
            }
            .frame(width: 40, height: 40)
            .background(Circle().fill(DesignTokens.Color.ink(area)))
            .foregroundStyle(DesignTokens.Color.surface(area))
        }
        .buttonStyle(.plain)
        .disabled(!canSend)
        .opacity(canSend ? 1 : 0.4)
        .accessibilityLabel("보내기")
        .accessibilityIdentifier("sessions.detail.send.button")
    }

    private var trimmedPrompt: String {
        prompt.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private var promptTooLarge: Bool {
        AgentSessionPrompt.exceedsLimit(prompt)
    }

    private var canSend: Bool {
        viewModel.canSend && !trimmedPrompt.isEmpty && !promptTooLarge
    }
}
