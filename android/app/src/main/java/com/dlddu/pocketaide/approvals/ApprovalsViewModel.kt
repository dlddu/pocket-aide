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

sealed interface ApprovalDestination {
    data object Pending : ApprovalDestination
    data class Detail(val item: ApprovalRequest) : ApprovalDestination
    data object Settings : ApprovalDestination
    data object History : ApprovalDestination
}

data class ApprovalsUiState(
    val items: List<ApprovalRequest> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val hasLoaded: Boolean = false,
    val autoResponse: ApprovalAutoResponse = ApprovalAutoResponse.OFF,
    val errorMessage: String? = null,
    val destination: ApprovalDestination = ApprovalDestination.Pending,
) {
    val pendingCount: Int get() = items.size
}

class ApprovalsViewModel(val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ApprovalsUiState())
    val state: StateFlow<ApprovalsUiState> = _state.asStateFlow()

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

    fun turnOffAutoResponse() {
        viewModelScope.launch {
            performTurnOffAutoResponse()
            reload()
        }
    }

    fun navigate(destination: ApprovalDestination) {
        _state.update { it.copy(destination = destination) }
    }

    fun open(approvalId: String) {
        viewModelScope.launch { performOpen(approvalId) }
    }

    internal suspend fun reload() {
        if (_state.value.isLoading) return
        _state.update { it.copy(isLoading = true) }
        try {
            val items = api.listPendingApprovals()
            val mode = api.approvalSettings().autoResponse
            _state.update { it.copy(items = items, autoResponse = mode, errorMessage = null, hasLoaded = true) }
            _pendingCount.value = items.size
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performTurnOffAutoResponse() {
        try {
            val mode = api.updateApprovalAutoResponse(ApprovalAutoResponse.OFF).autoResponse
            _state.update { it.copy(autoResponse = mode) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        }
    }

    internal suspend fun performOpen(approvalId: String) {
        val item = try {
            api.approval(approvalId)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            return
        }
        navigate(ApprovalDestination.Detail(item))
    }
}

data class ApprovalDetailUiState(
    val item: ApprovalRequest,
    val isDeciding: Boolean = false,
    val notice: String? = null,
    val loadError: String? = null,
) {
    fun canDecide(nowMillis: Long): Boolean =
        item.status == ApprovalStatus.PENDING && !item.isExpired(nowMillis) && !isDeciding
}

class ApprovalDetailViewModel(item: ApprovalRequest, private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ApprovalDetailUiState(item = item))
    val state: StateFlow<ApprovalDetailUiState> = _state.asStateFlow()

    fun reload() {
        viewModelScope.launch { performReload() }
    }

    fun decide(decision: ApprovalDecision) {
        viewModelScope.launch { performDecide(decision) }
    }

    internal suspend fun performReload() {
        try {
            val item = api.approval(_state.value.item.id)
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
            when (val outcome = api.decideApproval(_state.value.item.id, decision)) {
                is ApprovalDecisionOutcome.Decided -> _state.update { it.copy(item = outcome.item, notice = null) }
                is ApprovalDecisionOutcome.Conflict -> _state.update {
                    it.copy(item = outcome.item, notice = ApprovalFormatting.conflictNotice(outcome.item))
                }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(notice = ApprovalFormatting.FAILED_DECISION_NOTICE) }
        } finally {
            _state.update { it.copy(isDeciding = false) }
        }
    }
}
