package com.dlddu.pocketaide.sessions

import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.material.icons.outlined.ErrorOutline
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.testTag as semanticsTestTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

internal val SESSIONS_AREA = Area.PrMonitor

internal object SessionStyle {
    const val MUTED_OPACITY = 0.55f
    const val LABEL_OPACITY = 0.7f
    const val HINT_OPACITY = 0.6f
    const val FAINT_OPACITY = 0.45f
    const val DISABLED_OPACITY = 0.5f
    val badgeSize = 10.5.sp
    val badgeTracking = 0.5.sp
    val badgeGap = 5.dp
    val dotSize = 6.dp
    val borderWidth = 1.dp
    val bannerRadius = 6.dp
    val bannerVertical = 10.dp
    val spinnerSize = 10.dp
    val spinnerStroke = 1.5.dp
    val rowGap = 6.dp
    val bannerIconSize = 18.dp
}

@Composable
internal fun SessionRow(session: AgentSession, type: AgentSessionWorkloadType?, modifier: Modifier = Modifier) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(SESSIONS_AREA))
            .border(SessionStyle.borderWidth, DesignTokens.Colors.rule(SESSIONS_AREA), shape)
            .padding(DesignTokens.Spacing.lg),
        verticalArrangement = Arrangement.spacedBy(SessionStyle.rowGap),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
            SessionStateBadge(session.state)
            Spacer(Modifier.weight(1f))
            Text(
                "생성 ${AgentSessionFormatting.relative(session.createdAt)}",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
            )
        }
        Text(
            session.name,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
            color = ink,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            session.kindLabel(type),
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Medium),
            color = ink.copy(alpha = SessionStyle.LABEL_OPACITY),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            AgentSessionFormatting.stateDetail(session),
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.testTag("sessions.row.${session.id}.state"),
        )
    }
}

@Composable
internal fun SessionStateBadge(state: AgentSessionState) {
    val color = when (state) {
        AgentSessionState.ACTIVE -> DesignTokens.StatusColors.success()
        AgentSessionState.IDLE -> DesignTokens.StatusColors.inProgress()
        AgentSessionState.SNAPSHOT -> DesignTokens.Colors.accent(SESSIONS_AREA)
    }
    Row(
        Modifier.clearAndSetSemantics {
            contentDescription = state.label
            semanticsTestTag = "sessions.state.${state.wire}"
        },
        horizontalArrangement = Arrangement.spacedBy(SessionStyle.badgeGap),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(Modifier.size(SessionStyle.dotSize).background(color, CircleShape))
        Text(
            state.label,
            style = DesignTokens.Typography.style(SessionStyle.badgeSize, FontWeight.SemiBold)
                .copy(letterSpacing = SessionStyle.badgeTracking),
            color = color,
        )
    }
}

@Composable
internal fun SessionTurnStatusBadge(status: AgentSessionTurnStatus) {
    val color = when (status) {
        AgentSessionTurnStatus.Running -> DesignTokens.StatusColors.inProgress()
        AgentSessionTurnStatus.Ok -> DesignTokens.StatusColors.success()
        AgentSessionTurnStatus.Error, AgentSessionTurnStatus.Timeout -> DesignTokens.StatusColors.failure()
        AgentSessionTurnStatus.Canceled, is AgentSessionTurnStatus.Other ->
            DesignTokens.Colors.ink(SESSIONS_AREA).copy(alpha = SessionStyle.MUTED_OPACITY)
    }
    Row(
        Modifier.clearAndSetSemantics {
            contentDescription = status.label
            semanticsTestTag = "sessions.turn.status"
        },
        horizontalArrangement = Arrangement.spacedBy(SessionStyle.badgeGap),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (status.isFinished) {
            Box(Modifier.size(SessionStyle.dotSize).background(color, CircleShape))
        } else {
            SmallSpinner(color)
        }
        Text(
            status.label,
            style = DesignTokens.Typography.style(SessionStyle.badgeSize, FontWeight.SemiBold)
                .copy(letterSpacing = SessionStyle.badgeTracking),
            color = color,
        )
    }
}

@Composable
internal fun SmallSpinner(color: Color) {
    CircularProgressIndicator(
        color = color,
        strokeWidth = SessionStyle.spinnerStroke,
        modifier = Modifier.size(SessionStyle.spinnerSize),
    )
}

@Composable
internal fun SessionErrorBanner(error: AgentSessionError, onRetry: () -> Unit, modifier: Modifier = Modifier) {
    val accent = DesignTokens.Colors.accent(SESSIONS_AREA)
    val shape = RoundedCornerShape(SessionStyle.bannerRadius)
    Row(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.soft(SESSIONS_AREA))
            .border(SessionStyle.borderWidth, accent, shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = SessionStyle.bannerVertical),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.Top,
    ) {
        Icon(
            Icons.Outlined.ErrorOutline,
            contentDescription = null,
            tint = DesignTokens.StatusColors.failure(),
            modifier = Modifier.size(SessionStyle.bannerIconSize),
        )
        Text(
            error.message,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
            color = DesignTokens.Colors.ink(SESSIONS_AREA),
            modifier = Modifier.weight(1f),
        )
        Text(
            "다시 시도",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
            color = accent,
            modifier = Modifier
                .clickable(role = Role.Button, onClick = onRetry)
                .testTag("sessions.error.retry"),
        )
    }
}
