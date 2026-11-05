package com.dlddu.pocketaide.chat

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.ScreenHeader
import java.time.LocalDate

private val AREA = Area.AiChat
private val TITLE_SIZE = 26.sp
private val BUBBLE_MAX_WIDTH = 300.dp
private val BUBBLE_HORIZONTAL = 14.dp
private val BUBBLE_VERTICAL = 10.dp
private val AVATAR_SIZE = 24.dp
private val SEND_SIZE = 40.dp
private val BORDER = 1.dp
private val BODY_LINE_HEIGHT = 22.sp
private const val COMPOSER_MAX_LINES = 5

@Composable
fun ChatScreen(api: ApiClient) {
    val viewModel: ChatViewModel = viewModel(key = "chat") { ChatViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()

    LaunchedEffect(viewModel) { viewModel.load() }

    if (state.showsHistory) BackHandler { viewModel.hideHistory() }

    Column(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(AREA))
            .imePadding()
            .testTag("chat.screen"),
    ) {
        ScreenHeader(
            area = AREA,
            title = if (state.showsHistory) "대화 목록" else "AI 채팅",
            titleSize = TITLE_SIZE,
            subtitle = if (state.showsHistory) null else headerSubtitle(state),
            subtitleTag = "chat.header.subtitle",
            trailing = {
                Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
                    HeaderAction(
                        label = if (state.showsHistory) "닫기" else "목록",
                        tag = "chat.history.button",
                        onClick = { if (state.showsHistory) viewModel.hideHistory() else viewModel.showHistory() },
                    )
                    HeaderAction(label = "새 대화", tag = "chat.new.button", onClick = viewModel::startNewConversation)
                }
            },
        )
        state.errorMessage?.let { ErrorBanner(it, onDismiss = viewModel::dismissError) }
        if (state.showsHistory) {
            HistoryList(state, onOpen = viewModel::open, modifier = Modifier.weight(1f))
        } else {
            Conversation(state, modifier = Modifier.weight(1f))
            Composer(enabled = !state.isSending, onSend = viewModel::send)
        }
    }
}

private fun headerSubtitle(state: ChatUiState): String =
    if (state.current == null) "새 대화" else ChatCopy.messageCount(state.messages.size)

@Composable
private fun HeaderAction(label: String, tag: String, onClick: () -> Unit) {
    Text(
        label,
        style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm, FontWeight.SemiBold),
        color = DesignTokens.Colors.accent(AREA),
        modifier = Modifier.clickable(onClick = onClick).padding(vertical = DesignTokens.Spacing.sm).testTag(tag),
    )
}

@Composable
private fun ErrorBanner(message: String, onDismiss: () -> Unit) {
    Text(
        message,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = DesignTokens.StatusColors.failure(),
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onDismiss)
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.sm)
            .testTag("chat.error"),
    )
}

@Composable
private fun Conversation(state: ChatUiState, modifier: Modifier) {
    val listState = rememberLazyListState()
    val streaming = state.streamingReply
    val rowCount = state.messages.size + if (streaming != null) 1 else 0

    LaunchedEffect(rowCount, streaming?.length) {
        if (rowCount > 0) listState.animateScrollToItem(rowCount - 1)
    }

    if (rowCount == 0) {
        Box(modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
            Text(
                if (state.hasLoaded) "무엇이든 물어보세요" else "",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.5f),
                modifier = Modifier.testTag("chat.empty"),
            )
        }
        return
    }
    LazyColumn(
        state = listState,
        modifier = modifier.fillMaxWidth().testTag("chat.messages"),
        contentPadding = androidx.compose.foundation.layout.PaddingValues(
            horizontal = DesignTokens.Spacing.xl,
            vertical = DesignTokens.Spacing.md,
        ),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        items(state.messages, key = { it.id }) { message ->
            when (message.role) {
                ChatRole.USER -> MyBubble(message.content)
                ChatRole.ASSISTANT -> AiBubble(message.content, tag = "chat.bubble.ai")
            }
        }
        if (streaming != null) {
            item(key = "streaming") {
                if (streaming.isEmpty()) AiBubble("입력 중…", tag = "chat.bubble.typing") else AiBubble(streaming, tag = "chat.bubble.streaming")
            }
        }
    }
}

