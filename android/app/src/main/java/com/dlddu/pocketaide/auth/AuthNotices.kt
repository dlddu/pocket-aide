package com.dlddu.pocketaide.auth

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.getAndUpdate
import kotlinx.coroutines.flow.update

data class AuthNotice(val kind: Kind, val reason: SignOutReason) {
    enum class Kind(val title: String) {
        UNAUTHORIZED("요청이 거절됐어요 (401)"),
        SIGNED_OUT("로그인 화면으로 돌아왔어요"),
    }
}

class AuthNotices(
    private val store: SignOutReasonStore,
    private val clock: () -> Long = System::currentTimeMillis,
) {
    private val _current = MutableStateFlow<AuthNotice?>(null)
    val current: StateFlow<AuthNotice?> = _current.asStateFlow()

    fun onLaunch(hasToken: Boolean) {
        if (hasToken) {
            store.markSession()
            return
        }
        if (!store.hadSession()) return
        val reason = store.load() ?: SignOutReason.tokenMissing(clock()).also(store::save)
        _current.value = AuthNotice(AuthNotice.Kind.SIGNED_OUT, reason)
    }

    fun onUnauthorized(request: String, apiBody: String) {
        val reason = SignOutReason.unauthorized(store.load(), request, apiBody, clock())
        _current.update { it ?: AuthNotice(AuthNotice.Kind.UNAUTHORIZED, reason) }
    }

    fun onSessionExpired(request: String, apiBody: String) {
        val reason = SignOutReason.unauthorized(store.load(), request, apiBody, clock())
        store.save(reason)
        _current.value = AuthNotice(AuthNotice.Kind.SIGNED_OUT, reason)
    }

    fun onSignedIn() {
        store.clear()
        store.markSession()
        _current.value = null
    }

    fun onSignedOutByUser() {
        store.clear()
        store.forgetSession()
        _current.value = null
    }

    fun dismiss() {
        val dismissed = _current.getAndUpdate { null }
        if (dismissed?.kind == AuthNotice.Kind.SIGNED_OUT) {
            store.clear()
            store.forgetSession()
        }
    }
}
