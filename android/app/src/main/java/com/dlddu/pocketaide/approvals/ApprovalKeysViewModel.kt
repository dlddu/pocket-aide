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

data class ApprovalKeysUiState(
    val keys: List<ApprovalCallerKey> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val hasLoaded: Boolean = false,
    val isIssuing: Boolean = false,
    val issued: IssuedApprovalKey? = null,
    val isRegistering: Boolean = false,
    val registerOutcome: ApprovalServiceAccountOutcome? = null,
    val autoResponse: ApprovalAutoResponse? = null,
    val modeNotice: String? = null,
    val errorMessage: String? = null,
    val issueNotice: String? = null,
    val registerNotice: String? = null,
)

class ApprovalKeysViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ApprovalKeysUiState())
    val state: StateFlow<ApprovalKeysUiState> = _state.asStateFlow()

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

    fun issue(name: String) {
        viewModelScope.launch { performIssue(name) }
    }

    fun registerServiceAccount(name: String, namespace: String, account: String, onRegistered: () -> Unit) {
        viewModelScope.launch {
            if (performRegisterServiceAccount(name, namespace, account)) onRegistered()
        }
    }

    fun setAutoResponse(mode: ApprovalAutoResponse) {
        viewModelScope.launch { performSetAutoResponse(mode) }
    }

    fun revoke(key: ApprovalCallerKey) {
        viewModelScope.launch { performRevoke(key) }
    }

    fun finishIssuing() {
        _state.update { it.copy(issued = null, issueNotice = null) }
    }

    fun finishRegistering() {
        _state.update { it.copy(registerOutcome = null, registerNotice = null) }
    }

    internal suspend fun reload() {
        if (_state.value.isLoading) return
        _state.update { it.copy(isLoading = true) }
        try {
            val keys = api.listApprovalKeys()
            val mode = api.approvalSettings().autoResponse
            _state.update { it.copy(keys = keys, autoResponse = mode, errorMessage = null, hasLoaded = true) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performIssue(raw: String) {
        val name = ApprovalKeyFormatting.trimmedName(raw) ?: return
        if (_state.value.isIssuing) return
        _state.update { it.copy(isIssuing = true) }
        try {
            when (val outcome = api.issueApprovalKey(name)) {
                is ApprovalKeyIssueOutcome.Issued -> {
                    _state.update { it.copy(issued = outcome.issued, issueNotice = null) }
                    reload()
                }
                ApprovalKeyIssueOutcome.DuplicateName ->
                    _state.update { it.copy(issueNotice = ApprovalKeyFormatting.DUPLICATE_NAME_NOTICE) }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(issueNotice = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isIssuing = false) }
        }
    }

    internal suspend fun performRegisterServiceAccount(rawName: String, namespace: String, account: String): Boolean {
        if (_state.value.isRegistering) return false
        val name = ApprovalKeyFormatting.trimmedName(rawName) ?: return false
        val serviceAccount = ApprovalKeyFormatting.serviceAccount(namespace, account)
        if (serviceAccount == null) {
            _state.update {
                it.copy(
                    registerOutcome = ApprovalServiceAccountOutcome.InvalidServiceAccount,
                    registerNotice = ApprovalKeyFormatting.INVALID_SERVICE_ACCOUNT_NOTICE,
                )
            }
            return false
        }
        _state.update { it.copy(isRegistering = true) }
        try {
            val outcome = api.registerApprovalServiceAccount(name, serviceAccount)
            val notice = when (outcome) {
                is ApprovalServiceAccountOutcome.Registered -> null
                ApprovalServiceAccountOutcome.DuplicateName -> ApprovalKeyFormatting.SERVICE_ACCOUNT_DUPLICATE_NAME_NOTICE
                ApprovalServiceAccountOutcome.DuplicateServiceAccount -> ApprovalKeyFormatting.DUPLICATE_SERVICE_ACCOUNT_NOTICE
                ApprovalServiceAccountOutcome.InvalidServiceAccount -> ApprovalKeyFormatting.INVALID_SERVICE_ACCOUNT_NOTICE
            }
            _state.update { it.copy(registerOutcome = outcome, registerNotice = notice) }
            if (outcome is ApprovalServiceAccountOutcome.Registered) {
                _state.update { it.copy(isRegistering = false) }
                reload()
                return true
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(registerOutcome = null, registerNotice = e.message ?: e.toString()) }
        } finally {
            _state.update { it.copy(isRegistering = false) }
        }
        return false
    }

    internal suspend fun performSetAutoResponse(mode: ApprovalAutoResponse) {
        if (mode == _state.value.autoResponse) return
        try {
            val updated = api.updateApprovalAutoResponse(mode).autoResponse
            _state.update { it.copy(autoResponse = updated, modeNotice = null) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(modeNotice = e.message ?: e.toString()) }
        }
    }

    internal suspend fun performRevoke(key: ApprovalCallerKey) {
        try {
            api.revokeApprovalKey(key.id)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
        }
        reload()
    }
}
