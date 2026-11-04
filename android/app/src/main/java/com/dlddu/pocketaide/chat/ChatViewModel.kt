package com.dlddu.pocketaide.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ChatUiState(
    val sessions: List<ChatSession> = emptyList(),
    val current: ChatSession? = null,
    val messages: List<ChatMessage> = emptyList(),
    val streamingReply: String? = null,
    val isLoading: Boolean = false,
    val hasLoaded: Boolean = false,
    val showsHistory: Boolean = false,
    val errorMessage: String? = null,
) {
    val isSending: Boolean get() = streamingReply != null
}

class ChatViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ChatUiState())
    val state: StateFlow<ChatUiState> = _state.asStateFlow()

    private var localId = 0L

    private fun nextLocalId(): Long {
        localId -= 1
        return localId
    }

    fun load() {
        viewModelScope.launch { reload() }
    }

    fun send(text: String) {
        viewModelScope.launch { performSend(text) }
    }

    fun startNewConversation() {
        if (_state.value.isSending) return
        _state.update { it.copy(current = null, messages = emptyList(), showsHistory = false, errorMessage = null) }
    }

    fun open(session: ChatSession) {
        if (_state.value.isSending) return
        viewModelScope.launch { performOpen(session) }
    }

    fun showHistory() {
        _state.update { it.copy(showsHistory = true) }
    }

    fun hideHistory() {
        _state.update { it.copy(showsHistory = false) }
    }

    fun dismissError() {
        _state.update { it.copy(errorMessage = null) }
    }

    internal suspend fun reload() {
        _state.update { it.copy(isLoading = true, errorMessage = null) }
        try {
            val sessions = api.listChatSessions()
            _state.update { it.copy(sessions = sessions) }
            if (!_state.value.hasLoaded && _state.value.current == null) {
                sessions.firstOrNull()?.let { restore(it) }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        } finally {
            _state.update { it.copy(isLoading = false, hasLoaded = true) }
        }
    }

    internal suspend fun performOpen(session: ChatSession) {
        try {
            restore(session)
            _state.update { it.copy(showsHistory = false, errorMessage = null) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    internal suspend fun performSend(text: String) {
        val content = text.trim()
        if (content.isEmpty() || _state.value.isSending) return
        val outgoing = ChatMessage(id = nextLocalId(), role = ChatRole.USER, content = content)
        _state.update { it.copy(messages = it.messages + outgoing, streamingReply = "", errorMessage = null) }
        var streamed = ""
        var finished = false
        try {
            val session = _state.value.current ?: api.createChatSession().also { created ->
                _state.update { it.copy(current = created) }
            }
            api.sendChatMessage(session.id, content) { event ->
                when (event) {
                    is ChatStreamEvent.Delta -> {
                        streamed += event.text
                        _state.update { it.copy(streamingReply = streamed) }
                    }
                    is ChatStreamEvent.Done -> {
                        finished = true
                        val reply = event.message ?: ChatMessage(nextLocalId(), ChatRole.ASSISTANT, streamed)
                        _state.update { it.copy(messages = it.messages + reply, streamingReply = null) }
                    }
                    is ChatStreamEvent.Failed -> fail(ChatCopy.REPLY_FAILED)
                }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(ChatCopy.errorMessage(e))
        } finally {
            val partial = streamed
            _state.update {
                val kept = if (!finished && partial.isNotEmpty()) {
                    it.messages + ChatMessage(nextLocalId(), ChatRole.ASSISTANT, partial)
                } else {
                    it.messages
                }
                it.copy(messages = kept, streamingReply = null)
            }
        }
        refreshSessions()
    }

    private suspend fun restore(session: ChatSession) {
        val conversation = api.getChatConversation(session.id)
        _state.update { it.copy(current = conversation.session, messages = conversation.messages) }
    }

    private suspend fun refreshSessions() {
        try {
            val sessions = api.listChatSessions()
            _state.update { state ->
                state.copy(
                    sessions = sessions,
                    current = state.current?.let { cur -> sessions.firstOrNull { it.id == cur.id } ?: cur },
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    private fun fail(e: Exception) = fail(ChatCopy.errorMessage(e))

    private fun fail(message: String) {
        _state.update { it.copy(errorMessage = message) }
    }
}
