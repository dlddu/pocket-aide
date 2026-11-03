package com.dlddu.pocketaide.scratchpad

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
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
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.testTag as semanticsTestTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.dlddu.pocketaide.affirmations.PriorityEditMode
import com.dlddu.pocketaide.affirmations.PriorityEditSheet
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.ScreenHeader
import com.dlddu.pocketaide.designsystem.components.SwipeToDelete

private val AREA = Area.Scratchpad
private val TITLE_SIZE = 26.sp
private val LABEL_TRACKING = 2.2.sp
private val CARD_GAP = 10.dp
private val BADGE_HORIZONTAL = 6.dp
private val BADGE_VERTICAL = 2.dp
private val BADGE_RADIUS = 6.dp
private val ADD_ICON_SIZE = 13.dp
private val ADD_VERTICAL = 6.dp
private val DAY_HEADER_SIZE = 10.5.sp
private val DAY_HEADER_TRACKING = 2.1.sp
private val RULE_HEIGHT = 1.dp
private val SOURCE_DOT_SIZE = 6.dp
private val SOURCE_GAP = 6.dp
private val SOURCE_TRACKING = 0.5.sp
private val META_BOTTOM = 6.dp
private val BODY_LINE_HEIGHT = 1.625.em
private val CHIP_GAP = 6.dp
private val CHIP_HORIZONTAL = 10.dp
private val CHIP_BORDER = 1.dp
private val CARD_BORDER = 1.dp
private const val CHIP_FILL_OPACITY = 0.6f
private const val TABULAR_NUMS = "tnum"

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ScratchpadScreen(viewModel: ScratchpadViewModel) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val unclassified by viewModel.unclassifiedCount.collectAsStateWithLifecycle()
    var showsAddSheet by remember { mutableStateOf(false) }
    val sections = state.sections()
    val ink = DesignTokens.Colors.ink(AREA)

    LaunchedEffect(viewModel) { viewModel.load() }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(AREA))
            .testTag("scratchpad.screen"),
    ) {
        Column(Modifier.fillMaxSize()) {
            ScreenHeader(
                area = AREA,
                title = "분류되지 않은 메모",
                titleSize = TITLE_SIZE,
                labelTracking = LABEL_TRACKING,
                subtitle = "캡처 부담 없이 일단 던져두는 곳",
                bottomPadding = DesignTokens.Spacing.md,
                labelAccessory = {
                    Text(
                        "$unclassified",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold)
                            .copy(fontFeatureSettings = TABULAR_NUMS),
                        color = Color.White,
                        modifier = Modifier
                            .padding(start = DesignTokens.Spacing.sm)
                            .clip(RoundedCornerShape(BADGE_RADIUS))
                            .background(DesignTokens.Colors.accent(AREA))
                            .padding(horizontal = BADGE_HORIZONTAL, vertical = BADGE_VERTICAL)
                            .testTag("scratchpad.badge"),
                    )
                },
                trailing = { AddButton(onClick = { showsAddSheet = true }) },
            )
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                LazyColumn(Modifier.fillMaxSize()) {
                    state.errorMessage?.let { message ->
                        item(key = "error") {
                            Text(
                                message,
                                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                                color = DesignTokens.Colors.destructive(AREA),
                                modifier = Modifier
                                    .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.sm)
                                    .testTag("scratchpad.error"),
                            )
                        }
                    }
                    if (state.items.isEmpty()) {
                        item(key = "empty") {
                            Box(
                                Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xl),
                                contentAlignment = if (state.isLoading) Alignment.Center else Alignment.CenterStart,
                            ) {
                                if (state.isLoading) {
                                    CircularProgressIndicator(color = DesignTokens.Colors.accent(AREA))
                                } else {
                                    Text(
                                        "분류할 메모가 없습니다. 떠오르는 대로 새 메모에 던져두세요.",
                                        style = DesignTokens.Typography.style(DesignTokens.Typography.body)
                                            .copy(lineHeight = BODY_LINE_HEIGHT),
                                        color = ink.copy(alpha = 0.55f),
                                        modifier = Modifier.testTag("scratchpad.empty.state"),
                                    )
                                }
                            }
                        }
                    } else {
                        val lastId = sections.lastOrNull()?.items?.lastOrNull()?.id
                        sections.forEachIndexed { index, section ->
                            daySection(
                                section = section,
                                index = index,
                                lastId = lastId,
                                onMove = viewModel::move,
                                onDelete = { viewModel.delete(it.id) },
                            )
                        }
                    }
                }
            }
        }
    }

    if (showsAddSheet) {
        ScratchpadAddSheet(
            onSave = { text ->
                showsAddSheet = false
                viewModel.add(text)
            },
            onCancel = { showsAddSheet = false },
        )
    }

    state.priorityEdit?.let { affirmation ->
        PriorityEditSheet(
            mode = PriorityEditMode.Edit(affirmation),
            onSave = { text, priority ->
                viewModel.dismissPriorityEdit()
                viewModel.setPriority(affirmation, text, priority)
            },
            onCancel = viewModel::dismissPriorityEdit,
        )
    }
}

