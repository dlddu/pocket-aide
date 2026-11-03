package com.dlddu.pocketaide.approvals

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
internal fun ApprovalDetailScreen(item: ApprovalRequest, api: ApiClient, onBack: () -> Unit) {
    val viewModel: ApprovalDetailViewModel = viewModel(key = "approval.detail.${item.id}") {
        ApprovalDetailViewModel(item, api)
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var confirming by remember { mutableStateOf<ApprovalDecision?>(null) }
    val now = rememberNowMillis()
    val area = APPROVAL_AREA

    LaunchedEffect(viewModel) { viewModel.reload() }

    Column(Modifier.fillMaxSize()) {
        ApprovalBackBar(onBack = onBack) {
            CallerChip(state.item.keyName, Modifier.testTag("approval.detail.key.chip"))
        }
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(
                    start = DesignTokens.Spacing.xl,
                    end = DesignTokens.Spacing.xl,
                    top = DesignTokens.Spacing.md,
                    bottom = DesignTokens.Spacing.xxl,
                ),
            verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
        ) {
            Text(
                ApprovalFormatting.summary(state.item.context),
                style = DesignTokens.Typography.style(ApprovalMetrics.detailTitleSize, FontWeight.Bold),
                color = DesignTokens.Colors.ink(area),
                modifier = Modifier.testTag("approval.detail.title"),
            )
            DetailContent(
                state = state,
                nowMillis = now,
                onRetry = viewModel::reload,
                onDecide = { confirming = it },
            )
        }
    }

    confirming?.let { decision ->
        ApprovalConfirmDialog(
            title = ApprovalFormatting.confirmTitle(decision),
            message = ApprovalFormatting.confirmMessage(decision, state.item.context),
            confirmLabel = if (decision == ApprovalDecision.APPROVE) "승인" else "거절",
            destructive = false,
            onConfirm = { viewModel.decide(decision) },
            onDismiss = { confirming = null },
        )
    }
}

@Composable
private fun DetailContent(
    state: ApprovalDetailUiState,
    nowMillis: Long,
    onRetry: () -> Unit,
    onDecide: (ApprovalDecision) -> Unit,
) {
    val area = APPROVAL_AREA
    val item = state.item
    val expiredNow = item.isExpired(nowMillis)
    val deadline = item.expiresAt
    Column(verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical)) {
        if (item.status == ApprovalStatus.PENDING && !expiredNow && deadline != null) {
            Countdown(item = item, expiresAt = deadline, nowMillis = nowMillis)
        }
        val notice = state.notice
        if (notice != null) {
            NoticeBanner(notice)
        } else if (state.loadError != null) {
            ApprovalErrorBanner(
                title = "최신 상태를 불러오지 못했습니다",
                message = null,
                bannerTag = null,
                retryTag = "approval.detail.retry",
                onRetry = onRetry,
                showsIcon = false,
                outlined = false,
            )
        }
        if (item.keyRevoked) {
            Text(
                "폐기된 키로 생성됨 — 자동 처리되지 않음",
                style = DesignTokens.Typography.style(ApprovalMetrics.chipSize, FontWeight.Bold),
                color = DesignTokens.StatusColors.failure(),
                modifier = Modifier.testTag("approval.detail.revoked"),
            )
        }
        ApprovalFormatting.agentReason(item)?.let { AgentReason(it) }
        BorderedBlock(tag = "approval.detail.context") {
            Text(
                item.context,
                style = monoStyle(ApprovalMetrics.chipSize),
                color = DesignTokens.Colors.ink(area),
            )
        }
        Facts(item = item, status = if (expiredNow) ApprovalStatus.EXPIRED else item.status)
        if (item.status == ApprovalStatus.PENDING) {
            DecisionButtons(enabled = state.canDecide(nowMillis), onDecide = onDecide)
        }
    }
}

@Composable
private fun BorderedBlock(tag: String, content: @Composable () -> Unit) {
    val area = APPROVAL_AREA
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Box(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
            .testTag(tag),
    ) {
        content()
    }
}

