package com.dlddu.pocketaide.sessions

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
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
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel

private val TITLE_SIZE = 26.sp
private val HEADER_LABEL_GAP = 6.dp
private val CREATE_ICON_SIZE = 13.dp
private val CREATE_VERTICAL = 6.dp
private val ROW_VERTICAL = 5.dp

@Composable
fun SessionsScreen(api: ApiClient) {
    val viewModel: SessionsViewModel = viewModel(key = "sessions") { SessionsViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var openSessionId by rememberSaveable { mutableStateOf<String?>(null) }
    var openSnapshot by remember { mutableStateOf<AgentSession?>(null) }
    var showsCreate by remember { mutableStateOf(false) }
    val open: (AgentSession) -> Unit = { session ->
        openSnapshot = session
        openSessionId = session.id
    }

    LaunchedEffect(viewModel, openSessionId) {
        if (openSessionId == null) viewModel.load()
    }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (openSessionId == null) viewModel.load()
    }

    val openSession = openSessionId?.let { id ->
        openSnapshot?.takeIf { it.id == id } ?: state.sessions.firstOrNull { it.id == id }
    }
    LaunchedEffect(openSessionId, openSession, state.hasLoaded) {
        if (openSessionId != null && openSession == null && state.hasLoaded) openSessionId = null
    }
    if (openSession != null) {
        BackHandler { openSessionId = null }
        SessionDetailScreen(
            session = openSession,
            type = state.workloadType(openSession),
            api = viewModel.api,
            onClose = { openSessionId = null },
        )
        return
    }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(SESSIONS_AREA))
            .testTag("sessions.screen"),
    ) {
        Column(Modifier.fillMaxSize()) {
            Header(
                onCreate = {
                    showsCreate = true
                    viewModel.loadConfig()
                },
            )
            SessionList(
                state = state,
                onRefresh = viewModel::refresh,
                onRetry = viewModel::load,
                onOpen = open,
            )
        }
    }

    if (showsCreate) {
        SessionCreateSheet(
            config = state.config,
            configFailed = state.configFailed,
            onRetryConfig = viewModel::loadConfig,
            isCreating = state.isCreating,
            error = state.createError,
            onCreate = { name, workload, model ->
                viewModel.create(name, workload, model) { created ->
                    showsCreate = false
                    open(created)
                }
            },
            onCancel = {
                viewModel.clearCreateError()
                showsCreate = false
            },
        )
    }
}

@Composable
private fun Header(onCreate: () -> Unit) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val surface = DesignTokens.Colors.surface(SESSIONS_AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md),
        verticalAlignment = Alignment.Bottom,
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(HEADER_LABEL_GAP)) {
            AreaLabel(area = SESSIONS_AREA, text = "AGENT · SESSIONS", showsDot = true)
            Text(
                "에이전트 세션",
                style = DesignTokens.Typography.style(TITLE_SIZE, FontWeight.Bold),
                color = ink,
                modifier = Modifier.testTag("sessions.title"),
            )
        }
        Spacer(Modifier.weight(1f))
        Row(
            Modifier
                .clip(CircleShape)
                .background(ink)
                .clickable(role = Role.Button, onClick = onCreate)
                .padding(horizontal = DesignTokens.Spacing.md, vertical = CREATE_VERTICAL)
                .testTag("sessions.create.button"),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(Icons.Filled.Add, contentDescription = null, tint = surface, modifier = Modifier.size(CREATE_ICON_SIZE))
            Text(
                "새 세션",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                color = surface,
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun SessionList(
    state: SessionsUiState,
    onRefresh: () -> Unit,
    onRetry: () -> Unit,
    onOpen: (AgentSession) -> Unit,
) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    PullToRefreshBox(
        isRefreshing = state.isRefreshing,
        onRefresh = onRefresh,
        modifier = Modifier.fillMaxSize(),
    ) {
        LazyColumn(Modifier.fillMaxSize()) {
            state.error?.let { error ->
                item(key = "error") {
                    SessionErrorBanner(
                        error = error,
                        onRetry = onRetry,
                        modifier = Modifier.padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xs),
                    )
                }
            }
            if (state.sessions.isEmpty()) {
                item(key = "empty") {
                    Box(
                        Modifier
                            .fillMaxWidth()
                            .padding(horizontal = DesignTokens.Spacing.xl + DesignTokens.Spacing.xs, vertical = DesignTokens.Spacing.xl),
                        contentAlignment = Alignment.Center,
                    ) {
                        if (!state.hasLoaded || (state.isLoading && state.error == null)) {
                            CircularProgressIndicator(color = DesignTokens.Colors.accent(SESSIONS_AREA))
                        } else if (state.error == null) {
                            Text(
                                "아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요.",
                                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                                color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
                                modifier = Modifier.fillMaxWidth().testTag("sessions.empty.state"),
                            )
                        }
                    }
                }
            } else {
                items(state.sessions, key = { it.id }) { session ->
                    SessionRow(
                        session = session,
                        type = state.workloadType(session),
                        modifier = Modifier
                            .padding(horizontal = DesignTokens.Spacing.xl, vertical = ROW_VERTICAL)
                            .clickable(role = Role.Button) { onOpen(session) }
                            .testTag("sessions.row.${session.id}"),
                    )
                }
            }
        }
    }
}