@Composable
private fun AddButton(onClick: () -> Unit) {
    val surface = DesignTokens.Colors.surface(AREA)
    Row(
        Modifier
            .clip(CircleShape)
            .background(DesignTokens.Colors.ink(AREA))
            .clickable(role = Role.Button, onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ADD_VERTICAL)
            .testTag("scratchpad.add.button"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(Icons.Filled.Add, contentDescription = null, tint = surface, modifier = Modifier.size(ADD_ICON_SIZE))
        Text(
            "새 메모",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
            color = surface,
        )
    }
}

private fun LazyListScope.daySection(
    section: ScratchpadSection,
    index: Int,
    lastId: Long?,
    onMove: (ScratchpadItem, ScratchpadMoveTarget) -> Unit,
    onDelete: (ScratchpadItem) -> Unit,
) {
    item(key = "day.$index.${section.id}") {
        DayHeader(section = section, isFirst = index == 0)
    }
    items(section.items, key = { it.id }) { item ->
        val shape = RoundedCornerShape(DesignTokens.Radius.card)
        SwipeToDelete(
            area = AREA,
            shape = shape,
            onDelete = { onDelete(item) },
            modifier = Modifier.padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = CARD_GAP / 2,
                bottom = if (item.id == lastId) DesignTokens.Spacing.xxl else CARD_GAP / 2,
            ),
        ) {
            ScratchpadCard(item = item, onMove = { target -> onMove(item, target) })
        }
    }
}

@Composable
private fun DayHeader(section: ScratchpadSection, isFirst: Boolean) {
    val ink = DesignTokens.Colors.ink(AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = if (isFirst) DesignTokens.Spacing.xs + DesignTokens.Spacing.xs else CARD_GAP / 2 + DesignTokens.Spacing.md,
                bottom = CARD_GAP / 2,
            ),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            section.title.uppercase(),
            style = DesignTokens.Typography.style(DAY_HEADER_SIZE, FontWeight.SemiBold)
                .copy(letterSpacing = DAY_HEADER_TRACKING),
            color = ink.copy(alpha = 0.7f),
        )
        Box(Modifier.weight(1f).height(RULE_HEIGHT).background(DesignTokens.Colors.rule(AREA)))
        Text(
            "${section.items.size} ITEMS",
            style = DesignTokens.Typography.style(DAY_HEADER_SIZE),
            color = ink.copy(alpha = 0.5f),
        )
    }
}

@Composable
private fun ScratchpadCard(item: ScratchpadItem, onMove: (ScratchpadMoveTarget) -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    val rule = DesignTokens.Colors.rule(AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    val time = ScratchpadSections.timeLabel(item)
    val isShortcut = item.source == ScratchpadSource.SHORTCUT
    val chipFill = (if (isSystemInDarkTheme()) DesignTokens.Colors.card(AREA) else Color.White)
        .copy(alpha = CHIP_FILL_OPACITY)
    Column(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.surface(AREA), shape)
            .clip(shape)
            .background(DesignTokens.Colors.card(AREA))
            .border(CARD_BORDER, rule, shape)
            .padding(DesignTokens.Spacing.lg)
            .testTag("scratchpad.row.${item.id}"),
    ) {
        Row(
            Modifier
                .fillMaxWidth()
                .padding(bottom = META_BOTTOM)
                .clearAndSetSemantics {
                    contentDescription = "${item.source.displayName} · $time"
                    semanticsTestTag = "scratchpad.row.${item.id}.meta"
                },
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
            verticalAlignment = Alignment.Top,
        ) {
            Row(horizontalArrangement = Arrangement.spacedBy(SOURCE_GAP), verticalAlignment = Alignment.CenterVertically) {
                Box(
                    Modifier
                        .size(SOURCE_DOT_SIZE)
                        .background(if (isShortcut) accent else ink.copy(alpha = 0.55f), CircleShape),
                )
                Text(
                    item.source.displayName.uppercase(),
                    style = DesignTokens.Typography.style(DAY_HEADER_SIZE, FontWeight.SemiBold)
                        .copy(letterSpacing = SOURCE_TRACKING),
                    color = if (isShortcut) accent else ink.copy(alpha = 0.7f),
                )
            }
            Spacer(Modifier.weight(1f))
            Text(
                time,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = 0.55f),
            )
        }
        Text(
            item.text,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg).copy(lineHeight = BODY_LINE_HEIGHT),
            color = ink,
        )
        Row(
            Modifier.padding(top = DesignTokens.Spacing.md),
            horizontalArrangement = Arrangement.spacedBy(CHIP_GAP),
        ) {
            ScratchpadMoveTarget.entries.forEach { target ->
                Text(
                    target.chipLabel,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Medium),
                    color = ink,
                    maxLines = 1,
                    modifier = Modifier
                        .clip(CircleShape)
                        .background(chipFill)
                        .border(CHIP_BORDER, rule, CircleShape)
                        .clickable(role = Role.Button) { onMove(target) }
                        .padding(horizontal = CHIP_HORIZONTAL, vertical = DesignTokens.Spacing.xs)
                        .testTag("scratchpad.row.${item.id}.move.${target.wire}"),
                )
            }
        }
    }
}
