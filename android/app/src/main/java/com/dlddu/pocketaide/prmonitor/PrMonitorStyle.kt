package com.dlddu.pocketaide.prmonitor

import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.lerp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale

internal val PR_AREA = Area.PrMonitor

internal object PrMonitorStyle {
    const val SECONDARY_OPACITY = 0.6f
    const val TERTIARY_OPACITY = 0.4f
    const val MUTED_OPACITY = 0.55f
    const val ACKED_SURFACE_OPACITY = 0.6f
    val accentStripWidth = 3.dp
    val borderWidth = 1.dp
    val chipRadius = 6.dp
    val controlRadius = 8.dp
    val chipVertical = 4.dp
    val chipIconSize = 9.dp
    val chipIconGap = 4.dp
    val dashOn = 4.dp
    val dashOff = 3.dp

    private val timeFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("HH:mm", Locale.KOREA)

    fun clock(epochSeconds: Long, zone: ZoneId = ZoneId.systemDefault()): String =
        Instant.ofEpochSecond(epochSeconds).atZone(zone).format(timeFormatter)

    @Composable
    fun ciColor(outcome: CiOutcome): Color = when (outcome) {
        CiOutcome.SUCCESS -> DesignTokens.StatusColors.success()
        CiOutcome.IN_PROGRESS -> DesignTokens.StatusColors.inProgress()
        CiOutcome.FAILURE -> DesignTokens.StatusColors.failure()
    }
}

internal fun Modifier.accentStrip(color: Color, width: Dp): Modifier = drawWithContent {
    drawContent()
    drawRect(color, size = Size(width.toPx(), size.height))
}

internal fun Modifier.dashedBorder(color: Color, width: Dp, radius: Dp, on: Dp, off: Dp): Modifier = drawBehind {
    val stroke = width.toPx()
    drawRoundRect(
        color = color,
        topLeft = Offset(stroke / 2, stroke / 2),
        size = Size(size.width - stroke, size.height - stroke),
        cornerRadius = CornerRadius(radius.toPx()),
        style = Stroke(width = stroke, pathEffect = PathEffect.dashPathEffect(floatArrayOf(on.toPx(), off.toPx()))),
    )
}

internal fun Modifier.cardFrame(acknowledged: Boolean, radius: Dp, card: Color, surface: Color, rule: Color, accent: Color): Modifier {
    val shape = RoundedCornerShape(radius)
    return if (acknowledged) {
        clip(shape)
            .background(surface.copy(alpha = PrMonitorStyle.ACKED_SURFACE_OPACITY))
            .dashedBorder(rule, PrMonitorStyle.borderWidth, radius, PrMonitorStyle.dashOn, PrMonitorStyle.dashOff)
    } else {
        clip(shape)
            .background(card)
            .accentStrip(accent, PrMonitorStyle.accentStripWidth)
            .border(PrMonitorStyle.borderWidth, rule, shape)
    }
}

private const val PULSE_MILLIS = 1200
private const val PULSE_ALPHA_LOW = 0.55f
private const val PULSE_ALPHA_HIGH = 0.95f
private const val PULSE_SHADOW_LOW = 0.3f
private const val PULSE_SHADOW_HIGH = 0.6f
private val PULSE_WIDTH_LOW = 3.dp
private val PULSE_WIDTH_HIGH = 6.dp
private val PULSE_ELEVATION_LOW = 10.dp
private val PULSE_ELEVATION_HIGH = 22.dp
private const val STATIC_GLOW_ALPHA = 0.16f
private const val STATIC_SHADOW_ALPHA = 0.35f
private val STATIC_GLOW_WIDTH = 4.dp
private val STATIC_GLOW_ELEVATION = 12.dp

@Composable
internal fun Modifier.pulsingArrivalGlow(active: Boolean, shape: Shape): Modifier {
    if (!active) return this
    val glow = DesignTokens.StatusColors.arrivalGlow()
    val transition = rememberInfiniteTransition(label = "prmonitor.arrival")
    val progress by transition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(tween(PULSE_MILLIS, easing = FastOutSlowInEasing), RepeatMode.Reverse),
        label = "prmonitor.arrival.progress",
    )
    val shadowColor = glow.copy(alpha = PULSE_SHADOW_LOW + (PULSE_SHADOW_HIGH - PULSE_SHADOW_LOW) * progress)
    return this
        .shadow(lerp(PULSE_ELEVATION_LOW, PULSE_ELEVATION_HIGH, progress), shape, clip = false, ambientColor = shadowColor, spotColor = shadowColor)
        .border(
            lerp(PULSE_WIDTH_LOW, PULSE_WIDTH_HIGH, progress),
            glow.copy(alpha = PULSE_ALPHA_LOW + (PULSE_ALPHA_HIGH - PULSE_ALPHA_LOW) * progress),
            shape,
        )
}

@Composable
internal fun Modifier.staticArrivalGlow(active: Boolean, shape: Shape): Modifier {
    if (!active) return this
    val glow = DesignTokens.StatusColors.arrivalGlow()
    val shadowColor = glow.copy(alpha = STATIC_SHADOW_ALPHA)
    return this
        .shadow(STATIC_GLOW_ELEVATION, shape, clip = false, ambientColor = shadowColor, spotColor = shadowColor)
        .border(STATIC_GLOW_WIDTH, glow.copy(alpha = STATIC_GLOW_ALPHA), shape)
}

@Composable
internal fun PrLinkChip(label: String, icon: ImageVector, url: String?, tag: String) {
    val target = url?.takeIf { it.isNotBlank() } ?: return
    val uriHandler = LocalUriHandler.current
    val ink = DesignTokens.Colors.ink(PR_AREA).copy(alpha = PrMonitorStyle.SECONDARY_OPACITY)
    val shape = RoundedCornerShape(PrMonitorStyle.chipRadius)
    Row(
        Modifier
            .clip(shape)
            .border(PrMonitorStyle.borderWidth, DesignTokens.Colors.rule(PR_AREA), shape)
            .clickable(role = Role.Button) { runCatching { uriHandler.openUri(target) } }
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = PrMonitorStyle.chipVertical)
            .testTag(tag),
        horizontalArrangement = Arrangement.spacedBy(PrMonitorStyle.chipIconGap),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(icon, contentDescription = null, tint = ink, modifier = Modifier.size(PrMonitorStyle.chipIconSize))
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
            color = ink,
        )
    }
}
