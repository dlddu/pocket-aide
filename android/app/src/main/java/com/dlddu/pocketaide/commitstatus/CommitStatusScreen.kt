package com.dlddu.pocketaide.commitstatus

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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.CheckCircle
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
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
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.approvals.APPROVAL_AREA
import com.dlddu.pocketaide.approvals.ApprovalBackBar
import com.dlddu.pocketaide.approvals.ApprovalButton
import com.dlddu.pocketaide.approvals.ApprovalConfirmDialog
import com.dlddu.pocketaide.approvals.ApprovalDecision
import com.dlddu.pocketaide.approvals.ApprovalErrorBanner
import com.dlddu.pocketaide.approvals.ApprovalFormatting
import com.dlddu.pocketaide.approvals.ApprovalMetrics
import com.dlddu.pocketaide.approvals.ApprovalSectionHeader
import com.dlddu.pocketaide.approvals.CallerChip
import com.dlddu.pocketaide.approvals.CountBadge
import com.dlddu.pocketaide.approvals.monoStyle
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel

private val HEADER_DOT_SIZE = 8.dp

@Composable
fun CommitStatusScreen(
    viewModel: CommitStatusViewModel,
    openedRequestId: String? = null,
    onOpenedRequestConsumed: () -> Unit = {},
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val opened = state.opened

    LaunchedEffect(opened == null) { if (opened == null) viewModel.load() }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.load() }
    LaunchedEffect(openedRequestId) {
        val id = openedRequestId ?: return@LaunchedEffect
        onOpenedRequestConsumed()
        viewModel.open(id)
    }
    BackHandler(enabled = opened != null) { viewModel.show(null) }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(APPROVAL_AREA)),
    ) {
        if (opened == null) {
            RequestList(state = state, viewModel = viewModel)
        } else {
            CommitStatusDetail(item = opened, api = viewModel.api, onBack = { viewModel.show(null) })
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RequestList(state: CommitStatusUiState, viewModel: CommitStatusViewModel) {
    val area = APPROVAL_AREA
    val now = System.currentTimeMillis()
    Column(Modifier.fillMaxSize()) {
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
                AreaLabel(area = area, text = "COMMIT · STATUS", showsDot = true, dotSize = HEADER_DOT_SIZE)
                Text(
                    "커밋 상태",
                    style = DesignTokens.Typography.style(ApprovalMetrics.headerTitleSize, FontWeight.Bold),
                    color = DesignTokens.Colors.ink(area),
                    modifier = Modifier.testTag("commitStatus.title"),
                )
            }
            CountBadge(count = state.pendingCount, label = "대기", countTag = "commitStatus.pending.count")
        }
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
                            title = "커밋 상태 요청을 불러오지 못했습니다",
                            message = message,
                            bannerTag = "commitStatus.error.banner",
                            retryTag = "commitStatus.error.retry",
                            onRetry = viewModel::load,
                        )
                    }
                }
                if (state.items.isEmpty()) {
                    item(key = "empty") {
                        EmptyState(isLoading = state.isLoading || (!state.hasLoaded && state.errorMessage == null), showsMessage = state.errorMessage == null)
                    }
                } else {
                    item(key = "pending") {
                        ApprovalSectionHeader(
                            title = CommitStatusFormatting.pendingHeader(state.pendingCount),
                            tint = DesignTokens.Colors.accent(area),
                            tag = "commitStatus.section.pending",
                        )
                    }
                    items(state.items, key = { it.id }) { item ->
                        RequestRow(item = item, nowMillis = now, onClick = { viewModel.show(item) })
                    }
                }
                if (state.history.isNotEmpty()) {
                    item(key = "history") {
                        ApprovalSectionHeader(
                            title = CommitStatusFormatting.HISTORY_HEADER,
                            tint = DesignTokens.Colors.accent(area),
                            tag = "commitStatus.section.history",
                        )
                    }
                    items(state.history, key = { "h-${it.id}" }) { item ->
                        RequestRow(item = item, nowMillis = now, onClick = { viewModel.show(item) })
                    }
                }
            }
            if (state.isLoading && state.items.isNotEmpty() && !state.isRefreshing) {
                CircularProgressIndicator(
                    color = DesignTokens.Colors.accent(area),
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .padding(top = DesignTokens.Spacing.xs)
                        .size(ApprovalMetrics.iconSize),
                )
            }
        }
    }
}

