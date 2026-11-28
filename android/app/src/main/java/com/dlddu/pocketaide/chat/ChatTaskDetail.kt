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
import androidx.compose.material.icons.outlined.KeyboardArrowDown
import androidx.compose.material.icons.outlined.KeyboardArrowLeft
import androidx.compose.material.icons.outlined.KeyboardArrowUp
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
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
 * One work task, and the ways to ask about it.
 *
 * Everything above the request field is read straight from the task row the
 * panel last polled — the status, what the task is for, how it will be known to
 * be done, and what the session is asking — each marked 미기록 when the records
 * carry none. The pane adds no judgement of its own and states nothing the
 * records do not.
 *
 * When the session has sentences on offer, the turn that offered them says its
 * own piece directly above them (see [ChoiceTurn]): the request that went, the
 * explanation of what came back, and the question the sentences answer. Those
 * come from that turn, on the same task row as the choices, which is why they
 * are on screen before anybody opens the exchanges and why they can never be
 * some other turn's words.
 *
 * The task's exchanges *are* its history: the separate ledger history that
 * used to sit at the bottom is gone with the ledger it read (platform AC-N10),
 * because a past management event was never a turn and showing one as a turn
 * would put words in a person's mouth.
 *
 * Those exchanges are the last thing on the pane and start closed. A task that
 * has run for a while has a long one, and a person opening a task is almost
 * always there to answer what is being asked — so what is being asked and the
 * ways to answer it come first, and nothing of the exchanges is asked of the
 * session until somebody opens them. Everything the pane shows while they are
 * closed comes from the task list the panel already polls.
 *
 * Everything below the field stores one request, which becomes this task's next
 * turn. The suggestions fill the field rather than store, so what a button
 * promises is the text the user reads before it is kept. Storing adds nothing
 * to the conversation and starts no turn of the session's own — it works while
 * a reply is being written — and the parent picks it up when it next reads the
 * task's open turns, which leaves every tool approval and the human merge gate
 * exactly where they were.
 *
 * The screen says *saved and waiting*, never *done*: a turn is dealt with when
 * it ends, which only the parent writes.
 */
