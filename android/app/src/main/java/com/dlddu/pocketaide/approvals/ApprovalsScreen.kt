package com.dlddu.pocketaide.approvals

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Schedule
import androidx.compose.material.icons.outlined.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel

private val HEADER_DOT_SIZE = 8.dp

@Composable
fun ApprovalsScreen(
    viewModel: ApprovalsViewModel,
    openedApprovalId: String? = null,
    onOpenedApprovalConsumed: () -> Unit = {},
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val destination = state.destination
    val onPending = destination == ApprovalDestination.Pending

    LaunchedEffect(onPending) { if (onPending) viewModel.load() }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.load() }
    LaunchedEffect(openedApprovalId) {
        val id = openedApprovalId ?: return@LaunchedEffect
        onOpenedApprovalConsumed()
        viewModel.open(id)
    }
    BackHandler(enabled = !onPending) { viewModel.navigate(ApprovalDestination.Pending) }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(APPROVAL_AREA)),
    ) {
        when (destination) {
            ApprovalDestination.Pending -> PendingList(state = state, viewModel = viewModel)
            is ApprovalDestination.Detail -> ApprovalDetailScreen(
                item = destination.item,
                api = viewModel.api,
                onBack = { viewModel.navigate(ApprovalDestination.Pending) },
            )
            ApprovalDestination.Settings -> ApprovalKeysScreen(
                api = viewModel.api,
                onBack = { viewModel.navigate(ApprovalDestination.Pending) },
            )
            ApprovalDestination.History -> ApprovalHistoryScreen(
                api = viewModel.api,
                onBack = { viewModel.navigate(ApprovalDestination.Pending) },
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun PendingList(state: ApprovalsUiState, viewModel: ApprovalsViewModel) {
    val now = rememberNowMillis()
    Column(Modifier.fillMaxSize()) {
        PendingHeader(
            state = state,
            onHistory = { viewModel.navigate(ApprovalDestination.History) },
            onSettings = { viewModel.navigate(ApprovalDestination.Settings) },
        )
        PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
            LazyColumn(
                Modifier.fillMaxSize(),
                contentPadding = PaddingValues(
                    start = DesignTokens.Spacing.xl,
                    end = DesignTokens.Spacing.xl,
                    bottom = DesignTokens.Spacing.xxl,
                ),
                verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowGap * 2),
            ) {
                state.errorMessage?.let { message ->
                    item(key = "error") {
                        ApprovalErrorBanner(
                            title = "승인 요청을 불러오지 못했습니다",
                            message = message,
                            bannerTag = "approvals.error.banner",
                            retryTag = "approvals.error.retry",
                            onRetry = viewModel::load,
                        )
                    }
                }
                state.autoResponse.bannerTitle?.let { title ->
                    item(key = "autoResponse") {
                        ApprovalAutoResponseBanner(
                            title = title,
                            notice = state.autoResponse.outcomeNotice,
                            onTurnOff = viewModel::turnOffAutoResponse,
                        )
                    }
                }
                if (state.items.isEmpty()) {
                    item(key = "empty") {
                        ApprovalEmptyState(
                            isLoading = state.isLoading || (!state.hasLoaded && state.errorMessage == null),
                            showsMessage = state.errorMessage == null,
                            mode = state.autoResponse,
                        )
                    }
                } else {
                    item(key = "section") {
                        ApprovalSectionHeader(
                            title = ApprovalFormatting.pendingHeader(state.pendingCount),
                            tint = DesignTokens.Colors.accent(APPROVAL_AREA),
                            tag = "approvals.section.pending",
                        )
                    }
                    items(state.items, key = { it.id }) { item ->
                        ApprovalRowCard(
                            item = item,
                            nowMillis = now,
                            onClick = { viewModel.navigate(ApprovalDestination.Detail(item)) },
                        )
                    }
                }
            }
            if (state.isLoading && state.items.isNotEmpty() && !state.isRefreshing) {
                CircularProgressIndicator(
                    color = DesignTokens.Colors.accent(APPROVAL_AREA),
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .padding(top = DesignTokens.Spacing.xs)
                        .size(ApprovalMetrics.iconSize)
                        .testTag("approvals.loading"),
                )
            }
        }
    }
}

@Composable
private fun PendingHeader(state: ApprovalsUiState, onHistory: () -> Unit, onSettings: () -> Unit) {
    val area = APPROVAL_AREA
    Row(
        Modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = DesignTokens.Spacing.md,
                bottom = DesignTokens.Spacing.md,
            ),
        verticalAlignment = Alignment.Bottom,
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing)) {
            AreaLabel(area = area, text = "APPROVAL · GATE", showsDot = true, dotSize = HEADER_DOT_SIZE)
            Text(
                "승인",
                style = DesignTokens.Typography.style(ApprovalMetrics.headerTitleSize, FontWeight.Bold),
                color = DesignTokens.Colors.ink(area),
                modifier = Modifier.testTag("approvals.title"),
            )
            ApprovalAutoResponseHeader(mode = state.autoResponse)
        }
        CountBadge(count = state.pendingCount, label = "대기", countTag = "approvals.pending.count")
        HeaderIconButton(Icons.Outlined.Schedule, "처리 이력", "approvals.history.button", onHistory)
        HeaderIconButton(Icons.Outlined.Settings, "승인 게이트 설정", "approvals.settings.button", onSettings)
    }
}

