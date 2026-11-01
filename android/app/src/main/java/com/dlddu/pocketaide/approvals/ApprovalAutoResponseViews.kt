package com.dlddu.pocketaide.approvals

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.RadioButtonChecked
import androidx.compose.material.icons.outlined.ErrorOutline
import androidx.compose.material.icons.outlined.RadioButtonUnchecked
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
internal fun ApprovalAutoResponseHeader(mode: ApprovalAutoResponse) {
    val area = APPROVAL_AREA
    val style = monoStyle(ApprovalMetrics.noticeSize)
    Row(
        Modifier
            .semantics(mergeDescendants = true) {}
            .testTag("approvals.autoResponse.header"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Text(
            "자동 응답",
            style = style,
            color = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        )
        Text(
            mode.label,
            style = style.copy(fontWeight = FontWeight.Bold),
            color = DesignTokens.Colors.accent(area),
        )
    }
}

@Composable
internal fun ApprovalAutoResponseBanner(title: String, notice: String, onTurnOff: () -> Unit, modifier: Modifier = Modifier) {
    val area = APPROVAL_AREA
    val accent = DesignTokens.Colors.accent(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Row(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.soft(area))
            .border(ApprovalMetrics.ruleWidth, accent, shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
            .testTag("approvals.autoResponse.banner"),
        horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            Icons.Outlined.ErrorOutline,
            contentDescription = null,
            tint = accent,
            modifier = Modifier.size(ApprovalMetrics.iconSize),
        )
        Text(
            buildAnnotatedString {
                withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(title) }
                append(" ")
                append(notice)
            },
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
            color = DesignTokens.Colors.ink(area),
            modifier = Modifier.weight(1f).testTag("approvals.autoResponse.banner.text"),
        )
        Text(
            "끄기",
            style = monoStyle(DesignTokens.Typography.captionXs, FontWeight.Bold),
            color = accent,
            modifier = Modifier
                .clickable(role = Role.Button, onClick = onTurnOff)
                .testTag("approvals.autoResponse.off"),
        )
    }
}

@Composable
internal fun ApprovalEmptyState(isLoading: Boolean, showsMessage: Boolean, mode: ApprovalAutoResponse) {
    val area = APPROVAL_AREA
    val secondary = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY)
    Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
        when {
            isLoading -> CircularProgressIndicator(color = DesignTokens.Colors.accent(area))
            showsMessage -> Column(
                Modifier.padding(top = DesignTokens.Spacing.xxl),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing),
            ) {
                Text(
                    "대기 중인 승인 요청이 없습니다",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
                    color = secondary,
                    modifier = Modifier.testTag("approvals.empty.state"),
                )
                Text(
                    buildAnnotatedString {
                        append("자동 응답 모드 ")
                        withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(mode.label) }
                        append(" — ")
                        append(mode.outcomeNotice)
                    },
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = secondary,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.testTag("approvals.empty.mode"),
                )
            }
        }
    }
}

@Composable
internal fun ApprovalAutoResponseOption(mode: ApprovalAutoResponse, selected: Boolean, enabled: Boolean, onChoose: () -> Unit) {
    val area = APPROVAL_AREA
    Row(
        Modifier
            .fillMaxWidth()
            .alpha(if (enabled) 1f else ApprovalMetrics.DISABLED_OPACITY)
            .selectable(selected = selected, enabled = enabled, role = Role.RadioButton, onClick = onChoose)
            .padding(horizontal = ApprovalMetrics.rowPadding, vertical = DesignTokens.Spacing.md)
            .testTag("approval.settings.mode.${mode.wire}"),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            mode.optionLabel,
            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
            color = DesignTokens.Colors.ink(area),
            modifier = Modifier.weight(1f),
        )
        Icon(
            if (selected) Icons.Filled.RadioButtonChecked else Icons.Outlined.RadioButtonUnchecked,
            contentDescription = null,
            tint = if (selected) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area),
            modifier = Modifier.size(ApprovalMetrics.iconSize),
        )
    }
}
