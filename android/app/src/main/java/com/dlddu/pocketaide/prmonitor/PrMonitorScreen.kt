package com.dlddu.pocketaide.prmonitor

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.AccountTree
import androidx.compose.material.icons.outlined.FilterList
import androidx.compose.material.icons.outlined.Notifications
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.ScreenHeader
import kotlinx.coroutines.delay

const val PR_MONITOR_ARRIVAL_HIGHLIGHT_MILLIS = 5_000L

private val HEADER_ICON_BOX = 36.dp
private val HEADER_ICON_SIZE = 16.dp
private val BADGE_VERTICAL = 4.dp
private val BADGE_LINE_GAP = 2.dp
private val SECTION_MARK_WIDTH = 3.dp
private val SECTION_MARK_HEIGHT = 12.dp
private val SECTION_RULE_HEIGHT = 1.dp

private enum class PrMonitorSheet { None, Excluded, Settings, OpenPullRequests }

@Composable
fun PrMonitorScreen(
    api: ApiClient,
    highlightedEventId: Long? = null,
    onHighlightExpired: () -> Unit = {},
    pushDenied: Boolean = false,
    onSignOut: () -> Unit = {},
) {
    val context = LocalContext.current.applicationContext
    val viewModel: PrMonitorViewModel = viewModel(key = "prmonitor") { PrMonitorViewModel(api) }
    val openPullRequests: OpenPullRequestsViewModel = viewModel(key = "prmonitor.openprs") {
        OpenPullRequestsViewModel(
            store = KeystoreGitHubCredentialStore(context),
            client = GitHubClient(),
            filterStore = SharedPreferencesOpenPullRequestFilterStore(context),
        )
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var sheet by rememberSaveable { mutableStateOf(PrMonitorSheet.None) }
    val currentOnHighlightExpired by rememberUpdatedState(onHighlightExpired)

    LaunchedEffect(viewModel) { viewModel.load() }
    LaunchedEffect(highlightedEventId) {
        if (highlightedEventId == null) return@LaunchedEffect
        viewModel.load()
        delay(PR_MONITOR_ARRIVAL_HIGHLIGHT_MILLIS)
        currentOnHighlightExpired()
    }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(PR_AREA))
            .testTag("prmonitor.screen"),
    ) {
        Column(Modifier.fillMaxSize()) {
            ScreenHeader(
                area = PR_AREA,
                title = "PR 모니터",
                trailing = {
                    Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
                        if (state.totalUnacknowledgedCount > 0) UnreadBadge(state.totalUnacknowledgedCount)
                        HeaderIcon(Icons.Outlined.AccountTree, "열린 PR", "prmonitor.openprs.button") {
                            sheet = PrMonitorSheet.OpenPullRequests
                        }
                        HeaderIcon(Icons.Outlined.Notifications, "알림 설정", "prmonitor.settings.button") {
                            sheet = PrMonitorSheet.Settings
                            viewModel.loadNotificationSettings()
                        }
                        HeaderIcon(Icons.Outlined.FilterList, "제외 레포 관리", "prmonitor.excluded.button") {
                            sheet = PrMonitorSheet.Excluded
                            viewModel.loadExcludedRepos()
                        }
                    }
                },
            )
            Content(state = state, highlightedEventId = highlightedEventId, viewModel = viewModel)
        }
    }

    when (sheet) {
        PrMonitorSheet.None -> Unit
        PrMonitorSheet.Excluded -> PrMonitorExcludedReposSheet(
            state = state,
            onAdd = viewModel::excludeRepo,
            onRemove = viewModel::removeExcludedRepo,
            onDismiss = { sheet = PrMonitorSheet.None },
        )
        PrMonitorSheet.Settings -> PrMonitorNotificationSettingsSheet(
            state = state,
            pushDenied = pushDenied,
            onEnabledChange = viewModel::setNotificationsEnabled,
            onOutcomesChange = viewModel::setNotificationOutcomes,
            onDismiss = { sheet = PrMonitorSheet.None },
            onSignOut = {
                openPullRequests.disconnect()
                onSignOut()
            },
        )
        PrMonitorSheet.OpenPullRequests -> OpenPullRequestsSheet(
            viewModel = openPullRequests,
            onDismiss = { sheet = PrMonitorSheet.None },
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun Content(state: PrMonitorUiState, highlightedEventId: Long?, viewModel: PrMonitorViewModel) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val error = state.errorMessage
    when {
        state.isLoading && state.items.isEmpty() -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            CircularProgressIndicator(color = DesignTokens.Colors.accent(PR_AREA))
        }
        error != null && state.items.isEmpty() -> Column(
            Modifier
                .fillMaxSize()
                .padding(horizontal = DesignTokens.Spacing.xl)
                .testTag("prmonitor.error.state"),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm, Alignment.CenterVertically),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text("이력을 불러오지 못했습니다", style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold), color = ink)
            Text(
                error,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
                textAlign = TextAlign.Center,
            )
            Text(
                "다시 시도",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                color = DesignTokens.Colors.accent(PR_AREA),
                modifier = Modifier
                    .clickable(role = Role.Button, onClick = viewModel::load)
                    .padding(DesignTokens.Spacing.xs),
            )
        }
        state.items.isEmpty() -> Column(
            Modifier.fillMaxSize().testTag("prmonitor.empty.state"),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm, Alignment.CenterVertically),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text("아직 도착한 알림이 없습니다", style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold), color = ink)
            Text(
                "CI가 완료되면 여기에 표시됩니다.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
            )
        }
        else -> PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
            val unread = state.unacknowledgedGroups
            val read = state.acknowledgedGroups
            LazyColumn(Modifier.fillMaxSize()) {
                if (unread.isNotEmpty()) {
                    item(key = "section.unread") { SectionHeader("미확인 · ${unread.size}개 그룹", unreadSection = true) }
                    groupItems(unread, highlightedEventId, viewModel)
                }
                if (read.isNotEmpty()) {
                    item(key = "section.read") { SectionHeader("확인 완료", unreadSection = false) }
                    groupItems(read, highlightedEventId, viewModel)
                }
                item(key = "bottom") { Box(Modifier.height(DesignTokens.Spacing.xxl)) }
            }
        }
    }
}

