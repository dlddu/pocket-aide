package com.dlddu.pocketaide.auth

import java.security.MessageDigest
import java.security.SecureRandom
import java.util.Base64

object Pkce {
    private val random = SecureRandom()
    private val encoder = Base64.getUrlEncoder().withoutPadding()

    fun codeVerifier(): String = randomToken(64)

    fun state(): String = randomToken(16)

    fun codeChallenge(verifier: String): String =
        encoder.encodeToString(MessageDigest.getInstance("SHA-256").digest(verifier.toByteArray(Charsets.US_ASCII)))

    private fun randomToken(bytes: Int): String = encoder.encodeToString(ByteArray(bytes).also(random::nextBytes))
}
