package com.dlddu.pocketaide.sessions

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

data class SessionDetailUiState(
    val session: AgentSession,
    val type: AgentSessionWorkloadType? = null,
    val output: AgentSessionOutput = AgentSessionOutput(),
    val displayText: String = "",
    val transcript: AgentSessionTranscript = AgentSessionTranscript(),
    val phase: AgentSessionResponsePhase = AgentSessionResponsePhase.NONE,
    val hasLoadedOutput: Boolean = false,
    val isSending: Boolean = false,
    val isRestoring: Boolean = false,
    val isArchiving: Boolean = false,
    val isDeleting: Boolean = false,
    val deleted: Boolean = false,
    val error: AgentSessionError? = null,
) {
    val isTerminal: Boolean get() = type?.input == AgentSessionInput.TERMINAL
    val isBusy: Boolean get() = isRestoring || isArchiving || isDeleting
    val canSend: Boolean get() = session.state != AgentSessionState.SNAPSHOT && !isSending && !isBusy
    val canArchive: Boolean get() = session.state != AgentSessionState.SNAPSHOT && !isBusy
}

class SessionDetailViewModel(
    session: AgentSession,
    type: AgentSessionWorkloadType?,
    private val api: ApiClient,
    private val clock: () -> Long = System::currentTimeMillis,
) : ViewModel() {
    private val _state = MutableStateFlow(SessionDetailUiState(session = session, type = type))
    val state: StateFlow<SessionDetailUiState> = _state.asStateFlow()

    private var lastOutputAt = clock()
    private var lastPollAt: Long? = null
    private var catalogChecked = false
    private var awaitingThrough: Int? = null

    suspend fun pollWhileVisible() {
        while (coroutineContext.isActive) {
            tick()
            delay(POLL_INTERVAL_MILLIS)
        }
    }

    fun retry() {
        viewModelScope.launch { tick() }
    }

    fun send(prompt: String, onSent: () -> Unit) {
        viewModelScope.launch { if (performSend(prompt)) onSent() }
    }

    fun restore() {
        viewModelScope.launch { performRestore() }
    }

    fun archive() {
        viewModelScope.launch { performArchive() }
    }

    fun delete() {
        viewModelScope.launch { performDelete() }
    }

    internal suspend fun tick() {
        if (_state.value.type == null && !catalogChecked) {
            catalogChecked = true
            val found = try {
                api.agentSessionConfig().type(_state.value.session.workloadType)
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                null
            }
            _state.update { it.copy(type = found) }
        }
        val now = clock()
        val stale = lastPollAt?.let { now - it > STATE_RECHECK_AFTER_MILLIS } ?: true
        lastPollAt = now
        if (stale) refreshState()
        val current = _state.value
        if (current.session.state == AgentSessionState.SNAPSHOT || current.isBusy || current.deleted) return
        readOutput()
    }

    internal suspend fun refreshState() {
        try {
            val session = api.agentSession(_state.value.session.id)
            _state.update { it.copy(session = session) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(error = AgentSessionError.from(e)) }
        }
    }

    internal suspend fun readOutput() {
        val id = _state.value.session.id
        val requested = _state.value.output.nextOffset
        try {
            var result = api.readAgentSession(id, requested)
            var applied = _state.value.output.merge(result, requested)
            if (applied.outcome == AgentSessionOutput.Outcome.NEEDS_FULL_REREAD) {
                result = api.readAgentSession(id, 0)
                applied = _state.value.output.merge(result, 0)
            }
            val outcome = applied.outcome
            val output = applied.output
            val before = _state.value
            val grew = outcome == AgentSessionOutput.Outcome.GREW && before.hasLoadedOutput
            if (grew) lastOutputAt = clock()
            val refreshText = outcome == AgentSessionOutput.Outcome.GREW || !before.hasLoadedOutput
            _state.update { state ->
                val terminal = state.isTerminal
                val displayText = if (refreshText) {
                    if (terminal) AgentSessionTerminal.clean(output.text) else output.text
                } else {
                    state.displayText
                }
                val transcript = if (refreshText) {
                    if (terminal) AgentSessionTranscript() else AgentSessionTranscript.parse(output.text)
                } else {
                    state.transcript
                }
                val phase = if (terminal) {
                    state.phase
                } else {
                    AgentSessionResponsePhase.fromTurns(transcript, awaitingThrough)
                        ?: state.phase.next(grew, clock() - lastOutputAt)
                }
                state.copy(
                    session = result.session ?: state.session,
                    output = output,
                    displayText = displayText,
                    transcript = transcript,
                    phase = phase,
                    hasLoadedOutput = true,
                    error = null,
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            val category = AgentSessionError.from(e)
            _state.update { it.copy(error = category) }
            if (category == AgentSessionError.SNAPSHOTTED || category == AgentSessionError.NOT_FOUND) refreshState()
        }
    }

    internal suspend fun performSend(prompt: String): Boolean {
        if (!_state.value.canSend) return false
        if (AgentSessionPrompt.exceedsLimit(prompt)) {
            _state.update { it.copy(error = AgentSessionError.TOO_LARGE) }
            return false
        }
        _state.update { it.copy(isSending = true) }
        try {
            val terminal = _state.value.isTerminal
            val payload = if (terminal) AgentSessionTerminal.command(prompt) else prompt
            val result = api.writeAgentSession(_state.value.session.id, payload)
            if (!terminal) {
                awaitingThrough = maxOf(awaitingThrough ?: 0, _state.value.transcript.lastSeq ?: 0) + 1
            }
            lastOutputAt = clock()
            _state.update {
                it.copy(
                    session = result.session ?: it.session,
                    phase = if (terminal) it.phase else AgentSessionResponsePhase.QUEUED,
                    error = null,
                )
            }
            return true
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(error = AgentSessionError.from(e)) }
            return false
        } finally {
            _state.update { it.copy(isSending = false) }
        }
    }

    internal suspend fun performRestore() {
        val current = _state.value
        if (current.session.state != AgentSessionState.SNAPSHOT || current.isBusy) return
        _state.update { it.copy(isRestoring = true) }
        try {
            val session = api.restoreAgentSession(current.session.id)
            _state.update { it.copy(session = session, error = null) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(error = AgentSessionError.from(e)) }
        } finally {
            _state.update { it.copy(isRestoring = false) }
        }
    }

    internal suspend fun performArchive() {
        val current = _state.value
        if (!current.canArchive) return
        _state.update { it.copy(isArchiving = true) }
        try {
            val session = api.snapshotAgentSession(current.session.id)
            _state.update { it.copy(session = session, phase = AgentSessionResponsePhase.NONE, error = null) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(error = AgentSessionError.forSnapshot(e)) }
        } finally {
            _state.update { it.copy(isArchiving = false) }
        }
    }

    internal suspend fun performDelete() {
        val current = _state.value
        if (current.isBusy) return
        _state.update { it.copy(isDeleting = true) }
        try {
            api.deleteAgentSession(current.session.id)
            _state.update { it.copy(deleted = true) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            val category = AgentSessionError.from(e)
            _state.update { if (category == AgentSessionError.NOT_FOUND) it.copy(deleted = true) else it.copy(error = category) }
        } finally {
            _state.update { it.copy(isDeleting = false) }
        }
    }

    companion object {
        const val POLL_INTERVAL_MILLIS = 2_000L
        const val STATE_RECHECK_AFTER_MILLIS = 30_000L
    }
}
