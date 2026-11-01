package com.dlddu.pocketaide.routines

import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
fun RoutineHistoryRange(days: List<RoutineHistoryDay>, modifier: Modifier = Modifier) {
    val first = days.firstOrNull() ?: return
    val today = days.last()
    val style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs)
    val color = DesignTokens.Colors.ink(Area.Routines).copy(alpha = 0.55f)
    Row(modifier.fillMaxWidth()) {
        Text(RoutineDayFormat.title(first.day), style = style, color = color)
        Spacer(Modifier.weight(1f))
        if (today.scheduled) {
            Text(
                "오늘 · ${today.progressPercent}%",
                style = style,
                color = color,
                modifier = Modifier.testTag("routines.history.today"),
            )
        }
    }
}
