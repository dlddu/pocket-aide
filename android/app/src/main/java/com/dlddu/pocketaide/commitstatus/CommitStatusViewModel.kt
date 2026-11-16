package com.dlddu.pocketaide.commitstatus

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.approvals.ApprovalDecision
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class CommitStatusUiState(
    val items: List<CommitStatusRequest> = emptyList(),
    val history: List<CommitStatusRequest> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val hasLoaded: Boolean = false,
    val errorMessage: String? = null,
    val opened: CommitStatusRequest? = null,
) {
    val pendingCount: Int get() = items.size
}

class CommitStatusViewModel(val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(CommitStatusUiState())
    val state: StateFlow<CommitStatusUiState> = _state.asStateFlow()

    private val _pendingCount = MutableStateFlow(0)
    val pendingCount: StateFlow<Int> = _pendingCount.asStateFlow()

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

    fun show(item: CommitStatusRequest?) {
        _state.update { it.copy(opened = item) }
    }

    fun open(id: String) {
        viewModelScope.launch {
            val item = try {
                api.commitStatusRequest(id)
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                return@launch
            }
            show(item)
        }
    }

    internal suspend fun reload() {
        if (_state.value.isLoading) return
        _state.update { it.copy(isLoading = true) }
        try {
            val list = api.listCommitStatusRequests()
            _state.update { it.copy(items = list.items, history = list.history, errorMessage = null, hasLoaded = true) }
            _pendingCount.value = list.items.size
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }
}

data class CommitStatusDetailUiState(
    val item: CommitStatusRequest,
    val isDeciding: Boolean = false,
    val notice: String? = null,
    val loadError: String? = null,
) {
    val canDecide: Boolean get() = item.status == CommitStatusRequestStatus.PENDING && !isDeciding
}

class CommitStatusDetailViewModel(item: CommitStatusRequest, private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(CommitStatusDetailUiState(item = item))
    val state: StateFlow<CommitStatusDetailUiState> = _state.asStateFlow()

    fun reload() {
        viewModelScope.launch { performReload() }
    }

    fun decide(decision: ApprovalDecision) {
        viewModelScope.launch { performDecide(decision) }
    }

    internal suspend fun performReload() {
        try {
            val item = api.commitStatusRequest(_state.value.item.id)
            _state.update { it.copy(item = item, loadError = null) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(loadError = e.message ?: e.toString()) }
        }
    }

    internal suspend fun performDecide(decision: ApprovalDecision) {
        if (_state.value.isDeciding) return
        _state.update { it.copy(isDeciding = true) }
        try {
            when (val outcome = api.decideCommitStatusRequest(_state.value.item.id, decision)) {
                is CommitStatusDecisionOutcome.Decided -> _state.update { it.copy(item = outcome.item, notice = null) }
                is CommitStatusDecisionOutcome.Conflict -> _state.update {
                    it.copy(item = outcome.item, notice = CommitStatusFormatting.conflictNotice(outcome.item))
                }
                is CommitStatusDecisionOutcome.WriteFailed -> _state.update {
                    it.copy(item = outcome.item, notice = CommitStatusFormatting.writeFailedNotice(outcome.message))
                }
                is CommitStatusDecisionOutcome.Unavailable -> _state.update {
                    it.copy(notice = CommitStatusFormatting.UNAVAILABLE_NOTICE)
                }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(notice = CommitStatusFormatting.FAILED_DECISION_NOTICE) }
        } finally {
            _state.update { it.copy(isDeciding = false) }
        }
    }
}
