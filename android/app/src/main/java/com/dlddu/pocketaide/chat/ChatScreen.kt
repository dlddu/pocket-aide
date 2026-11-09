package com.dlddu.pocketaide.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowUpward
import androidx.compose.material.icons.outlined.Refresh
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.LinkAnnotation
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextLinkStyles
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.text.withLink
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.repeatOnLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel
import com.dlddu.pocketaide.designsystem.components.PlainTextField

// docs/mockups/screen-chat-text.html — header, bubbles, composer. The header's
// action refreshes plugin skills (one conversation, no "new chat"), and the
// composer's round button sends (voice mode is not built yet).
private val CHAT_AREA = Area.AiChat
private val HEADER_BUTTON_SIZE = 36.dp
private val HEADER_ICON_SIZE = 16.dp
private val HEADER_LABEL_GAP = 2.dp
private val HEADER_COUNT_BASELINE = 3.dp
private val AVATAR_SIZE = 28.dp
private val BUBBLE_TAIL = 6.dp
private val BUBBLE_VERTICAL = 10.dp
private val COMPOSER_VERTICAL = 10.dp
private val SEND_SIZE = 44.dp
private val SEND_ICON_SIZE = 18.dp
private val SPINNER_SIZE = 16.dp
private val SPINNER_STROKE = 2.dp
private val RULE_WIDTH = 1.dp
private val DOT_SIZE = 6.dp
private const val ME_MAX_FRACTION = 0.80f
private const val AI_MAX_FRACTION = 0.82f
private const val COMPOSER_MAX_LINES = 6
private const val AVATAR_FILL_OPACITY = 0.15f
private const val MUTED_OPACITY = 0.55f
private const val CHIP_TEXT_OPACITY = 0.75f
private const val DISABLED_OPACITY = 0.4f

@Composable
fun ChatScreen(api: ApiClient) {
    val viewModel: ChatViewModel = viewModel(key = "chat") { ChatViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current
    var draft by rememberSaveable { mutableStateOf("") }

    LaunchedEffect(viewModel, lifecycleOwner) {
        lifecycleOwner.lifecycle.repeatOnLifecycle(Lifecycle.State.RESUMED) { viewModel.pollWhileVisible() }
    }

    Column(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(CHAT_AREA))
            .imePadding()
            .testTag("chat.screen"),
    ) {
        Header(state, viewModel::refreshPlugins)
        Box(Modifier.fillMaxWidth().height(RULE_WIDTH).background(DesignTokens.Colors.rule(CHAT_AREA)))
        state.notice?.let {
            Banner(it, DesignTokens.Colors.ink(CHAT_AREA), tag = "chat.notice", onDismiss = viewModel::dismissNotice)
        }
        state.error?.let { Banner(it.message, DesignTokens.StatusColors.failure(), tag = "chat.error", onDismiss = null) }
        Messages(state, Modifier.weight(1f), onCancelQueued = viewModel::cancelQueued)
        Composer(
            draft = draft,
            onDraftChange = { draft = it },
            state = state,
            onSend = { viewModel.send(draft) { draft = "" } },
        )
    }
}

@Composable
private fun Header(state: ChatUiState, onRefresh: () -> Unit) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = DesignTokens.Spacing.sm,
                bottom = DesignTokens.Spacing.md,
            ),
        verticalAlignment = Alignment.Bottom,
    ) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(HEADER_LABEL_GAP)) {
            AreaLabel(area = CHAT_AREA, text = "AI 채팅")
            Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
                Text(
                    state.conversation?.title?.takeIf { it.isNotBlank() } ?: "새 대화",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.h2, FontWeight.Bold),
                    color = ink,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f, fill = false).testTag("chat.header.title"),
                )
                Text(
                    "${state.messages.size}개의 메시지",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.padding(bottom = HEADER_COUNT_BASELINE).testTag("chat.header.count"),
                )
            }
        }
        RefreshButton(state, onRefresh)
    }
}

@Composable
private fun RefreshButton(state: ChatUiState, onRefresh: () -> Unit) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Box(
        Modifier
            .size(HEADER_BUTTON_SIZE)
            .alpha(if (state.canRefresh) 1f else DISABLED_OPACITY)
            .clip(CircleShape)
            .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), CircleShape)
            .clickable(enabled = state.canRefresh, role = Role.Button, onClick = onRefresh)
            .semantics { contentDescription = "스킬 새로고침" }
            .testTag("chat.refresh.button"),
        contentAlignment = Alignment.Center,
    ) {
        if (state.isRefreshing) {
            CircularProgressIndicator(color = ink, strokeWidth = SPINNER_STROKE, modifier = Modifier.size(SPINNER_SIZE))
        } else {
            Icon(Icons.Outlined.Refresh, contentDescription = null, tint = ink, modifier = Modifier.size(HEADER_ICON_SIZE))
        }
    }
}

