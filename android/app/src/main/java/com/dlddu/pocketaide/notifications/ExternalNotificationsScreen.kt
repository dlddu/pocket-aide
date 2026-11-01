package com.dlddu.pocketaide.notifications

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Notifications
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.approvals.ApprovalErrorBanner
import com.dlddu.pocketaide.approvals.ApprovalMetrics
import com.dlddu.pocketaide.approvals.ApprovalSectionHeader
import com.dlddu.pocketaide.approvals.CallerChip
import com.dlddu.pocketaide.approvals.CountBadge
import com.dlddu.pocketaide.approvals.monoStyle
import com.dlddu.pocketaide.approvals.rememberNowMillis
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel
import com.dlddu.pocketaide.widget.WidgetRefresher

private val AREA = Area.PrMonitor
private val HEADER_DOT_SIZE = 8.dp
private val EMPTY_ICON_SIZE = 22.dp
private val EMPTY_CIRCLE_SIZE = 56.dp
private val EMPTY_TOP = 96.dp
private val EMPTY_TITLE_SIZE = 15.sp
private val UNREAD_BADGE_SIZE = 10.sp
private val UNREAD_BADGE_HORIZONTAL = 6.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ExternalNotificationsScreen(
    api: ApiClient,
    reloadToken: Int = 0,
    openedNotificationId: Long? = null,
    onOpenedNotificationConsumed: () -> Unit = {},
) {
    val appContext = LocalContext.current.applicationContext
    val viewModel: ExternalNotificationsViewModel = viewModel(key = "notifications") {
        ExternalNotificationsViewModel(api, onMarkedRead = { WidgetRefresher.reloadAll(appContext) })
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val now = rememberNowMillis()

    LaunchedEffect(reloadToken) { viewModel.load() }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.load() }
    LaunchedEffect(openedNotificationId) {
        val id = openedNotificationId ?: return@LaunchedEffect
        onOpenedNotificationConsumed()
        viewModel.open(id)
    }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(AREA))
            .testTag("notifications.screen"),
    ) {
        Column(Modifier.fillMaxSize()) {
            Header(unreadCount = state.unreadCount)
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
                                title = "알림 기록을 불러오지 못했습니다",
                                message = message,
                                bannerTag = "notifications.error.banner",
                                retryTag = "notifications.error.retry",
                                onRetry = viewModel::load,
                            )
                        }
                    }
                    if (state.items.isEmpty()) {
                        item(key = "empty") {
                            EmptyState(
                                isLoading = state.isLoading || (!state.hasLoaded && state.errorMessage == null),
                                showsMessage = state.errorMessage == null,
                            )
                        }
                    } else {
                        item(key = "section") {
                            ApprovalSectionHeader(title = "최근 알림", tint = DesignTokens.Colors.accent(AREA))
                        }
                        items(state.items, key = { it.id }) { item ->
                            NotificationRowCard(item = item, nowMillis = now, onClick = { viewModel.select(item) })
                        }
                        if (state.hasMore) {
                            item(key = "loadingMore") {
                                LaunchedEffect(state.items.lastOrNull()?.id) { viewModel.loadMore() }
                                Text(
                                    "이전 기록 불러오는 중…",
                                    style = monoStyle(ApprovalMetrics.loadingMoreSize),
                                    color = DesignTokens.Colors.ink(AREA).copy(alpha = ApprovalMetrics.FAINT_OPACITY),
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = DesignTokens.Spacing.sm),
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    state.selected?.let { item ->
        ExternalNotificationDetailSheet(item = item, onClose = viewModel::dismissDetail)
    }
}

@Composable
private fun Header(unreadCount: Int) {
    val ink = DesignTokens.Colors.ink(AREA)
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
    ) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(ApprovalMetrics.rowSpacing)) {
            AreaLabel(area = AREA, text = "NOTIFY · LOG", showsDot = true, dotSize = HEADER_DOT_SIZE)
            Text(
                "알림 기록",
                style = DesignTokens.Typography.style(ApprovalMetrics.headerTitleSize, FontWeight.Bold),
                color = ink,
                modifier = Modifier.testTag("notifications.title"),
            )
            Text(
                "자동화가 보낸 알림 · 보낸 호출자와 요청 주체",
                style = DesignTokens.Typography.style(ApprovalMetrics.noticeSize),
                color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
            )
        }
        if (unreadCount > 0) {
            CountBadge(count = unreadCount, label = "새 알림", countTag = "notifications.unread.count")
        }
    }
}

