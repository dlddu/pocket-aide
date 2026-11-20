package com.dlddu.pocketaide.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.selection.toggleable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

// docs/design-system/tokens.md §6.2 — the wide layout's right pane. No mockup
// exists for this pane (user decision, 2026-11-18: build it from the design
// system and the requirements), so every value here comes from a token.
private val PANEL_AREA = Area.AiChat
private val SPINNER_SIZE = 18.dp
private val SPINNER_STROKE = 2.dp
private const val MUTED_OPACITY = 0.55f
private const val CHIP_OPACITY = 0.12f
private const val CHIP_SELECTED_OPACITY = 0.22f
private val CHIP_BORDER = 1.dp

/**
 * The task ledger of the session answering this conversation.
 *
 * The panel reads the ledger and never writes it: picking a status pill or
 * hiding completed work changes what is drawn, not what the ledger says, and
 * choosing a task opens a detail pane whose only outward action is a message
 * sent into this same conversation.
 */
@Composable
fun ChatTaskPanel(
    state: ChatTasksUiState,
    filter: ChatTaskFilter,
    onFilterChange: (ChatTaskFilter) -> Unit,
    onSelect: (ChatTask) -> Unit,
    // Held by the caller so the reading position survives collapsing the panel
    // and stepping into a task's detail and back.
    listState: LazyListState,
    modifier: Modifier = Modifier,
) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    Column(
        modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(PANEL_AREA))
            .testTag("chat.tasks.panel"),
    ) {
        PanelHeader(state, filter, onFilterChange)
        when {
            // Loading, "cannot read", "no tasks" and a list are four different
            // answers. None of them is allowed to look like another.
            state.isLoading -> PanelCentre {
                CircularProgressIndicator(
                    strokeWidth = SPINNER_STROKE,
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.size(SPINNER_SIZE).testTag("chat.tasks.loading"),
                )
            }

            state.error != null -> PanelCentre {
                Text(
                    state.error.message,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = if (state.error == ChatTaskError.UNREACHABLE) {
                        DesignTokens.StatusColors.failure()
                    } else {
                        ink.copy(alpha = MUTED_OPACITY)
                    },
                    modifier = Modifier.padding(DesignTokens.Spacing.xl).testTag("chat.tasks.error"),
                )
            }

            else -> when (val empty = filter.emptiness(state.tasks)) {
                null -> TaskList(filter.select(state.tasks), listState, onSelect)
                else -> EmptyNote(empty, filter, onFilterChange)
            }
        }
    }
}

/**
 * Nothing to draw, said in the words of the reason. A ledger with no tasks is
 * not the same as work the user chose to hide, and the two that still have work
 * behind them offer the one tap that brings it back.
 */
@Composable
private fun EmptyNote(empty: ChatTaskEmptiness, filter: ChatTaskFilter, onFilterChange: (ChatTaskFilter) -> Unit) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val chosen = filter.status?.let { ChatTaskStatus.label(it) }
    val (tag, text, undo) = when (empty) {
        ChatTaskEmptiness.NO_TASKS ->
            Triple("chat.tasks.empty", "아직 기록된 작업이 없어요.", null)

        ChatTaskEmptiness.FILTERED -> Triple(
            "chat.tasks.empty.filtered",
            if (chosen != null) "「$chosen」 작업이 없어요." else "보여 줄 작업이 없어요.",
            "전체 보기" to filter.copy(status = null),
        )

        // The rows are completed and hidden — they are still in the ledger, so
        // this says so and offers to show them rather than looking like none.
        ChatTaskEmptiness.COMPLETED_HIDDEN -> Triple(
            "chat.tasks.empty.completedHidden",
            "완료한 작업을 숨기고 있어요. 지워진 것은 아니에요.",
            "완료 표시하기" to filter.copy(hideCompleted = false),
        )
    }
    PanelCentre {
        Column(
            Modifier.padding(DesignTokens.Spacing.xl),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        ) {
            Text(
                text,
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = ink.copy(alpha = MUTED_OPACITY),
                textAlign = TextAlign.Center,
                modifier = Modifier.testTag(tag),
            )
            undo?.let { (label, restored) ->
                val shape = RoundedCornerShape(DesignTokens.Radius.chip)
                Text(
                    label,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = DesignTokens.Colors.accent(PANEL_AREA),
                    modifier = Modifier
                        .clip(shape)
                        .border(CHIP_BORDER, DesignTokens.Colors.accent(PANEL_AREA), shape)
                        .clickable(role = Role.Button) { onFilterChange(restored) }
                        .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                        .testTag("chat.tasks.empty.undo"),
                )
            }
        }
    }
}