@Composable
internal fun CountBadge(count: Int, label: String, countTag: String) {
    Column(
        Modifier
            .clip(RoundedCornerShape(DesignTokens.Radius.key))
            .background(DesignTokens.Colors.accent(APPROVAL_AREA))
            .padding(horizontal = ApprovalMetrics.badgeHorizontal, vertical = DesignTokens.Spacing.xs),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.chipVertical),
    ) {
        Text(
            "$count",
            style = monoStyle(ApprovalMetrics.badgeCountSize, FontWeight.Bold),
            color = Color.White,
            modifier = Modifier.testTag(countTag),
        )
        Text(label, style = monoStyle(ApprovalMetrics.badgeLabelSize, FontWeight.Bold), color = Color.White)
    }
}

@Composable
private fun HeaderIconButton(icon: ImageVector, label: String, tag: String, onClick: () -> Unit) {
    val area = APPROVAL_AREA
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Box(
        Modifier
            .size(ApprovalMetrics.iconButtonSize)
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Icon(icon, contentDescription = label, tint = DesignTokens.Colors.ink(area), modifier = Modifier.size(ApprovalMetrics.iconSize))
    }
}

@Composable
private fun ApprovalRowCard(item: ApprovalRequest, nowMillis: Long, onClick: () -> Unit) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val accent = DesignTokens.Colors.accent(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    val created = ApprovalFormatting.requestTime(item.createdAt, nowMillis)
    Row(
        Modifier
            .fillMaxWidth()
            .height(IntrinsicSize.Min)
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .semantics(mergeDescendants = true) {}
            .testTag("approvals.row.${item.externalId}"),
    ) {
        Box(Modifier.width(ApprovalMetrics.accentStripWidth).fillMaxHeight().background(accent))
        Column(
            Modifier
                .weight(1f)
                .padding(
                    start = ApprovalMetrics.rowPadding - ApprovalMetrics.accentStripWidth,
                    end = ApprovalMetrics.rowPadding,
                    top = DesignTokens.Spacing.md,
                    bottom = DesignTokens.Spacing.md,
                ),
            verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                CallerChip(item.keyName)
                Spacer(Modifier.weight(1f))
                Text(
                    ApprovalFormatting.remaining(item.expiresAt, nowMillis),
                    style = monoStyle(ApprovalMetrics.chipSize, FontWeight.Bold),
                    color = accent,
                )
            }
            Text(
                ApprovalFormatting.summary(item.context),
                style = DesignTokens.Typography.style(ApprovalMetrics.rowTitleSize, FontWeight.SemiBold),
                color = ink,
                maxLines = 2,
            )
            Text(
                if (item.requesterName.isEmpty()) created else "${item.requesterName} · $created",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
            )
            if (item.keyRevoked) {
                Text(
                    "폐기된 키로 생성됨 — 자동 처리되지 않음",
                    style = DesignTokens.Typography.style(ApprovalMetrics.chipSize, FontWeight.Bold),
                    color = DesignTokens.StatusColors.failure(),
                )
            }
        }
    }
}
