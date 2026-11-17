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
    val isSyncing: Boolean = false,
    val notice: String? = null,
    val error: ChatError? = null,
) {
    val isReplying: Boolean get() = pending != null
    val canSend: Boolean get() = hasLoaded && !isSending && !isReplying && !isRefreshing
    val canRefresh: Boolean get() = hasLoaded && !isSending && !isReplying && !isRefreshing && messages.isNotEmpty()
    // A plain re-read of the conversation, so it stays available even while a
    // reply is in progress or sending — only guarded against itself (rapid taps).
    val canSync: Boolean get() = hasLoaded && !isSyncing
}

class ChatViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ChatUiState())
    val state: StateFlow<ChatUiState> = _state.asStateFlow()

    suspend fun pollWhileVisible() {
        while (coroutineContext.isActive) {
            load()
            delay(if (_state.value.isReplying) REPLY_POLL_MILLIS else IDLE_POLL_MILLIS)
        }
    }

    fun send(content: String, onSent: () -> Unit) {
        viewModelScope.launch { if (performSend(content)) onSent() }
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

    // Pulls the conversation right away instead of waiting for the next poll
    // tick — for history the server added while the tab wasn't visible. The
    // background poll loop keeps calling load() on its own schedule the whole
    // time; both calls are plain reads that each replace state with the
    // server's current truth, so running at the same time is harmless.
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

        fun refreshNotice(result: ChatPluginRefresh): String = when {
            result.state == "unconfigured" -> "이 대화에는 새로 받을 스킬이 없어요."
            result.changedCount == 0 -> "스킬이 이미 최신이에요."
            else -> "스킬을 새로 받았어요 — 추가 ${result.added.size} · 수정 ${result.updated.size} · 삭제 ${result.removed.size}. 다음 답부터 반영돼요."
        }
    }
}
