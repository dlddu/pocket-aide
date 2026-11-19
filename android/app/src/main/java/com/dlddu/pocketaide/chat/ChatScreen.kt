package com.dlddu.pocketaide.chat

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowUpward
import androidx.compose.material.icons.outlined.KeyboardArrowLeft
import androidx.compose.material.icons.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.Refresh
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.LinkAnnotation
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextLinkStyles
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.text.withLink
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.repeatOnLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel
import com.dlddu.pocketaide.designsystem.components.PlainTextField

// docs/mockups/screen-chat-text.html — header, bubbles, composer. The header's
// action refreshes plugin skills (one conversation, no "new chat"), and the
// composer's round button sends (voice mode is not built yet).
private val CHAT_AREA = Area.AiChat
private val HEADER_BUTTON_SIZE = 36.dp
private val HEADER_ICON_SIZE = 16.dp
private val HEADER_LABEL_GAP = 2.dp
private val HEADER_COUNT_BASELINE = 3.dp
private val AVATAR_SIZE = 28.dp
private val BUBBLE_TAIL = 6.dp
private val BUBBLE_VERTICAL = 10.dp
private val COMPOSER_VERTICAL = 10.dp
private val SEND_SIZE = 44.dp
private val SEND_ICON_SIZE = 18.dp
private val SPINNER_SIZE = 16.dp
private val SPINNER_STROKE = 2.dp
private val RULE_WIDTH = 1.dp
private val DOT_SIZE = 6.dp
private const val ME_MAX_FRACTION = 0.80f
private const val AI_MAX_FRACTION = 0.82f
private const val COMPOSER_MAX_LINES = 6
private const val AVATAR_FILL_OPACITY = 0.15f
private const val MUTED_OPACITY = 0.55f
private const val CHIP_TEXT_OPACITY = 0.75f
private const val DISABLED_OPACITY = 0.4f