@Composable
private fun AgentReason(reason: String) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        Row(verticalAlignment = Alignment.Bottom) {
            Text(
                "에이전트 설명",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold),
                color = DesignTokens.Colors.accent(area),
            )
            Spacer(Modifier.weight(1f))
            Text(
                "에이전트가 작성 · 검증되지 않음",
                style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                modifier = Modifier.padding(start = DesignTokens.Spacing.sm).testTag("approval.detail.reason.caption"),
            )
        }
        BorderedBlock(tag = "approval.detail.reason") {
            Text(reason, style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm), color = ink)
        }
    }
}

@Composable
private fun Countdown(item: ApprovalRequest, expiresAt: Long, nowMillis: Long) {
    val area = APPROVAL_AREA
    val accent = DesignTokens.Colors.accent(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    val progress = ApprovalFormatting.progress(item.createdAt, expiresAt, nowMillis)
    Row(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm),
        horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            "남은 시간",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        )
        Text(
            ApprovalFormatting.countdown(expiresAt, nowMillis),
            style = monoStyle(ApprovalMetrics.detailTitleSize, FontWeight.Bold),
            color = accent,
            modifier = Modifier.testTag("approval.detail.countdown"),
        )
        LinearProgressIndicator(
            progress = { progress },
            color = accent,
            trackColor = DesignTokens.Colors.rule(area),
            modifier = Modifier.weight(1f),
        )
    }
}

@Composable
private fun Facts(item: ApprovalRequest, status: ApprovalStatus) {
    val facts = buildList {
        add(Triple("보낸 키", item.keyName, "key"))
        add(Triple("요청자", item.requesterName.ifEmpty { "—" }, "requester"))
        ApprovalFormatting.sessionLabel(item)?.let { add(Triple("세션", it, "session")) }
        add(Triple("요청 ID", item.externalId, "externalId"))
        add(Triple("요청 시각", ApprovalFormatting.clockTime(item.createdAt), "createdAt"))
        add(Triple("상태", ApprovalFormatting.statusLabel(status), "status"))
        if (status != ApprovalStatus.PENDING) {
            add(Triple("처리 방식", ApprovalFormatting.decisionModeLabel(status, item.decisionMode), "mode"))
            add(Triple("처리 시각", ApprovalFormatting.clockTime(item.processedAt ?: item.expiresAt), "processedAt"))
        }
    }
    Column(verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing)) {
        facts.chunked(2).forEach { pair ->
            Row(Modifier.fillMaxWidth()) {
                pair.forEach { (label, value, id) -> Fact(label, value, id, Modifier.weight(1f)) }
                if (pair.size == 1) Spacer(Modifier.weight(1f))
            }
        }
    }
}

@Composable
private fun Fact(label: String, value: String, id: String, modifier: Modifier) {
    val ink = DesignTokens.Colors.ink(APPROVAL_AREA)
    Column(modifier, verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.chipVertical)) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        )
        Text(
            value,
            style = monoStyle(DesignTokens.Typography.captionXs),
            color = ink,
            modifier = Modifier.testTag("approval.detail.$id"),
        )
    }
}

@Composable
private fun DecisionButtons(enabled: Boolean, onDecide: (ApprovalDecision) -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .padding(top = DesignTokens.Spacing.xs),
        horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
    ) {
        ApprovalButton(
            title = "거절",
            primary = false,
            outlineAccent = true,
            enabled = enabled,
            tag = "approval.detail.reject",
            modifier = Modifier.weight(1f),
        ) { onDecide(ApprovalDecision.REJECT) }
        ApprovalButton(
            title = "승인",
            primary = true,
            enabled = enabled,
            tag = "approval.detail.approve",
            modifier = Modifier.weight(1f),
        ) { onDecide(ApprovalDecision.APPROVE) }
    }
}

@Composable
private fun NoticeBanner(text: String) {
    val area = APPROVAL_AREA
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
        color = DesignTokens.Colors.ink(area),
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(DesignTokens.Radius.key))
            .background(DesignTokens.Colors.soft(area))
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
            .testTag("approval.detail.notice"),
    )
}
