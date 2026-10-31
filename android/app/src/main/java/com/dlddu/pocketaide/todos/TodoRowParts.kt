package com.dlddu.pocketaide.todos

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.DesignTokens

private val CHECK_STROKE = 1.6.dp
private val CHECKMARK_SIZE = 12.dp
private val PRIORITY_TAG_SIZE = 10.sp
private val PRIORITY_TAG_TRACKING = 0.5.sp
private val PRIORITY_TAG_BOTTOM = 2.dp
private val CHIP_GAP = 6.dp
private val CHIP_BORDER = 1.dp

@Composable
internal fun TodoCheckBox(area: TodoArea, isDone: Boolean, modifier: Modifier = Modifier) {
    val accent = DesignTokens.Colors.accent(area.designArea)
    val shape = if (area.checkIsRound) CircleShape else RectangleShape
    Box(
        modifier
            .size(area.checkSize)
            .clip(shape)
            .background(if (isDone) accent else Color.Transparent)
            .border(CHECK_STROKE, accent, shape),
        contentAlignment = Alignment.Center,
    ) {
        if (isDone) {
            Icon(
                Icons.Filled.Check,
                contentDescription = null,
                tint = Color.White,
                modifier = Modifier.size(CHECKMARK_SIZE),
            )
        }
    }
}

@Composable
internal fun TodoPriorityTag(area: TodoArea, item: TodoItem) {
    Text(
        TodoPriority.HIGH.displayName.uppercase(),
        style = DesignTokens.Typography.style(PRIORITY_TAG_SIZE, FontWeight.Bold)
            .copy(letterSpacing = PRIORITY_TAG_TRACKING),
        color = DesignTokens.Colors.accent(area.designArea),
        modifier = Modifier
            .padding(bottom = PRIORITY_TAG_BOTTOM)
            .testTag("todos.${area.wire}.row.${item.id}.priority"),
    )
}

@Composable
internal fun TodoDueFilterChips(
    area: TodoArea,
    total: Int,
    selection: TodoDueFilter,
    onSelect: (TodoDueFilter) -> Unit,
) {
    val tone = area.designArea
    val ink = DesignTokens.Colors.ink(tone)
    val soft = DesignTokens.Colors.soft(tone)
    val rule = DesignTokens.Colors.rule(tone)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(bottom = DesignTokens.Spacing.md)
            .horizontalScroll(rememberScrollState())
            .padding(horizontal = DesignTokens.Spacing.xl),
        horizontalArrangement = Arrangement.spacedBy(CHIP_GAP),
    ) {
        TodoDueFilter.entries.forEach { filter ->
            val selected = filter == selection
            Text(
                filter.chipTitle(total),
                style = DesignTokens.Typography.style(
                    DesignTokens.Typography.captionSm,
                    if (selected) FontWeight.SemiBold else FontWeight.Medium,
                ),
                color = if (selected) ink else ink.copy(alpha = 0.7f),
                maxLines = 1,
                modifier = Modifier
                    .clip(CircleShape)
                    .background(if (selected) soft else Color.Transparent)
                    .border(CHIP_BORDER, if (selected) Color.Transparent else rule, CircleShape)
                    .selectable(selected = selected, role = Role.Tab, onClick = { onSelect(filter) })
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
                    .testTag("todos.${area.wire}.filter.${filter.wire}"),
            )
        }
    }
}