@Composable
private fun EmptyState(isLoading: Boolean, showsMessage: Boolean) {
    val area = APPROVAL_AREA
    Column(
        Modifier
            .fillMaxWidth()
            .padding(vertical = DesignTokens.Spacing.xxl)
            .testTag("commitStatus.empty"),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        if (isLoading) {
            CircularProgressIndicator(color = DesignTokens.Colors.accent(area))
        } else if (showsMessage) {
            Icon(Icons.Outlined.CheckCircle, contentDescription = null, tint = DesignTokens.Colors.accent(area))
            Text(
                "기록을 기다리는 커밋 상태가 없습니다",
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm, FontWeight.SemiBold),
                color = DesignTokens.Colors.ink(area),
            )
            Text(
                "에이전트가 커밋 상태를 요청하면 여기에 쌓이고 알림이 옵니다.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                textAlign = TextAlign.Center,
            )
        }
    }
}

@Composable
private fun RequestRow(item: CommitStatusRequest, nowMillis: Long, onClick: () -> Unit) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val accent = DesignTokens.Colors.accent(area)
    val pending = item.status == CommitStatusRequestStatus.PENDING
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    Row(
        Modifier
            .fillMaxWidth()
            .height(IntrinsicSize.Min)
            .alpha(if (pending) 1f else ApprovalMetrics.REVOKED_OPACITY)
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .semantics(mergeDescendants = true) {}
            .testTag("commitStatus.row.${item.id}"),
    ) {
        Box(
            Modifier
                .width(ApprovalMetrics.accentStripWidth)
                .fillMaxHeight()
                .background(if (pending) accent else DesignTokens.Colors.rule(area)),
        )
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
                    if (pending) ApprovalFormatting.requestTime(item.createdAt, nowMillis) else CommitStatusFormatting.statusLabel(item.status),
                    style = monoStyle(ApprovalMetrics.chipSize, FontWeight.Bold),
                    color = accent,
                )
            }
            Text(
                CommitStatusFormatting.title(item),
                style = DesignTokens.Typography.style(ApprovalMetrics.rowTitleSize, FontWeight.SemiBold),
                color = ink,
                maxLines = 2,
            )
            Text(CommitStatusFormatting.target(item), style = monoStyle(DesignTokens.Typography.captionXs), color = ink)
            Text(
                item.requester,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                maxLines = 1,
            )
            if (pending && item.lastError.isNotEmpty()) {
                Text(
                    CommitStatusFormatting.writeFailedNotice(item.lastError),
                    style = DesignTokens.Typography.style(ApprovalMetrics.chipSize, FontWeight.Bold),
                    color = DesignTokens.StatusColors.failure(),
                    maxLines = 2,
                )
            }
        }
    }
}

