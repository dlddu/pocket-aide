package com.dlddu.pocketaide.auth

import kotlinx.serialization.Serializable

@Serializable
data class TokenBundle(
    val accessToken: String,
    val refreshToken: String?,
    val expiresAtMillis: Long,
)

interface TokenStore {
    fun load(): TokenBundle?
    fun save(bundle: TokenBundle)
    fun clear()
}

fun interface TokenRefresher {
    suspend fun refresh(): TokenBundle
}

class InMemoryTokenStore(initial: TokenBundle? = null) : TokenStore {
    @Volatile
    private var bundle: TokenBundle? = initial

    override fun load(): TokenBundle? = bundle

    override fun save(bundle: TokenBundle) {
        this.bundle = bundle
    }

    override fun clear() {
        bundle = null
    }
}
