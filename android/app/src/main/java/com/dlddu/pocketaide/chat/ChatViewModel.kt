package com.dlddu.pocketaide.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.api.Conditional
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
    val queued: ChatQueued? = null,
    val hasLoaded: Boolean = false,
    val isSending: Boolean = false,
    val isRefreshing: Boolean = false,
    val isSyncing: Boolean = false,
    val notice: String? = null,
    val error: ChatError? = null,
) {
    val isReplying: Boolean get() = pending != null

    // Sending while a reply is in progress queues instead of failing (server
    // decides), so only an already-queued message blocks the composer.
    val canSend: Boolean get() = hasLoaded && !isSending && queued == null && !isRefreshing
    val canRefresh: Boolean get() = hasLoaded && !isSending && !isReplying && !isRefreshing && messages.isNotEmpty()
    // A plain re-read of the conversation, so it stays available even while a
    // reply is in progress or sending — only guarded against itself (rapid taps).
    val canSync: Boolean get() = hasLoaded && !isSyncing
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
 *
 * Nothing in here is read until somebody opens the exchanges. The pane's
 * other fields — what the task is for, what is being asked, what there is to
 * say back — come from the task list the panel already polls, so a pane whose
 * exchanges are closed holds this at its default and has asked for nothing.
 *
 * [signal] is what the task list said about the task's exchanges when this
 * page was read. Re-opening a page whose signal still matches reads nothing;
 * a signal that moved is an exchange that happened, and the page is read
 * again.
 */
data class ChatTaskTurnsUiState(
    val taskId: String = "",
    val taskVersion: Long = 0,
    val signal: ChatTaskHistorySignal = ChatTaskHistorySignal.none,
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

    /**
     * Whether this already holds [taskId]'s exchanges as of [want], so opening
     * them again reads nothing. A failure is not holding them: somebody who
     * opens the section again is asking for another try.
     */
    fun holds(taskId: String, want: ChatTaskHistorySignal): Boolean =
        this.taskId == taskId && signal == want && error == null && (hasLoaded || isLoading)
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

    // The newest message id a read has delivered — not one appended by send,
    // which could be newer than a reply the polls have not fetched yet.
    private var cursor: Long? = null
    private var etag: String? = null

    // readSeq orders the conversation's reads the way taskReadSeq orders the
    // panel's. The poll loop and the sync button both call load(), so two reads
    // are in the air whenever somebody taps while a poll is out; a reply that is
    // no longer the newest read's is dropped rather than applied, because its
    // pending/queued would put the screen back to a state it has already left.
    private var readSeq = 0L

    // taskReadSeq orders the panel's reads. A reply that is not the newest
    // read's is dropped, so a slow answer cannot overwrite a fresher one.
    private var taskReadSeq = 0L

    // turnsReadSeq does the same for the exchanges, which are read per task and
    // only when somebody opens them. Switching tasks or closing the pane moves
    // it, so a page that was still in the air lands nowhere instead of
    // appearing under whichever task is open by then.
    private var turnsReadSeq = 0L

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

    fun cancelQueued() {
        viewModelScope.launch { performCancelQueued() }
    }

    fun refreshPlugins() {
        viewModelScope.launch { performRefreshPlugins() }
    }

    fun sync() {
        viewModelScope.launch { performSync() }
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

    /**
     * Reads [task]'s exchanges, and only when the pane does not already hold
     * them. Called when somebody opens the section and whenever the panel's
     * poll says something happened on the task whose section is open — so a
     * closed section reads nothing, the first open reads one page of that one
     * task, and opening it again after nothing changed reads nothing.
     *
     * The "nothing changed" answer comes from the task list the panel already
     * polls ([ChatTaskHistorySignal]), so nothing is asked of the session to
     * find out whether something is worth asking for.
     */
    fun loadTaskTurns(task: ChatTask) {
        viewModelScope.launch { openTaskTurns(task) }
    }

    /**
     * Opening them, as one suspending step so it can be driven without a main
     * looper. True when a page was actually read — false is "the pane already
     * holds this", which is the answer that costs nothing.
     */
    internal suspend fun openTaskTurns(task: ChatTask): Boolean {
        val want = ChatTaskHistorySignal.of(task)
        if (_taskTurns.value.holds(task.id, want)) return false
        return performTaskTurns(task.id, want, more = false)
    }

    /** Reads this task's exchanges again, whatever the pane is holding — the retry. */
    fun reloadTaskTurns(task: ChatTask) {
        viewModelScope.launch { performTaskTurns(task.id, ChatTaskHistorySignal.of(task), more = false) }
    }

    fun loadMoreTaskTurns() {
        val state = _taskTurns.value
        if (!state.hasMore || state.isLoading) return
        viewModelScope.launch { performTaskTurns(state.taskId, state.signal, more = true) }
    }

    /**
     * Forgets the exchanges once that task's pane is closed, and gives up on a
     * page still in the air — the next pane opens with nothing read, which is
     * what it shows.
     */
    fun clearTaskTurns() {
        turnsReadSeq++
        _taskTurns.value = ChatTaskTurnsUiState()
    }

    internal suspend fun performTaskTurns(
        taskId: String,
        signal: ChatTaskHistorySignal,
        more: Boolean,
    ): Boolean {
        val current = _taskTurns.value
        // Paging is same-task and additive, so a second one while the first is
        // in flight would append the same page twice. A first page for another
        // task is not blocked: it supersedes, and the older read's reply is
        // dropped by the sequence below.
        if (more && current.isLoading) return false
        if (!more && current.isLoading && current.taskId == taskId && current.signal == signal) return false
        val mine = ++turnsReadSeq
        val sameTask = current.taskId == taskId
        val after = if (more && sameTask) current.nextCursor else 0
        val kept = if (more && sameTask) current.turns else emptyList()
        // What is already on screen stays there until the new page replaces it.
        // A re-read happens while somebody is reading — a turn arrived on the
        // task whose section is open — and blanking a long exchange to a
        // spinner would take away what they were reading to announce that
        // there is now more of it. Another task's page is not kept: that one is
        // not this task's to show.
        val shown = if (sameTask) current.turns else emptyList()
        _taskTurns.value = current.copy(
            taskId = taskId,
            signal = signal,
            turns = shown,
            // A first page is page one: the previous read's cursor is not this
            // read's, and keeping it would offer "더 보기" for a page nobody has.
            nextCursor = if (more) current.nextCursor else 0,
            isLoading = true,
            // A first page that never arrives leaves the pane unread rather
            // than holding an empty history it never read — unless something is
            // still on screen from the read before it, which *was* read.
            hasLoaded = more || shown.isNotEmpty(),
            error = null,
        )
        return try {
            val page = api.chatTaskTurns(taskId, after, CHAT_TASK_TURNS_PAGE)
            // A page that lost its race belongs to a task the pane has left, or
            // to a pane that is closed. Showing it would put one task's
            // exchanges under another's name.
            if (mine != turnsReadSeq) return false
            _taskTurns.value = ChatTaskTurnsUiState(
                taskId = taskId,
                taskVersion = page.taskVersion,
                signal = signal,
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
            if (mine != turnsReadSeq) return false
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

    /**
     * Reads the conversation: only what came after the last read, and nothing
     * at all when the server answers 304. The whole conversation is read when
     * there is no read to continue from — before the first one, and whenever an
     * answer turns out to be a different conversation than the one on screen,
     * whose ids this cursor says nothing about.
     */
    internal suspend fun load() {
        val mine = ++readSeq
        try {
            val after = cursor
            val known = _state.value.conversation?.id
            if (after == null || known == null) {
                loadAll(mine)
            } else {
                when (val reply = api.chat(after, etag)) {
                    // Nothing changed, so nothing is applied — in particular the
                    // messages on screen are not replaced by the empty body.
                    Conditional.NotModified -> Unit
                    is Conditional.Fresh ->
                        if (reply.value.conversation.id != known) {
                            loadAll(mine)
                        } else if (mine == readSeq) {
                            applyView(reply.value, merging = true)
                            etag = reply.etag
                        }
                }
            }
            if (mine != readSeq) return
            _state.update {
                it.copy(hasLoaded = true, error = if (it.error == ChatError.UNREACHABLE) null else it.error)
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            // A superseded read's failure is not the screen's: a newer read has
            // answered, and its outcome is the one the screen is showing.
            if (mine != readSeq) return
            _state.update { it.copy(hasLoaded = true, error = ChatError.from(e)) }
        }
    }

    private suspend fun loadAll(mine: Long) {
        val view = api.chat()
        if (mine != readSeq) return
        applyView(view, merging = false)
        etag = null
    }

    private fun applyView(view: ChatView, merging: Boolean) {
        _state.update {
            it.copy(
                conversation = view.conversation,
                messages = if (merging) it.messages.merging(view.messages) else view.messages,
                pending = view.pending,
                queued = view.queued,
            )
        }
        val newest = view.messages.lastOrNull()?.id ?: 0L
        cursor = if (merging) maxOf(cursor ?: 0L, newest) else newest
    }

    // Reads right away instead of waiting for the next poll tick, and reads the
    // same way a poll does: what came after the last read. A tap is therefore
    // the cost of one poll rather than of the whole conversation, which on a
    // long one was megabytes (PRD-AI채팅.AC1, 2026-11-28 개정). The background
    // poll loop keeps calling load() on its own schedule the whole time; the
    // two overlapping is what readSeq is for.
    internal suspend fun performSync(): Boolean {
        if (!_state.value.canSync) return false
        _state.update { it.copy(isSyncing = true) }
        try {
            load()
            return true
        } finally {
            _state.update { it.copy(isSyncing = false) }
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
            when (val outcome = api.sendChatMessage(text)) {
                is ChatSendOutcome.Sent -> _state.update {
                    it.copy(messages = it.messages + outcome.message, pending = it.pending ?: ChatPending())
                }
                is ChatSendOutcome.Queued -> _state.update { it.copy(queued = outcome.queued) }
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

    /** A 404 means the queued message was already dispatched just as the
     * user cancelled — not an error worth showing, since it is about to
     * appear as a sent message. */
    internal suspend fun performCancelQueued() {
        val queued = _state.value.queued ?: return
        try {
            api.cancelQueuedChatMessage(queued.id)
        } catch (e: CancellationException) {
            throw e
        } catch (e: ApiException.BadStatus) {
            if (e.status != 404) {
                _state.update { it.copy(error = ChatError.from(e)) }
                return
            }
        } catch (e: Exception) {
            _state.update { it.copy(error = ChatError.from(e)) }
            return
        }
        _state.update { it.copy(queued = null) }
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