@Composable
private fun PanelHeader(
    state: ChatTasksUiState,
    filter: ChatTaskFilter,
    onFilterChange: (ChatTaskFilter) -> Unit,
) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    Column(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Text(
            "작업",
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
            color = ink,
            modifier = Modifier.testTag("chat.tasks.title"),
        )
        // The statuses the ledger actually has, in display order, so what needs
        // a person is read first — and the chosen one stays even when its count
        // falls to zero, or there would be no way to see what is narrowing the
        // list. The counts are the whole ledger's, not the filtered view's.
        val present = ChatTaskStatus.order.filter { (state.counts[it] ?: 0) > 0 || filter.status == it }
        if (present.isNotEmpty()) {
            Row(
                Modifier.horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            ) {
                StatusPill(
                    label = "전체",
                    count = state.tasks.size,
                    selected = filter.status == null,
                    tone = ink,
                    tag = "chat.tasks.filter.all",
                ) { onFilterChange(filter.copy(status = null)) }
                present.forEach { status ->
                    StatusPill(
                        label = ChatTaskStatus.label(status),
                        count = state.counts[status] ?: 0,
                        selected = filter.status == status,
                        tone = if (ChatTaskStatus.needsAttention(status)) {
                            DesignTokens.Colors.accent(PANEL_AREA)
                        } else {
                            ink
                        },
                        tag = "chat.tasks.filter.$status",
                    ) { onFilterChange(filter.toggled(status)) }
                }
            }
        }
        // Offered while the ledger has completed work, and while the hide is on
        // even if it has none — a choice the user cannot see is a choice they
        // cannot undo.
        if ((state.counts[ChatTaskStatus.COMPLETED] ?: 0) > 0 || filter.hideCompleted) {
            HideCompletedToggle(filter.hideCompleted) { onFilterChange(filter.copy(hideCompleted = it)) }
        }
        if (state.truncated) {
            // A prefix must not read as the whole ledger.
            Text(
                "작업이 많아 일부만 보여요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.tasks.truncated"),
            )
        }
    }
    Box(
        Modifier
            .fillMaxWidth()
            .height(DesignTokens.Split.rule)
            .background(DesignTokens.Colors.rule(PANEL_AREA)),
    )
}

/**
 * One status and how many the ledger has of it. Tapping narrows the list to
 * that status; tapping the chosen one again goes back to 전체. It is a way of
 * looking, so it carries no label that suggests a change to the work.
 */
@Composable
private fun StatusPill(
    label: String,
    count: Int,
    selected: Boolean,
    tone: androidx.compose.ui.graphics.Color,
    tag: String,
    onClick: () -> Unit,
) {
    val shape = RoundedCornerShape(DesignTokens.Radius.chip)
    Text(
        "$label $count",
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = tone,
        modifier = Modifier
            .clip(shape)
            .background(tone.copy(alpha = if (selected) CHIP_SELECTED_OPACITY else CHIP_OPACITY))
            .then(if (selected) Modifier.border(CHIP_BORDER, tone, shape) else Modifier)
            .selectable(selected = selected, role = Role.RadioButton, onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = DesignTokens.Spacing.xs)
            .testTag(tag),
    )
}

/**
 * Hides or shows completed work. Display only — the label says so, because a
 * control next to a task list could otherwise read as finishing or removing
 * something.
 */
