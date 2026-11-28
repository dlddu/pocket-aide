import DesignSystem
import PocketAideAPI
import SwiftUI

// docs/mockups/screen-chat-text.html — header, bubbles, composer. The header's
// three actions sync the conversation on demand, refresh plugin skills (one
// conversation, no "new chat") and archive the session answering it, and the
// composer's round button sends (voice mode is not built yet).
struct ChatTab: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Environment(\.scenePhase) private var scenePhase
    @StateObject private var viewModel = ChatViewModel(api: nil)
    @State private var draft = ""
    @State private var followsLatest = true
    // The session the confirmation was drawn for. Held rather than read again
    // when the person confirms: drawing the confirmation and tapping it are two
    // steps, and if another session took over the conversation in between the
    // server must refuse rather than archive the one that took over (AC11).
    @State private var archiveTarget: String?

    private let area: DesignTokens.Area = .aiChat
    private static let bottomAnchor = "chat.messages.bottom"
    private static let tail: CGFloat = 6
    private static let avatarSize: CGFloat = 28
    private static let headerButtonSize: CGFloat = 36
    private static let sendSize: CGFloat = 44
    private static let mutedOpacity = 0.55
    // Bubbles stop short of the trailing edge (patterns.md §2: 78–82% wide).
    private static let bubbleInset: CGFloat = 56
    private static let followSlack: CGFloat = 48

    var body: some View {
        VStack(spacing: 0) {
            ScreenHeader(
                area: area,
                title: (viewModel.conversation?.title).flatMap { $0.isEmpty ? nil : $0 } ?? "새 대화",
                subtitle: "\(viewModel.messages.count)개의 메시지"
            ) {
                HStack(spacing: DesignTokens.Spacing.sm) {
                    syncButton
                    refreshButton
                    archiveButton
                }
            }
            Rectangle().fill(DesignTokens.Color.rule(area)).frame(height: 1)
            if let notice = viewModel.notice {
                banner(notice, color: DesignTokens.Color.ink(area), identifier: "chat.notice")
                    .onTapGesture { viewModel.notice = nil }
            }
            if let error = viewModel.error {
                banner(error.message, color: DesignTokens.StatusColor.failure, identifier: "chat.error")
            }
            // Its own banner, dismissable: an archive that did not happen is
            // about the action the person just took, not about the conversation
            // being broken — and 「아카이브 꺼짐」 is not a failure at all
            // (PRD-11 AC8).
            if let archiveError = viewModel.archiveError {
                banner(
                    archiveError.message,
                    color: archiveError == .disabled ? DesignTokens.Color.ink(area) : DesignTokens.StatusColor.failure,
                    identifier: "chat.archive.error"
                )
                .onTapGesture { viewModel.archiveError = nil }
            }
            messageList
            composer
        }
        .background(DesignTokens.Color.surface(area).ignoresSafeArea())
        // Asks before handing the environment back (PRD-1 AC11). It says what
        // goes and what stays: the environment is returned, and the
        // conversation is not — the next message restores this very session and
        // carries on (AC4). It is not the sessions tab's delete, so it does not
        // borrow its words about something that cannot be undone.
        .confirmationDialog(
            "이 대화를 맡은 세션을 아카이브할까요?",
            isPresented: Binding(get: { archiveTarget != nil }, set: { if !$0 { archiveTarget = nil } }),
            titleVisibility: .visible
        ) {
            Button("아카이브") {
                guard let target = archiveTarget else { return }
                archiveTarget = nil
                Task { await viewModel.archive(sessionID: target) }
            }
            .accessibilityIdentifier("chat.archive.confirm")
            Button("취소", role: .cancel) { archiveTarget = nil }
                .accessibilityIdentifier("chat.archive.cancel")
        } message: {
            Text("환경을 반납하고 작업 내용은 보관해요. 대화 기록은 그대로이고, 다음 메시지를 보내면 이 세션이 복원되어 이어져요.")
        }
        .task(id: scenePhase) {
            viewModel.replaceAPI(auth.api)
            guard scenePhase == .active else { return }
            await viewModel.pollWhileVisible()
        }
    }

    private var syncButton: some View {
        headerIconButton(
            systemImage: "arrow.triangle.2.circlepath",
            isLoading: viewModel.isSyncing,
            enabled: viewModel.canSync,
            accessibilityLabel: "대화 동기화",
            accessibilityIdentifier: "chat.sync.button"
        ) {
            Task { await viewModel.sync() }
        }
    }

    private var refreshButton: some View {
        headerIconButton(
            systemImage: "arrow.clockwise",
            isLoading: viewModel.isRefreshing,
            enabled: viewModel.canRefresh,
            accessibilityLabel: "스킬 새로고침",
            accessibilityIdentifier: "chat.refresh.button"
        ) {
            Task { await viewModel.refreshPlugins() }
        }
    }

    // Disabled until a session answers this conversation (before the first
    // message there is nothing to hand back) and while the server would refuse
    // anyway. A tap only opens the confirmation, and the loading state covers
    // the archive itself — so repeated taps while one is in flight send
    // nothing (PRD-1 AC11).
    private var archiveButton: some View {
        headerIconButton(
            systemImage: "archivebox",
            isLoading: viewModel.isArchiving,
            enabled: viewModel.canArchive,
            accessibilityLabel: "이 대화를 맡은 세션 아카이브",
            accessibilityIdentifier: "chat.archive.button"
        ) {
            let session = viewModel.sessionID
            guard !session.isEmpty else { return }
            archiveTarget = session
        }
    }

    private func headerIconButton(
        systemImage: String,
        isLoading: Bool,
        enabled: Bool,
        accessibilityLabel: String,
        accessibilityIdentifier: String,
        action: @escaping () -> Void
    ) -> some View {
        Button(action: action) {
            Group {
                if isLoading {
                    ProgressView().tint(DesignTokens.Color.ink(area))
                } else {
                    Image(systemName: systemImage)
                        .font(.system(size: 14, weight: .semibold))
                }
            }
            .frame(width: Self.headerButtonSize, height: Self.headerButtonSize)
            .overlay(Circle().stroke(DesignTokens.Color.rule(area), lineWidth: 1))
            .foregroundStyle(DesignTokens.Color.ink(area))
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
        .opacity(enabled ? 1 : 0.4)
        .accessibilityLabel(accessibilityLabel)
        .accessibilityIdentifier(accessibilityIdentifier)
    }

    private func banner(_ text: String, color: Color, identifier: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
            .foregroundStyle(color)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.vertical, DesignTokens.Spacing.sm)
            .accessibilityIdentifier(identifier)
    }

    @ViewBuilder
    private var messageList: some View {
        if viewModel.hasLoaded && viewModel.messages.isEmpty && !viewModel.isReplying {
            Text("무엇이든 물어보세요. 첫 메시지를 보내면 ChatGPT 로그인 안내가 이 대화에 나와요.")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                .padding(DesignTokens.Spacing.xl)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .accessibilityIdentifier("chat.empty")
        } else {
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
                        ForEach(viewModel.messages) { message in
                            if message.isMine {
                                myBubble(message.content)
                            } else {
                                aiBubble(message.content, pending: false)
                            }
                        }
                        if let pending = viewModel.pending {
                            aiBubble(pending.content, pending: true)
                        }
                        if let queued = viewModel.queued {
                            queuedBubble(queued.content)
                        }
                        Color.clear.frame(height: 1).id(Self.bottomAnchor)
                    }
                    .padding(.horizontal, DesignTokens.Spacing.xl)
                    .padding(.vertical, DesignTokens.Spacing.lg)
                }
                .onScrollPhaseChange { _, phase, context in
                    if phase == .idle {
                        guard !followsLatest else { return }
                        followsLatest = Self.isAtEnd(
                            visibleMaxY: context.geometry.visibleRect.maxY,
                            contentHeight: context.geometry.contentSize.height
                        )
                    } else if phase != .animating {
                        followsLatest = false
                    }
                }
                .onChange(of: viewModel.messages.count) { _, _ in follow(proxy) }
                .onChange(of: viewModel.pending) { _, _ in follow(proxy) }
                // A message held for after the reply is the newest thing on
                // screen, so following the end (AC9) covers it too.
                .onChange(of: viewModel.queued) { _, _ in follow(proxy) }
                .onAppear { follow(proxy) }
            }
        }
    }

    /// Keeps the end of a streaming reply in view — instantly, because a reply
    /// changes every poll and an animation restarts on each change instead of
    /// arriving. A drag hands the position over until a gesture settles back
    /// at the end.
    private func follow(_ proxy: ScrollViewProxy) {
        guard followsLatest else { return }
        proxy.scrollTo(Self.bottomAnchor, anchor: .bottom)
    }

    private static func isAtEnd(visibleMaxY: CGFloat, contentHeight: CGFloat) -> Bool {
        visibleMaxY >= contentHeight - followSlack
    }

    /// Web addresses become links the system opens in the browser.
    private func linked(_ text: String, color: Color) -> AttributedString {
        var attributed = AttributedString(text)
        for range in chatLinkRanges(in: text) {
            guard let url = URL(string: String(text[range])),
                  let lower = AttributedString.Index(range.lowerBound, within: attributed),
                  let upper = AttributedString.Index(range.upperBound, within: attributed) else { continue }
            attributed[lower..<upper].link = url
            attributed[lower..<upper].foregroundColor = color
            attributed[lower..<upper].underlineStyle = .single
        }
        return attributed
    }

    private func myBubble(_ text: String) -> some View {
        Text(linked(text, color: DesignTokens.Color.card(area)))
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
            .foregroundStyle(DesignTokens.Color.card(area))
            .textSelection(.enabled)
            .padding(.horizontal, DesignTokens.Spacing.lg)
            .padding(.vertical, 10)
            .background(
                UnevenRoundedRectangle(
                    topLeadingRadius: DesignTokens.Radius.bubble,
                    bottomLeadingRadius: DesignTokens.Radius.bubble,
                    bottomTrailingRadius: Self.tail,
                    topTrailingRadius: DesignTokens.Radius.bubble,
                    style: .continuous
                )
                .fill(DesignTokens.Color.ink(area))
            )
            .fixedSize(horizontal: false, vertical: true)
            .padding(.trailing, Self.bubbleInset)
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityIdentifier("chat.message.mine")
    }

    /// A message held for after the reply in progress (PRD-1 AC1, 2026-11-09
    /// 개정) — dimmer than a sent bubble, with a cancel button while it waits.
    private func queuedBubble(_ text: String) -> some View {
        VStack(alignment: .trailing, spacing: DesignTokens.Spacing.xs) {
            Text(text)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                .foregroundStyle(DesignTokens.Color.card(area))
                .textSelection(.enabled)
                .padding(.horizontal, DesignTokens.Spacing.lg)
                .padding(.vertical, 10)
                .background(
                    UnevenRoundedRectangle(
                        topLeadingRadius: DesignTokens.Radius.bubble,
                        bottomLeadingRadius: DesignTokens.Radius.bubble,
                        bottomTrailingRadius: Self.tail,
                        topTrailingRadius: DesignTokens.Radius.bubble,
                        style: .continuous
                    )
                    .fill(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                )
                .fixedSize(horizontal: false, vertical: true)
            HStack(spacing: DesignTokens.Spacing.xs) {
                Text("대기 중")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                Button {
                    Task { await viewModel.cancelQueued() }
                } label: {
                    Text("취소")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                        .padding(.horizontal, DesignTokens.Spacing.md)
                        .padding(.vertical, DesignTokens.Spacing.xs)
                        .background(Capsule().fill(DesignTokens.Color.card(area)))
                        .overlay(Capsule().stroke(DesignTokens.Color.rule(area), lineWidth: 1))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("chat.queued.cancel")
            }
        }
        .padding(.leading, Self.bubbleInset)
        .frame(maxWidth: .infinity, alignment: .trailing)
        .accessibilityIdentifier("chat.message.queued")
    }

    private func aiBubble(_ text: String, pending: Bool) -> some View {
        let shape = UnevenRoundedRectangle(
            topLeadingRadius: DesignTokens.Radius.bubble,
            bottomLeadingRadius: Self.tail,
            bottomTrailingRadius: DesignTokens.Radius.bubble,
            topTrailingRadius: DesignTokens.Radius.bubble,
            style: .continuous
        )
        return HStack(alignment: .bottom, spacing: DesignTokens.Spacing.sm) {
            Text("PA")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))
                .foregroundStyle(DesignTokens.Color.accent(area))
                .frame(width: Self.avatarSize, height: Self.avatarSize)
                .background(Circle().fill(DesignTokens.Color.accent(area).opacity(0.15)))
                .padding(.bottom, DesignTokens.Spacing.xs)
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
                if !trimmed.isEmpty {
                    Text(linked(trimmed, color: DesignTokens.Color.accent(area)))
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                        .foregroundStyle(DesignTokens.Color.ink(area))
                        .textSelection(.enabled)
                }
                if pending {
                    HStack(spacing: DesignTokens.Spacing.xs) {
                        Text("입력 중")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                            .italic()
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                        ForEach(0..<3, id: \.self) { _ in
                            Circle()
                                .fill(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                                .frame(width: 6, height: 6)
                        }
                    }
                } else {
                    Rectangle().fill(DesignTokens.Color.rule(area)).frame(height: 1)
                    Button {
                        UIPasteboard.general.string = text
                    } label: {
                        Text("복사")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                            .padding(.horizontal, DesignTokens.Spacing.md)
                            .padding(.vertical, DesignTokens.Spacing.xs)
                            .background(Capsule().fill(DesignTokens.Color.card(area)))
                            .overlay(Capsule().stroke(DesignTokens.Color.rule(area), lineWidth: 1))
                    }
                    .buttonStyle(.plain)
                    .accessibilityIdentifier("chat.message.copy")
                }
            }
            .padding(.horizontal, DesignTokens.Spacing.lg)
            .padding(.vertical, DesignTokens.Spacing.md)
            .background(shape.fill(DesignTokens.Color.card(area)))
            .overlay(shape.stroke(DesignTokens.Color.rule(area), lineWidth: 1))
            .fixedSize(horizontal: false, vertical: true)
        }
        .padding(.trailing, Self.bubbleInset)
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityIdentifier(pending ? "chat.message.pending" : "chat.message.ai")
    }

    private var composerHint: String {
        if viewModel.queued != nil {
            return "대기 중인 메시지가 있어요 — 답이 끝나면 보낼게요"
        }
        if viewModel.isReplying {
            return "답을 쓰는 중이에요 — 보내면 끝난 뒤 바로 전달돼요"
        }
        return "ChatGPT 계정의 Codex가 답해요"
    }

    private var composer: some View {
        let canSend = viewModel.canSend && !draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        return VStack(spacing: DesignTokens.Spacing.xs) {
            HStack(alignment: .bottom, spacing: DesignTokens.Spacing.sm) {
                TextField("메시지를 입력하세요", text: $draft, axis: .vertical)
                    .textInputAutocapitalization(.sentences)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                    .lineLimit(1...6)
                    .padding(.horizontal, DesignTokens.Spacing.lg)
                    .padding(.vertical, 10)
                    .background(
                        RoundedRectangle(cornerRadius: DesignTokens.Radius.cardLarge, style: .continuous)
                            .fill(DesignTokens.Color.card(area))
                    )
                    .overlay(
                        RoundedRectangle(cornerRadius: DesignTokens.Radius.cardLarge, style: .continuous)
                            .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
                    )
                    .accessibilityIdentifier("chat.composer.field")
                Button {
                    let text = draft
                    Task {
                        if await viewModel.send(text) { draft = "" }
                    }
                } label: {
                    Group {
                        if viewModel.isSending {
                            ProgressView().tint(DesignTokens.Color.card(area))
                        } else {
                            Image(systemName: "arrow.up")
                                .font(.system(size: 16, weight: .bold))
                        }
                    }
                    .frame(width: Self.sendSize, height: Self.sendSize)
                    .background(Circle().fill(DesignTokens.Color.accent(area)))
                    .foregroundStyle(DesignTokens.Color.card(area))
                }
                .buttonStyle(.plain)
                .disabled(!canSend)
                .opacity(canSend ? 1 : 0.4)
                .accessibilityLabel("보내기")
                .accessibilityIdentifier("chat.composer.send")
            }
            Text(composerHint)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                .accessibilityIdentifier("chat.composer.hint")
        }
        .padding(.horizontal, DesignTokens.Spacing.lg)
        .padding(.vertical, DesignTokens.Spacing.sm)
        .background(DesignTokens.Color.surface(area))
    }
}