private fun LazyListScope.groupItems(groups: List<HistoryGroup>, highlightedEventId: Long?, viewModel: PrMonitorViewModel) {
    items(groups, key = { "group.${it.id}" }) { group ->
        PrMonitorGroupCard(
            group = group,
            highlightedEventId = highlightedEventId,
            onAcknowledge = viewModel::acknowledge,
            onAcknowledgeGroup = { viewModel.acknowledgeGroup(group) },
            modifier = Modifier.padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xs),
        )
    }
}

@Composable
private fun SectionHeader(title: String, unreadSection: Boolean) {
    val accent = DesignTokens.Colors.accent(PR_AREA)
    val rule = DesignTokens.Colors.rule(PR_AREA)
    val ink = DesignTokens.Colors.ink(PR_AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.xl)
            .padding(top = DesignTokens.Spacing.sm, bottom = DesignTokens.Spacing.xs),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(Modifier.size(width = SECTION_MARK_WIDTH, height = SECTION_MARK_HEIGHT).background(if (unreadSection) accent else rule))
        Text(
            title.uppercase(),
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold),
            color = if (unreadSection) accent else ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
        )
        Box(Modifier.weight(1f).height(SECTION_RULE_HEIGHT).background(rule))
    }
}

@Composable
private fun UnreadBadge(count: Int) {
    Column(
        Modifier
            .clip(RoundedCornerShape(PrMonitorStyle.chipRadius))
            .background(DesignTokens.Colors.accent(PR_AREA))
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = BADGE_VERTICAL)
            .testTag("prmonitor.unread.badge"),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(BADGE_LINE_GAP),
    ) {
        Text("$count", style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold), color = Color.White)
        Text("미확인", style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold), color = Color.White)
    }
}

@Composable
private fun HeaderIcon(icon: ImageVector, label: String, tag: String, onClick: () -> Unit) {
    val shape = RoundedCornerShape(PrMonitorStyle.controlRadius)
    Box(
        Modifier
            .size(HEADER_ICON_BOX)
            .clip(shape)
            .background(DesignTokens.Colors.card(PR_AREA))
            .border(PrMonitorStyle.borderWidth, DesignTokens.Colors.rule(PR_AREA), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .semantics { contentDescription = label }
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Icon(icon, contentDescription = null, tint = DesignTokens.Colors.ink(PR_AREA), modifier = Modifier.size(HEADER_ICON_SIZE))
    }
}
