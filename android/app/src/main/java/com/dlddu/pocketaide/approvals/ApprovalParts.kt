package com.dlddu.pocketaide.approvals

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.ArrowBack
import androidx.compose.material.icons.outlined.ErrorOutline
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.produceState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import kotlinx.coroutines.delay

internal val APPROVAL_AREA = Area.PrMonitor

internal object ApprovalMetrics {
    val chipSize: TextUnit = 10.5.sp
    val rowTitleSize: TextUnit = 13.5.sp
    val headerTitleSize: TextUnit = 26.sp
    val detailTitleSize: TextUnit = 18.sp
    val sheetTitleSize: TextUnit = 17.sp
    val noticeSize: TextUnit = 11.5.sp
    val badgeCountSize: TextUnit = 15.sp
    val badgeLabelSize: TextUnit = 8.sp
    val loadingMoreSize: TextUnit = 11.sp
    val chipRadius: Dp = 2.dp
    val buttonRadius: Dp = 10.dp
    val rowPadding: Dp = 14.dp
    val bannerVertical: Dp = 10.dp
    val rowGap: Dp = 5.dp
    val rowSpacing: Dp = 6.dp
    val ruleWidth: Dp = 1.dp
    val accentStripWidth: Dp = 3.dp
    val sectionMarkWidth: Dp = 4.dp
    val sectionMarkHeight: Dp = 12.dp
    val iconButtonSize: Dp = 36.dp
    val iconSize: Dp = 18.dp
    val badgeHorizontal: Dp = 10.dp
    val chipVertical: Dp = 2.dp
    val dashOn: Dp = 4.dp
    val dashOff: Dp = 3.dp
    val sectionTracking: TextUnit = 0.3.sp
    val noticeLineHeight: TextUnit = 18.sp
    const val SECONDARY_OPACITY = 0.55f
    const val FAINT_OPACITY = 0.45f
    const val DISABLED_OPACITY = 0.35f
    const val REVOKED_OPACITY = 0.6f
    const val TICK_MILLIS = 1_000L
}

internal fun monoStyle(size: TextUnit, weight: FontWeight = FontWeight.Normal): TextStyle =
    DesignTokens.Typography.style(size, weight).copy(fontFamily = FontFamily.Monospace, fontFeatureSettings = "tnum")

@Composable
internal fun rememberNowMillis(): Long {
    val now by produceState(System.currentTimeMillis()) {
        while (true) {
            delay(ApprovalMetrics.TICK_MILLIS)
            value = System.currentTimeMillis()
        }
    }
    return now
}

@Composable
internal fun CallerChip(text: String, modifier: Modifier = Modifier, revoked: Boolean = false) {
    val area = APPROVAL_AREA
    val shape = RoundedCornerShape(ApprovalMetrics.chipRadius)
    val ink = DesignTokens.Colors.ink(area)
    Text(
        text,
        style = monoStyle(ApprovalMetrics.chipSize, FontWeight.Bold).copy(
            textDecoration = if (revoked) TextDecoration.LineThrough else null,
        ),
        color = if (revoked) ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY) else DesignTokens.StatusColors.arrivalGlow(),
        maxLines = 1,
        overflow = TextOverflow.Ellipsis,
        modifier = modifier
            .clip(shape)
            .background(if (revoked) Color.Transparent else DesignTokens.Colors.soft(area))
            .border(ApprovalMetrics.ruleWidth, if (revoked) DesignTokens.Colors.rule(area) else Color.Transparent, shape)
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = ApprovalMetrics.chipVertical),
    )
}

@Composable
internal fun ApprovalSectionHeader(title: String, tint: Color, modifier: Modifier = Modifier, tag: String? = null) {
    Row(
        modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier
                .size(width = ApprovalMetrics.sectionMarkWidth, height = ApprovalMetrics.sectionMarkHeight)
                .background(tint),
        )
        Text(
            title,
            style = monoStyle(DesignTokens.Typography.captionXs, FontWeight.Bold).copy(letterSpacing = ApprovalMetrics.sectionTracking),
            color = tint,
            modifier = if (tag != null) Modifier.testTag(tag) else Modifier,
        )
        Box(Modifier.weight(1f).height(ApprovalMetrics.ruleWidth).background(DesignTokens.Colors.rule(APPROVAL_AREA)))
    }
}

