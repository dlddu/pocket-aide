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
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
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
import androidx.compose.ui.text.style.TextOverflow
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
private const val HISTORY_COLLAPSED_LINES = 2

/**
 * One task of the ledger, and the ways to ask about it.
 *
 * Everything above the request field is read straight from the ledger row the
 * panel last read — status, stage, the reason it stands where it does and the
 * next action, each marked 미기록 when the ledger carries none. The pane adds no
 * judgement of its own and states nothing the ledger does not.
 *
 * Everything below it stores one request on this task's ledger row. The
 * suggestions fill the field rather than store, so what a button promises is the
 * text the user reads before it is kept. Storing adds nothing to the
 * conversation and starts no turn — it works while a reply is being written —
 * and the session picks the request up when it next reads its ledger, which
 * leaves every tool approval and the human merge gate exactly where they were.
 *
 * The screen says *saved and waiting*, never *done*: only the ledger's own
 * status says the session dealt with it.
 */
@Composable
fun ChatTaskDetailPane(
    task: ChatTask,
    draft: String,
    onDraftChange: (String) -> Unit,
    onChoicePicked: (ChatTaskChoice) -> Unit,
    request: ChatTaskRequestUiState,
    onStore: () -> Unit,
    history: ChatTaskHistoryUiState,
    onLoadHistory: () -> Unit,
    onLoadMoreHistory: () -> Unit,
    turns: ChatTaskTurnsUiState,
    onLoadTurns: () -> Unit,
    onLoadMoreTurns: () -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    // The outcome belongs to the task it was sent about; another task's result
    // must not appear under this one.
    val mine = request.taskId == task.id
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
            // What the task is for and how it will be known to be done. The
            // older ledger shape had no column for either, so a carried-over
            // task says "not recorded" rather than showing a blank.
            NullableField("무엇을 · 왜", turns.task.description, "chat.tasks.detail.description")
            NullableField("완료 기준", turns.task.completionCriteria, "chat.tasks.detail.criteria")
            Field("진행 내용", task.stage, "chat.tasks.detail.stage")
            Field("검토·차단 사유", task.reason, "chat.tasks.detail.reason")
            Field("남은 다음 행동", task.nextAction, "chat.tasks.detail.next")
            task.requests?.takeIf { it.open > 0 }?.let { Waiting(it) }
            Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
            Turns(task.id, turns, onLoadTurns, onLoadMoreTurns)
            Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
            Choices(task.choices, onChoicePicked)
            RequestField(
                draft = draft,
                onDraftChange = onDraftChange,
                request = if (mine) request else ChatTaskRequestUiState(),
                onStore = onStore,
            )
            // Only a task carried over from the ledger has an older record to
            // show. One created after the cutover has none, and the pane does
            // not imply one is missing.
            if (turns.ledgerHistory) {
                Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
                History(task.id, history, onLoadHistory, onLoadMoreHistory)
            }
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
 * What the session itself offers to say about this task.
 *
 * The app writes none of these: they are the sentences the session wrote
 * against this very task version, and a task that has moved offers none until
 * the session writes new ones. Tapping one fills the request box — it does not
 * store and does not send — so the words that go are the words that were read.
 */
@Composable
private fun Choices(choices: List<ChatTaskChoice>, onPicked: (ChatTaskChoice) -> Unit) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            "이 작업에 보낼 요청",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        if (choices.isEmpty()) {
            // Saying there are none is the honest answer: the app has no list
            // of its own to fall back on, and free text is still there.
            Text(
                "세션이 올려 둔 선택지가 없어요. 직접 써서 보낼 수 있어요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.tasks.detail.noChoices"),
            )
            return@Column
        }
        choices.forEach { choice ->
            val shape = RoundedCornerShape(DesignTokens.Radius.chip)
            Text(
                choice.label,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink,
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(shape)
                    .background(DesignTokens.Colors.card(DETAIL_AREA))
                    .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), shape)
                    .clickable(role = Role.Button) { onPicked(choice) }
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm)
                    .testTag("chat.tasks.detail.choice"),
            )
        }
    }
}

