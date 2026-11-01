package com.dlddu.pocketaide.approvals

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun ApprovalKeysScreen(api: ApiClient, onBack: () -> Unit) {
    val viewModel: ApprovalKeysViewModel = viewModel(key = "approval.keys") { ApprovalKeysViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var showsIssueSheet by remember { mutableStateOf(false) }
    var showsRegisterSheet by remember { mutableStateOf(false) }
    var revoking by remember { mutableStateOf<ApprovalCallerKey?>(null) }
    val now = rememberNowMillis()
    val area = APPROVAL_AREA
    val accent = DesignTokens.Colors.accent(area)

    LaunchedEffect(viewModel) { viewModel.load() }

    Column(Modifier.fillMaxSize()) {
        ApprovalBackBar(onBack = onBack)
        PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
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
            ) {
                Text(
                    "승인 게이트 설정",
                    style = DesignTokens.Typography.style(ApprovalMetrics.detailTitleSize, FontWeight.Bold),
                    color = DesignTokens.Colors.ink(area),
                    modifier = Modifier
                        .padding(bottom = DesignTokens.Spacing.md)
                        .testTag("approval.keys.title"),
                )
                ApprovalSectionHeader(
                    title = "자동 응답 모드",
                    tint = accent,
                    modifier = Modifier.padding(top = DesignTokens.Spacing.xs, bottom = DesignTokens.Spacing.sm),
                )
                AutoResponseOptions(
                    selected = state.autoResponse,
                    notice = state.modeNotice,
                    onChoose = viewModel::setAutoResponse,
                )
                ApprovalSectionHeader(
                    title = "호출자 키",
                    tint = accent,
                    modifier = Modifier.padding(top = DesignTokens.Spacing.xs, bottom = DesignTokens.Spacing.sm),
                )
                if (state.errorMessage != null) {
                    ApprovalErrorBanner(
                        title = "호출자 키를 불러오지 못했습니다",
                        message = null,
                        bannerTag = null,
                        retryTag = "approval.keys.retry",
                        onRetry = viewModel::load,
                        modifier = Modifier.padding(bottom = DesignTokens.Spacing.sm),
                        showsIcon = false,
                        outlined = false,
                    )
                }
                KeyList(state = state, nowMillis = now, onRevoke = { revoking = it })
                DashedActionButton(
                    title = "+ 호출자 키 발급",
                    tag = "approval.keys.issue",
                    modifier = Modifier.padding(top = DesignTokens.Spacing.md),
                ) { showsIssueSheet = true }
                DashedActionButton(
                    title = "+ k8s 서비스 어카운트 등록",
                    tag = "approval.keys.register-sa",
                    modifier = Modifier.padding(top = DesignTokens.Spacing.sm),
                ) { showsRegisterSheet = true }
            }
        }
    }

    if (showsIssueSheet) {
        ApprovalKeyIssueSheet(
            state = state,
            onIssue = viewModel::issue,
            onClose = {
                showsIssueSheet = false
                viewModel.finishIssuing()
            },
        )
    }
    if (showsRegisterSheet) {
        ApprovalServiceAccountSheet(
            state = state,
            onRegister = { name, namespace, account ->
                viewModel.registerServiceAccount(name, namespace, account) {
                    showsRegisterSheet = false
                    viewModel.finishRegistering()
                }
            },
            onClose = {
                showsRegisterSheet = false
                viewModel.finishRegistering()
            },
        )
    }
    revoking?.let { key ->
        ApprovalConfirmDialog(
            title = ApprovalKeyFormatting.revokeTitle(key),
            message = ApprovalKeyFormatting.revokeMessage(key),
            confirmLabel = "폐기",
            destructive = true,
            onConfirm = { viewModel.revoke(key) },
            onDismiss = { revoking = null },
        )
    }
}

@Composable
private fun CardList(content: @Composable () -> Unit) {
    val area = APPROVAL_AREA
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Column(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape),
    ) {
        content()
    }
}

@Composable
private fun RuleDivider() {
    Box(Modifier.fillMaxWidth().height(ApprovalMetrics.ruleWidth).background(DesignTokens.Colors.rule(APPROVAL_AREA)))
}

@Composable
private fun AutoResponseOptions(selected: ApprovalAutoResponse?, notice: String?, onChoose: (ApprovalAutoResponse) -> Unit) {
    Column(Modifier.padding(bottom = DesignTokens.Spacing.lg)) {
        CardList {
            ApprovalAutoResponse.entries.forEachIndexed { index, mode ->
                if (index > 0) RuleDivider()
                ApprovalAutoResponseOption(
                    mode = mode,
                    selected = selected == mode,
                    enabled = selected != null,
                    onChoose = { onChoose(mode) },
                )
            }
        }
        if (notice != null) {
            Text(
                notice,
                style = DesignTokens.Typography.style(ApprovalMetrics.noticeSize),
                color = DesignTokens.StatusColors.failure(),
                modifier = Modifier
                    .padding(top = DesignTokens.Spacing.sm)
                    .testTag("approval.settings.mode.notice"),
            )
        }
    }
}

@Composable
private fun KeyList(state: ApprovalKeysUiState, nowMillis: Long, onRevoke: (ApprovalCallerKey) -> Unit) {
    val area = APPROVAL_AREA
    if (state.keys.isEmpty()) {
        Box(
            Modifier
                .fillMaxWidth()
                .padding(vertical = DesignTokens.Spacing.lg),
            contentAlignment = Alignment.Center,
        ) {
            if (state.isLoading || !state.hasLoaded) {
                CircularProgressIndicator(color = DesignTokens.Colors.accent(area))
            } else {
                Text(
                    "발급한 호출자 키가 없습니다",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                    color = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                    modifier = Modifier.testTag("approval.keys.empty"),
                )
            }
        }
        return
    }
    CardList {
        state.keys.forEachIndexed { index, key ->
            if (index > 0) RuleDivider()
            KeyRow(key = key, nowMillis = nowMillis, onRevoke = { onRevoke(key) })
        }
    }
}

@Composable
private fun KeyRow(key: ApprovalCallerKey, nowMillis: Long, onRevoke: () -> Unit) {
    val area = APPROVAL_AREA
    val secondary = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY)
    Column(
        Modifier
            .fillMaxWidth()
            .alpha(if (key.isRevoked) ApprovalMetrics.REVOKED_OPACITY else 1f)
            .padding(horizontal = ApprovalMetrics.rowPadding, vertical = DesignTokens.Spacing.md)
            .testTag("approval.keys.row.${key.name}"),
        verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            CallerChip(key.name, revoked = key.isRevoked)
            Spacer(Modifier.weight(1f))
            val revoked = ApprovalKeyFormatting.revokedLabel(key)
            if (revoked != null) {
                Text(
                    revoked,
                    style = monoStyle(ApprovalMetrics.chipSize),
                    color = secondary,
                    modifier = Modifier.testTag("approval.keys.revoked.${key.name}"),
                )
            } else {
                Text(
                    "폐기",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                    color = DesignTokens.StatusColors.failure(),
                    modifier = Modifier
                        .clickable(role = Role.Button, onClick = onRevoke)
                        .testTag("approval.keys.revoke.${key.name}"),
                )
            }
        }
        Text(
            ApprovalKeyFormatting.meta(key, nowMillis),
            style = monoStyle(ApprovalMetrics.chipSize),
            color = secondary,
            modifier = Modifier.testTag("approval.keys.meta.${key.name}"),
        )
    }
}
