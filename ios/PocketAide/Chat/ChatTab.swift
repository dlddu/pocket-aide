import DesignSystem
import PocketAideAPI
import SwiftUI

// docs/mockups/screen-chat-text.html — header, bubbles, composer. The header's
// action refreshes plugin skills (one conversation, no "new chat"), and the
// composer's round button sends (voice mode is not built yet).
struct ChatTab: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Environment(\.scenePhase) private var scenePhase
    @StateObject private var viewModel = ChatViewModel(api: nil)
    @State private var draft = ""

    private let area: DesignTokens.Area = .aiChat
    private static let bottomAnchor = "chat.messages.bottom"
    private static let tail: CGFloat = 6
    private static let avatarSize: CGFloat = 28
    private static let headerButtonSize: CGFloat = 36
    private static let sendSize: CGFloat = 44
    private static let mutedOpacity = 0.55
    // Bubbles stop short of the trailing edge (patterns.md §2: 78–82% wide).
    private static let bubbleInset: CGFloat = 56

    var body: some View {
        VStack(spacing: 0) {
            ScreenHeader(
                area: area,
                title: (viewModel.conversation?.title).flatMap { $0.isEmpty ? nil : $0 } ?? "새 대화",
                subtitle: "\(viewModel.messages.count)개의 메시지"
            ) {
                refreshButton
            }
            Rectangle().fill(DesignTokens.Color.rule(area)).frame(height: 1)
            if let notice = viewModel.notice {
                banner(notice, color: DesignTokens.Color.ink(area), identifier: "chat.notice")
                    .onTapGesture { viewModel.notice = nil }
            }
            if let error = viewModel.error {
                banner(error.message, color: DesignTokens.StatusColor.failure, identifier: "chat.error")
            }
            messageList
            composer
        }
        .background(DesignTokens.Color.surface(area).ignoresSafeArea())
        .task(id: scenePhase) {
            viewModel.replaceAPI(auth.api)
            guard scenePhase == .active else { return }
            await viewModel.pollWhileVisible()
        }
    }

    private var refreshButton: some View {
        Button {
            Task { await viewModel.refreshPlugins() }
        } label: {
            Group {
                if viewModel.isRefreshing {
                    ProgressView().tint(DesignTokens.Color.ink(area))
                } else {
                    Image(systemName: "arrow.clockwise")
                        .font(.system(size: 14, weight: .semibold))
                }
            }
            .frame(width: Self.headerButtonSize, height: Self.headerButtonSize)
            .overlay(Circle().stroke(DesignTokens.Color.rule(area), lineWidth: 1))
            .foregroundStyle(DesignTokens.Color.ink(area))
        }
        .buttonStyle(.plain)
        .disabled(!viewModel.canRefresh)
        .opacity(viewModel.canRefresh ? 1 : 0.4)
        .accessibilityLabel("스킬 새로고침")
        .accessibilityIdentifier("chat.refresh.button")
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
                        Color.clear.frame(height: 1).id(Self.bottomAnchor)
                    }
                    .padding(.horizontal, DesignTokens.Spacing.xl)
                    .padding(.vertical, DesignTokens.Spacing.lg)
                }
                .onChange(of: viewModel.messages.count) { _, _ in
                    withAnimation { proxy.scrollTo(Self.bottomAnchor, anchor: .bottom) }
                }
                .onChange(of: viewModel.pending) { _, _ in
                    withAnimation { proxy.scrollTo(Self.bottomAnchor, anchor: .bottom) }
                }
                .onAppear { proxy.scrollTo(Self.bottomAnchor, anchor: .bottom) }
            }
        }
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
            Text(viewModel.isReplying ? "답을 쓰는 중이에요 — 끝나면 이어서 보낼 수 있어요" : "ChatGPT 계정의 Codex가 답해요")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(Self.mutedOpacity))
                .accessibilityIdentifier("chat.composer.hint")
        }
        .padding(.horizontal, DesignTokens.Spacing.lg)
        .padding(.vertical, DesignTokens.Spacing.sm)
        .background(DesignTokens.Color.surface(area))
    }
}
