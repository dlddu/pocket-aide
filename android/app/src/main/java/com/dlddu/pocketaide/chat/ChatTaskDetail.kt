package com.dlddu.pocketaide.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.KeyboardArrowLeft
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.PlainTextField

// docs/design-system/tokens.md §6.2 — the wide layout's right pane, showing one
// task instead of the list. No mockup exists for this pane (the panel it belongs
// to has none either, user decision 2026-11-18), so every value is a token.
private val DETAIL_AREA = Area.AiChat
private val BACK_BUTTON_SIZE = 28.dp
private val BACK_ICON_SIZE = 16.dp
private val SPINNER_SIZE = 14.dp
private val SPINNER_STROKE = 2.dp
private val RULE = 1.dp
private const val MUTED_OPACITY = 0.55f
private const val DISABLED_OPACITY = 0.4f
private const val QUOTE_OPACITY = 0.06f
private const val REQUEST_MAX_LINES = 6

/**
 * One task of the ledger, and the ways to ask about it.
 *
 * Everything above the request field is read straight from the ledger row the
 * panel last read — status, stage, the reason it stands where it does and the
 * next action, each marked 미기록 when the ledger carries none. The pane adds no
 * judgement of its own and states nothing the ledger does not.
 *
 * Everything below it is one message in this conversation. The suggestions fill
 * the field rather than send, so what a button promises is the text the user
 * reads before it goes; the session receives it as a request and decides, which
 * leaves every tool approval and the human merge gate exactly where they were.
 */
@Composable
fun ChatTaskDetailPane(
    task: ChatTask,
    draft: String,
    onDraftChange: (String) -> Unit,
    followUp: ChatFollowUpUiState,
    // A conversation carries one reply at a time (PRD-1 AC1), so a follow-up
    // waits for the one in flight instead of being refused by the backend.
    replying: Boolean,
    onSend: () -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    // The outcome belongs to the task it was sent about; another task's result
    // must not appear under this one.
    val mine = followUp.taskId == task.id
    Column(
        modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(DETAIL_AREA))
            .testTag("chat.tasks.detail"),
    ) {
        DetailHeader(task, onBack)
        Column(
            Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.lg),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        ) {
            Field("진행 내용", task.stage, "chat.tasks.detail.stage")
            Field("검토·차단 사유", task.reason, "chat.tasks.detail.reason")
            Field("남은 다음 행동", task.nextAction, "chat.tasks.detail.next")
            Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
            Suggestions(task, onDraftChange)
            Request(
                task = task,
                draft = draft,
                onDraftChange = onDraftChange,
                followUp = if (mine) followUp else ChatFollowUpUiState(),
                replying = replying,
                onSend = onSend,
            )
        }
    }
}

@Composable
private fun DetailHeader(task: ChatTask, onBack: () -> Unit) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
            Box(
                Modifier
                    .size(BACK_BUTTON_SIZE)
                    .clip(RoundedCornerShape(DesignTokens.Radius.chip))
                    .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), RoundedCornerShape(DesignTokens.Radius.chip))
                    .clickable(role = Role.Button, onClick = onBack)
                    .semantics { contentDescription = "작업 목록으로" }
                    .testTag("chat.tasks.detail.back"),
                contentAlignment = Alignment.Center,
            ) {
                Icon(
                    Icons.Outlined.KeyboardArrowLeft,
                    contentDescription = null,
                    tint = ink,
                    modifier = Modifier.size(BACK_ICON_SIZE),
                )
            }
            Text(
                "작업 목록",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
            )
        }
        Row(
            Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            verticalAlignment = Alignment.Top,
        ) {
            Text(
                task.title.ifBlank { task.id },
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold),
                color = ink,
                modifier = Modifier.weight(1f).testTag("chat.tasks.detail.title"),
            )
            StatusBadge(task.status, tag = "chat.tasks.detail.status.${task.status}")
        }
        // The id is what the follow-up names, so it is on screen rather than
        // hidden inside the message.
        Text(
            "작업 ID ${task.id}",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
            modifier = Modifier.testTag("chat.tasks.detail.id"),
        )
    }
    Box(Modifier.fillMaxWidth().height(DesignTokens.Split.rule).background(DesignTokens.Colors.rule(DETAIL_AREA)))
}

/** One ledger field. An unrecorded one says so — it is never left blank, and never filled in. */
@Composable
private fun Field(label: String, value: String, tag: String) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    val recorded = value.isNotBlank()
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        Text(
            if (recorded) value else "원장에 적혀 있지 않아요.",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
            color = ink,
            modifier = Modifier
                .then(if (recorded) Modifier else Modifier.alpha(MUTED_OPACITY))
                .testTag(tag),
        )
    }
}