@Composable
private fun Banner(text: String, color: Color, tag: String, onDismiss: (() -> Unit)?) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = color,
        modifier = Modifier
            .fillMaxWidth()
            .then(if (onDismiss != null) Modifier.clickable(onClick = onDismiss) else Modifier)
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.sm)
            .testTag(tag),
    )
}

@Composable
private fun Messages(state: ChatUiState, modifier: Modifier, onCancelQueued: () -> Unit) {
    val listState = rememberLazyListState()
    val count = state.messages.size + (if (state.isReplying) 1 else 0) + (if (state.queued != null) 1 else 0)
    LaunchedEffect(count, state.pending?.content, state.queued) {
        if (count > 0) listState.animateScrollToItem(count - 1)
    }
    if (state.hasLoaded && count == 0) {
        Box(modifier.fillMaxWidth().padding(DesignTokens.Spacing.xl), contentAlignment = Alignment.Center) {
            Text(
                "무엇이든 물어보세요. 첫 메시지를 보내면 ChatGPT 로그인 안내가 이 대화에 나와요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = DesignTokens.Colors.ink(CHAT_AREA).copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.empty"),
            )
        }
        return
    }
    BoxWithConstraints(modifier.fillMaxWidth()) {
        val width = maxWidth
        LazyColumn(
            state = listState,
            modifier = Modifier.fillMaxSize().testTag("chat.messages"),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(
                horizontal = DesignTokens.Spacing.xl,
                vertical = DesignTokens.Spacing.lg,
            ),
        ) {
            items(state.messages, key = { it.id }) { message ->
                if (message.isMine) {
                    MyBubble(message.content, Modifier.widthIn(max = width * ME_MAX_FRACTION))
                } else {
                    AiBubble(message.content, pending = false, Modifier.widthIn(max = width * AI_MAX_FRACTION))
                }
            }
            state.pending?.let { pending ->
                item(key = "pending") {
                    AiBubble(pending.content, pending = true, Modifier.widthIn(max = width * AI_MAX_FRACTION))
                }
            }
            state.queued?.let { queued ->
                item(key = "queued") {
                    QueuedBubble(queued.content, onCancelQueued, Modifier.widthIn(max = width * ME_MAX_FRACTION))
                }
            }
        }
    }
}

@Composable
private fun MyBubble(text: String, modifier: Modifier) {
    val shape = RoundedCornerShape(
        topStart = DesignTokens.Radius.bubble,
        topEnd = DesignTokens.Radius.bubble,
        bottomEnd = BUBBLE_TAIL,
        bottomStart = DesignTokens.Radius.bubble,
    )
    SelectionContainer {
        Text(
            withLinks(text, DesignTokens.Colors.card(CHAT_AREA)),
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
            color = DesignTokens.Colors.card(CHAT_AREA),
            modifier = modifier
                .clip(shape)
                .background(DesignTokens.Colors.ink(CHAT_AREA))
                .padding(horizontal = DesignTokens.Spacing.lg, vertical = BUBBLE_VERTICAL)
                .testTag("chat.message.mine"),
        )
    }
}

/** A message held for after the reply in progress (PRD-1 AC1, 2026-11-09
 * 개정) — dimmer than a sent bubble, with a cancel button while it waits. */
@Composable
private fun QueuedBubble(text: String, onCancel: () -> Unit, modifier: Modifier) {
    val shape = RoundedCornerShape(
        topStart = DesignTokens.Radius.bubble,
        topEnd = DesignTokens.Radius.bubble,
        bottomEnd = BUBBLE_TAIL,
        bottomStart = DesignTokens.Radius.bubble,
    )
    Column(
        Modifier.fillMaxWidth().testTag("chat.message.queued"),
        horizontalAlignment = Alignment.End,
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Text(
            text,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
            color = DesignTokens.Colors.card(CHAT_AREA),
            modifier = modifier
                .clip(shape)
                .background(DesignTokens.Colors.ink(CHAT_AREA).copy(alpha = MUTED_OPACITY))
                .padding(horizontal = DesignTokens.Spacing.lg, vertical = BUBBLE_VERTICAL),
        )
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs), verticalAlignment = Alignment.CenterVertically) {
            Text(
                "대기 중",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = DesignTokens.Colors.ink(CHAT_AREA).copy(alpha = MUTED_OPACITY),
            )
            Chip("취소", tag = "chat.queued.cancel", onClick = onCancel)
        }
    }
}

@Composable
private fun AiBubble(text: String, pending: Boolean, modifier: Modifier) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    val rule = DesignTokens.Colors.rule(CHAT_AREA)
    val shape = RoundedCornerShape(
        topStart = DesignTokens.Radius.bubble,
        topEnd = DesignTokens.Radius.bubble,
        bottomEnd = DesignTokens.Radius.bubble,
        bottomStart = BUBBLE_TAIL,
    )
    val clipboard = LocalClipboardManager.current
    Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
        Avatar()
        Column(
            modifier
                .clip(shape)
                .background(DesignTokens.Colors.card(CHAT_AREA))
                .border(RULE_WIDTH, rule, shape)
                .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.md)
                .then(if (pending) Modifier.testTag("chat.message.pending") else Modifier.testTag("chat.message.ai")),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        ) {
            if (text.isNotBlank()) {
                SelectionContainer {
                    Text(
                        withLinks(text.trim(), DesignTokens.Colors.accent(CHAT_AREA)),
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                        color = ink,
                    )
                }
            }
            if (pending) {
                Typing()
            } else {
                Box(Modifier.fillMaxWidth().height(RULE_WIDTH).background(rule))
                Chip("복사") { clipboard.setText(AnnotatedString(text)) }
            }
        }
    }
}