@Composable
private fun MyBubble(text: String) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
        Text(
            text,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg).copy(lineHeight = BODY_LINE_HEIGHT),
            color = Color.White,
            modifier = Modifier
                .widthIn(max = BUBBLE_MAX_WIDTH)
                .clip(RoundedCornerShape(DesignTokens.Radius.bubble))
                .background(DesignTokens.Colors.accent(AREA))
                .padding(horizontal = BUBBLE_HORIZONTAL, vertical = BUBBLE_VERTICAL)
                .testTag("chat.bubble.me"),
        )
    }
}

@Composable
private fun AiBubble(text: String, tag: String) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Box(
            Modifier.size(AVATAR_SIZE).clip(CircleShape).background(DesignTokens.Colors.soft(AREA)),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                "PA",
                style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold),
                color = DesignTokens.Colors.accent(AREA),
            )
        }
        Card(area = AREA, padding = CardPadding.Medium, modifier = Modifier.widthIn(max = BUBBLE_MAX_WIDTH)) {
            Text(
                text,
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg).copy(lineHeight = BODY_LINE_HEIGHT),
                color = DesignTokens.Colors.ink(AREA),
                modifier = Modifier.testTag(tag),
            )
        }
    }
}

@Composable
private fun Composer(enabled: Boolean, onSend: (String) -> Unit) {
    var text by remember { mutableStateOf("") }
    val ink = DesignTokens.Colors.ink(AREA)
    val canSend = enabled && text.isNotBlank()
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.md),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier
                .weight(1f)
                .clip(RoundedCornerShape(DesignTokens.Radius.card))
                .background(DesignTokens.Colors.card(AREA))
                .border(BORDER, DesignTokens.Colors.rule(AREA), RoundedCornerShape(DesignTokens.Radius.card))
                .padding(horizontal = BUBBLE_HORIZONTAL, vertical = BUBBLE_VERTICAL),
        ) {
            PlainTextField(
                value = text,
                onValueChange = { text = it },
                placeholder = "메시지를 입력하거나…",
                textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                textColor = ink,
                placeholderColor = ink.copy(alpha = 0.45f),
                maxLines = COMPOSER_MAX_LINES,
                modifier = Modifier.fillMaxWidth().testTag("chat.composer.field"),
            )
        }
        Box(
            Modifier
                .size(SEND_SIZE)
                .clip(CircleShape)
                .background(DesignTokens.Colors.accent(AREA).copy(alpha = if (canSend) 1f else 0.35f))
                .clickable(enabled = canSend) {
                    onSend(text)
                    text = ""
                }
                .testTag("chat.send.button"),
            contentAlignment = Alignment.Center,
        ) {
            Text("↑", style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.Bold), color = Color.White)
        }
    }
}

@Composable
private fun HistoryList(state: ChatUiState, onOpen: (ChatSession) -> Unit, modifier: Modifier) {
    val today = remember { LocalDate.now() }
    if (state.sessions.isEmpty()) {
        Box(modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
            Text(
                "지난 대화가 없어요",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.5f),
                modifier = Modifier.testTag("chat.history.empty"),
            )
        }
        return
    }
    LazyColumn(
        modifier = modifier.fillMaxWidth().testTag("chat.history.list"),
        contentPadding = androidx.compose.foundation.layout.PaddingValues(
            horizontal = DesignTokens.Spacing.xl,
            vertical = DesignTokens.Spacing.md,
        ),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.cardGap),
    ) {
        items(state.sessions, key = { it.id }) { session ->
            Card(
                area = AREA,
                emphasized = session.id == state.current?.id,
                modifier = Modifier.fillMaxWidth().clickable { onOpen(session) }.testTag("chat.history.row.${session.id}"),
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
                    Text(
                        ChatCopy.sessionTitle(session),
                        style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
                        color = DesignTokens.Colors.ink(AREA),
                    )
                    Text(
                        ChatTimes.stamp(session.updatedAt, today),
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.55f),
                    )
                }
            }
        }
    }
}