@Composable
fun ChatScreen(api: ApiClient) {
    val viewModel: ChatViewModel = viewModel(key = "chat") { ChatViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val tasks by viewModel.tasks.collectAsStateWithLifecycle()
    val taskRequest by viewModel.taskRequest.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current
    // The draft and the panel's open/closed choice survive a fold, a rotation
    // and a trip through the back stack, so neither is lost by the layout
    // changing under the user.
    var draft by rememberSaveable { mutableStateOf("") }
    var panelOpen by rememberSaveable { mutableStateOf(true) }
    // How the panel is being looked at, and which task is open, kept the same
    // way for the same reason: a fold or a rotation must not silently widen a
    // narrowed list or close a task the user was working on.
    var filterStatus by rememberSaveable { mutableStateOf<String?>(null) }
    var hideCompleted by rememberSaveable { mutableStateOf(false) }
    var openTaskId by rememberSaveable { mutableStateOf<String?>(null) }
    // The request being written is its own draft, separate from the chat
    // composer's — storing one never touches the other. The key beside it is
    // this draft's identity on the ledger: the backend stores one row per key,
    // so a retry after an answer the app never saw cannot leave two requests,
    // and a new key is drawn only once a request has actually been stored.
    var requestDraft by rememberSaveable { mutableStateOf("") }
    var requestKey by rememberSaveable { mutableStateOf(newTaskRequestKey()) }
    // One list state each for the conversation and the task list, held above the
    // layout branch, so both reading positions survive folding, unfolding,
    // collapsing the panel and stepping into a task and back.
    val messageList = rememberLazyListState()
    val taskList = rememberLazyListState()

    LaunchedEffect(viewModel, lifecycleOwner) {
        lifecycleOwner.lifecycle.repeatOnLifecycle(Lifecycle.State.RESUMED) { viewModel.pollWhileVisible() }
    }

    val openTask = tasks.tasks.firstOrNull { it.id == openTaskId }
    // A task the ledger no longer carries — finished and dropped, or a session
    // we can no longer read — closes for good rather than waiting to reappear
    // under a request the user wrote against the row it used to be. Only once
    // the panel has an answer, so a read in flight closes nothing.
    LaunchedEffect(openTaskId, tasks.hasLoaded, openTask == null) {
        if (openTaskId != null && openTask == null && tasks.hasLoaded) {
            openTaskId = null
            requestDraft = ""
            requestKey = newTaskRequestKey()
            viewModel.dismissTaskRequest()
        }
    }

    ChatPanes(
        state = state,
        tasks = tasks,
        draft = draft,
        onDraftChange = { draft = it },
        onSend = { viewModel.send(draft) { draft = "" } },
        onRefresh = viewModel::refreshPlugins,
        onDismissNotice = viewModel::dismissNotice,
        listState = messageList,
        panelOpen = panelOpen,
        onTogglePanel = { panelOpen = !panelOpen },
        pollTasks = {
            lifecycleOwner.lifecycle.repeatOnLifecycle(Lifecycle.State.RESUMED) {
                viewModel.pollTasksWhileVisible()
            }
        },
        taskListState = taskList,
        filter = ChatTaskFilter(filterStatus, hideCompleted),
        onFilterChange = {
            filterStatus = it.status
            hideCompleted = it.hideCompleted
        },
        openTask = openTask,
        onOpenTask = { task ->
            openTaskId = task.id
            // A new task starts with an empty request rather than the last
            // one's words, which belonged to a different task — and with its
            // own key, so the two are never stored as one.
            requestDraft = ""
            requestKey = newTaskRequestKey()
            viewModel.dismissTaskRequest()
        },
        onCloseTask = {
            openTaskId = null
            viewModel.dismissTaskRequest()
        },
        taskRequest = taskRequest,
        requestDraft = requestDraft,
        onRequestDraftChange = { requestDraft = it },
        // The draft is cleared only once the backend has stored the request,
        // and the next request gets a new key then; a failure leaves both as
        // they are, so the same words retry as the same request.
        onStoreRequest = {
            openTask?.let { task ->
                viewModel.storeTaskRequest(task, requestKey, requestDraft) {
                    requestDraft = ""
                    requestKey = newTaskRequestKey()
                }
            }
        },
    )
}

/**
 * The chat screen's layout: one pane in a narrow window, two in a wide one.
 *
 * The branch is taken from the window's own available width
 * (docs/design-system/tokens.md §6.1) — not from the device, the screen size or
 * the orientation — so a folded phone and a split window stay narrow while an
 * unfolded foldable goes wide even in portrait.
 */
@Composable
internal fun ChatPanes(
    state: ChatUiState,
    tasks: ChatTasksUiState,
    draft: String,
    onDraftChange: (String) -> Unit,
    onSend: () -> Unit,
    onRefresh: () -> Unit,
    onDismissNotice: () -> Unit,
    listState: androidx.compose.foundation.lazy.LazyListState,
    panelOpen: Boolean,
    onTogglePanel: () -> Unit,
    // Runs for exactly as long as the panel is on screen.
    pollTasks: suspend () -> Unit,
    taskListState: androidx.compose.foundation.lazy.LazyListState,
    filter: ChatTaskFilter,
    onFilterChange: (ChatTaskFilter) -> Unit,
    // The task whose detail the right pane is showing, or null for the list.
    // Resolved from the ledger each read, so a task that leaves the ledger
    // closes rather than freezing on a row that no longer exists.
    openTask: ChatTask?,
    onOpenTask: (ChatTask) -> Unit,
    onCloseTask: () -> Unit,
    taskRequest: ChatTaskRequestUiState,
    requestDraft: String,
    onRequestDraftChange: (String) -> Unit,
    onStoreRequest: () -> Unit,
) {
    BoxWithConstraints(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(CHAT_AREA))
            .imePadding()
            .testTag("chat.screen"),
    ) {
        val wide = isWideChatWindow(maxWidth)
        val panelShown = wide && panelOpen

        // Polling follows what is on screen: this effect exists only while the
        // panel does, so a narrow window or a collapsed panel costs no requests
        // at all — cancelling it is what stops the reads.
        if (panelShown) {
            LaunchedEffect(Unit) { pollTasks() }
        }

        Row(Modifier.fillMaxSize()) {
            Conversation(
                state = state,
                draft = draft,
                onDraftChange = onDraftChange,
                onSend = onSend,
                onRefresh = onRefresh,
                onDismissNotice = onDismissNotice,
                listState = listState,
                // The toggle belongs to the wide layout only; a narrow window
                // has no second pane to open.
                panelOpen = if (wide) panelOpen else null,
                onTogglePanel = onTogglePanel,
                modifier = Modifier.weight(DesignTokens.Split.paneWeight),
            )
            if (panelShown) {
                Box(
                    Modifier
                        .fillMaxHeight()
                        .width(DesignTokens.Split.rule)
                        .background(DesignTokens.Colors.rule(CHAT_AREA)),
                )
                // 1:1 — the same weight as the conversation beside it. The pane
                // shows the list or one task's detail, never both: the detail is
                // a step into the same pane, so the conversation keeps its half.
                Box(Modifier.weight(DesignTokens.Split.paneWeight).fillMaxHeight()) {
                    if (openTask == null) {
                        ChatTaskPanel(
                            state = tasks,
                            filter = filter,
                            onFilterChange = onFilterChange,
                            onSelect = onOpenTask,
                            listState = taskListState,
                        )
                    } else {
                        // Back belongs to the step the user took, so it returns
                        // to the list instead of leaving the chat tab.
                        BackHandler(onBack = onCloseTask)
                        ChatTaskDetailPane(
                            task = openTask,
                            draft = requestDraft,
                            onDraftChange = onRequestDraftChange,
                            request = taskRequest,
                            // No reply gate: the request goes to the ledger,
                            // not into the turn being written.
                            onStore = onStoreRequest,
                            onBack = onCloseTask,
                        )
                    }
                }
            }
        }
    }
}

