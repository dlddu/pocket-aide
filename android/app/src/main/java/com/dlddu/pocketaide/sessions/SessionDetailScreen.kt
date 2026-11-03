package com.dlddu.pocketaide.sessions

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
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.ArrowBack
import androidx.compose.material.icons.filled.ArrowUpward
import androidx.compose.material.icons.outlined.Archive
import androidx.compose.material.icons.outlined.Delete
import androidx.compose.material.icons.outlined.MoreHoriz
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.withFrameNanos
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.repeatOnLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.PlainTextField

private val TOP_BAR_ICON_SIZE = 22.dp
private val TOP_BAR_TOUCH = 40.dp
private val STATUS_LINE_GAP = 2.dp
private val PHASE_GAP = 4.dp
private val COMPOSER_RADIUS = 18.dp
private val COMPOSER_VERTICAL = 10.dp
private val SEND_SIZE = 40.dp
private val SEND_ICON_SIZE = 18.dp
private val RESTORE_VERTICAL = 12.dp
private val INLINE_SPINNER_SIZE = 16.dp
private val INLINE_SPINNER_STROKE = 2.dp
private const val COMPOSER_MAX_LINES = 6
private const val PHASE_OPACITY = 0.75f
private const val TRIMMED_OPACITY = 0.5f
private const val SNAPSHOT_BODY_OPACITY = 0.65f
private const val SEND_DISABLED_OPACITY = 0.4f

@Composable
fun SessionDetailScreen(
    session: AgentSession,
    type: AgentSessionWorkloadType?,
    api: ApiClient,
    onClose: () -> Unit,
) {
    val viewModel: SessionDetailViewModel = viewModel(key = "sessions.detail.${session.id}") {
        SessionDetailViewModel(session, type, api)
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current
    var prompt by rememberSaveable(session.id) { mutableStateOf("") }
    var confirmsDelete by remember { mutableStateOf(false) }
    val currentOnClose by rememberUpdatedState(onClose)

    LaunchedEffect(viewModel, lifecycleOwner) {
        lifecycleOwner.lifecycle.repeatOnLifecycle(Lifecycle.State.RESUMED) { viewModel.pollWhileVisible() }
    }
    LaunchedEffect(state.deleted) {
        if (state.deleted) currentOnClose()
    }

    Column(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(SESSIONS_AREA))
            .imePadding()
            .testTag("sessions.detail.screen"),
    ) {
        TopBar(
            title = state.session.name,
            state = state,
            onBack = onClose,
            onArchive = viewModel::archive,
            onDelete = { confirmsDelete = true },
        )
        StatusStrip(state)
        state.error?.let { error ->
            SessionErrorBanner(
                error = error,
                onRetry = viewModel::retry,
                modifier = Modifier
                    .padding(horizontal = DesignTokens.Spacing.lg)
                    .padding(bottom = DesignTokens.Spacing.sm),
            )
        }
        if (state.session.state == AgentSessionState.SNAPSHOT) {
            SnapshotPanel(isRestoring = state.isRestoring, isBusy = state.isBusy, onRestore = viewModel::restore)
        } else {
            OutputArea(state)
        }
        Composer(
            prompt = prompt,
            onPromptChange = { prompt = it },
            state = state,
            onSend = {
                val text = prompt.trim()
                viewModel.send(text) { prompt = "" }
            },
        )
    }

    if (confirmsDelete) {
        val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
        AlertDialog(
            onDismissRequest = { confirmsDelete = false },
            containerColor = DesignTokens.Colors.surface(SESSIONS_AREA),
            titleContentColor = ink,
            textContentColor = ink.copy(alpha = SNAPSHOT_BODY_OPACITY),
            title = { Text("세션을 삭제할까요?") },
            text = {
                Text("되돌릴 수 없어요. 환경과 세션 기록이 지워져요. 이미 보관된 아카이브는 보존 정책에 따라 바로 지워지지 않을 수 있어요.")
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        confirmsDelete = false
                        viewModel.delete()
                    },
                    modifier = Modifier.testTag("sessions.detail.delete.confirm"),
                ) {
                    Text("삭제", color = DesignTokens.StatusColors.failure())
                }
            },
            dismissButton = {
                TextButton(onClick = { confirmsDelete = false }) {
                    Text("취소", color = DesignTokens.Colors.accent(SESSIONS_AREA))
                }
            },
        )
    }
}

