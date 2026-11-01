package com.dlddu.pocketaide.approvals

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens

private val STATUS_DOT_SIZE = 10.dp
private val STATUS_DOT_TOP = 6.dp
private val HISTORY_TITLE_SIZE = DesignTokens.Typography.bodySm

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun ApprovalHistoryScreen(api: ApiClient, onBack: () -> Unit) {
    val viewModel: ApprovalHistoryViewModel = viewModel(key = "approval.history") { ApprovalHistoryViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val now = rememberNowMillis()
    val ink = DesignTokens.Colors.ink(APPROVAL_AREA)

    LaunchedEffect(viewModel) { viewModel.load() }

    Column(Modifier.fillMaxSize()) {
        ApprovalBackBar(onBack = onBack)
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
                item(key = "section") {
                    ApprovalSectionHeader(
                        title = "처리 이력",
                        tint = ink.copy(alpha = ApprovalMetrics.FAINT_OPACITY),
                        tag = "approvals.history.title",
                    )
                }
                state.errorMessage?.let { message ->
                    item(key = "error") {
                        ApprovalErrorBanner(
                            title = null,
                            message = message,
                            bannerTag = "approvals.history.error",
                            retryTag = null,
                            onRetry = viewModel::load,
                            outlined = false,
                        )
                    }
                }
                if (state.items.isEmpty()) {
                    if (state.isLoading || (!state.hasLoaded && state.errorMessage == null)) {
                        item(key = "loading") {
                            Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                                CircularProgressIndicator(color = DesignTokens.Colors.accent(APPROVAL_AREA))
                            }
                        }
                    }
                } else {
                    items(state.items, key = { it.id }) { item ->
                        HistoryRowCard(item = item, nowMillis = now)
                    }
                    if (state.hasMore) {
                        item(key = "loadingMore") {
                            LaunchedEffect(state.nextCursor) { viewModel.loadMore() }
                            Text(
                                "이전 이력 불러오는 중…",
                                style = monoStyle(ApprovalMetrics.loadingMoreSize),
                                color = ink.copy(alpha = ApprovalMetrics.FAINT_OPACITY),
                                textAlign = TextAlign.Center,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = DesignTokens.Spacing.sm)
                                    .testTag("approvals.history.loadingMore"),
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun HistoryRowCard(item: ApprovalRequest, nowMillis: Long) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val tint: Color = when (item.status) {
        ApprovalStatus.APPROVED -> DesignTokens.StatusColors.success()
        ApprovalStatus.REJECTED -> DesignTokens.StatusColors.failure()
        ApprovalStatus.PENDING, ApprovalStatus.EXPIRED -> ink.copy(alpha = ApprovalMetrics.FAINT_OPACITY)
    }
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Row(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .padding(horizontal = ApprovalMetrics.rowPadding, vertical = DesignTokens.Spacing.md)
            .semantics(mergeDescendants = true) {}
            .testTag("approvals.history.row.${item.externalId}"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        verticalAlignment = Alignment.Top,
    ) {
        Box(
            Modifier
                .padding(top = STATUS_DOT_TOP)
                .size(STATUS_DOT_SIZE)
                .background(tint, CircleShape),
        )
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    ApprovalFormatting.statusLabel(item.status),
                    style = monoStyle(ApprovalMetrics.chipSize, FontWeight.Bold),
                    color = tint,
                )
                Spacer(Modifier.weight(1f))
                Text(
                    item.processedAt?.let { ApprovalFormatting.historyTime(it, nowMillis) }.orEmpty(),
                    style = monoStyle(ApprovalMetrics.chipSize),
                    color = ink.copy(alpha = ApprovalMetrics.FAINT_OPACITY),
                )
            }
            Text(
                ApprovalFormatting.summary(item.context),
                style = DesignTokens.Typography.style(HISTORY_TITLE_SIZE, FontWeight.SemiBold),
                color = ink,
                maxLines = 2,
            )
            Row(
                horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                CallerChip(item.keyName)
                Text(
                    ApprovalFormatting.historyMeta(item),
                    style = monoStyle(ApprovalMetrics.chipSize),
                    color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                )
            }
        }
    }
}
