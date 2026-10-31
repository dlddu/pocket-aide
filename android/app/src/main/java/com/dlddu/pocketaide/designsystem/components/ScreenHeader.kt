package com.dlddu.pocketaide.designsystem.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

private val DEFAULT_TITLE_SIZE = 24.sp
private val DEFAULT_LABEL_SPACING = 2.dp
private val TITLE_SUBTITLE_SPACING = 2.dp
private val TITLE_TRACKING = (-0.025).em

@Composable
fun ScreenHeader(
    area: Area,
    title: String,
    modifier: Modifier = Modifier,
    titleFamily: DesignTokens.Typography.Family = DesignTokens.Typography.Family.Sans,
    titleSize: TextUnit = DEFAULT_TITLE_SIZE,
    labelTracking: TextUnit = AreaLabelDefaults.tracking,
    labelShowsDot: Boolean = false,
    labelDotShape: AreaLabelDotShape = AreaLabelDotShape.Circle,
    labelDotSize: Dp = AreaLabelDefaults.dotSize,
    labelSpacing: Dp = DEFAULT_LABEL_SPACING,
    subtitle: String? = null,
    subtitleTag: String = "screen.header.subtitle",
    subtitleMonospaced: Boolean = false,
    bottomPadding: Dp = DesignTokens.Spacing.sm,
    labelAccessory: @Composable RowScope.() -> Unit = {},
    trailing: @Composable () -> Unit = {},
) {
    val ink = DesignTokens.Colors.ink(area)
    Row(
        modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = DesignTokens.Spacing.md,
                bottom = bottomPadding,
            ),
        verticalAlignment = Alignment.Bottom,
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(labelSpacing)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                AreaLabel(
                    area = area,
                    showsDot = labelShowsDot,
                    dotShape = labelDotShape,
                    dotSize = labelDotSize,
                    tracking = labelTracking,
                )
                labelAccessory()
            }
            Column(verticalArrangement = Arrangement.spacedBy(TITLE_SUBTITLE_SPACING)) {
                Text(
                    title,
                    style = DesignTokens.Typography.style(titleSize, FontWeight.Bold, titleFamily)
                        .copy(letterSpacing = TITLE_TRACKING),
                    color = ink,
                    modifier = Modifier.testTag("screen.header.title"),
                )
                if (subtitle != null) {
                    val base = DesignTokens.Typography.style(DesignTokens.Typography.captionSm)
                    Text(
                        subtitle,
                        style = if (subtitleMonospaced) base.copy(fontFamily = FontFamily.Monospace) else base,
                        color = ink.copy(alpha = 0.55f),
                        modifier = Modifier.testTag(subtitleTag),
                    )
                }
            }
        }
        Spacer(Modifier.weight(1f))
        trailing()
    }
}