/**
 * A fresh key for one request draft. It is the app's idempotency key on the
 * ledger, so it has to be unique per request and stable across a retry — which
 * is why it is drawn when a draft starts, not when the store button is tapped.
 */
private fun newTaskRequestKey(): String = "app-" + java.util.UUID.randomUUID().toString().replace("-", "")

/** The conversation pane: the whole chat screen in a narrow window, the left half in a wide one. */
@Composable
private fun Conversation(
    state: ChatUiState,
    draft: String,
    onDraftChange: (String) -> Unit,
    onSend: () -> Unit,
    onRefresh: () -> Unit,
    onDismissNotice: () -> Unit,
    listState: androidx.compose.foundation.lazy.LazyListState,
    panelOpen: Boolean?,
    onTogglePanel: () -> Unit,
    modifier: Modifier,
) {
    Column(modifier.fillMaxHeight().testTag("chat.messages.pane")) {
        Header(state, onRefresh, panelOpen = panelOpen, onTogglePanel = onTogglePanel)
        Box(Modifier.fillMaxWidth().height(RULE_WIDTH).background(DesignTokens.Colors.rule(CHAT_AREA)))
        state.notice?.let {
            Banner(it, DesignTokens.Colors.ink(CHAT_AREA), tag = "chat.notice", onDismiss = onDismissNotice)
        }
        state.error?.let { Banner(it.message, DesignTokens.StatusColors.failure(), tag = "chat.error", onDismiss = null) }
        Messages(state, Modifier.weight(1f), listState)
        Composer(draft = draft, onDraftChange = onDraftChange, state = state, onSend = onSend)
    }
}