@Composable
private fun TopBar(
    title: String,
    state: SessionDetailUiState,
    onBack: () -> Unit,
    onArchive: () -> Unit,
    onDelete: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    var menuOpen by remember { mutableStateOf(false) }
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.sm, vertical = DesignTokens.Spacing.xs),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier
                .size(TOP_BAR_TOUCH)
                .clip(CircleShape)
                .clickable(role = Role.Button, onClick = onBack)
                .semantics { contentDescription = "뒤로" }
                .testTag("sessions.detail.back.button"),
            contentAlignment = Alignment.Center,
        ) {
            Icon(Icons.AutoMirrored.Outlined.ArrowBack, contentDescription = null, tint = ink, modifier = Modifier.size(TOP_BAR_ICON_SIZE))
        }
        Text(
            title,
            style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.SemiBold),
            color = ink,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f).padding(horizontal = DesignTokens.Spacing.xs),
        )
        Box {
            Box(
                Modifier
                    .size(TOP_BAR_TOUCH)
                    .clip(CircleShape)
                    .clickable(role = Role.Button) { menuOpen = true }
                    .semantics { contentDescription = "세션 동작" }
                    .testTag("sessions.detail.actions.menu"),
                contentAlignment = Alignment.Center,
            ) {
                if (state.isArchiving || state.isDeleting) {
                    CircularProgressIndicator(
                        color = ink,
                        strokeWidth = INLINE_SPINNER_STROKE,
                        modifier = Modifier.size(INLINE_SPINNER_SIZE),
                    )
                } else {
                    Icon(Icons.Outlined.MoreHoriz, contentDescription = null, tint = ink, modifier = Modifier.size(TOP_BAR_ICON_SIZE))
                }
            }
            DropdownMenu(
                expanded = menuOpen,
                onDismissRequest = { menuOpen = false },
                containerColor = DesignTokens.Colors.card(SESSIONS_AREA),
            ) {
                DropdownMenuItem(
                    text = { Text("지금 아카이브", color = ink) },
                    leadingIcon = { Icon(Icons.Outlined.Archive, contentDescription = null, tint = ink) },
                    enabled = state.canArchive,
                    onClick = {
                        menuOpen = false
                        onArchive()
                    },
                    modifier = Modifier.testTag("sessions.detail.archive.button"),
                )
                val failure = DesignTokens.StatusColors.failure()
                DropdownMenuItem(
                    text = { Text("세션 삭제", color = failure) },
                    leadingIcon = { Icon(Icons.Outlined.Delete, contentDescription = null, tint = failure) },
                    enabled = !state.isBusy,
                    onClick = {
                        menuOpen = false
                        onDelete()
                    },
                    modifier = Modifier.testTag("sessions.detail.delete.button"),
                )
            }
        }
    }
}

@Composable
private fun StatusStrip(state: SessionDetailUiState) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    Column(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.sm),
        verticalArrangement = Arrangement.spacedBy(STATUS_LINE_GAP),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically) {
            SessionStateBadge(state.session.state)
            Text(
                state.session.kindLabel(state.type),
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f, fill = false),
            )
            Spacer(Modifier.weight(1f))
            state.phase.label?.let { label ->
                Row(
                    Modifier.testTag("sessions.detail.phase"),
                    horizontalArrangement = Arrangement.spacedBy(PHASE_GAP),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    if (state.phase == AgentSessionResponsePhase.QUEUED || state.phase == AgentSessionResponsePhase.STREAMING) {
                        SmallSpinner(ink.copy(alpha = PHASE_OPACITY))
                    }
                    Text(
                        label,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                        color = ink.copy(alpha = PHASE_OPACITY),
                        maxLines = 1,
                    )
                }
            }
        }
        Text(
            if (state.isArchiving) {
                "아카이브하는 중… 환경을 반납하고 작업 내용을 보관해요"
            } else {
                AgentSessionFormatting.stateDetail(state.session)
            },
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.testTag("sessions.detail.state"),
        )
    }
}

@Composable
private fun ColumnScope.OutputArea(state: SessionDetailUiState) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    val scroll = rememberScrollState()
    LaunchedEffect(state.output.nextOffset) {
        withFrameNanos { }
        scroll.animateScrollTo(scroll.maxValue)
    }
    Box(
        Modifier
            .weight(1f)
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.lg)
            .clip(shape)
            .background(DesignTokens.Colors.card(SESSIONS_AREA))
            .border(SessionStyle.borderWidth, DesignTokens.Colors.rule(SESSIONS_AREA), shape),
    ) {
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(scroll)
                .padding(DesignTokens.Spacing.lg),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        ) {
            if (state.output.isTrimmed) {
                Text(
                    "앞부분 출력은 생략했어요.",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                    color = ink.copy(alpha = TRIMMED_OPACITY),
                )
            }
            when {
                !state.hasLoadedOutput -> Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = DesignTokens.Colors.accent(SESSIONS_AREA))
                }
                state.displayText.isEmpty() -> Text(
                    if (state.isTerminal) {
                        "아직 출력이 없어요. 아래에 명령을 입력해 보세요."
                    } else {
                        "아직 출력이 없어요. 아래에 프롬프트를 적어 에이전트에게 일을 맡겨 보세요."
                    },
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                    color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
                    modifier = Modifier.testTag("sessions.detail.output.empty"),
                )
                !state.isTerminal && state.transcript.hasTurns -> SessionTranscriptView(state.transcript)
                else -> SelectionContainer {
                    Text(
                        state.displayText,
                        style = monospaceStyle(),
                        color = ink,
                        modifier = Modifier.fillMaxWidth().testTag("sessions.detail.output"),
                    )
                }
            }
        }
    }
}