@Composable
private fun HideCompletedToggle(hidden: Boolean, onChange: (Boolean) -> Unit) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.chip)
    Text(
        if (hidden) "완료 표시하기" else "완료 숨기기",
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = ink.copy(alpha = if (hidden) 1f else MUTED_OPACITY),
        modifier = Modifier
            .clip(shape)
            .border(CHIP_BORDER, DesignTokens.Colors.rule(PANEL_AREA), shape)
            .toggleable(value = hidden, role = Role.Switch) { onChange(it) }
            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
            .testTag("chat.tasks.hideCompleted"),
    )
}

@Composable
private fun TaskList(tasks: List<ChatTask>, listState: LazyListState, onSelect: (ChatTask) -> Unit) {
    LazyColumn(
        state = listState,
        modifier = Modifier.fillMaxSize().testTag("chat.tasks.list"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        contentPadding = PaddingValues(
            horizontal = DesignTokens.Spacing.xl,
            vertical = DesignTokens.Spacing.lg,
        ),
    ) {
        items(tasks, key = { it.id }) { task -> TaskCard(task, onSelect) }
    }
}

@Composable
private fun TaskCard(task: ChatTask, onSelect: (ChatTask) -> Unit) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(PANEL_AREA))
            .border(DesignTokens.Split.rule, DesignTokens.Colors.rule(PANEL_AREA), shape)
            .clickable(role = Role.Button) { onSelect(task) }
            .padding(DesignTokens.Spacing.md)
            .testTag("chat.tasks.item"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Row(
            Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            verticalAlignment = Alignment.Top,
        ) {
            Text(
                task.title.ifBlank { task.id },
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = ink,
                modifier = Modifier.weight(1f).testTag("chat.tasks.item.title"),
            )
            StatusBadge(task.status)
        }
        if (task.description.isNotBlank()) {
            Text(
                task.description,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.tasks.item.description"),
            )
        }
        // A blocked or reviewing task is the one a person has to act on, so its
        // reason and next action are spelled out — and when the ledger recorded
        // neither, the panel says so instead of leaving a blank.
        // What the session is asking, and what it would take to be done. The
        // ledger's free-text reason and next action were the parent's own
        // bookkeeping; these are the records a person actually reads.
        if (ChatTaskStatus.needsAttention(task.status) ||
            task.waiting?.question?.isNotBlank() == true || task.completionCriteria.isNotBlank()
        ) {
            Note("묻는 것", task.waiting?.question.orEmpty(), "chat.tasks.item.question")
            Note("완료 기준", task.completionCriteria, "chat.tasks.item.criteria")
        }
    }
}

/** Where one task stands. Blocked and review carry the area's accent, because they are the two a person is meant to act on. */
@Composable
internal fun StatusBadge(status: String, tag: String = "chat.tasks.item.status.$status") {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val tone = if (ChatTaskStatus.needsAttention(status)) DesignTokens.Colors.accent(PANEL_AREA) else ink.copy(alpha = MUTED_OPACITY)
    Text(
        ChatTaskStatus.label(status),
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = tone,
        modifier = Modifier
            .clip(RoundedCornerShape(DesignTokens.Radius.chip))
            .background(tone.copy(alpha = CHIP_OPACITY))
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = DesignTokens.Spacing.xs)
            .testTag(tag),
    )
}

@Composable
private fun Note(label: String, value: String, tag: String) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val recorded = value.isNotBlank()
    Text(
        "$label · ${if (recorded) value else "미기록"}",
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = ink,
        modifier = Modifier
            .then(if (recorded) Modifier else Modifier.alpha(MUTED_OPACITY))
            .testTag(tag),
    )
}

@Composable
private fun PanelCentre(content: @Composable () -> Unit) {
    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { content() }
}

/**
 * Whether the window is wide enough for the two-pane layout
 * (docs/design-system/tokens.md §6.1). The caller passes the window's own
 * available width — not a device or orientation query — so a folded phone, a
 * split window and a small window are narrow, while an unfolded foldable is
 * wide even in portrait.
 */
fun isWideChatWindow(availableWidth: androidx.compose.ui.unit.Dp): Boolean =
    availableWidth >= DesignTokens.Breakpoints.wide