@Composable
internal fun ApprovalErrorBanner(
    title: String?,
    message: String?,
    bannerTag: String?,
    retryTag: String?,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
    showsIcon: Boolean = true,
    outlined: Boolean = true,
) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val accent = DesignTokens.Colors.accent(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Row(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.soft(area))
            .border(ApprovalMetrics.ruleWidth, if (outlined) accent else Color.Transparent, shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
            .then(if (bannerTag != null) Modifier.testTag(bannerTag) else Modifier),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.Top,
    ) {
        if (showsIcon) {
            Icon(
                Icons.Outlined.ErrorOutline,
                contentDescription = null,
                tint = DesignTokens.StatusColors.failure(),
                modifier = Modifier.size(ApprovalMetrics.iconSize),
            )
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.chipVertical)) {
            if (title != null) {
                Text(
                    title,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                    color = ink,
                )
            }
            if (message != null) {
                Text(
                    message,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                    color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                    maxLines = 2,
                )
            }
        }
        Text(
            "다시 시도",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
            color = accent,
            modifier = Modifier
                .clickable(role = Role.Button, onClick = onRetry)
                .then(if (retryTag != null) Modifier.testTag(retryTag) else Modifier),
        )
    }
}

@Composable
internal fun ApprovalBackBar(onBack: () -> Unit, trailing: @Composable RowScope.() -> Unit = {}) {
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = DesignTokens.Spacing.xs),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier
                .size(ApprovalMetrics.iconButtonSize)
                .clip(RoundedCornerShape(DesignTokens.Radius.key))
                .clickable(role = Role.Button, onClick = onBack)
                .testTag("approvals.back"),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                Icons.AutoMirrored.Outlined.ArrowBack,
                contentDescription = "뒤로",
                tint = DesignTokens.Colors.accent(APPROVAL_AREA),
            )
        }
        Spacer(Modifier.weight(1f))
        trailing()
        Spacer(Modifier.width(DesignTokens.Spacing.md))
    }
}

@Composable
internal fun ApprovalButton(
    title: String,
    primary: Boolean,
    tag: String,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    outlineAccent: Boolean = false,
    onClick: () -> Unit,
) {
    val area = APPROVAL_AREA
    val accent = DesignTokens.Colors.accent(area)
    val shape = RoundedCornerShape(ApprovalMetrics.buttonRadius)
    val outline = when {
        primary -> accent
        outlineAccent -> accent
        else -> DesignTokens.Colors.rule(area)
    }
    Box(
        modifier
            .alpha(if (enabled) 1f else ApprovalMetrics.DISABLED_OPACITY)
            .clip(shape)
            .background(if (primary) accent else DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, outline, shape)
            .clickable(enabled = enabled, role = Role.Button, onClick = onClick)
            .padding(vertical = DesignTokens.Spacing.md)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            title,
            style = DesignTokens.Typography.style(
                DesignTokens.Typography.bodyLg,
                if (primary || outlineAccent) FontWeight.Bold else FontWeight.Normal,
            ),
            color = when {
                primary -> Color.White
                outlineAccent -> accent
                else -> DesignTokens.Colors.ink(area)
            },
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
internal fun DashedActionButton(title: String, tag: String, modifier: Modifier = Modifier, onClick: () -> Unit) {
    val accent = DesignTokens.Colors.accent(APPROVAL_AREA)
    val shape = RoundedCornerShape(ApprovalMetrics.buttonRadius)
    Box(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .drawBehind {
                val inset = ApprovalMetrics.ruleWidth.toPx() / 2
                drawRoundRect(
                    color = accent,
                    topLeft = Offset(inset, inset),
                    size = Size(size.width - inset * 2, size.height - inset * 2),
                    cornerRadius = CornerRadius(ApprovalMetrics.buttonRadius.toPx()),
                    style = Stroke(
                        width = ApprovalMetrics.ruleWidth.toPx(),
                        pathEffect = PathEffect.dashPathEffect(
                            floatArrayOf(ApprovalMetrics.dashOn.toPx(), ApprovalMetrics.dashOff.toPx()),
                        ),
                    ),
                )
            }
            .clickable(role = Role.Button, onClick = onClick)
            .padding(vertical = DesignTokens.Spacing.md)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            title,
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
            color = accent,
        )
    }
}

@Composable
internal fun ApprovalConfirmDialog(
    title: String,
    message: String,
    confirmLabel: String,
    destructive: Boolean,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    AlertDialog(
        onDismissRequest = onDismiss,
        containerColor = DesignTokens.Colors.card(area),
        titleContentColor = ink,
        textContentColor = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        title = {
            Text(title, style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.Bold))
        },
        text = {
            Text(message, style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm))
        },
        dismissButton = {
            TextButton(onClick = onDismiss, modifier = Modifier.testTag("approvals.confirm.cancel")) {
                Text(
                    "취소",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = DesignTokens.Colors.accent(area),
                )
            }
        },
        confirmButton = {
            TextButton(
                onClick = {
                    onDismiss()
                    onConfirm()
                },
                modifier = Modifier.testTag("approvals.confirm.ok"),
            ) {
                Text(
                    confirmLabel,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
                    color = if (destructive) DesignTokens.StatusColors.failure() else DesignTokens.Colors.accent(area),
                )
            }
        },
    )
}
