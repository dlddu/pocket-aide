package com.dlddu.pocketaide.sessions

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class SessionsUiState(
    val sessions: List<AgentSession> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val hasLoaded: Boolean = false,
    val error: AgentSessionError? = null,
    val config: AgentSessionConfig? = null,
    val configFailed: Boolean = false,
    val isCreating: Boolean = false,
    val createError: AgentSessionError? = null,
) {
    fun workloadType(session: AgentSession): AgentSessionWorkloadType? = config?.type(session.workloadType)
}

class SessionsViewModel(val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(SessionsUiState())
    val state: StateFlow<SessionsUiState> = _state.asStateFlow()

    fun load() {
        viewModelScope.launch { reload() }
    }

    fun refresh() {
        viewModelScope.launch {
            _state.update { it.copy(isRefreshing = true) }
            try {
                reload()
            } finally {
                _state.update { it.copy(isRefreshing = false) }
            }
        }
    }

    fun loadConfig() {
        viewModelScope.launch { reloadConfig() }
    }

    fun create(name: String, workload: AgentSessionWorkloadType, model: String?, onCreated: (AgentSession) -> Unit) {
        viewModelScope.launch { performCreate(name, workload, model)?.let(onCreated) }
    }

    fun clearCreateError() {
        _state.update { it.copy(createError = null) }
    }

    internal suspend fun reload() {
        if (_state.value.isLoading) return
        _state.update { it.copy(isLoading = true) }
        try {
            if (_state.value.config == null) reloadConfig()
            try {
                val sessions = api.listAgentSessions()
                _state.update { it.copy(sessions = sessions, error = null) }
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                _state.update { it.copy(error = AgentSessionError.from(e)) }
            }
            _state.update { it.copy(hasLoaded = true) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun reloadConfig() {
        try {
            val config = api.agentSessionConfig()
            _state.update { it.copy(config = config, configFailed = false) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(configFailed = it.config == null) }
        }
    }

    internal suspend fun performCreate(name: String, workload: AgentSessionWorkloadType, model: String?): AgentSession? {
        if (_state.value.isCreating) return null
        _state.update { it.copy(isCreating = true, createError = null) }
        return try {
            val created = api.createAgentSession(name, workload, model)
            _state.update { state ->
                state.copy(sessions = listOf(created) + state.sessions.filterNot { it.id == created.id })
            }
            created
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(createError = AgentSessionError.from(e)) }
            null
        } finally {
            _state.update { it.copy(isCreating = false) }
        }
    }
}
