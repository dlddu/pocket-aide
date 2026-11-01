package com.dlddu.pocketaide.notifications

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.approvals.ApprovalMetrics
import com.dlddu.pocketaide.approvals.CallerChip
import com.dlddu.pocketaide.approvals.monoStyle
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetDefaults

private val AREA = Area.PrMonitor
private val META_VALUE_SIZE = 12.5.sp
private val BODY_LINE_HEIGHT = 20.sp

@Composable
fun ExternalNotificationDetailSheet(item: ExternalNotification, onClose: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    Sheet(area = AREA, onClose = onClose) {
        SelectionContainer {
            Column(
                Modifier
                    .fillMaxWidth()
                    .padding(
                        start = SheetDefaults.horizontalPadding,
                        end = SheetDefaults.horizontalPadding,
                        top = SheetDefaults.titleTopPadding,
                        bottom = DesignTokens.Spacing.xl,
                    )
                    .testTag("notifications.detail"),
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.weight(1f)) {
                        CallerChip(ExternalNotificationFormatting.callerLabel(item))
                    }
                    Spacer(Modifier.width(DesignTokens.Spacing.sm))
                    Text(
                        "닫기",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
                        color = DesignTokens.Colors.accent(AREA),
                        modifier = Modifier
                            .clickable(role = Role.Button, onClick = onClose)
                            .testTag("notifications.detail.close"),
                    )
                }
                Text(
                    item.title,
                    style = DesignTokens.Typography.style(ApprovalMetrics.sheetTitleSize, FontWeight.Bold),
                    color = ink,
                    modifier = Modifier
                        .padding(top = DesignTokens.Spacing.md)
                        .testTag("notifications.detail.title"),
                )
                if (item.body.isNotEmpty()) {
                    Text(
                        item.body,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.body).copy(lineHeight = BODY_LINE_HEIGHT),
                        color = ink,
                        modifier = Modifier
                            .padding(top = DesignTokens.Spacing.sm)
                            .testTag("notifications.detail.body"),
                    )
                }
                Column(Modifier.padding(top = DesignTokens.Spacing.xl)) {
                    Rule()
                    MetaRow("보낸 호출자", ExternalNotificationFormatting.callerLabel(item), "caller")
                    Rule()
                    MetaRow("요청 주체", ExternalNotificationFormatting.requesterValue(item), "requester")
                    Rule()
                    MetaRow("받은 시각", ExternalNotificationFormatting.fullTime(item), "time")
                    Rule()
                    MetaRow("도달", ExternalNotificationFormatting.deliveryLabel(item), "delivery")
                }
            }
        }
    }
}

@Composable
private fun Rule() {
    Box(Modifier.fillMaxWidth().height(ApprovalMetrics.ruleWidth).background(DesignTokens.Colors.rule(AREA)))
}

@Composable
private fun MetaRow(label: String, value: String, id: String) {
    val ink = DesignTokens.Colors.ink(AREA)
    Column(
        Modifier
            .fillMaxWidth()
            .padding(vertical = ApprovalMetrics.bannerVertical),
        verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.chipVertical),
    ) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        )
        Text(
            value,
            style = monoStyle(META_VALUE_SIZE),
            color = ink,
            modifier = Modifier.testTag("notifications.detail.$id"),
        )
    }
}