@Composable
private fun Header(
    state: ChatUiState,
    onRefresh: () -> Unit,
    // panelOpen is null in a narrow window, which has no second pane to open,
    // so no toggle is drawn there at all.
    panelOpen: Boolean? = null,
    onTogglePanel: () -> Unit = {},
) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = DesignTokens.Spacing.sm,
                bottom = DesignTokens.Spacing.md,
            ),
        verticalAlignment = Alignment.Bottom,
    ) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(HEADER_LABEL_GAP)) {
            AreaLabel(area = CHAT_AREA, text = "AI 채팅")
            Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
                Text(
                    state.conversation?.title?.takeIf { it.isNotBlank() } ?: "새 대화",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.h2, FontWeight.Bold),
                    color = ink,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f, fill = false).testTag("chat.header.title"),
                )
                Text(
                    "${state.messages.size}개의 메시지",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = ink.copy(alpha = MUTED_OPACITY),
                    modifier = Modifier.padding(bottom = HEADER_COUNT_BASELINE).testTag("chat.header.count"),
                )
            }
        }
        if (panelOpen != null) {
            PanelToggle(panelOpen, onTogglePanel, Modifier.padding(end = DesignTokens.Spacing.sm))
        }
        RefreshButton(state, onRefresh)
    }
}

/** Collapses and reopens the task panel. Wide layout only. */
@Composable
private fun PanelToggle(open: Boolean, onToggle: () -> Unit, modifier: Modifier = Modifier) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Box(
        modifier
            .size(HEADER_BUTTON_SIZE)
            .clip(CircleShape)
            .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), CircleShape)
            .clickable(role = Role.Button, onClick = onToggle)
            .semantics { contentDescription = if (open) "작업 패널 접기" else "작업 패널 펼치기" }
            .testTag("chat.tasks.toggle"),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            if (open) Icons.Outlined.KeyboardArrowRight else Icons.Outlined.KeyboardArrowLeft,
            contentDescription = null,
            tint = ink,
            modifier = Modifier.size(HEADER_ICON_SIZE),
        )
    }
}

@Composable
private fun RefreshButton(state: ChatUiState, onRefresh: () -> Unit) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Box(
        Modifier
            .size(HEADER_BUTTON_SIZE)
            .alpha(if (state.canRefresh) 1f else DISABLED_OPACITY)
            .clip(CircleShape)
            .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), CircleShape)
            .clickable(enabled = state.canRefresh, role = Role.Button, onClick = onRefresh)
            .semantics { contentDescription = "스킬 새로고침" }
            .testTag("chat.refresh.button"),
        contentAlignment = Alignment.Center,
    ) {
        if (state.isRefreshing) {
            CircularProgressIndicator(color = ink, strokeWidth = SPINNER_STROKE, modifier = Modifier.size(SPINNER_SIZE))
        } else {
            Icon(Icons.Outlined.Refresh, contentDescription = null, tint = ink, modifier = Modifier.size(HEADER_ICON_SIZE))
        }
    }
}

@Composable
private fun Banner(text: String, color: Color, tag: String, onDismiss: (() -> Unit)?) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = color,
        modifier = Modifier
            .fillMaxWidth()
            .then(if (onDismiss != null) Modifier.clickable(onClick = onDismiss) else Modifier)
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.sm)
            .testTag(tag),
    )
}