/**
 * Ready-made follow-ups for this task's state. Tapping one writes its sentence
 * into the request field — it does not send — so the user sees the words that
 * will go and can change them first.
 */
@Composable
private fun Suggestions(task: ChatTask, onDraftChange: (String) -> Unit) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            "이 작업에 보낼 요청",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        chatTaskSuggestions(task).forEach { suggestion ->
            val shape = RoundedCornerShape(DesignTokens.Radius.chip)
            Text(
                suggestion.label,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink,
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(shape)
                    .background(DesignTokens.Colors.card(DETAIL_AREA))
                    .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), shape)
                    .clickable(role = Role.Button) { onDraftChange(suggestion.request) }
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm)
                    .testTag("chat.tasks.detail.suggestion"),
            )
        }
    }
}

/**
 * The request itself: free text, the line the task is named on, and one send.
 *
 * The whole outgoing message is on screen — the quoted line above the field is
 * exactly what is prepended — so nothing is sent that the user did not read.
 */
@Composable
private fun Request(
    task: ChatTask,
    draft: String,
    onDraftChange: (String) -> Unit,
    followUp: ChatFollowUpUiState,
    replying: Boolean,
    onSend: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    val fieldShape = RoundedCornerShape(DesignTokens.Radius.cardLarge)
    // Blocked while one is in flight (so a second tap cannot send a second
    // copy) and while a reply is still being written.
    val canSend = draft.isNotBlank() && !followUp.isSending && !replying
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            chatTaskFollowUpMessage(task, "").trim(),
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(DesignTokens.Radius.chip))
                .background(ink.copy(alpha = QUOTE_OPACITY))
                .padding(DesignTokens.Spacing.sm)
                .testTag("chat.tasks.detail.context"),
        )
        Box(
            Modifier
                .fillMaxWidth()
                .clip(fieldShape)
                .background(DesignTokens.Colors.card(DETAIL_AREA))
                .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), fieldShape)
                .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm),
        ) {
            PlainTextField(
                value = draft,
                onValueChange = onDraftChange,
                placeholder = "직접 쓰거나 위에서 골라요",
                textStyle = DesignTokens.Typography.style(DesignTokens.Typography.body),
                textColor = ink,
                placeholderColor = ink.copy(alpha = MUTED_OPACITY),
                maxLines = REQUEST_MAX_LINES,
                keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Sentences),
                modifier = Modifier.fillMaxWidth().testTag("chat.tasks.detail.field"),
            )
        }
        val shape = RoundedCornerShape(DesignTokens.Radius.chip)
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
            Box(
                Modifier
                    .alpha(if (canSend) 1f else DISABLED_OPACITY)
                    .clip(shape)
                    .background(DesignTokens.Colors.accent(DETAIL_AREA))
                    .clickable(enabled = canSend, role = Role.Button, onClick = onSend)
                    .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.sm)
                    .testTag("chat.tasks.detail.send"),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    // Says what the tap does, including that a failed one is a
                    // retry rather than a second request.
                    if (followUp.error != null) "다시 보내기" else "대화로 보내기",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                    color = DesignTokens.Colors.card(DETAIL_AREA),
                )
            }
            if (followUp.isSending) {
                CircularProgressIndicator(
                    color = ink.copy(alpha = MUTED_OPACITY),
                    strokeWidth = SPINNER_STROKE,
                    modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.detail.sending"),
                )
            }
        }
        // Taken, refused, or waiting on the reply in flight — three different
        // sentences, and a refusal never clears what was typed.
        when {
            followUp.error != null -> Note(followUp.error.message, DesignTokens.StatusColors.failure(), "chat.tasks.detail.error")
            followUp.accepted -> Note("보냈어요 — 왼쪽 대화에 올라갔어요.", ink.copy(alpha = MUTED_OPACITY), "chat.tasks.detail.accepted")
            replying -> Note("답을 쓰는 중이에요 — 끝나면 보낼 수 있어요.", ink.copy(alpha = MUTED_OPACITY), "chat.tasks.detail.waiting")
            else -> Note("보내면 왼쪽 대화의 메시지로 올라가요. 원장은 이 화면에서 바뀌지 않아요.", ink.copy(alpha = MUTED_OPACITY), "chat.tasks.detail.hint")
        }
    }
}

@Composable
private fun Note(text: String, color: androidx.compose.ui.graphics.Color, tag: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
        color = color,
        modifier = Modifier.testTag(tag),
    )
}
