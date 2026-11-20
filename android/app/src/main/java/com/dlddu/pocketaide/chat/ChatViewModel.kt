package com.dlddu.pocketaide.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlin.coroutines.coroutineContext

data class ChatUiState(
    val conversation: ChatConversation? = null,
    val messages: List<ChatMessage> = emptyList(),
    val pending: ChatPending? = null,
    val hasLoaded: Boolean = false,
    val isSending: Boolean = false,
    val isRefreshing: Boolean = false,
    val notice: String? = null,
    val error: ChatError? = null,
) {
    val isReplying: Boolean get() = pending != null
    val canSend: Boolean get() = hasLoaded && !isSending && !isReplying && !isRefreshing
    val canRefresh: Boolean get() = hasLoaded && !isSending && !isReplying && !isRefreshing && messages.isNotEmpty()
}

/**
 * The task panel's state. Loading, empty, error and "no ledger here" are
 * separate on purpose: an empty list means the session really has no tasks, and
 * nothing else may be shown as if it did.
 */
data class ChatTasksUiState(
    val sessionId: String = "",
    val tasks: List<ChatTask> = emptyList(),
    val counts: Map<String, Int> = emptyMap(),
    val truncated: Boolean = false,
    val hasLoaded: Boolean = false,
    val error: ChatTaskError? = null,
) {
    val isLoading: Boolean get() = !hasLoaded && error == null
    val isEmpty: Boolean get() = hasLoaded && error == null && tasks.isEmpty()
}

/**
 * The request the panel is storing on a task's ledger.
 *
 * It carries its own in-flight flag so a second tap cannot store a second copy,
 * and its own outcome so the panel can say the request is *saved and waiting* —
 * never that the session has dealt with it, which only the ledger's own status
 * says.
 */
/**
 * One task's turns as the detail pane holds them, which is that task's whole
 * history.
 *
 * This is the work itself — what was asked, what went to the session doing it,
 * that session's own answer and the parent's explanation. There is no separate
 * ledger history beside it any more: a past management event was never a turn,
 * and the route that served them is gone with the ledger (platform AC-N10).
 *
 * Loading, empty, a page and a failure are separate on purpose: an empty list
 * means this task has had no exchange yet, and a failure must not look like
 * that.
 */
data class ChatTaskTurnsUiState(
    val taskId: String = "",
    val task: ChatTaskTurnTask = ChatTaskTurnTask(),
    val taskVersion: Long = 0,
    val turns: List<ChatTaskTurn> = emptyList(),
    val nextCursor: Long = 0,
    val isLoading: Boolean = false,
    val hasLoaded: Boolean = false,
    val error: ChatTaskTurnsError? = null,
) {
    val isEmpty: Boolean get() = hasLoaded && error == null && turns.isEmpty()
    val hasMore: Boolean get() = nextCursor > 0

    /** The choices the latest turn offers, which are the only current ones. */
    val offered: List<ChatTaskTurnChoice> get() = turns.lastOrNull()?.choices ?: emptyList()
}

data class ChatTaskRequestUiState(
    val taskId: String = "",
    val isStoring: Boolean = false,
    val stored: ChatTaskTurnRef? = null,
    val replayed: Boolean = false,
    val error: ChatTaskRequestError? = null,
)

class ChatViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ChatUiState())
    val state: StateFlow<ChatUiState> = _state.asStateFlow()

    private val _tasks = MutableStateFlow(ChatTasksUiState())
    val tasks: StateFlow<ChatTasksUiState> = _tasks.asStateFlow()

    private val _taskRequest = MutableStateFlow(ChatTaskRequestUiState())
    val taskRequest: StateFlow<ChatTaskRequestUiState> = _taskRequest.asStateFlow()

    private val _taskTurns = MutableStateFlow(ChatTaskTurnsUiState())
    val taskTurns: StateFlow<ChatTaskTurnsUiState> = _taskTurns.asStateFlow()

    // taskReadSeq orders the panel's reads. A reply that is not the newest
    // read's is dropped, so a slow answer cannot overwrite a fresher one.
    private var taskReadSeq = 0L

    suspend fun pollWhileVisible() {
        while (coroutineContext.isActive) {
            load()
            delay(if (_state.value.isReplying) REPLY_POLL_MILLIS else IDLE_POLL_MILLIS)
        }
    }

    /**
     * Polls the task panel. The screen starts this only while the panel is on
     * screen, so a hidden or collapsed panel costs no requests — cancelling
     * this coroutine is what stops the polling.
     */
    suspend fun pollTasksWhileVisible() {
        while (coroutineContext.isActive) {
            loadTasks()
            delay(if (_state.value.isReplying) TASKS_REPLY_POLL_MILLIS else TASKS_IDLE_POLL_MILLIS)
        }
    }

    fun send(content: String, onSent: () -> Unit) {
        viewModelScope.launch { if (performSend(content)) onSent() }
    }

    fun refreshPlugins() {
        viewModelScope.launch { performRefreshPlugins() }
    }

    fun dismissNotice() {
        _state.update { it.copy(notice = null) }
    }

    /**
     * Stores one request about [task], where it becomes that task's next turn.
     * [onStored] runs only when the backend stored it, so the caller clears its
     * draft then and keeps it on every failure.
     *
     * [choiceTurnId] is the turn the person picked a sentence from, or zero for
     * their own words. It is the turn rather than a choice id because what
     * makes a pick current is that turn still being the task's latest.
     */
    fun storeTaskRequest(
        task: ChatTask,
        requestID: String,
        request: String,
        choiceTurnId: Long = 0,
        onStored: () -> Unit,
        onChoiceStale: () -> Unit = {},
    ) {
        viewModelScope.launch {
            runTaskRequest(task, requestID, request, choiceTurnId, onStored, onChoiceStale)
        }
    }

    /**
     * The store and what follows it, as one suspending step so it can be driven
     * without a main looper. [onStored] runs only on a store; a pick from a turn
     * the task has moved past runs [onChoiceStale] instead and re-reads the
     * task, so the pane shows the choices that hold now rather than the ones it
     * was refused on. Neither path touches the draft — the caller clears it only
     * on [onStored].
     */
    internal suspend fun runTaskRequest(
        task: ChatTask,
        requestID: String,
        request: String,
        choiceTurnId: Long,
        onStored: () -> Unit,
        onChoiceStale: () -> Unit = {},
    ) {
        if (performTaskRequest(task, requestID, request, choiceTurnId)) {
            onStored()
            return
        }
        if (_taskRequest.value.error == ChatTaskRequestError.STALE_CHOICE) {
            onChoiceStale()
            loadTasks()
        }
    }

    fun loadTaskTurns(taskId: String) {
        viewModelScope.launch { performTaskTurns(taskId, more = false) }
    }

    fun loadMoreTaskTurns() {
        val state = _taskTurns.value
        if (!state.hasMore || state.isLoading) return
        viewModelScope.launch { performTaskTurns(state.taskId, more = true) }
    }

    /** Forgets the turns once that task's pane is closed. */
    fun clearTaskTurns() {
        _taskTurns.value = ChatTaskTurnsUiState()
    }

    internal suspend fun performTaskTurns(taskId: String, more: Boolean): Boolean {
        val current = _taskTurns.value
        if (current.isLoading) return false
        val after = if (more && current.taskId == taskId) current.nextCursor else 0
        val kept = if (more && current.taskId == taskId) current.turns else emptyList()
        _taskTurns.value = current.copy(taskId = taskId, turns = kept, isLoading = true, error = null)
        return try {
            val page = api.chatTaskTurns(taskId, after, CHAT_TASK_TURNS_PAGE)
            _taskTurns.value = ChatTaskTurnsUiState(
                taskId = taskId,
                task = page.task,
                taskVersion = page.taskVersion,
                // Pages arrive oldest first and are appended in that order, so
                // the pane reads top to bottom as it happened.
                turns = kept + page.turns,
                nextCursor = page.nextCursor,
                hasLoaded = true,
            )
            true
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            // The platform's four answers stay four different things: a task
            // this slot does not have, a session with no live pod, a
            // deployment that keeps no work records, and anything else. None
            // of them is "this task has no work".
            _taskTurns.value = _taskTurns.value.copy(
                isLoading = false,
                hasLoaded = true,
                error = ChatTaskTurnsError.from(e),
            )
            false
        }
    }

    /** Clears the last request's outcome once the panel has shown it. */
    fun dismissTaskRequest() {
        _taskRequest.update { it.copy(stored = null, replayed = false, error = null) }
    }

    internal suspend fun performTaskRequest(
        task: ChatTask,
        requestID: String,
        request: String,
        choiceTurnId: Long = 0,
    ): Boolean {
        val text = chatTaskRequestText(request)
        if (text.isEmpty()) return false
        // A second tap while the first is in flight would store the request
        // twice if the backend had not seen the first one yet; the panel also
        // disables its button, and this is the guard that holds when a tap gets
        // through anyway. The request id makes the retry of a lost answer safe,
        // and this makes the double tap unnecessary.
        if (_taskRequest.value.isStoring) return false
        if (text.toByteArray(Charsets.UTF_8).size > CHAT_TASK_REQUEST_MAX_BYTES) {
            _taskRequest.value = ChatTaskRequestUiState(taskId = task.id, error = ChatTaskRequestError.INVALID)
            return false
        }
        _taskRequest.value = ChatTaskRequestUiState(taskId = task.id, isStoring = true)
        try {
            val stored = api.storeChatTaskRequest(task.id, requestID, text, choiceTurnId)
            // Nothing is added to the conversation and no reply is awaited: the
            // request is the task's next turn now, and the parent finds it when
            // it next reads the task's open turns. Storing therefore works
            // while a reply is being written.
            _taskRequest.value = ChatTaskRequestUiState(
                taskId = task.id,
                stored = stored.turn,
                replayed = stored.replayed,
            )
            return true
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _taskRequest.value = ChatTaskRequestUiState(taskId = task.id, error = ChatTaskRequestError.from(e))
            return false
        }
    }

    internal suspend fun load() {
        try {
            val view = api.chat()
            _state.update {
                it.copy(
                    conversation = view.conversation,
                    messages = view.messages,
                    pending = view.pending,
                    hasLoaded = true,
                    error = if (it.error == ChatError.UNREACHABLE) null else it.error,
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(hasLoaded = true, error = ChatError.from(e)) }
        }
    }

    internal suspend fun loadTasks() {
        val mine = ++taskReadSeq
        try {
            val list = api.chatTasks()
            // A reply that lost its race is stale; applying it would move the
            // panel backwards.
            if (mine != taskReadSeq) return
            _tasks.update {
                // A reply for another session replaces the list outright — two
                // sessions' tasks are never shown together, and the conversation
                // that moved on does not keep the old session's rows.
                it.copy(
                    sessionId = list.sessionId,
                    tasks = list.tasks,
                    counts = list.counts,
                    truncated = list.truncated,
                    hasLoaded = true,
                    error = null,
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            if (mine != taskReadSeq) return
            val failure = ChatTaskError.from(e)
            // The list that is on screen belongs to a session we can no longer
            // read, so it is cleared rather than left to look current.
            _tasks.update { it.copy(tasks = emptyList(), counts = emptyMap(), truncated = false, hasLoaded = true, error = failure) }
        }
    }

    internal suspend fun performSend(content: String): Boolean {
        val text = content.trim()
        if (text.isEmpty() || !_state.value.canSend) return false
        if (text.toByteArray(Charsets.UTF_8).size > CHAT_MAX_MESSAGE_BYTES) {
            _state.update { it.copy(error = ChatError.TOO_LARGE) }
            return false
        }
        _state.update { it.copy(isSending = true, error = null, notice = null) }
        try {
            val sent = api.sendChatMessage(text)
            _state.update {
                it.copy(messages = it.messages + sent, pending = it.pending ?: ChatPending())
            }
            return true
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(error = ChatError.from(e)) }
            return false
        } finally {
            _state.update { it.copy(isSending = false) }
        }
    }

    internal suspend fun performRefreshPlugins() {
        if (!_state.value.canRefresh) return
        _state.update { it.copy(isRefreshing = true, error = null, notice = null) }
        try {
            val result = api.refreshChatPlugins()
            _state.update { it.copy(notice = refreshNotice(result)) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(error = ChatError.from(e, refreshing = true)) }
        } finally {
            _state.update { it.copy(isRefreshing = false) }
        }
    }

    companion object {
        // What one stored request may carry. The backend and the ledger bound
        // it too; this is the app saying so before the hop.
        const val CHAT_TASK_REQUEST_MAX_BYTES = 4_000

        // One history page: small enough that the first page is quick and the
        // rest is one tap away.
        const val CHAT_TASK_TURNS_PAGE = 20

        const val REPLY_POLL_MILLIS = 1_500L
        const val IDLE_POLL_MILLIS = 15_000L

        // The ledger moves with a turn, so the panel reads more often while a
        // reply is coming and rarely otherwise. It is slower than the chat's
        // own poll because a task list changes in steps, not words.
        const val TASKS_REPLY_POLL_MILLIS = 3_000L
        const val TASKS_IDLE_POLL_MILLIS = 20_000L

        fun refreshNotice(result: ChatPluginRefresh): String = when {
            result.state == "unconfigured" -> "이 대화에는 새로 받을 스킬이 없어요."
            result.changedCount == 0 -> "스킬이 이미 최신이에요."
            else -> "스킬을 새로 받았어요 — 추가 ${result.added.size} · 수정 ${result.updated.size} · 삭제 ${result.removed.size}. 다음 답부터 반영돼요."
        }
    }
}
