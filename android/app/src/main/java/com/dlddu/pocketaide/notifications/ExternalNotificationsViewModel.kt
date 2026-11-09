package com.dlddu.pocketaide.notifications

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ExternalNotificationsUiState(
    val items: List<ExternalNotification> = emptyList(),
    val unreadCount: Int = 0,
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val isMarkingAllRead: Boolean = false,
    val hasLoaded: Boolean = false,
    val hasMore: Boolean = false,
    val errorMessage: String? = null,
    val selected: ExternalNotification? = null,
)

class ExternalNotificationsViewModel(
    private val api: ApiClient,
    private val onMarkedRead: () -> Unit = {},
    private val clock: () -> Long = System::currentTimeMillis,
) : ViewModel() {
    private val _state = MutableStateFlow(ExternalNotificationsUiState())
    val state: StateFlow<ExternalNotificationsUiState> = _state.asStateFlow()

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

    fun loadMore() {
        viewModelScope.launch { performLoadMore() }
    }

    fun select(item: ExternalNotification) {
        _state.update { it.copy(selected = item) }
        viewModelScope.launch { markRead(item) }
    }

    fun dismissDetail() {
        _state.update { it.copy(selected = null) }
    }

    fun open(id: Long) {
        viewModelScope.launch {
            val item = notification(id) ?: return@launch
            select(item)
        }
    }

    fun markAllRead() {
        viewModelScope.launch { performMarkAllRead() }
    }

    internal suspend fun reload() {
        if (_state.value.isLoading) return
        _state.update { it.copy(isLoading = true) }
        try {
            val page = api.externalNotifications(limit = ExternalNotificationFormatting.PAGE_SIZE)
            _state.update {
                it.copy(
                    items = page.items,
                    unreadCount = page.unreadCount,
                    hasMore = page.items.size >= ExternalNotificationFormatting.PAGE_SIZE,
                    errorMessage = null,
                    hasLoaded = true,
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performLoadMore() {
        val current = _state.value
        val oldest = current.items.lastOrNull()
        if (current.isLoading || !current.hasMore || oldest == null) return
        _state.update { it.copy(isLoading = true) }
        try {
            val page = api.externalNotifications(limit = ExternalNotificationFormatting.PAGE_SIZE, before = oldest.id)
            _state.update { state ->
                val known = state.items.map { it.id }.toSet()
                state.copy(
                    items = state.items + page.items.filterNot { it.id in known },
                    hasMore = page.items.size >= ExternalNotificationFormatting.PAGE_SIZE,
                    errorMessage = null,
                )
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun notification(id: Long): ExternalNotification? {
        _state.value.items.firstOrNull { it.id == id }?.let { return it }
        return try {
            api.externalNotification(id)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
            null
        }
    }

    internal suspend fun markRead(item: ExternalNotification) {
        if (!item.isUnread) return
        try {
            api.markExternalNotificationRead(item.id)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            return
        }
        val now = clock() / 1000L
        _state.update { state ->
            val index = state.items.indexOfFirst { it.id == item.id }
            if (index < 0 || !state.items[index].isUnread) return@update state
            val items = state.items.toMutableList().also { it[index] = it[index].markedRead(now) }
            state.copy(items = items, unreadCount = maxOf(0, state.unreadCount - 1))
        }
        onMarkedRead()
    }

    internal suspend fun performMarkAllRead() {
        if (_state.value.isMarkingAllRead || _state.value.unreadCount == 0) return
        _state.update { it.copy(isMarkingAllRead = true) }
        try {
            api.markAllExternalNotificationsRead()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
            return
        } finally {
            _state.update { it.copy(isMarkingAllRead = false) }
        }
        val now = clock() / 1000L
        _state.update { state ->
            state.copy(
                items = state.items.map { if (it.isUnread) it.markedRead(now) else it },
                unreadCount = 0,
                errorMessage = null,
            )
        }
        onMarkedRead()
    }
}