@Composable
private fun CommitStatusDetail(item: CommitStatusRequest, api: ApiClient, onBack: () -> Unit) {
    val viewModel: CommitStatusDetailViewModel = viewModel(key = "commitStatus.detail.${item.id}") {
        CommitStatusDetailViewModel(item, api)
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var confirming by remember { mutableStateOf<ApprovalDecision?>(null) }
    val area = APPROVAL_AREA
    val current = state.item

    LaunchedEffect(viewModel) { viewModel.reload() }

    Column(Modifier.fillMaxSize()) {
        ApprovalBackBar(onBack = onBack) { CallerChip(current.keyName) }
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
                CommitStatusFormatting.title(current),
                style = DesignTokens.Typography.style(ApprovalMetrics.detailTitleSize, FontWeight.Bold),
                color = DesignTokens.Colors.ink(area),
                modifier = Modifier.testTag("commitStatus.detail.title"),
            )
            Text(
                CommitStatusFormatting.target(current),
                style = monoStyle(DesignTokens.Typography.captionSm),
                color = DesignTokens.Colors.ink(area).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
            )
            val notice = state.notice
            when {
                notice != null -> Banner(notice, "commitStatus.detail.notice")
                current.status == CommitStatusRequestStatus.PENDING && current.lastError.isNotEmpty() ->
                    Banner(CommitStatusFormatting.writeFailedNotice(current.lastError), "commitStatus.detail.lastError")
                state.loadError != null -> ApprovalErrorBanner(
                    title = "최신 상태를 불러오지 못했습니다",
                    message = null,
                    bannerTag = null,
                    retryTag = "commitStatus.detail.retry",
                    onRetry = viewModel::reload,
                    showsIcon = false,
                    outlined = false,
                )
            }
            if (current.reason.isNotEmpty()) AgentReason(current.reason)
            Facts(current)
            if (current.status == CommitStatusRequestStatus.PENDING) {
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
                        enabled = state.canDecide,
                        tag = "commitStatus.detail.reject",
                        modifier = Modifier.weight(1f),
                    ) { confirming = ApprovalDecision.REJECT }
                    ApprovalButton(
                        title = "승인",
                        primary = true,
                        enabled = state.canDecide,
                        tag = "commitStatus.detail.approve",
                        modifier = Modifier.weight(1f),
                    ) { confirming = ApprovalDecision.APPROVE }
                }
            }
        }
    }

    confirming?.let { decision ->
        ApprovalConfirmDialog(
            title = CommitStatusFormatting.confirmTitle(decision),
            message = CommitStatusFormatting.confirmMessage(decision, current),
            confirmLabel = if (decision == ApprovalDecision.APPROVE) "승인" else "거절",
            destructive = false,
            onConfirm = { viewModel.decide(decision) },
            onDismiss = { confirming = null },
        )
    }
}

@Composable
private fun Banner(text: String, tag: String) {
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
            .testTag(tag),
    )
}

@Composable
private fun AgentReason(reason: String) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
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
            )
        }
        Text(
            reason,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
            color = ink,
            modifier = Modifier
                .fillMaxWidth()
                .clip(shape)
                .background(DesignTokens.Colors.card(area))
                .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
                .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
                .testTag("commitStatus.detail.reason"),
        )
    }
}

@Composable
private fun Facts(item: CommitStatusRequest) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    val uriHandler = LocalUriHandler.current
    val facts = buildList {
        add(Triple("저장소", "${item.owner}/${item.repo}", "repo"))
        add(Triple("커밋", item.sha, "sha"))
        add(Triple("context", item.context, "context"))
        add(Triple("state", item.state, "state"))
        if (item.description.isNotEmpty()) add(Triple("설명", item.description, "description"))
        add(Triple("보낸 키", item.keyName, "key"))
        add(Triple("요청자", item.requester, "requester"))
        val session = item.sessionName.ifEmpty { if (item.sessionId.isEmpty()) "" else "제목 없음 · ${item.sessionId.take(8)}" }
        if (session.isNotEmpty()) add(Triple("세션", session, "session"))
        add(Triple("요청 시각", ApprovalFormatting.clockTime(item.createdAt), "createdAt"))
        add(Triple("상태", CommitStatusFormatting.statusLabel(item.status), "status"))
        if (item.status != CommitStatusRequestStatus.PENDING) {
            add(Triple("처리 시각", ApprovalFormatting.clockTime(item.processedAt), "processedAt"))
        }
        item.githubStatusId?.let { add(Triple("GitHub status id", it.toString(), "githubStatusId")) }
    }
    Column(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(area), shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        facts.forEach { (label, value, id) ->
            Fact(label = label, value = value, tag = "commitStatus.detail.$id")
        }
        if (item.targetUrl.isNotEmpty()) {
            Column(verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.chipVertical)) {
                Text(
                    "링크",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                    color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                )
                Text(
                    item.targetUrl,
                    style = monoStyle(DesignTokens.Typography.captionXs).copy(textDecoration = TextDecoration.Underline),
                    color = DesignTokens.Colors.accent(area),
                    modifier = Modifier
                        .clickable(role = Role.Button) { runCatching { uriHandler.openUri(item.targetUrl) } }
                        .testTag("commitStatus.detail.targetUrl"),
                )
            }
        }
    }
}

@Composable
private fun Fact(label: String, value: String, tag: String) {
    val ink = DesignTokens.Colors.ink(APPROVAL_AREA)
    Column(verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.chipVertical)) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        )
        Text(value, style = monoStyle(DesignTokens.Typography.captionXs), color = ink, modifier = Modifier.testTag(tag))
    }
}
