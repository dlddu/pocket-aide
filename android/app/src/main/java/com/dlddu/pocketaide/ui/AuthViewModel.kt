package com.dlddu.pocketaide.ui

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.PocketAideApplication
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.api.MeResponse
import com.dlddu.pocketaide.auth.AuthBrowser
import com.dlddu.pocketaide.auth.OidcException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class AuthUiState(
    val signedIn: Boolean,
    val isSigningIn: Boolean = false,
    val signInError: String? = null,
    val healthOk: Boolean = false,
    val healthError: String? = null,
    val me: MeResponse? = null,
    val meError: String? = null,
    val pushDenied: Boolean = false,
)

class AuthViewModel(application: Application) : AndroidViewModel(application) {
    private val container = (application as PocketAideApplication).container
    val api get() = container.api
    val canSignIn: Boolean get() = container.backendConfigured

    private val _state = MutableStateFlow(
        AuthUiState(
            signedIn = container.tokenStore.load() != null,
            healthError = if (container.backendConfigured) null else "BackendBaseURL missing",
        ),
    )
    val state: StateFlow<AuthUiState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            refreshHealth()
            if (_state.value.signedIn) {
                refreshMe()
                registerForPush()
            }
        }
    }

    fun signIn(browser: AuthBrowser) {
        if (_state.value.isSigningIn) return
        viewModelScope.launch {
            _state.update { it.copy(isSigningIn = true, signInError = null) }
            try {
                container.oidc.signIn(browser)
                _state.update { it.copy(signedIn = true) }
                refreshMe()
                registerForPush()
            } catch (e: OidcException.Cancelled) {
                _state.update { it.copy(signedIn = container.tokenStore.load() != null) }
            } catch (e: Exception) {
                _state.update {
                    it.copy(signInError = e.message ?: e.toString(), signedIn = container.tokenStore.load() != null)
                }
            } finally {
                _state.update { it.copy(isSigningIn = false) }
            }
        }
    }

    fun signOut() {
        container.oidc.signOut()
        container.pushRegistrar.unregister()
        _state.update { it.copy(signedIn = false, me = null, meError = null) }
    }

    fun refreshPushPermission() {
        val denied = container.pushRegistrar.isConfigured && !container.pushRegistrar.notificationsPermitted()
        _state.update { it.copy(pushDenied = denied) }
    }

    private suspend fun refreshHealth() {
        if (!container.backendConfigured) return
        try {
            container.api.health()
            _state.update { it.copy(healthOk = true, healthError = null) }
        } catch (e: Exception) {
            _state.update { it.copy(healthOk = false, healthError = e.message ?: e.toString()) }
        }
    }

    private suspend fun refreshMe() {
        _state.update { it.copy(meError = null) }
        try {
            val me = container.api.me()
            _state.update { it.copy(me = me) }
        } catch (e: ApiException.BadStatus) {
            if (e.status == 401) signOut() else _state.update { it.copy(me = null, meError = e.message) }
        } catch (e: Exception) {
            _state.update { it.copy(me = null, meError = e.message ?: e.toString()) }
        }
    }

    private suspend fun registerForPush() {
        container.pushRegistrar.register(container.api)
        refreshPushPermission()
    }
}
