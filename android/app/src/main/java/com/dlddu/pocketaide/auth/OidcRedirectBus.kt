package com.dlddu.pocketaide.auth

import kotlinx.coroutines.CompletableDeferred

class OidcRedirectBus {
    private var pending: CompletableDeferred<String>? = null
    private var leftForBrowser = false

    @Synchronized
    fun begin(): CompletableDeferred<String> {
        pending?.completeExceptionally(OidcException.Cancelled())
        leftForBrowser = false
        return CompletableDeferred<String>().also { pending = it }
    }

    @Synchronized
    fun deliver(uri: String): Boolean {
        val target = pending ?: return false
        pending = null
        return target.complete(uri)
    }

    @Synchronized
    fun onHostPaused() {
        if (pending != null) leftForBrowser = true
    }

    @Synchronized
    fun onHostResumed() {
        val target = pending ?: return
        if (!leftForBrowser) return
        pending = null
        leftForBrowser = false
        target.completeExceptionally(OidcException.Cancelled())
    }
}
