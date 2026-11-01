package com.dlddu.pocketaide.prmonitor

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.outlined.AccountTree
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens

private val STATUS_DOT_SIZE = 6.dp
private val RULE_HEIGHT = 1.dp
private const val ROW_RULE_OPACITY = 0.6f

@Composable
internal fun PrMonitorGroupCard(
    group: HistoryGroup,
    highlightedEventId: Long?,
    onAcknowledge: (Long) -> Unit,
    onAcknowledgeGroup: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    val rule = DesignTokens.Colors.rule(PR_AREA)
    Column(
        modifier
            .fillMaxWidth()
            .staticArrivalGlow(group.contains(highlightedEventId), shape)
            .cardFrame(
                acknowledged = group.allAcknowledged,
                radius = DesignTokens.Radius.card,
                card = DesignTokens.Colors.card(PR_AREA),
                surface = DesignTokens.Colors.surface(PR_AREA),
                rule = rule,
                accent = DesignTokens.Colors.accent(PR_AREA),
            )
            .testTag("prmonitor.group.${group.id}"),
    ) {
        GroupHeader(group, isHighlighted = group.contains(highlightedEventId), onAcknowledgeGroup = onAcknowledgeGroup)
        if (!group.allAcknowledged) {
            Box(Modifier.fillMaxWidth().height(RULE_HEIGHT).background(rule))
            group.items.forEachIndexed { index, item ->
                if (index > 0) {
                    Box(Modifier.fillMaxWidth().height(RULE_HEIGHT).background(rule.copy(alpha = ROW_RULE_OPACITY)))
                }
                PrMonitorHistoryRow(
                    item = item,
                    isHighlighted = highlightedEventId == item.id,
                    onAcknowledge = { onAcknowledge(item.id) },
                    modifier = Modifier
                        .padding(horizontal = DesignTokens.Spacing.sm, vertical = DesignTokens.Spacing.xs)
                        .testTag("prmonitor.row.${item.id}"),
                )
            }
        }
    }
}

@Composable
private fun GroupHeader(group: HistoryGroup, isHighlighted: Boolean, onAcknowledgeGroup: () -> Unit) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val accent = DesignTokens.Colors.accent(PR_AREA)
    Row(
        Modifier.fillMaxWidth().padding(DesignTokens.Spacing.md),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.Top,
    ) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
            if (isHighlighted) {
                Text(
                    "방금 진입",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold),
                    color = accent,
                    modifier = Modifier.testTag("prmonitor.group.${group.id}.arrival"),
                )
            }
            if (group.prNumber == null && group.headBranch.isNotEmpty()) {
                Text(
                    "PR 없음 · ${group.headBranch} 직접 푸시",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs).copy(fontStyle = FontStyle.Italic),
                    color = ink.copy(alpha = PrMonitorStyle.TERTIARY_OPACITY),
                )
            }
            GroupTitle(group)
            StatusSummary(group)
            if (!group.allAcknowledged) {
                PrLinkChip("PR", Icons.Outlined.AccountTree, group.prUrl, "prmonitor.group.${group.id}.link.pr")
            }
        }
        Column(
            horizontalAlignment = Alignment.End,
            verticalArrangement = Arrangement.spacedBy(PrMonitorStyle.chipVertical),
        ) {
            GroupBadge(group)
            if (!group.allAcknowledged) AckGroupButton(group, onAcknowledgeGroup)
        }
    }
}