@Composable
fun ChatTaskDetailPane(
    task: ChatTask,
    draft: String,
    onDraftChange: (String) -> Unit,
    onChoicePicked: (ChatTaskChoice) -> Unit,
    request: ChatTaskRequestUiState,
    onStore: () -> Unit,
    turns: ChatTaskTurnsUiState,
    turnsOpen: Boolean,
    onToggleTurns: () -> Unit,
    onLoadTurns: () -> Unit,
    onRetryTurns: () -> Unit,
    onLoadMoreTurns: () -> Unit,
    materials: ChatTaskMaterialsUiState,
    onLoadMaterials: () -> Unit,
    onRetryMaterials: () -> Unit,
    onOpenArtifact: (ChatTaskArtifact) -> Unit,
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
            // What the task is for and how it will be known to be done. Both
            // come from the task row the panel polls rather than from the
            // exchanges, so they are on screen before anybody opens those —
            // and they stay as fresh as the panel's own reads. The older
            // ledger shape had no column for either, so a carried-over task
            // says "not recorded" rather than showing a blank.
            Field("무엇을 · 왜", task.description, "chat.tasks.detail.description")
            Field("완료 기준", task.completionCriteria, "chat.tasks.detail.criteria")
            // The materials for reviewing this task, read once for this task
            // when the pane opens. They are up here because a task in review is
            // opened to decide something, and the exchanges that hold the same
            // facts at length stay closed at the bottom.
            ReviewMaterials(
                task = task,
                state = materials,
                onLoad = onLoadMaterials,
                onRetry = onRetryMaterials,
                onOpenArtifact = onOpenArtifact,
            )
            // The turn whose sentences are on offer, when there are any. It is
            // read from the task row beside the choices themselves, so it is
            // that turn's words and not whichever turn happens to be open or
            // last.
            val offering = task.choiceTurn?.takeIf { task.choices.isNotEmpty() && it.hasAny }
            // What the session is asking, which is the nearest thing the
            // agreed records have to the ledger's free-text reason and next
            // action — those were the parent's own bookkeeping and never part
            // of what a person reads.
            //
            // It moves next to the choices when the offering turn asked one,
            // because that is the question those sentences answer and one
            // question read twice in two places is a person wondering which of
            // them is being asked. When the offering turn asked none — and on
            // a task with no choices at all — this stays exactly where it was.
            if (offering?.question.isNullOrBlank()) {
                Field("묻는 것", task.waiting?.question.orEmpty(), "chat.tasks.detail.question")
            }
            task.waiting?.let { Waiting(it) }
            Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
            offering?.let { ChoiceTurn(it) }
            Choices(task.choices, onChoicePicked)
            RequestField(
                draft = draft,
                onDraftChange = onDraftChange,
                request = if (mine) request else ChatTaskRequestUiState(),
                onStore = onStore,
            )
            Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
            // Last, and closed until asked for: a long exchange must not push
            // what is being asked off the screen, and a closed one is read
            // from nowhere.
            Turns(
                task = task,
                open = turnsOpen,
                onToggle = onToggleTurns,
                turns = turns,
                onLoad = onLoadTurns,
                onRetry = onRetryTurns,
                onLoadMore = onLoadMoreTurns,
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
 * What there is to review on this task, in one place near the top of the pane
 * (PRD-AI채팅 AC10, 검토 자료).
 *
 * Three different things, each under its own heading and each read from where
 * it is actually recorded:
 *
 *  - **연결된 PR** — what the work produced, across the task's whole history
 *    rather than its latest exchange. A task whose newest turn is "수정해 주세요"
 *    is exactly the case this exists for: the pull requests under review were
 *    produced three turns ago, and a screen that showed only the latest turn
 *    would show none of them. The same pull request named on several turns is
 *    one row here, because it is one pull request.
 *  - **이미 고른 것** — the sentences the session offered that came back as the
 *    next turn's request. That is a decision the records hold.
 *  - **아직 묻고 있는 것** — what the *latest* turn is asking. An earlier
 *    question that has turns after it is history, and is not offered here as
 *    something to decide.
 *
 * What it will not do is show a fourth: an alternative's advantages and
 * drawbacks. The agreed turn has no field for them, so this says they are not
 * recorded rather than reading them out of something else — a parent's
 * explanation is its account of a result, not a decision somebody took, and the
 * words on a choice button are a sentence to send, not a trade-off. Saying
 * nothing at all would read as "there were no trade-offs", which is a different
 * claim and not one the records make.
 *
 * None of it costs the task list anything. It is one read for the one task
 * somebody opened, triggered by the same signal the exchanges use, so the list
 * — and a panel of any number of rows — reads no history at all.
 */
@Composable
private fun ReviewMaterials(
    task: ChatTask,
    state: ChatTaskMaterialsUiState,
    onLoad: () -> Unit,
    onRetry: () -> Unit,
    onOpenArtifact: (ChatTaskArtifact) -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    // Reading starts here: on opening the task, and again when the panel's own
    // poll says this task's exchanges moved.
    LaunchedEffect(task.id, ChatTaskHistorySignal.of(task)) { onLoad() }
    // An answer belongs to the task it was read for. Switching tasks quickly
    // leaves the previous task's answer in state for an instant, and drawing it
    // under this title would put one task's pull requests under another's name.
    val mine = state.taskId == task.id
    val materials = if (mine) state.materials else ChatTaskMaterials()
    Column(
        Modifier.fillMaxWidth().testTag("chat.tasks.detail.materials"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Text(
            "검토 자료",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold),
            color = ink.copy(alpha = MUTED_OPACITY),
            modifier = Modifier.testTag("chat.tasks.detail.materials.label"),
        )
        when {
            !mine || (state.isLoading && !state.hasLoaded) -> CircularProgressIndicator(
                color = ink.copy(alpha = MUTED_OPACITY),
                strokeWidth = SPINNER_STROKE,
                modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.detail.materials.loading"),
            )

            state.error != null -> {
                // Why it could not be read, which is never "this task produced
                // nothing" — the records are still on the slot.
                Note(state.error.message, DesignTokens.StatusColors.failure(), "chat.tasks.detail.materials.error")
                val shape = RoundedCornerShape(DesignTokens.Radius.chip)
                Text(
                    "다시 시도",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = DesignTokens.Colors.accent(DETAIL_AREA),
                    modifier = Modifier
                        .clip(shape)
                        .border(RULE, DesignTokens.Colors.accent(DETAIL_AREA), shape)
                        .clickable(role = Role.Button, onClick = onRetry)
                        .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                        .testTag("chat.tasks.detail.materials.retry"),
                )
            }

            else -> {
                Artifacts(materials.artifacts, onOpenArtifact)
                Tradeoffs(materials.tradeoffs)
                Decisions(materials.decisions)
                OpenPoints(materials.open)
                if (materials.truncated) {
                    Note(
                        "기록이 길어 일부만 모았어요. 전체는 아래 「주고받은 내용」에 있어요.",
                        ink.copy(alpha = MUTED_OPACITY),
                        "chat.tasks.detail.materials.truncated",
                    )
                }
            }
        }
    }
    Box(Modifier.fillMaxWidth().height(RULE).background(DesignTokens.Colors.rule(DETAIL_AREA)))
}

/**
 * The pull requests and other artifacts the work produced, newest first.
 *
 * A row shows the recorded title when one is recorded and the identifier when
 * none is — never a title made up from a number, because the same number is a
 * different pull request in every repository. A row with an address is a tap
 * that opens it; one without is the identifier the records hold, and says so
 * instead of being a tap that goes nowhere.
 */
@Composable
private fun Artifacts(artifacts: List<ChatTaskArtifact>, onOpen: (ChatTaskArtifact) -> Unit) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            "연결된 PR · 산출물",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        if (artifacts.isEmpty()) {
            Text(
                "이 작업 기록에 적힌 산출물이 없어요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.alpha(MUTED_OPACITY).testTag("chat.tasks.detail.materials.noArtifacts"),
            )
            return@Column
        }
        artifacts.forEachIndexed { index, artifact ->
            val shape = RoundedCornerShape(DesignTokens.Radius.chip)
            Column(
                Modifier
                    .fillMaxWidth()
                    .clip(shape)
                    .background(DesignTokens.Colors.card(DETAIL_AREA))
                    .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), shape)
                    .then(
                        if (artifact.openable) {
                            Modifier.clickable(role = Role.Button) { onOpen(artifact) }
                        } else {
                            Modifier
                        },
                    )
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm)
                    .testTag("chat.tasks.detail.materials.artifact.$index"),
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
            ) {
                Text(
                    artifact.label,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = if (artifact.openable) DesignTokens.Colors.accent(DETAIL_AREA) else ink,
                    modifier = Modifier.testTag("chat.tasks.detail.materials.artifact.$index.label"),
                )
                // The second line is what the first one is not: the identifier
                // under a title, or the plain statement that there is nothing
                // to open. Either way nothing on this row is a guess.
                Text(
                    when {
                        artifact.titled && artifact.repo.isNotBlank() && artifact.number > 0 ->
                            "${artifact.repo}#${artifact.number}"
                        artifact.openable -> artifact.ref
                        else -> "링크가 적혀 있지 않아 열 수 없어요 · 기록된 식별자"
                    },
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.testTag("chat.tasks.detail.materials.artifact.$index.ref"),
                )
            }
        }
    }
}