/**
 * How many requests on this task the session has not settled yet, and what the
 * newest of them is doing. It is the ledger's own count — the panel reports it
 * and never decides that something is finished.
 */
@Composable
private fun Waiting(requests: ChatTaskRequests) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        Text(
            "보낸 요청",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        Text(
            buildString {
                append("확인 대기 ").append(requests.open).append("건")
                if (requests.reconciling > 0) append(" · 결과 대조 중 ").append(requests.reconciling).append("건")
            },
            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
            color = ink,
            modifier = Modifier.testTag("chat.tasks.detail.waitingRequests"),
        )
    }
}

/**
 * The request itself: free text and one store.
 *
 * Storing writes the words to this task's ledger row and nothing else — no
 * message, no turn — so it is available while a reply is being written. A
 * refusal keeps what was typed so the same words can be stored again.
 */
@Composable
private fun RequestField(
    draft: String,
    onDraftChange: (String) -> Unit,
    request: ChatTaskRequestUiState,
    onStore: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    val fieldShape = RoundedCornerShape(DesignTokens.Radius.cardLarge)
    // Blocked only while one is in flight: a reply being written does not stop
    // a request from being stored.
    val canStore = draft.isNotBlank() && !request.isStoring
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
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
                    .alpha(if (canStore) 1f else DISABLED_OPACITY)
                    .clip(shape)
                    .background(DesignTokens.Colors.accent(DETAIL_AREA))
                    .clickable(enabled = canStore, role = Role.Button, onClick = onStore)
                    .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.sm)
                    .testTag("chat.tasks.detail.store"),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    // Says what the tap does, including that a failed one is a
                    // retry rather than a second request.
                    if (request.error != null) "다시 저장하기" else "요청 등록",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                    color = DesignTokens.Colors.card(DETAIL_AREA),
                )
            }
            if (request.isStoring) {
                CircularProgressIndicator(
                    color = ink.copy(alpha = MUTED_OPACITY),
                    strokeWidth = SPINNER_STROKE,
                    modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.detail.storing"),
                )
            }
        }
        // Saved, refused, or nothing yet — three different sentences, and a
        // refusal never clears what was typed. "Saved" is deliberately not
        // "done": the session has not seen it yet.
        when {
            request.error != null -> Note(request.error.message, DesignTokens.StatusColors.failure(), "chat.tasks.detail.error")
            request.stored != null -> Note(
                if (request.replayed) {
                    "이미 저장된 요청이에요 · 부모 확인 대기"
                } else {
                    "요청 저장됨 · 부모 확인 대기"
                },
                ink.copy(alpha = MUTED_OPACITY),
                "chat.tasks.detail.stored",
            )
            else -> Note(
                "등록하면 이 작업의 원장에 남고, 부모가 다음 확인 때 접수해요. 대화로는 보내지 않아요.",
                ink.copy(alpha = MUTED_OPACITY),
                "chat.tasks.detail.hint",
            )
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

/**
 * This task's history, oldest first: what the person asked, what the session
 * did with it, and the task's own transitions.
 *
 * It shows only what the ledger recorded. A task from before the ledger kept a
 * history has none, and the pane says so rather than reconstructing one from
 * the task's current state. A long line opens in place rather than being cut.
 */
@Composable
private fun History(
    taskID: String,
    history: ChatTaskHistoryUiState,
    onLoad: () -> Unit,
    onLoadMore: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    // The first page is read when this task's pane opens, and again when the
    // pane switches to another task.
    LaunchedEffect(taskID) { onLoad() }
    Column(
        Modifier.fillMaxWidth().testTag("chat.tasks.detail.history"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Text(
            "작업 이력",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        when {
            history.error != null -> {
                Note(history.error.message, DesignTokens.StatusColors.failure(), "chat.tasks.detail.history.error")
                val shape = RoundedCornerShape(DesignTokens.Radius.chip)
                Text(
                    "다시 시도",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = DesignTokens.Colors.accent(DETAIL_AREA),
                    modifier = Modifier
                        .clip(shape)
                        .border(RULE, DesignTokens.Colors.accent(DETAIL_AREA), shape)
                        .clickable(role = Role.Button, onClick = onLoad)
                        .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                        .testTag("chat.tasks.detail.history.retry"),
                )
            }

            history.isLoading && history.events.isEmpty() -> CircularProgressIndicator(
                color = ink.copy(alpha = MUTED_OPACITY),
                strokeWidth = SPINNER_STROKE,
                modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.detail.history.loading"),
            )

            history.isEmpty -> Note(
                "아직 남은 기록이 없어요. 이 시점 이전의 일은 원장에 적혀 있지 않아요.",
                ink.copy(alpha = MUTED_OPACITY),
                "chat.tasks.detail.history.empty",
            )

            else -> {
                history.events.forEach { HistoryLine(it) }
                if (history.hasMore) {
                    val shape = RoundedCornerShape(DesignTokens.Radius.chip)
                    Text(
                        if (history.isLoading) "불러오는 중" else "이전 기록 더 보기",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = DesignTokens.Colors.accent(DETAIL_AREA),
                        modifier = Modifier
                            .clip(shape)
                            .border(RULE, DesignTokens.Colors.accent(DETAIL_AREA), shape)
                            .clickable(enabled = !history.isLoading, role = Role.Button, onClick = onLoadMore)
                            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                            .testTag("chat.tasks.detail.history.more"),
                    )
                }
            }
        }
    }
}

/** One history line. Its long body opens in place so nothing is read half. */
@Composable
private fun HistoryLine(event: ChatTaskEvent) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    var open by remember(event.seq) { mutableStateOf(false) }
    val body = listOf(event.text, event.result).firstOrNull { it.isNotBlank() }.orEmpty()
    Column(
        Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(DesignTokens.Radius.card))
            .background(DesignTokens.Colors.card(DETAIL_AREA))
            .then(if (body.isBlank()) Modifier else Modifier.clickable(role = Role.Button) { open = !open })
            .padding(DesignTokens.Spacing.md)
            .testTag("chat.tasks.detail.history.line"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Text(
            ChatTaskEventKind.label(event.kind) + " · " + event.at,
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        Text(
            event.summary,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
            color = ink,
        )
        if (body.isNotBlank()) {
            Text(
                body,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                maxLines = if (open) Int.MAX_VALUE else HISTORY_COLLAPSED_LINES,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.testTag("chat.tasks.detail.history.body"),
            )
        }
    }
}

/**
 * A field the agreed records may genuinely not have recorded.
 *
 * null is not an empty string here: a task carried over from a ledger that
 * never kept the field has nothing to show, and saying so is more honest than
 * a blank line a person would read as the answer.
 */
@Composable
private fun NullableField(label: String, value: String?, tag: String) {
    Field(label, value ?: "", tag)
}

/**
 * One task's turns, oldest first — the work itself.
 *
 * Each turn is shown as the exchange it was: the person's words, the
 * instruction that went to the session doing the work, that session's own
 * answer, and the parent's explanation of it, each labelled so none of them is
 * read as another. The parent's summary is never shown as the session's words.
 */
@Composable
private fun Turns(
    taskID: String,
    turns: ChatTaskTurnsUiState,
    onLoad: () -> Unit,
    onLoadMore: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    // The first page is read when this task's pane opens, and again when the
    // pane switches to another task.
    LaunchedEffect(taskID) { onLoad() }
    Column(
        Modifier.fillMaxWidth().testTag("chat.tasks.detail.turns"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Text(
            "주고받은 내용",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        when {
            turns.error != null -> {
                Note(turns.error.message, DesignTokens.StatusColors.failure(), "chat.tasks.detail.turns.error")
                val shape = RoundedCornerShape(DesignTokens.Radius.chip)
                Text(
                    "다시 시도",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = DesignTokens.Colors.accent(DETAIL_AREA),
                    modifier = Modifier
                        .clip(shape)
                        .border(RULE, DesignTokens.Colors.accent(DETAIL_AREA), shape)
                        .clickable(role = Role.Button, onClick = onLoad)
                        .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                        .testTag("chat.tasks.detail.turns.retry"),
                )
            }

            turns.isLoading && turns.turns.isEmpty() -> CircularProgressIndicator(
                color = ink.copy(alpha = MUTED_OPACITY),
                strokeWidth = SPINNER_STROKE,
                modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.detail.turns.loading"),
            )

            turns.isEmpty -> Text(
                "아직 주고받은 내용이 없어요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.tasks.detail.turns.empty"),
            )

            else -> {
                turns.turns.forEach { TurnCard(it) }
                if (turns.hasMore) {
                    val shape = RoundedCornerShape(DesignTokens.Radius.chip)
                    Text(
                        "이전 내용 더 보기",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = DesignTokens.Colors.accent(DETAIL_AREA),
                        modifier = Modifier
                            .clip(shape)
                            .border(RULE, DesignTokens.Colors.accent(DETAIL_AREA), shape)
                            .clickable(role = Role.Button, onClick = onLoadMore)
                            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                            .testTag("chat.tasks.detail.turns.more"),
                    )
                }
            }
        }
    }
}

/** One turn, with every part of it labelled for what it is. */
@Composable
private fun TurnCard(turn: ChatTaskTurn) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(DETAIL_AREA))
            .padding(DesignTokens.Spacing.md)
            .testTag("chat.tasks.detail.turn.${turn.id}"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Row(
            Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                turnWhen(turn),
                style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.tasks.detail.turn.${turn.id}.when"),
            )
            if (turn.isCarriedOver) {
                // A carried-over exchange is not presented as something written
                // here: the older ledger is where it came from.
                Text(
                    "이전 기록",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.testTag("chat.tasks.detail.turn.${turn.id}.carried"),
                )
            }
        }
        TurnPart("보낸 요청", turn.request, "chat.tasks.detail.turn.${turn.id}.request")
        TurnPart("세션에 전달한 지시", turn.instruction, "chat.tasks.detail.turn.${turn.id}.instruction")
        TurnPart("세션의 답(원문)", turn.response, "chat.tasks.detail.turn.${turn.id}.response")
        TurnPart("설명", turn.explanation, "chat.tasks.detail.turn.${turn.id}.explanation")
        TurnPart("질문", turn.question, "chat.tasks.detail.turn.${turn.id}.question")
        if (turn.choices.isNotEmpty()) {
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
                turn.choices.forEachIndexed { index, choice ->
                    Text(
                        choice.label,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = ink,
                        modifier = Modifier.testTag("chat.tasks.detail.turn.${turn.id}.choice.$index"),
                    )
                }
            }
        }
        if (turn.artifacts.isNotEmpty()) {
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
                Text(
                    "산출물",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = MUTED_OPACITY),
                )
                turn.artifacts.forEachIndexed { index, artifact ->
                    Text(
                        artifact,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = ink,
                        modifier = Modifier.testTag("chat.tasks.detail.turn.${turn.id}.artifact.$index"),
                    )
                }
            }
        }
        if (turn.isOpen) {
            Note(
                "아직 답을 기다리고 있어요.",
                DesignTokens.StatusColors.inProgress(),
                "chat.tasks.detail.turn.${turn.id}.open",
            )
        }
    }
}

/**
 * One labelled part of a turn. A part the records do not have is left out
 * entirely rather than shown as an empty row: a turn is written as it happens,
 * so a turn with only a request is the ordinary case and not a gap.
 */
@Composable
private fun TurnPart(label: String, value: String?, tag: String) {
    if (value.isNullOrBlank()) return
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        Text(
            value,
            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
            color = ink,
            modifier = Modifier.testTag(tag),
        )
    }
}

/** When a turn happened, as far as the records say. */
private fun turnWhen(turn: ChatTaskTurn): String = when {
    turn.endedAt != null -> turn.endedAt
    turn.startedAt != null -> turn.startedAt
    turn.requestedAt.isNotBlank() -> turn.requestedAt
    else -> "시각 기록 없음"
}
