package com.dlddu.pocketaide.designsystem.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

enum class AreaLabelDotShape { Circle, Square }

object AreaLabelDefaults {
    val dotSize: Dp = 6.dp
    val tracking: TextUnit = 2.4.sp
    internal val dotGap: Dp = 6.dp

    fun text(area: Area): String = when (area) {
        Area.Personal -> "Personal"
        Area.Work -> "Work"
        Area.AiChat -> "채팅"
        Area.Scratchpad -> "임시 공간"
        Area.Routines -> "루틴"
        Area.Affirmations -> "다짐"
        Area.Voice -> "Voice"
        Area.System -> "System"
        Area.PrMonitor -> "PR · Monitor"
    }
}

@Composable
fun AreaLabel(
    area: Area,
    modifier: Modifier = Modifier,
    text: String? = null,
    showsDot: Boolean = false,
    dotShape: AreaLabelDotShape = AreaLabelDotShape.Circle,
    dotSize: Dp = AreaLabelDefaults.dotSize,
    tracking: TextUnit = AreaLabelDefaults.tracking,
) {
    val accent = DesignTokens.Colors.accent(area)
    Row(
        modifier,
        horizontalArrangement = Arrangement.spacedBy(AreaLabelDefaults.dotGap),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (showsDot) {
            val shape = if (dotShape == AreaLabelDotShape.Circle) CircleShape else RectangleShape
            Box(Modifier.size(dotSize).background(accent, shape))
        }
        Text(
            (text ?: AreaLabelDefaults.text(area)).uppercase(),
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold)
                .copy(letterSpacing = tracking),
            color = accent,
        )
    }
}