@Composable
internal fun Messages(
    state: ChatUiState,
    modifier: Modifier,
    // The caller owns the list state so the reading position survives the
    // layout changing — folding, unfolding, or collapsing the task panel.
    listState: androidx.compose.foundation.lazy.LazyListState = rememberLazyListState(),
) {
    val count = state.messages.size + if (state.isReplying) 1 else 0
    FollowLatestMessage(listState, count, state.pending?.content)
    if (state.hasLoaded && count == 0) {
        Box(modifier.fillMaxWidth().padding(DesignTokens.Spacing.xl), contentAlignment = Alignment.Center) {
            Text(
                "무엇이든 물어보세요. 첫 메시지를 보내면 ChatGPT 로그인 안내가 이 대화에 나와요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = DesignTokens.Colors.ink(CHAT_AREA).copy(alpha = MUTED_OPACITY),
                modifier = Modifier.testTag("chat.empty"),
            )
        }
        return
    }
    BoxWithConstraints(modifier.fillMaxWidth()) {
        val width = maxWidth
        LazyColumn(
            state = listState,
            modifier = Modifier.fillMaxSize().testTag("chat.messages"),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(
                horizontal = DesignTokens.Spacing.xl,
                vertical = DesignTokens.Spacing.lg,
            ),
        ) {
            items(state.messages, key = { it.id }) { message ->
                if (message.isMine) {
                    MyBubble(message.content, Modifier.widthIn(max = width * ME_MAX_FRACTION))
                } else {
                    AiBubble(message.content, pending = false, Modifier.widthIn(max = width * AI_MAX_FRACTION))
                }
            }
            state.pending?.let { pending ->
                item(key = "pending") {
                    AiBubble(pending.content, pending = true, Modifier.widthIn(max = width * AI_MAX_FRACTION))
                }
            }
        }
    }
}

@Composable
private fun MyBubble(text: String, modifier: Modifier) {
    val shape = RoundedCornerShape(
        topStart = DesignTokens.Radius.bubble,
        topEnd = DesignTokens.Radius.bubble,
        bottomEnd = BUBBLE_TAIL,
        bottomStart = DesignTokens.Radius.bubble,
    )
    SelectionContainer {
        Text(
            withLinks(text, DesignTokens.Colors.card(CHAT_AREA)),
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
            color = DesignTokens.Colors.card(CHAT_AREA),
            modifier = modifier
                .clip(shape)
                .background(DesignTokens.Colors.ink(CHAT_AREA))
                .padding(horizontal = DesignTokens.Spacing.lg, vertical = BUBBLE_VERTICAL)
                .testTag("chat.message.mine"),
        )
    }
}

@Composable
private fun AiBubble(text: String, pending: Boolean, modifier: Modifier) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    val rule = DesignTokens.Colors.rule(CHAT_AREA)
    val shape = RoundedCornerShape(
        topStart = DesignTokens.Radius.bubble,
        topEnd = DesignTokens.Radius.bubble,
        bottomEnd = DesignTokens.Radius.bubble,
        bottomStart = BUBBLE_TAIL,
    )
    val clipboard = LocalClipboardManager.current
    Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
        Avatar()
        Column(
            modifier
                .clip(shape)
                .background(DesignTokens.Colors.card(CHAT_AREA))
                .border(RULE_WIDTH, rule, shape)
                .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.md)
                .then(if (pending) Modifier.testTag("chat.message.pending") else Modifier.testTag("chat.message.ai")),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        ) {
            if (text.isNotBlank()) {
                SelectionContainer {
                    Text(
                        withLinks(text.trim(), DesignTokens.Colors.accent(CHAT_AREA)),
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                        color = ink,
                    )
                }
            }
            if (pending) {
                Typing()
            } else {
                Box(Modifier.fillMaxWidth().height(RULE_WIDTH).background(rule))
                Chip("복사") { clipboard.setText(AnnotatedString(text)) }
            }
        }
    }
}

/** Web addresses become links the system opens in the browser. */
private fun withLinks(text: String, linkColor: Color): AnnotatedString {
    val ranges = chatLinkRanges(text)
    if (ranges.isEmpty()) return AnnotatedString(text)
    val style = TextLinkStyles(SpanStyle(color = linkColor, textDecoration = TextDecoration.Underline))
    return buildAnnotatedString {
        var at = 0
        for (range in ranges) {
            append(text.substring(at, range.first))
            val url = text.substring(range)
            withLink(LinkAnnotation.Url(url, style)) { append(url) }
            at = range.last + 1
        }
        append(text.substring(at))
    }
}

