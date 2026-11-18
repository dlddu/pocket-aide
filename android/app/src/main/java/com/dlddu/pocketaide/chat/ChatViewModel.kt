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
 * A follow-up request sent from the task panel. It carries its own in-flight
 * flag so a second tap cannot send a second copy, and its own outcome so the
 * panel can say the request was taken — or say it failed and keep the draft.
 */
data class ChatFollowUpUiState(
    val taskId: String = "",
    val isSending: Boolean = false,
    val accepted: Boolean = false,
    val error: ChatError? = null,
)

class ChatViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ChatUiState())
    val state: StateFlow<ChatUiState> = _state.asStateFlow()

    private val _tasks = MutableStateFlow(ChatTasksUiState())
    val tasks: StateFlow<ChatTasksUiState> = _tasks.asStateFlow()

    private val _followUp = MutableStateFlow(ChatFollowUpUiState())
    val followUp: StateFlow<ChatFollowUpUiState> = _followUp.asStateFlow()

    // taskRequest orders the panel's reads. A reply that is not the newest
    // request's is dropped, so a slow answer cannot overwrite a fresher one.
    private var taskRequest = 0L

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
     * Sends one follow-up about [task] as a message in this conversation.
     * [onAccepted] runs only when the backend took it, so the caller clears its
     * draft then and keeps it on every failure.
     */
    fun sendTaskFollowUp(task: ChatTask, request: String, onAccepted: () -> Unit) {
        viewModelScope.launch { if (performTaskFollowUp(task, request)) onAccepted() }
    }

    /** Clears the last follow-up's outcome once the panel has shown it. */
    fun dismissFollowUp() {
        _followUp.update { it.copy(accepted = false, error = null) }
    }

    internal suspend fun performTaskFollowUp(task: ChatTask, request: String): Boolean {
        val text = request.trim()
        if (text.isEmpty()) return false
        // A second tap while the first is in flight would send the request
        // twice; the panel also disables its button, and this is the guard that
        // holds when a tap gets through anyway.
        if (_followUp.value.isSending) return false
        val message = chatTaskFollowUpMessage(task, text)
        if (message.toByteArray(Charsets.UTF_8).size > CHAT_MAX_MESSAGE_BYTES) {
            _followUp.value = ChatFollowUpUiState(taskId = task.id, error = ChatError.TOO_LARGE)
            return false
        }
        _followUp.value = ChatFollowUpUiState(taskId = task.id, isSending = true)
        try {
            val sent = api.sendChatMessage(message)
            // The request lands in the conversation beside the panel, the same
            // as a typed message — the user can see what was sent and the reply
            // arrives in the usual place. The chat's own draft is not touched.
            _state.update { it.copy(messages = it.messages + sent, pending = it.pending ?: ChatPending()) }
            _followUp.value = ChatFollowUpUiState(taskId = task.id, accepted = true)
            return true
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _followUp.value = ChatFollowUpUiState(taskId = task.id, error = ChatError.from(e))
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
        val mine = ++taskRequest
        try {
            val list = api.chatTasks()
            // A reply that lost its race is stale; applying it would move the
            // panel backwards.
            if (mine != taskRequest) return
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
            if (mine != taskRequest) return
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
