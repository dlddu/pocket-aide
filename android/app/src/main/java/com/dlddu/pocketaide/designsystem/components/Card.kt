package com.dlddu.pocketaide.designsystem.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

private val SMALL_PADDING = 14.dp
private val LARGE_RADIUS = 28.dp
private val BORDER_WIDTH = 1.dp
private val EMPHASIZED_BORDER_WIDTH = 2.dp

enum class CardPadding(val value: Dp, val radius: Dp) {
    Small(SMALL_PADDING, DesignTokens.Radius.card),
    Medium(DesignTokens.Spacing.lg, DesignTokens.Radius.card),
    Large(DesignTokens.Spacing.xxl, LARGE_RADIUS),
}

@Composable
fun Card(
    area: Area,
    modifier: Modifier = Modifier,
    padding: CardPadding = CardPadding.Medium,
    emphasized: Boolean = false,
    content: @Composable () -> Unit,
) {
    val shape = RoundedCornerShape(padding.radius)
    Box(
        modifier
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(
                width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH,
                color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area),
                shape = shape,
            )
            .padding(padding.value),
    ) {
        content()
    }
}
