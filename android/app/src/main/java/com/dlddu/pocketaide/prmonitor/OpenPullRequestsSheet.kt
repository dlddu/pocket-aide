package com.dlddu.pocketaide.prmonitor

import android.text.format.DateUtils
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.outlined.AccountCircle
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
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.FilterPills
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.time.format.FormatStyle

private const val DISABLED_OPACITY = 0.5f
private const val CANCEL_OPACITY = 0.7f
private const val NO_CHECKS_OPACITY = 0.4f
private const val BANNER_BORDER_OPACITY = 0.4f
private const val TITLE_MAX_LINES = 2
private val CI_DOT_SIZE = 8.dp
private val BANNER_ICON_SIZE = 16.dp
private val ACCOUNT_ICON_SIZE = 20.dp
private val CENTERED_TOP = 120.dp

@Composable
fun OpenPullRequestsSheet(viewModel: OpenPullRequestsViewModel, onDismiss: () -> Unit) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var tokenInput by remember { mutableStateOf("") }
    var replacingToken by remember { mutableStateOf(false) }

    LaunchedEffect(viewModel) { viewModel.refreshIfNeeded() }

    PrMonitorSheetScaffold(title = "열린 PR", onDone = onDismiss) {
        if (state.isConnected && !replacingToken) {
            ConnectedContent(
                state = state,
                viewModel = viewModel,
                onReplaceToken = {
                    tokenInput = ""
                    replacingToken = true
                },
            )
        } else {
            ConnectForm(
                state = state,
                tokenInput = tokenInput,
                onTokenChange = { tokenInput = it },
                replacingToken = replacingToken,
                onSubmit = {
                    viewModel.connect(tokenInput) {
                        tokenInput = ""
                        replacingToken = false
                    }
                },
                onCancel = {
                    replacingToken = false
                    tokenInput = ""
                },
            )
        }
    }
}

@Composable
private fun ConnectForm(
    state: OpenPullRequestsUiState,
    tokenInput: String,
    onTokenChange: (String) -> Unit,
    replacingToken: Boolean,
    onSubmit: () -> Unit,
    onCancel: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val accent = DesignTokens.Colors.accent(PR_AREA)
    val canSubmit = !state.isConnecting && tokenInput.isNotBlank()
    Column(
        Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = DesignTokens.Spacing.xl)
            .padding(top = DesignTokens.Spacing.lg),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Text(
            "GitHub 계정을 연결하면 내가 작성했거나 리뷰어로 지정된 열린 PR과 CI 상태를 볼 수 있습니다.",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
            color = ink,
            modifier = Modifier.testTag("openprs.connect.prompt"),
        )
        Text(
            "Personal Access Token(classic은 repo 범위)을 붙여 넣으세요. 토큰은 이 기기의 키체인에만 저장되고 GitHub에만 전송됩니다.",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
        )
        val textStyle = DesignTokens.Typography.style(DesignTokens.Typography.body)
        BasicTextField(
            value = tokenInput,
            onValueChange = onTokenChange,
            singleLine = true,
            textStyle = textStyle.copy(color = ink),
            cursorBrush = SolidColor(ink),
            visualTransformation = PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(
                capitalization = KeyboardCapitalization.None,
                autoCorrectEnabled = false,
                keyboardType = KeyboardType.Password,
            ),
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(PrMonitorStyle.controlRadius))
                .background(DesignTokens.Colors.card(PR_AREA))
                .padding(DesignTokens.Spacing.md)
                .testTag("openprs.token.field"),
            decorationBox = { inner ->
                Box {
                    if (tokenInput.isEmpty()) {
                        Text("ghp_…", style = textStyle, color = ink.copy(alpha = PrMonitorStyle.TERTIARY_OPACITY))
                    }
                    inner()
                }
            },
        )
        state.connectError?.let { message ->
            Text(
                message,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = DesignTokens.StatusColors.failure(),
                modifier = Modifier.testTag("openprs.connect.error"),
            )
        }
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
            Text(
                if (state.isConnecting) "확인 중…" else "연결",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
                color = accent,
                modifier = Modifier
                    .alpha(if (canSubmit) 1f else DISABLED_OPACITY)
                    .clickable(enabled = canSubmit, role = Role.Button, onClick = onSubmit)
                    .padding(vertical = DesignTokens.Spacing.xs)
                    .testTag("openprs.connect.button"),
            )
            if (replacingToken) {
                Text(
                    "취소",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = ink.copy(alpha = CANCEL_OPACITY),
                    modifier = Modifier
                        .clickable(role = Role.Button, onClick = onCancel)
                        .padding(vertical = DesignTokens.Spacing.xs)
                        .testTag("openprs.connect.cancel"),
                )
            }
        }
    }
}