@Composable
private fun Avatar() {
    val accent = DesignTokens.Colors.accent(CHAT_AREA)
    Box(
        Modifier
            .padding(bottom = DesignTokens.Spacing.xs)
            .size(AVATAR_SIZE)
            .clip(CircleShape)
            .background(accent.copy(alpha = AVATAR_FILL_OPACITY)),
        contentAlignment = Alignment.Center,
    ) {
        Text("PA", style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold), color = accent)
    }
}

@Composable
private fun Typing() {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs), verticalAlignment = Alignment.CenterVertically) {
        Text(
            "입력 중",
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm).copy(fontStyle = FontStyle.Italic),
            color = ink.copy(alpha = MUTED_OPACITY),
        )
        repeat(3) {
            Box(Modifier.size(DOT_SIZE).clip(CircleShape).background(ink.copy(alpha = MUTED_OPACITY)))
        }
    }
}

@Composable
private fun Chip(label: String, onClick: () -> Unit) {
    val shape = RoundedCornerShape(DesignTokens.Radius.chip)
    Text(
        label,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
        color = DesignTokens.Colors.ink(CHAT_AREA).copy(alpha = CHIP_TEXT_OPACITY),
        modifier = Modifier
            .clip(shape)
            .background(DesignTokens.Colors.card(CHAT_AREA))
            .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.xs)
            .testTag("chat.message.copy"),
    )
}

@Composable
private fun Composer(draft: String, onDraftChange: (String) -> Unit, state: ChatUiState, onSend: () -> Unit) {
    val ink = DesignTokens.Colors.ink(CHAT_AREA)
    val canSend = state.canSend && draft.isNotBlank()
    val fieldShape = RoundedCornerShape(DesignTokens.Radius.cardLarge)
    Column(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.surface(CHAT_AREA))
            .padding(horizontal = DesignTokens.Spacing.lg, vertical = DesignTokens.Spacing.sm),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.Bottom) {
            Box(
                Modifier
                    .weight(1f)
                    .clip(fieldShape)
                    .background(DesignTokens.Colors.card(CHAT_AREA))
                    .border(RULE_WIDTH, DesignTokens.Colors.rule(CHAT_AREA), fieldShape)
                    .padding(horizontal = DesignTokens.Spacing.lg, vertical = COMPOSER_VERTICAL),
            ) {
                PlainTextField(
                    value = draft,
                    onValueChange = onDraftChange,
                    placeholder = "메시지를 입력하세요",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = MUTED_OPACITY),
                    maxLines = COMPOSER_MAX_LINES,
                    keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Sentences),
                    modifier = Modifier.fillMaxWidth().testTag("chat.composer.field"),
                )
            }
            Box(
                Modifier
                    .size(SEND_SIZE)
                    .alpha(if (canSend) 1f else DISABLED_OPACITY)
                    .clip(CircleShape)
                    .background(DesignTokens.Colors.accent(CHAT_AREA))
                    .clickable(enabled = canSend, role = Role.Button, onClick = onSend)
                    .semantics { contentDescription = "보내기" }
                    .testTag("chat.composer.send"),
                contentAlignment = Alignment.Center,
            ) {
                if (state.isSending) {
                    CircularProgressIndicator(
                        color = DesignTokens.Colors.card(CHAT_AREA),
                        strokeWidth = SPINNER_STROKE,
                        modifier = Modifier.size(SPINNER_SIZE),
                    )
                } else {
                    Icon(
                        Icons.Filled.ArrowUpward,
                        contentDescription = null,
                        tint = DesignTokens.Colors.card(CHAT_AREA),
                        modifier = Modifier.size(SEND_ICON_SIZE),
                    )
                }
            }
        }
        Text(
            if (state.isReplying) "답을 쓰는 중이에요 — 끝나면 이어서 보낼 수 있어요" else "ChatGPT 계정의 Codex가 답해요",
            style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
            color = ink.copy(alpha = MUTED_OPACITY),
            modifier = Modifier.testTag("chat.composer.hint"),
        )
    }
}