/**
 * The alternatives somebody weighed, each with the advantages and drawbacks
 * *they wrote* — the one place on this screen where two options can be compared
 * side by side.
 *
 * Each row is one alternative under its own name, with 「장점」 and 「단점」 as
 * their own labelled lines. They are told apart by those labels rather than by
 * colour: the pane labels every other field the same way, and the status colours
 * this design system has mean a run passed or failed, which is not what an
 * advantage is.
 *
 * A side the author did not write says **단점 미기록** for that side. That is
 * not the same as "no drawbacks", and leaving the line blank would read as the
 * second — a claim the records do not make. The source turn is on every row, so
 * a reader can open that exchange and read the whole thing.
 *
 * When nothing was marked at all, this says so and says *why*, because silence
 * here reads as "there were no trade-offs".
 */
@Composable
private fun Tradeoffs(tradeoffs: List<ChatTaskTradeoff>) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            "대안별 장단점",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        if (tradeoffs.isEmpty()) {
            Text(
                "기록에 대안별 장점·단점이 적혀 있지 않아요. " +
                    "선택지 문구나 설명을 장단점으로 보여 주지는 않아요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.alpha(MUTED_OPACITY)
                    .testTag("chat.tasks.detail.materials.noTradeoffs"),
            )
            return@Column
        }
        tradeoffs.forEachIndexed { index, entry ->
            val shape = RoundedCornerShape(DesignTokens.Radius.card)
            Column(
                Modifier
                    .fillMaxWidth()
                    .clip(shape)
                    .background(DesignTokens.Colors.card(DETAIL_AREA))
                    .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), shape)
                    .padding(DesignTokens.Spacing.md)
                    .testTag("chat.tasks.detail.materials.tradeoff.$index"),
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            ) {
                Text(
                    entry.alternative,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
                    color = ink,
                    modifier = Modifier.testTag("chat.tasks.detail.materials.tradeoff.$index.alternative"),
                )
                TradeoffSide("장점", entry.pros, "chat.tasks.detail.materials.tradeoff.$index.pros")
                TradeoffSide("단점", entry.cons, "chat.tasks.detail.materials.tradeoff.$index.cons")
                Text(
                    "출처 · 턴 ${entry.turnId}${sourceSuffix(entry.source)}",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.testTag("chat.tasks.detail.materials.tradeoff.$index.source"),
                )
            }
        }
    }
}