@Composable
private fun ColumnScope.ConnectedContent(
    state: OpenPullRequestsUiState,
    viewModel: OpenPullRequestsViewModel,
    onReplaceToken: () -> Unit,
) {
    state.alert?.let { alert ->
        AlertBanner(
            alert = alert,
            isLoading = state.isLoading,
            onAction = { if (alert is GitHubAlert.RateLimited) viewModel.refresh() else onReplaceToken() },
        )
    }
    AccountBar(login = state.login.orEmpty(), onReplaceToken = onReplaceToken, onDisconnect = viewModel::disconnect)
    StatusLine(state)
    FilterPills(
        area = PR_AREA,
        options = OpenPullRequestFilter.entries,
        selection = state.filter,
        onSelect = viewModel::setFilter,
        optionTag = { it.wire },
        modifier = Modifier
            .padding(horizontal = DesignTokens.Spacing.xl)
            .padding(bottom = DesignTokens.Spacing.sm),
    ) { option, _ ->
        Text(
            option.label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
    Box(Modifier.fillMaxWidth().weight(1f)) {
        PullRequestList(state = state, onRefresh = viewModel::refresh)
    }
}

@Composable
private fun AlertBanner(alert: GitHubAlert, isLoading: Boolean, onAction: () -> Unit) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val failure = DesignTokens.StatusColors.failure()
    val shape = RoundedCornerShape(PrMonitorStyle.controlRadius)
    val formatter = remember { DateTimeFormatter.ofLocalizedTime(FormatStyle.SHORT) }
    Column(
        Modifier
            .padding(horizontal = DesignTokens.Spacing.xl)
            .padding(top = DesignTokens.Spacing.md)
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(PR_AREA))
            .border(PrMonitorStyle.borderWidth, failure.copy(alpha = BANNER_BORDER_OPACITY), shape)
            .padding(DesignTokens.Spacing.md),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
            Icon(Icons.Filled.Warning, contentDescription = null, tint = failure, modifier = Modifier.size(BANNER_ICON_SIZE))
            Text(
                alert.title,
                style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
                color = failure,
                modifier = Modifier.testTag("openprs.banner.${alert.kind}"),
            )
        }
        Text(
            alert.message { formatter.format(it.atZone(ZoneId.systemDefault())) },
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = CANCEL_OPACITY),
        )
        Text(
            alert.actionTitle,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
            color = DesignTokens.Colors.accent(PR_AREA),
            modifier = Modifier
                .alpha(if (isLoading) DISABLED_OPACITY else 1f)
                .clickable(enabled = !isLoading, role = Role.Button, onClick = onAction)
                .padding(vertical = DesignTokens.Spacing.xs)
                .testTag("openprs.banner.action"),
        )
    }
}

@Composable
private fun AccountBar(login: String, onReplaceToken: () -> Unit, onDisconnect: () -> Unit) {
    val accent = DesignTokens.Colors.accent(PR_AREA)
    val actionStyle = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(Icons.Outlined.AccountCircle, contentDescription = null, tint = accent, modifier = Modifier.size(ACCOUNT_ICON_SIZE))
        Text(
            "@$login",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
            color = DesignTokens.Colors.ink(PR_AREA),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f).testTag("openprs.account.login"),
        )
        Text(
            "토큰 바꾸기",
            style = actionStyle,
            color = accent,
            modifier = Modifier
                .clickable(role = Role.Button, onClick = onReplaceToken)
                .padding(vertical = DesignTokens.Spacing.xs)
                .testTag("openprs.token.replace"),
        )
        Text(
            "연결 해제",
            style = actionStyle,
            color = DesignTokens.StatusColors.failure(),
            modifier = Modifier
                .clickable(role = Role.Button, onClick = onDisconnect)
                .padding(vertical = DesignTokens.Spacing.xs)
                .testTag("openprs.disconnect.button"),
        )
    }
}

