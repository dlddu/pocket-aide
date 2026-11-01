package com.dlddu.pocketaide.sessions

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
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens

private val RULE_HEIGHT = 1.dp
private val CHEVRON_SIZE = 14.dp
private const val CHEVRON_EXPANDED_DEGREES = 90f
private const val TEXT_SEGMENT_OPACITY = 0.65f
private const val TIMING_OPACITY = 0.45f
private const val PREVIEW_OPACITY = 0.8f
private const val PREVIEW_MAX_LINES = 2

internal fun monospaceStyle(): TextStyle =
    DesignTokens.Typography.style(DesignTokens.Typography.bodySm).copy(fontFamily = FontFamily.Monospace)

@Composable
internal fun SessionTranscriptView(transcript: AgentSessionTranscript) {
    val expanded = remember { mutableStateListOf<String>() }
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val rule = DesignTokens.Colors.rule(SESSIONS_AREA)
    Column(
        Modifier.fillMaxWidth().testTag("sessions.detail.transcript"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        transcript.segments.forEachIndexed { index, segment ->
            if (index > 0) Box(Modifier.fillMaxWidth().height(RULE_HEIGHT).background(rule))
            when (segment) {
                is AgentSessionTranscript.Segment.Text -> SelectionContainer {
                    Text(
                        segment.text.trim('\n', '\r'),
                        style = monospaceStyle(),
                        color = ink.copy(alpha = TEXT_SEGMENT_OPACITY),
                        modifier = Modifier.fillMaxWidth().testTag("sessions.detail.output.text"),
                    )
                }
                is AgentSessionTranscript.Segment.Turn -> TurnView(
                    turn = segment.turn,
                    isExpanded = { it in expanded },
                    onToggle = { key -> if (!expanded.remove(key)) expanded.add(key) },
                )
            }
        }
    }
}

@Composable
private fun TurnView(turn: AgentSessionTurn, isExpanded: (String) -> Boolean, onToggle: (String) -> Unit) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val key = "sessions.detail.turn.${turn.seq}"
    Column(Modifier.fillMaxWidth().testTag(key)) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
            Text(
                "#${turn.seq}",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
            )
            AgentSessionTurnTiming.label(turn)?.let { timing ->
                Text(
                    timing,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                    color = ink.copy(alpha = TIMING_OPACITY),
                    maxLines = 1,
                )
            }
            Spacer(Modifier.weight(1f))
            SessionTurnStatusBadge(turn.status)
        }
        TurnSection(
            title = "입력",
            preview = turn.inputPreview,
            full = turn.input,
            placeholder = if (turn.hasInput) "빈 프롬프트" else "앞부분이 생략돼 입력을 볼 수 없어요",
            isExpanded = isExpanded("$key.input"),
            identifier = "$key.input",
            onToggle = { onToggle("$key.input") },
        )
        TurnSection(
            title = "출력",
            preview = turn.outputPreview,
            full = turn.output,
            placeholder = if (turn.status.isFinished) "출력이 없어요" else "응답을 기다리는 중…",
            isExpanded = isExpanded("$key.output"),
            identifier = "$key.output",
            onToggle = { onToggle("$key.output") },
        )
    }
}

@Composable
private fun TurnSection(
    title: String,
    preview: String,
    full: String,
    placeholder: String,
    isExpanded: Boolean,
    identifier: String,
    onToggle: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val summary = preview.ifEmpty { placeholder }
    Column(Modifier.fillMaxWidth()) {
        Row(
            Modifier
                .fillMaxWidth()
                .clickable(role = Role.Button, onClick = onToggle)
                .semantics {
                    contentDescription = "$title ${if (isExpanded) "접기" else "펼치기"}"
                    stateDescription = summary
                }
                .padding(vertical = DesignTokens.Spacing.sm)
                .testTag("$identifier.toggle"),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            verticalAlignment = Alignment.Top,
        ) {
            Icon(
                Icons.AutoMirrored.Outlined.KeyboardArrowRight,
                contentDescription = null,
                tint = ink.copy(alpha = TIMING_OPACITY),
                modifier = Modifier.size(CHEVRON_SIZE).rotate(if (isExpanded) CHEVRON_EXPANDED_DEGREES else 0f),
            )
            Text(
                title,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold),
                color = DesignTokens.Colors.accent(SESSIONS_AREA),
            )
            if (!isExpanded) {
                Text(
                    summary,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = ink.copy(alpha = if (preview.isEmpty()) TIMING_OPACITY else PREVIEW_OPACITY),
                    maxLines = PREVIEW_MAX_LINES,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f),
                )
            }
        }
        if (isExpanded) {
            SelectionContainer {
                Text(
                    full.ifEmpty { placeholder },
                    style = monospaceStyle(),
                    color = if (full.isEmpty()) ink.copy(alpha = TIMING_OPACITY) else ink,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(start = DesignTokens.Spacing.lg, bottom = DesignTokens.Spacing.sm)
                        .testTag("$identifier.full"),
                )
            }
        }
    }
}