/**
 * One side of one alternative. A side with nothing recorded says so under its
 * own label — the reader learns that this side was not written, rather than
 * reading an empty space as "none".
 */
@Composable
private fun TradeoffSide(label: String, values: List<String>, tag: String) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        if (values.isEmpty()) {
            Text(
                "$label 미기록",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.alpha(MUTED_OPACITY).testTag("$tag.missing"),
            )
            return@Column
        }
        values.forEachIndexed { at, value ->
            Text(
                "· $value",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = ink,
                modifier = Modifier.testTag("$tag.$at"),
            )
        }
    }
}

/**
 * What has been settled, as the author recorded it under a 「결정 사항」 marker.
 *
 * Only that. A parent's explanation, however much it reads like a conclusion,
 * is not in this list — and neither is a request that happens to repeat an
 * offered sentence, which is not evidence anybody pressed the button and not an
 * acceptance of anything.
 */
@Composable
private fun Decisions(decisions: List<ChatTaskDecision>) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            "이미 결정된 것",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        if (decisions.isEmpty()) {
            Text(
                "기록에 「결정 사항」으로 적힌 것이 없어요. 설명을 결정으로 보여 주지는 않아요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.alpha(MUTED_OPACITY).testTag("chat.tasks.detail.materials.noDecisions"),
            )
            return@Column
        }
        decisions.forEachIndexed { index, decision ->
            Column(
                Modifier.fillMaxWidth().testTag("chat.tasks.detail.materials.decision.$index"),
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
            ) {
                Text(
                    decision.text,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = ink,
                    modifier = Modifier.testTag("chat.tasks.detail.materials.decision.$index.text"),
                )
                Text(
                    "출처 · 턴 ${decision.turnId}${sourceSuffix(decision.source)}",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.testTag("chat.tasks.detail.materials.decision.$index.source"),
                )
            }
        }
    }
}

/**
 * Which field of the source turn said it, when the records name one. An
 * unknown field adds nothing rather than guessing — the turn alone is still a
 * place a reader can go.
 */