@Composable
private fun StatusLine(state: OpenPullRequestsUiState) {
    val muted = DesignTokens.Colors.ink(PR_AREA).copy(alpha = PrMonitorStyle.MUTED_OPACITY)
    val style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)
    val formatter = remember { DateTimeFormatter.ofLocalizedTime(FormatStyle.SHORT) }
    val refreshed = state.lastRefreshedAt
    val error = state.errorMessage?.takeIf { state.alert == null && state.pullRequests.isNotEmpty() }
    if (refreshed == null && error == null) return
    Column(
        Modifier
            .padding(horizontal = DesignTokens.Spacing.xl)
            .padding(bottom = DesignTokens.Spacing.sm),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        refreshed?.let {
            Text(
                "마지막 갱신 ${formatter.format(it.atZone(ZoneId.systemDefault()))}",
                style = style,
                color = muted,
                modifier = Modifier.testTag("openprs.lastRefreshed"),
            )
        }
        error?.let {
            Text(it, style = style, color = DesignTokens.StatusColors.failure(), modifier = Modifier.testTag("openprs.refresh.error"))
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun PullRequestList(state: OpenPullRequestsUiState, onRefresh: () -> Unit) {
    val error = state.errorMessage
    if (state.isLoading && !state.hasLoaded) {
        Box(Modifier.fillMaxSize().testTag("openprs.loading"), contentAlignment = Alignment.Center) {
            CircularProgressIndicator(color = DesignTokens.Colors.accent(PR_AREA))
        }
        return
    }
    PullToRefreshBox(
        isRefreshing = state.isLoading && state.hasLoaded,
        onRefresh = onRefresh,
        modifier = Modifier.fillMaxSize(),
    ) {
        when {
            error != null && state.pullRequests.isEmpty() ->
                CenteredMessage("열린 PR을 불러오지 못했습니다", error, "openprs.error.state")
            state.pullRequests.isEmpty() ->
                CenteredMessage(
                    "열려 있는 PR이 없습니다",
                    "작성했거나 리뷰어로 지정된 PR이 열리면 여기에 표시됩니다.",
                    "openprs.empty.state",
                )
            state.visiblePullRequests.isEmpty() ->
                CenteredMessage(
                    "조건에 맞는 PR이 없습니다",
                    "「${state.filter.label}」 필터를 끄면 열린 PR ${state.pullRequests.size}개를 모두 볼 수 있습니다.",
                    "openprs.filter.empty",
                )
            else -> {
                val uriHandler = LocalUriHandler.current
                LazyColumn(Modifier.fillMaxSize()) {
                    items(state.visiblePullRequests, key = { it.id }) { pullRequest ->
                        OpenPullRequestRow(
                            pullRequest = pullRequest,
                            modifier = Modifier
                                .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xs)
                                .clickable(role = Role.Button) {
                                    pullRequest.url?.let { url -> runCatching { uriHandler.openUri(url) } }
                                }
                                .testTag("openprs.row.${pullRequest.id}"),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun CenteredMessage(title: String, detail: String, tag: String) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    Column(
        Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = DesignTokens.Spacing.xl)
            .padding(top = CENTERED_TOP)
            .testTag(tag),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Text(title, style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold), color = ink)
        Text(
            detail,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
private fun OpenPullRequestRow(pullRequest: OpenPullRequest, modifier: Modifier = Modifier) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val shape = RoundedCornerShape(PrMonitorStyle.controlRadius)
    val meta = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)
    val muted = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY)
    Column(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(PR_AREA))
            .border(PrMonitorStyle.borderWidth, DesignTokens.Colors.rule(PR_AREA), shape)
            .padding(DesignTokens.Spacing.md),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
            Text(
                pullRequest.repository,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                color = ink.copy(alpha = CANCEL_OPACITY),
                modifier = Modifier.weight(1f),
            )
            CiBadge(pullRequest.ciStatus)
        }
        Text(
            "#${pullRequest.number} ${pullRequest.title}",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
            color = ink,
            maxLines = TITLE_MAX_LINES,
            overflow = TextOverflow.Ellipsis,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
            Text(pullRequest.author, style = meta, color = muted)
            Text("·", style = meta, color = muted)
            Text(pullRequest.role.label, style = meta, color = muted)
            Spacer(Modifier.weight(1f))
            Text(relativeUpdated(pullRequest.updatedAt), style = meta, color = muted)
        }
    }
}

@Composable
private fun CiBadge(status: PullRequestCiStatus) {
    val color = when (status) {
        PullRequestCiStatus.SUCCESS -> DesignTokens.StatusColors.success()
        PullRequestCiStatus.FAILURE -> DesignTokens.StatusColors.failure()
        PullRequestCiStatus.IN_PROGRESS -> DesignTokens.StatusColors.inProgress()
        PullRequestCiStatus.NO_CHECKS -> DesignTokens.Colors.ink(PR_AREA).copy(alpha = NO_CHECKS_OPACITY)
    }
    Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs), verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(CI_DOT_SIZE).background(color, CircleShape))
        Text(status.label, style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold), color = color)
    }
}

private fun relativeUpdated(instant: Instant): String =
    DateUtils.getRelativeTimeSpanString(
        instant.toEpochMilli(),
        System.currentTimeMillis(),
        DateUtils.MINUTE_IN_MILLIS,
    ).toString()