/** Web addresses become links the system opens in the browser. */
private fun withLinks(text: String, linkColor: Color): AnnotatedString {
    val ranges = chatLinkRanges(text)
    if (ranges.isEmpty()) return AnnotatedString(text)
    val style = TextLinkStyles(SpanStyle(color = linkColor, textDecoration = TextDecoration.Underline))
    return buildAnnotatedString {
        var at = 0
        for (range in ranges) {
            append(text.substring(at, range.first))
            val url = text.substring(range)
            withLink(LinkAnnotation.Url(url, style)) { append(url) }
            at = range.last + 1
        }
        append(text.substring(at))
    }
}

@Composable
private fun Avatar() {
    val accent = DesignTokens.Colors.accent(CHAT_AREA)
    Box(
        Modifier
            .padding(bottom = DesignTokens.Spacing.xs)
            .size(AVATAR_SIZE)
            .clip(CircleShape)
            .background(accent.copy(alpha = AVATAR_FILL_OPACITY)),
        contentAlignment = Alignment.Center,
    ) {
        Text("PA", style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold), color = accent)
    }
}

@Composable
private fun Typing() {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs), verticalAlignment = Alignment.CenterVertically) {
        Text(
            "입력 중",
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm).copy(fontStyle = FontStyle.Italic),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        repeat(3) {
            Box(Modifier.size(DOT_SIZE).clip(CircleShape).background(ink.copy(alpha = MUTED_OPACITY)))
        }
    }
}

@Composable
private fun Chip(label: String, tag: String = "chat.message.copy", onClick: () -> Unit) {
    val shape = RoundedCornerShape(DesignTokens.Radius.chip)
    Text(
        label,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = DesignTokens.Colors.ink(CHAT_AREA).copy(alpha = CHIP_TEXT_OPACITY),
        modifier = Modifier
            .clip(shape)
            .background(DesignTokens.Colors.card(CHAT_AREA))
            .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
            .testTag(tag),
    )
}

private fun composerHint(state: ChatUiState): String = when {
    state.queued != null -> "대기 중인 메시지가 있어요 — 답이 끝나면 보낼게요"
    state.isReplying -> "답을 쓰는 중이에요 — 보내면 끝난 뒤 바로 전달돼요"
    else -> "ChatGPT 계정의 Codex가 답해요"
}

@Composable
private fun Composer(draft: String, onDraftChange: (String) -> Unit, state: ChatUiState, onSend: () -> Unit) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    val canSend = state.canSend && draft.isNotBlank()
    val fieldShape = RoundedCornerShape(DesignTokens.Radius.cardLarge)
    Column(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.surface(CHAT_AREA))
            .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.sm),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
            Box(
                Modifier
                    .weight(1f)
                    .clip(fieldShape)
                    .background(DesignTokens.Colors.card(CHAT_AREA))
                    .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), fieldShape)
                    .padding(horizontal = DesignTokens.Spacing.lg, vertical = COMPOSER_VERTICAL),
            ) {
                PlainTextField(
                    value = draft,
                    onValueChange = onDraftChange,
                    placeholder = "메시지를 입력하세요",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = MUTED_OPACITY),
                    maxLines = COMPOSER_MAX_LINES,
                    keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Sentences),
                    modifier = Modifier.fillMaxWidth().testTag("chat.composer.field"),
                )
            }
            Box(
                Modifier
                    .size(SEND_SIZE)
                    .alpha(if (canSend) 1f else DISABLED_OPACITY)
                    .clip(CircleShape)
                    .background(DesignTokens.Colors.accent(CHAT_AREA))
                    .clickable(enabled = canSend, role = Role.Button, onClick = onSend)
                    .semantics { contentDescription = "보내기" }
                    .testTag("chat.composer.send"),
                contentAlignment = Alignment.Center,
            ) {
                if (state.isSending) {
                    CircularProgressIndicator(
                        color = DesignTokens.Colors.card(CHAT_AREA),
                        strokeWidth = SPINNER_STROKE,
                        modifier = Modifier.size(SPINNER_SIZE),
                    )
                } else {
                    Icon(
                        Icons.Filled.ArrowUpward,
                        contentDescription = null,
                        tint = DesignTokens.Colors.card(CHAT_AREA),
                        modifier = Modifier.size(SEND_ICON_SIZE),
                    )
                }
            }
        }
        Text(
            composerHint(state),
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
            modifier = Modifier.testTag("chat.composer.hint"),
        )
    }
}