private fun sourceSuffix(source: String): String = when (source) {
    "response" -> " · 세션의 답"
    "explanation" -> " · 설명"
    else -> ""
}

/**
 * What is still being asked, from the task's latest turn.
 *
 * The sentences are listed as what was offered, not as buttons: the buttons
 * that act are further down the pane beside the request field, and two sets of
 * them would be two places to answer the same question.
 */
@Composable
private fun OpenPoints(points: List<ChatTaskOpenPoint>) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            "아직 판단이 필요한 것",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        if (points.isEmpty()) {
            Text(
                "기록에 아직 묻고 있는 것이 없어요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.alpha(MUTED_OPACITY).testTag("chat.tasks.detail.materials.noOpen"),
            )
            return@Column
        }
        points.forEachIndexed { index, point ->
            Column(
                Modifier.fillMaxWidth().testTag("chat.tasks.detail.materials.open.$index"),
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
            ) {
                if (point.question.isNotBlank()) {
                    Text(
                        point.question,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                        color = ink,
                        modifier = Modifier.testTag("chat.tasks.detail.materials.open.$index.question"),
                    )
                }
                point.choices.forEachIndexed { at, choice ->
                    Text(
                        "· ${choice.label}",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = ink.copy(alpha = MUTED_OPACITY),
                        modifier = Modifier
                            .testTag("chat.tasks.detail.materials.open.$index.choice.$at"),
                    )
                }
            }
        }
    }
}

/**
 * The turn the offered sentences belong to, in its own frame directly above
 * them: what was sent, what came back as the parent tells it, and the question
 * those sentences answer — in that order, because that is the order it
 * happened in and the question is what the buttons under it are answers to.
 *
 * It is here because the sentences alone are not a conversation. A person
 * opening a task sees "전부 반영해 줘" and has to know what was asked, what came
 * of it and what is being asked now before any of the buttons means anything —
 * and the exchanges that hold all of it start closed, deliberately, at the
 * bottom of the pane.
 *
 * These are [ChatTask.choiceTurn]'s own words, read with the choices from the
 * same turn of the same task row. Not the last turn of a history page, not the
 * open turn below: those can be a different turn, and one turn's words over
 * another turn's choices would tell somebody they are answering something they
 * are not. A part the records do not have is left out rather than filled in
 * from a turn that does have it.
 *
 * It is framed and labelled as one turn's words so it is not read as three more
 * fields of the task itself.
 */
@Composable
private fun ChoiceTurn(turn: ChatTaskChoiceTurn) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.ink(DETAIL_AREA).copy(alpha = QUOTE_OPACITY))
            .border(RULE, DesignTokens.Colors.rule(DETAIL_AREA), shape)
            .padding(DesignTokens.Spacing.md)
            .testTag("chat.tasks.detail.choiceTurn"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Text(
            "이 선택지를 올린 턴",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold),
            color = ink.copy(alpha = MUTED_OPACITY),
            modifier = Modifier.testTag("chat.tasks.detail.choiceTurn.label"),
        )
        TurnPart("보낸 요청", turn.request, "chat.tasks.detail.choiceTurn.request")
        TurnPart("설명", turn.explanation, "chat.tasks.detail.choiceTurn.explanation")
        TurnPart("질문", turn.question, "chat.tasks.detail.choiceTurn.question")
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
 * The open turn on this task: what was asked and whether anybody has picked it
 * up. The pane reports it and never decides that something is finished — a
 * turn is settled when it ends, which only the parent writes.
 *
 * This replaced a tally of stored request rows. A person's words are the
 * task's next turn now, so what is waiting is one turn rather than a count,
 * and there is no request lifecycle to report.
 */
