package com.dlddu.pocketaide.approvals

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ApprovalHistoryUiState(
    val items: List<ApprovalRequest> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val hasLoaded: Boolean = false,
    val nextCursor: String? = null,
    val errorMessage: String? = null,
) {
    val hasMore: Boolean get() = nextCursor != null
}

class ApprovalHistoryViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ApprovalHistoryUiState())
    val state: StateFlow<ApprovalHistoryUiState> = _state.asStateFlow()

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

    internal suspend fun reload() {
        if (_state.value.isLoading) return
        _state.update { it.copy(isLoading = true) }
        try {
            val page = api.approvalHistory()
            _state.update {
                it.copy(items = page.items, nextCursor = page.nextCursor, errorMessage = null, hasLoaded = true)
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
        val cursor = current.nextCursor
        if (current.isLoading || cursor == null) return
        _state.update { it.copy(isLoading = true) }
        try {
            val page = api.approvalHistory(cursor)
            _state.update { it.copy(items = it.items + page.items, nextCursor = page.nextCursor, errorMessage = null) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }
}
