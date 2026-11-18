package com.dlddu.pocketaide.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
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

/**
 * The task ledger of the session answering this conversation. Read-only: the
 * panel shows what the ledger says and offers nothing that changes it.
 */
@Composable
fun ChatTaskPanel(state: ChatTasksUiState, modifier: Modifier = Modifier) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    Column(
        modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(PANEL_AREA))
            .testTag("chat.tasks.panel"),
    ) {
        PanelHeader(state)
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

            state.isEmpty -> PanelCentre {
                Text(
                    "아직 기록된 작업이 없어요.",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.padding(DesignTokens.Spacing.xl).testTag("chat.tasks.empty"),
                )
            }

            else -> TaskList(state)
        }
    }
}

@Composable
private fun PanelHeader(state: ChatTasksUiState) {
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
        // Counts for the statuses the ledger actually has, in display order, so
        // what needs a person is read first.
        val present = ChatTaskStatus.order.filter { (state.counts[it] ?: 0) > 0 }
        if (present.isNotEmpty()) {
            Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
                present.forEach { status -> CountChip(status, state.counts[status] ?: 0) }
            }
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

@Composable
private fun CountChip(status: String, count: Int) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val tone = if (ChatTaskStatus.needsAttention(status)) DesignTokens.Colors.accent(PANEL_AREA) else ink
    Text(
        "${ChatTaskStatus.label(status)} $count",
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = tone,
        modifier = Modifier
            .clip(RoundedCornerShape(DesignTokens.Radius.chip))
            .background(tone.copy(alpha = CHIP_OPACITY))
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = DesignTokens.Spacing.xs)
            .testTag("chat.tasks.count.$status"),
    )
}

@Composable
private fun TaskList(state: ChatTasksUiState) {
    // The panel keeps its own scroll position, so collapsing and reopening it
    // does not move the conversation beside it.
    val listState = rememberLazyListState()
    LazyColumn(
        state = listState,
        modifier = Modifier.fillMaxSize().testTag("chat.tasks.list"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        contentPadding = PaddingValues(
            horizontal = DesignTokens.Spacing.xl,
            vertical = DesignTokens.Spacing.lg,
        ),
    ) {
        items(state.tasks, key = { it.id }) { task -> TaskCard(task) }
    }
}

@Composable
private fun TaskCard(task: ChatTask) {
    val ink = DesignTokens.Colors.ink(PANEL_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(PANEL_AREA))
            .border(DesignTokens.Split.rule, DesignTokens.Colors.rule(PANEL_AREA), shape)
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
        if (task.stage.isNotBlank()) {
            Text(
                task.stage,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                color = ink.copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.tasks.item.stage"),
            )
        }
        // A blocked or reviewing task is the one a person has to act on, so its
        // reason and next action are spelled out — and when the ledger recorded
        // neither, the panel says so instead of leaving a blank.
        if (ChatTaskStatus.needsAttention(task.status) || task.reason.isNotBlank() || task.nextAction.isNotBlank()) {
            Note("사유", task.reason, "chat.tasks.item.reason")
            Note("다음", task.nextAction, "chat.tasks.item.next")
        }
    }
}

/** Where one task stands. Blocked and review carry the area's accent, because they are the two a person is meant to act on. */
@Composable
private fun StatusBadge(status: String) {
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
            .testTag("chat.tasks.item.status.$status"),
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