@Composable
private fun Waiting(waiting: ChatTaskWaiting) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        Text(
            // Says it is the state of the open turn, not the words of one. The
            // words now have a place of their own above the choices, and two
            // things under one label is a person reading a status as a request.
            "보낸 요청의 상태",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        Text(
            if (waiting.running) "확인 중" else "확인 대기",
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
 * One task's turns, oldest first — the work itself, and the last thing on the
 * pane.
 *
 * Each turn is shown as the exchange it was: the person's words, the
 * instruction that went to the session doing the work, that session's own
 * answer, and the parent's explanation of it, each labelled so none of them is
 * read as another. The parent's summary is never shown as the session's words.
 *
 * It starts closed, and closed means *not read*: the page is asked for when
 * somebody opens it and never before, so stepping into a task — or polling it,
 * or storing a request on it — asks the session for no exchange at all. The
 * header says how many there are without reading any, because the task row the
 * panel already polls carries the count.
 *
 * While it is open, the panel's own poll is what says something happened:
 * [onLoad] runs again when the task row's exchange signal moves, so a turn
 * that arrives while somebody is reading shows up. Opening it again after
 * nothing changed reads nothing — [ChatViewModel.loadTaskTurns] holds that
 * decision, so the pane asks once per thing worth asking about.
 */
@Composable
private fun Turns(
    task: ChatTask,
    open: Boolean,
    onToggle: () -> Unit,
    turns: ChatTaskTurnsUiState,
    onLoad: () -> Unit,
    onRetry: () -> Unit,
    onLoadMore: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    Column(
        Modifier.fillMaxWidth().testTag("chat.tasks.detail.turns"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Row(
            Modifier
                .fillMaxWidth()
                .clickable(role = Role.Button, onClick = onToggle)
                // The handle is one button, so its own description is what gets
                // announced and the number beside it would not be — the count
                // is said here too, or somebody listening would hear "펼치기"
                // and not how much there is to open. A task with no exchange
                // says no number, here as on screen.
                .semantics {
                    val what = if (task.turnsTotal > 0) "주고받은 내용 ${task.turnsTotal}개" else "주고받은 내용"
                    contentDescription = if (open) "$what 접기" else "$what 펼치기"
                }
                .testTag("chat.tasks.detail.turns.toggle"),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                "주고받은 내용",
                style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.weight(1f),
            )
            // The count is the task row's own, so it is on screen without any
            // exchange being read. A task with none says nothing here rather
            // than "0".
            if (task.turnsTotal > 0) {
                Text(
                    task.turnsTotal.toString(),
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.testTag("chat.tasks.detail.turns.count"),
                )
            }
            Icon(
                if (open) Icons.Outlined.KeyboardArrowUp else Icons.Outlined.KeyboardArrowDown,
                contentDescription = null,
                tint = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.size(BACK_ICON_SIZE),
            )
        }
        // Nothing below the header exists while it is closed — no card is
        // composed, so nothing of the exchanges had to be read to draw one.
        if (open) {
            // Reading starts here and only here: on opening, and again when
            // the panel's poll says this task's exchanges moved. Closing the
            // section leaves this branch and cancels the effect, so nothing is
            // in flight behind a closed header.
            LaunchedEffect(task.id, ChatTaskHistorySignal.of(task)) { onLoad() }
            OpenTurns(task.id, turns, onRetry, onLoadMore)
        }
    }
}

/**
 * The exchanges once somebody has opened them: the page, or the one honest
 * answer about why there is none to draw.
 *
 * A page belongs to the task it was read for. Switching tasks quickly leaves
 * the previous task's page in state for an instant, and drawing it under this
 * title would put one task's work under another's name — so a page that is not
 * [taskID]'s is treated as one that has not arrived.
 */
@Composable
private fun OpenTurns(
    taskID: String,
    turns: ChatTaskTurnsUiState,
    onRetry: () -> Unit,
    onLoadMore: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(DETAIL_AREA)
    val mine = turns.taskId == taskID
    Column(
        Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        when {
            !mine -> CircularProgressIndicator(
                color = ink.copy(alpha = MUTED_OPACITY),
                strokeWidth = SPINNER_STROKE,
                modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.detail.turns.loading"),
            )

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
                        .clickable(role = Role.Button, onClick = onRetry)
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
