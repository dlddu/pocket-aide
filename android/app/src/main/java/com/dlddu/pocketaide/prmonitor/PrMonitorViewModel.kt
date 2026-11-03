package com.dlddu.pocketaide.prmonitor

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class PrMonitorUiState(
    val items: List<NotificationHistoryItem> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val errorMessage: String? = null,
    val excludedRepos: List<ExcludedRepo> = emptyList(),
    val isLoadingExcluded: Boolean = false,
    val excludedRepoError: String? = null,
    val notificationSettings: NotificationSettings = NotificationSettings(),
    val isLoadingSettings: Boolean = false,
    val settingsError: String? = null,
) {
    val groups: List<HistoryGroup> get() = HistoryGrouping.group(items)
    val totalUnacknowledgedCount: Int get() = items.count { !it.isAcknowledged }
    val unacknowledgedGroups: List<HistoryGroup> get() = groups.filterNot { it.allAcknowledged }
    val acknowledgedGroups: List<HistoryGroup> get() = groups.filter { it.allAcknowledged }
}

class PrMonitorViewModel(
    private val api: ApiClient,
    private val clock: () -> Long = System::currentTimeMillis,
    private val onAcknowledged: () -> Unit = {},
) : ViewModel() {
    private val _state = MutableStateFlow(PrMonitorUiState())
    val state: StateFlow<PrMonitorUiState> = _state.asStateFlow()

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

    fun loadExcludedRepos() {
        viewModelScope.launch { reloadExcludedRepos() }
    }

    fun loadNotificationSettings() {
        viewModelScope.launch { reloadNotificationSettings() }
    }

    fun setNotificationsEnabled(enabled: Boolean) {
        viewModelScope.launch { performSetNotificationsEnabled(enabled) }
    }

    fun setNotificationOutcomes(outcomes: NotificationOutcomes) {
        viewModelScope.launch { performSetNotificationOutcomes(outcomes) }
    }

    fun acknowledge(id: Long) {
        viewModelScope.launch { performAcknowledge(id) }
    }

    fun acknowledgeGroup(group: HistoryGroup) {
        viewModelScope.launch { performAcknowledgeGroup(group) }
    }

    fun excludeRepo(repoFullName: String) {
        viewModelScope.launch { performExcludeRepo(repoFullName) }
    }

    fun removeExcludedRepo(id: Long) {
        viewModelScope.launch { performRemoveExcludedRepo(id) }
    }

    internal suspend fun reload() {
        _state.update { it.copy(isLoading = true, errorMessage = null) }
        try {
            val items = api.listNotificationHistory()
            _state.update { it.copy(items = items) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = describe(e)) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun reloadExcludedRepos() {
        _state.update { it.copy(isLoadingExcluded = true, excludedRepoError = null) }
        try {
            val repos = api.listExcludedRepos()
            _state.update { it.copy(excludedRepos = repos) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(excludedRepoError = describe(e)) }
        } finally {
            _state.update { it.copy(isLoadingExcluded = false) }
        }
    }

    internal suspend fun reloadNotificationSettings() {
        _state.update { it.copy(isLoadingSettings = true, settingsError = null) }
        try {
            val settings = api.notificationSettings()
            _state.update { it.copy(notificationSettings = settings) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(settingsError = describe(e)) }
        } finally {
            _state.update { it.copy(isLoadingSettings = false) }
        }
    }

    internal suspend fun performSetNotificationsEnabled(enabled: Boolean) {
        updateSettings(NotificationSettingsPatch(enabled = enabled)) { it.copy(enabled = enabled) }
    }

    internal suspend fun performSetNotificationOutcomes(outcomes: NotificationOutcomes) {
        updateSettings(NotificationSettingsPatch(outcomes = outcomes)) { it.copy(outcomes = outcomes) }
    }

    private suspend fun updateSettings(
        change: NotificationSettingsPatch,
        apply: (NotificationSettings) -> NotificationSettings,
    ) {
        val prior = _state.value.notificationSettings
        _state.update { it.copy(settingsError = null, notificationSettings = apply(it.notificationSettings)) }
        try {
            val saved = api.updateNotificationSettings(change)
            _state.update { it.copy(notificationSettings = saved) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(notificationSettings = prior, settingsError = describe(e)) }
        }
    }

    internal suspend fun performAcknowledge(id: Long) {
        val original = _state.value.items.firstOrNull { it.id == id } ?: return
        if (original.isAcknowledged) return
        val stamped = original.copy(acknowledgedAt = clock() / MILLIS_PER_SECOND)
        replaceItem(id, stamped)
        try {
            api.acknowledgeNotification(id)
            onAcknowledged()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            replaceItem(id, original)
            _state.update { it.copy(errorMessage = describe(e)) }
        }
    }

    internal suspend fun performAcknowledgeGroup(group: HistoryGroup) {
        group.items.filterNot { it.isAcknowledged }.map { it.id }.forEach { performAcknowledge(it) }
    }

    internal suspend fun performExcludeRepo(repoFullName: String) {
        _state.update { it.copy(excludedRepoError = null) }
        try {
            val created = api.addExcludedRepo(repoFullName)
            _state.update { it.copy(excludedRepos = listOf(created) + it.excludedRepos) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(excludedRepoError = describe(e)) }
        }
    }

    internal suspend fun performRemoveExcludedRepo(id: Long) {
        val prior = _state.value.excludedRepos
        _state.update { state -> state.copy(excludedRepoError = null, excludedRepos = state.excludedRepos.filterNot { it.id == id }) }
        try {
            api.removeExcludedRepo(id)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(excludedRepos = prior, excludedRepoError = describe(e)) }
        }
    }

    private fun replaceItem(id: Long, item: NotificationHistoryItem) {
        _state.update { state -> state.copy(items = state.items.map { if (it.id == id) item else it }) }
    }

    private fun describe(e: Exception): String = e.message ?: e.toString()

    private companion object {
        const val MILLIS_PER_SECOND = 1000L
    }
}
