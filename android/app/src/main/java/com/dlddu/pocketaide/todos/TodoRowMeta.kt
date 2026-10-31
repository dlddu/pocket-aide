package com.dlddu.pocketaide.todos

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.CalendarToday
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.testTag
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens

private val CALENDAR_ICON_SIZE = 11.dp

@Composable
internal fun TodoRowMeta(area: TodoArea, item: TodoItem) {
    val parts = TodoRowMetaText.visibleParts(area, item)
    if (parts.isEmpty()) return
    val label = TodoRowMetaText.labelParts(area, item).joinToString(" · ")
    val tag = "todos.${area.wire}.row.${item.id}.meta"
    val color = DesignTokens.Colors.ink(area.designArea).copy(alpha = 0.55f)
    val base = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)
    val style: TextStyle = if (area == TodoArea.WORK) base.copy(fontFamily = FontFamily.Monospace) else base
    Row(
        Modifier.clearAndSetSemantics {
            contentDescription = label
            testTag = tag
        },
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        parts.forEachIndexed { index, part ->
            val text = if (area == TodoArea.WORK) part.uppercase() else part
            if (area == TodoArea.WORK && index > 0) {
                Text("·", style = style, color = color)
            }
            if (area == TodoArea.PERSONAL && index == 0 && item.dueDate != null) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Icon(
                        Icons.Outlined.CalendarToday,
                        contentDescription = null,
                        tint = color,
                        modifier = Modifier.size(CALENDAR_ICON_SIZE),
                    )
                    Text(text, style = style, color = color)
                }
            } else {
                Text(text, style = style, color = color)
            }
        }
    }
}