@Composable
private fun GroupTitle(group: HistoryGroup) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val accent = DesignTokens.Colors.accent(PR_AREA)
    val tertiary = ink.copy(alpha = PrMonitorStyle.TERTIARY_OPACITY)
    val text = buildAnnotatedString {
        withStyle(SpanStyle(color = accent)) { append(group.repoFullName) }
        withStyle(SpanStyle(color = tertiary)) { append(" · ") }
        val number = group.prNumber
        if (number != null) {
            withStyle(SpanStyle(color = ink)) {
                append("#$number")
                append(PrMonitorText.prTitleSuffix(group.prTitle))
            }
        } else {
            withStyle(SpanStyle(color = ink.copy(alpha = ACKED_BRANCH_OPACITY))) { append(group.headBranch.ifEmpty { "—" }) }
            withStyle(SpanStyle(color = tertiary)) { append(PrMonitorText.shortShaSuffix(group.headSha)) }
        }
    }
    Text(
        text,
        style = DesignTokens.Typography.style(
            DesignTokens.Typography.body,
            if (group.allAcknowledged) FontWeight.Medium else FontWeight.SemiBold,
        ).copy(textDecoration = if (group.allAcknowledged) TextDecoration.LineThrough else TextDecoration.None),
    )
}

private const val ACKED_BRANCH_OPACITY = 0.7f

@Composable
private fun StatusSummary(group: HistoryGroup) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs)
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (group.successCount > 0) StatusDot(DesignTokens.StatusColors.success(), "통과 ${group.successCount}")
        if (group.failureCount > 0) StatusDot(DesignTokens.StatusColors.failure(), "실패 ${group.failureCount}")
        if (group.inProgressCount > 0) StatusDot(DesignTokens.StatusColors.inProgress(), "진행 ${group.inProgressCount}")
        Text("CI ${group.items.size}건", style = style, color = ink.copy(alpha = PrMonitorStyle.SECONDARY_OPACITY))
        Spacer(Modifier.weight(1f))
        Text(
            "최근 ${PrMonitorStyle.clock(group.latestCreatedAt)}",
            style = style,
            color = ink.copy(alpha = PrMonitorStyle.TERTIARY_OPACITY),
        )
    }
}

@Composable
private fun StatusDot(color: Color, label: String) {
    Row(horizontalArrangement = Arrangement.spacedBy(PrMonitorStyle.chipIconGap), verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(STATUS_DOT_SIZE).background(color, CircleShape))
        Text(label, style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold), color = color)
    }
}

@Composable
private fun GroupBadge(group: HistoryGroup) {
    val shape = RoundedCornerShape(PrMonitorStyle.chipRadius)
    val style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold)
    val tag = "prmonitor.group.${group.id}.badge"
    if (group.allAcknowledged) {
        Text(
            "모두 확인",
            style = style,
            color = DesignTokens.Colors.ink(PR_AREA).copy(alpha = PrMonitorStyle.SECONDARY_OPACITY),
            modifier = Modifier
                .border(PrMonitorStyle.borderWidth, DesignTokens.Colors.rule(PR_AREA), shape)
                .padding(horizontal = DesignTokens.Spacing.sm, vertical = PrMonitorStyle.chipVertical)
                .testTag(tag),
        )
    } else {
        Text(
            "미확인 ${group.unacknowledgedCount}",
            style = style,
            color = Color.White,
            modifier = Modifier
                .clip(shape)
                .background(DesignTokens.Colors.accent(PR_AREA))
                .padding(horizontal = DesignTokens.Spacing.sm, vertical = PrMonitorStyle.chipVertical)
                .testTag(tag),
        )
    }
}

@Composable
private fun AckGroupButton(group: HistoryGroup, onClick: () -> Unit) {
    val accent = DesignTokens.Colors.accent(PR_AREA)
    val shape = RoundedCornerShape(PrMonitorStyle.chipRadius)
    Row(
        Modifier
            .clip(shape)
            .border(PrMonitorStyle.borderWidth, accent, shape)
            .clickable(role = Role.Button, onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = PrMonitorStyle.chipVertical)
            .testTag("prmonitor.group.${group.id}.ack-all.button"),
        horizontalArrangement = Arrangement.spacedBy(PrMonitorStyle.chipIconGap),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(Icons.Filled.Check, contentDescription = null, tint = accent, modifier = Modifier.size(PrMonitorStyle.chipIconSize))
        Text("모두 확인", style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold), color = accent)
    }
}