@Composable
private fun ColumnScope.SnapshotPanel(isRestoring: Boolean, isBusy: Boolean, onRestore: () -> Unit) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val surface = DesignTokens.Colors.surface(SESSIONS_AREA)
    Column(
        Modifier
            .weight(1f)
            .fillMaxWidth()
            .padding(DesignTokens.Spacing.xl),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Text(
            "아카이브된 세션이에요",
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
            color = ink,
        )
        Text(
            "오래 쓰지 않아 환경을 반납한 상태예요. 복원하면 새 환경이 뜨고 이전 출력이 그대로 이어져요. 수십 초 걸릴 수 있어요.",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
            color = ink.copy(alpha = SNAPSHOT_BODY_OPACITY),
        )
        Row(
            Modifier
                .fillMaxWidth()
                .alpha(if (isBusy) SessionStyle.DISABLED_OPACITY else 1f)
                .clip(CircleShape)
                .background(ink)
                .clickable(enabled = !isBusy, role = Role.Button, onClick = onRestore)
                .padding(vertical = RESTORE_VERTICAL)
                .testTag("sessions.detail.restore.button"),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm, Alignment.CenterHorizontally),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            if (isRestoring) {
                CircularProgressIndicator(color = surface, strokeWidth = INLINE_SPINNER_STROKE, modifier = Modifier.size(INLINE_SPINNER_SIZE))
            }
            Text(
                if (isRestoring) "복원하는 중…" else "복원",
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold),
                color = surface,
            )
        }
    }
}

@Composable
private fun Composer(prompt: String, onPromptChange: (String) -> Unit, state: SessionDetailUiState, onSend: () -> Unit) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val surface = DesignTokens.Colors.surface(SESSIONS_AREA)
    val trimmed = prompt.trim()
    val tooLarge = AgentSessionPrompt.exceedsLimit(prompt)
    val canSend = state.canSend && trimmed.isNotEmpty() && !tooLarge
    val snapshot = state.session.state == AgentSessionState.SNAPSHOT
    val fieldShape = RoundedCornerShape(COMPOSER_RADIUS)
    Column(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.sm),
        verticalArrangement = Arrangement.spacedBy(PHASE_GAP),
    ) {
        if (tooLarge) {
            Text(
                "1 MiB를 넘는 프롬프트는 보낼 수 없어요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = DesignTokens.StatusColors.failure(),
                modifier = Modifier.testTag("sessions.detail.prompt.toolarge"),
            )
        }
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
            Box(
                Modifier
                    .weight(1f)
                    .alpha(if (snapshot) SessionStyle.DISABLED_OPACITY else 1f)
                    .clip(fieldShape)
                    .background(DesignTokens.Colors.card(SESSIONS_AREA))
                    .border(SessionStyle.borderWidth, DesignTokens.Colors.rule(SESSIONS_AREA), fieldShape)
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = COMPOSER_VERTICAL),
            ) {
                PlainTextField(
                    value = prompt,
                    onValueChange = { if (!snapshot) onPromptChange(it) },
                    placeholder = if (state.isTerminal) "명령 입력" else "에이전트에게 맡길 일",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = SessionStyle.FAINT_OPACITY),
                    maxLines = COMPOSER_MAX_LINES,
                    keyboardOptions = if (state.isTerminal) {
                        KeyboardOptions(capitalization = KeyboardCapitalization.None, autoCorrectEnabled = false)
                    } else {
                        KeyboardOptions(capitalization = KeyboardCapitalization.Sentences)
                    },
                    modifier = Modifier.fillMaxWidth().testTag("sessions.detail.prompt.field"),
                )
            }
            Box(
                Modifier
                    .size(SEND_SIZE)
                    .alpha(if (canSend) 1f else SEND_DISABLED_OPACITY)
                    .clip(CircleShape)
                    .background(ink)
                    .clickable(enabled = canSend, role = Role.Button, onClick = onSend)
                    .semantics { contentDescription = "보내기" }
                    .testTag("sessions.detail.send.button"),
                contentAlignment = Alignment.Center,
            ) {
                if (state.isSending) {
                    CircularProgressIndicator(color = surface, strokeWidth = INLINE_SPINNER_STROKE, modifier = Modifier.size(INLINE_SPINNER_SIZE))
                } else {
                    Icon(Icons.Filled.ArrowUpward, contentDescription = null, tint = surface, modifier = Modifier.size(SEND_ICON_SIZE))
                }
            }
        }
    }
}
