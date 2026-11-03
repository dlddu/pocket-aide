package com.dlddu.pocketaide.prmonitor

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.Instant

data class OpenPullRequestsUiState(
    val login: String? = null,
    val pullRequests: List<OpenPullRequest> = emptyList(),
    val inaccessibleCount: Int = 0,
    val lastRefreshedAt: Instant? = null,
    val isLoading: Boolean = false,
    val hasLoaded: Boolean = false,
    val isConnecting: Boolean = false,
    val lastError: GitHubError? = null,
    val errorMessage: String? = null,
    val connectError: String? = null,
    val filter: OpenPullRequestFilter = OpenPullRequestFilter.ALL,
) {
    val isConnected: Boolean get() = login != null
    val visiblePullRequests: List<OpenPullRequest> get() = filter.apply(pullRequests)
    val alert: GitHubAlert? get() = GitHubAlert.make(lastError, inaccessibleCount)
}

class OpenPullRequestsViewModel(
    private val store: GitHubCredentialStore,
    private val client: GitHubClient,
    private val filterStore: OpenPullRequestFilterStore,
    private val clock: () -> Instant = Instant::now,
) : ViewModel() {
    private val _state = MutableStateFlow(
        OpenPullRequestsUiState(
            login = runCatching { store.load() }.getOrNull()?.login,
            filter = filterStore.load() ?: OpenPullRequestFilter.ALL,
        ),
    )
    val state: StateFlow<OpenPullRequestsUiState> = _state.asStateFlow()

    fun setFilter(filter: OpenPullRequestFilter) {
        filterStore.save(filter)
        _state.update { it.copy(filter = filter) }
    }

    fun connect(token: String, onConnected: () -> Unit = {}) {
        viewModelScope.launch { if (performConnect(token)) onConnected() }
    }

    fun refresh() {
        viewModelScope.launch { performRefresh() }
    }

    fun refreshIfNeeded() {
        val current = _state.value
        if (current.isConnected && !current.hasLoaded && !current.isLoading) refresh()
    }

    fun disconnect() {
        runCatching { store.clear() }
        _state.update { resetList(it).copy(login = null) }
    }

    internal suspend fun performConnect(raw: String): Boolean {
        val token = raw.trim()
        if (token.isEmpty()) return false
        _state.update { it.copy(isConnecting = true, connectError = null) }
        try {
            val viewer = client.viewerLogin(token)
            store.save(GitHubCredential(token = token, login = viewer))
            _state.update { resetList(it).copy(login = viewer) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(connectError = message(e)) }
            return false
        } finally {
            _state.update { it.copy(isConnecting = false) }
        }
        performRefresh()
        return true
    }

    internal suspend fun performRefresh() {
        val credential = runCatching { store.load() }.getOrNull()
        if (credential == null) {
            _state.update { resetList(it).copy(login = null) }
            return
        }
        _state.update { it.copy(isLoading = true, errorMessage = null, lastError = null) }
        try {
            val result = client.openPullRequests(credential.token, credential.login)
            _state.update {
                it.copy(
                    pullRequests = result.pullRequests,
                    inaccessibleCount = result.inaccessibleCount,
                    lastRefreshedAt = clock(),
                    hasLoaded = true,
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(lastError = e as? GitHubError, errorMessage = message(e)) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    private fun resetList(state: OpenPullRequestsUiState): OpenPullRequestsUiState = state.copy(
        pullRequests = emptyList(),
        inaccessibleCount = 0,
        lastRefreshedAt = null,
        hasLoaded = false,
        errorMessage = null,
        lastError = null,
    )

    private fun message(e: Exception): String = (e as? GitHubError)?.description ?: (e.message ?: e.toString())
}
