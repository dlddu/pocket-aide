package com.dlddu.pocketaide.prmonitor

import android.text.format.DateUtils
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Circle
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens

private val ROW_RADIUS = 12.dp
private val STATUS_BADGE_SIZE = 18.dp
private val BADGE_ICON_SIZE = 11.dp
private val ACK_VERTICAL = 6.dp
private val ACK_ICON_SIZE = 12.dp
private const val ACKED_ROW_OPACITY = 0.7f

@Composable
internal fun PrMonitorHistoryRow(
    item: NotificationHistoryItem,
    isHighlighted: Boolean,
    onAcknowledge: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val acked = item.isAcknowledged
    val shape = RoundedCornerShape(ROW_RADIUS)
    Column(
        modifier
            .fillMaxWidth()
            .pulsingArrivalGlow(isHighlighted, shape)
            .cardFrame(
                acknowledged = acked,
                radius = ROW_RADIUS,
                card = DesignTokens.Colors.card(PR_AREA),
                surface = DesignTokens.Colors.surface(PR_AREA),
                rule = DesignTokens.Colors.rule(PR_AREA),
                accent = DesignTokens.Colors.accent(PR_AREA),
            )
            .alpha(if (acked) ACKED_ROW_OPACITY else 1f)
            .padding(DesignTokens.Spacing.md),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        RowHeader(item)
        RowTitle(item)
        RowMeta(item)
        if (!acked) RowActions(item, onAcknowledge)
    }
}

@Composable
private fun RowHeader(item: NotificationHistoryItem) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val outcome = CiConclusion.outcome(item.conclusion)
    val statusColor = PrMonitorStyle.ciColor(outcome)
    val verdict = CiConclusion.verdictLabel(item.conclusion)
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier.size(STATUS_BADGE_SIZE).background(statusColor, CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                when (outcome) {
                    CiOutcome.SUCCESS -> Icons.Filled.Check
                    CiOutcome.IN_PROGRESS -> Icons.Filled.Schedule
                    CiOutcome.FAILURE -> Icons.Filled.Close
                },
                contentDescription = null,
                tint = Color.White,
                modifier = Modifier.size(BADGE_ICON_SIZE),
            )
        }
        Text(
            if (item.isAcknowledged) "$verdict · 확인됨" else verdict,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold),
            color = if (item.isAcknowledged) ink.copy(alpha = PrMonitorStyle.SECONDARY_OPACITY) else statusColor,
            modifier = Modifier.testTag("prmonitor.row.${item.id}.status"),
        )
        Spacer(Modifier.weight(1f))
        Text(
            PrMonitorStyle.clock(item.createdAt),
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = PrMonitorStyle.SECONDARY_OPACITY),
        )
    }
}

@Composable
private fun RowTitle(item: NotificationHistoryItem) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val acked = item.isAcknowledged
    Text(
        PrMonitorText.rowTitle(item),
        style = DesignTokens.Typography.style(
            DesignTokens.Typography.bodyLg,
            if (acked) FontWeight.Medium else FontWeight.SemiBold,
        ).copy(textDecoration = if (acked) TextDecoration.LineThrough else TextDecoration.None),
        color = if (acked) ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY) else ink,
        modifier = Modifier.testTag("prmonitor.row.${item.id}.title"),
    )
}

@Composable
private fun RowMeta(item: NotificationHistoryItem) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val secondary = ink.copy(alpha = PrMonitorStyle.SECONDARY_OPACITY)
    val tertiary = ink.copy(alpha = PrMonitorStyle.TERTIARY_OPACITY)
    val style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs)
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (item.headSha.isNotEmpty()) {
            Text("commit ${PrMonitorText.shortSha(item.headSha)}", style = style, color = secondary)
            Text("·", style = style, color = tertiary)
        }
        Text(item.workflowName.ifEmpty { "workflow" }, style = style, color = secondary, maxLines = 1)
        if (item.headBranch.isNotEmpty()) {
            Text("·", style = style, color = tertiary)
            Text(item.headBranch, style = style, color = secondary, maxLines = 1, modifier = Modifier.weight(1f, fill = false))
        }
        item.acknowledgedAt?.let { ackedAt ->
            Spacer(Modifier.weight(1f))
            Text("확인 ${relativeAck(ackedAt)}", style = style, color = secondary)
        }
    }
}

@Composable
private fun RowActions(item: NotificationHistoryItem, onAcknowledge: () -> Unit) {
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        PrLinkChip("커밋", Icons.Filled.Circle, item.commitUrl, "prmonitor.row.${item.id}.link.commit")
        PrLinkChip("런", Icons.Filled.PlayArrow, item.runUrl, "prmonitor.row.${item.id}.link.run")
        Spacer(Modifier.weight(1f))
        Row(
            Modifier
                .clip(RoundedCornerShape(PrMonitorStyle.controlRadius))
                .background(DesignTokens.Colors.accent(PR_AREA))
                .clickable(role = Role.Button, onClick = onAcknowledge)
                .padding(horizontal = DesignTokens.Spacing.md, vertical = ACK_VERTICAL)
                .testTag("prmonitor.row.${item.id}.ack.button"),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(Icons.Filled.Check, contentDescription = null, tint = Color.White, modifier = Modifier.size(ACK_ICON_SIZE))
            Text(
                "확인",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                color = Color.White,
            )
        }
    }
}

private fun relativeAck(epochSeconds: Long): String =
    DateUtils.getRelativeTimeSpanString(
        epochSeconds * MILLIS_PER_SECOND,
        System.currentTimeMillis(),
        DateUtils.MINUTE_IN_MILLIS,
    ).toString()

private const val MILLIS_PER_SECOND = 1000L