@Composable
private fun EmptyState(isLoading: Boolean, showsMessage: Boolean) {
    val ink = DesignTokens.Colors.ink(AREA)
    Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
        when {
            isLoading -> CircularProgressIndicator(color = DesignTokens.Colors.accent(AREA))
            showsMessage -> Column(
                Modifier.padding(top = EMPTY_TOP),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Box(
                    Modifier
                        .size(EMPTY_CIRCLE_SIZE)
                        .background(DesignTokens.Colors.soft(AREA), CircleShape),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        Icons.Outlined.Notifications,
                        contentDescription = null,
                        tint = DesignTokens.Colors.accent(AREA),
                        modifier = Modifier.size(EMPTY_ICON_SIZE),
                    )
                }
                Text(
                    "받은 알림이 없어요.",
                    style = DesignTokens.Typography.style(EMPTY_TITLE_SIZE, FontWeight.SemiBold),
                    color = ink,
                    modifier = Modifier
                        .padding(top = DesignTokens.Spacing.lg)
                        .testTag("notifications.empty"),
                )
                Text(
                    "등록한 호출자가 보낸 알림이 여기에 남습니다.",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                    modifier = Modifier.padding(top = ApprovalMetrics.rowSpacing),
                )
            }
        }
    }
}

@Composable
private fun NotificationRowCard(item: ExternalNotification, nowMillis: Long, onClick: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    val faint = ink.copy(alpha = ApprovalMetrics.FAINT_OPACITY)
    val secondary = ink.copy(alpha = ApprovalMetrics.SECONDARY_OPACITY)
    Row(
        Modifier
            .fillMaxWidth()
            .height(IntrinsicSize.Min)
            .clip(shape)
            .background(DesignTokens.Colors.card(AREA))
            .border(ApprovalMetrics.ruleWidth, DesignTokens.Colors.rule(AREA), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .semantics(mergeDescendants = true) {}
            .testTag("notifications.row.${item.id}"),
    ) {
        Box(
            Modifier
                .width(ApprovalMetrics.accentStripWidth)
                .fillMaxHeight()
                .background(if (item.isUnread) accent else Color.Transparent),
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
            Row(
                horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(Modifier.weight(1f)) {
                    CallerChip(ExternalNotificationFormatting.callerLabel(item))
                }
                if (item.isUnread) {
                    Text(
                        "새 알림",
                        style = monoStyle(UNREAD_BADGE_SIZE, FontWeight.Bold),
                        color = Color.White,
                        modifier = Modifier
                            .clip(RoundedCornerShape(ApprovalMetrics.chipRadius))
                            .background(accent)
                            .padding(horizontal = UNREAD_BADGE_HORIZONTAL, vertical = ApprovalMetrics.chipVertical)
                            .testTag("notifications.row.unread"),
                    )
                }
                Text(
                    ExternalNotificationFormatting.time(item, nowMillis),
                    style = monoStyle(ApprovalMetrics.chipSize),
                    color = faint,
                )
            }
            Text(
                item.title,
                style = DesignTokens.Typography.style(ApprovalMetrics.rowTitleSize, FontWeight.SemiBold),
                color = ink,
                maxLines = 2,
            )
            if (item.body.isNotEmpty()) {
                Text(
                    item.body,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = secondary,
                    maxLines = 3,
                )
            }
            Row(
                horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                verticalAlignment = Alignment.Top,
            ) {
                Text(
                    ExternalNotificationFormatting.requesterLabel(item),
                    style = monoStyle(ApprovalMetrics.chipSize),
                    color = if (item.requester.isEmpty()) faint else secondary,
                    maxLines = 2,
                    modifier = Modifier.weight(1f),
                )
                Text(
                    ExternalNotificationFormatting.deliveryLabel(item),
                    style = monoStyle(ApprovalMetrics.chipSize),
                    color = faint,
                )
            }
        }
    }
}
