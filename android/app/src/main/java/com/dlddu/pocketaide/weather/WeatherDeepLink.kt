package com.dlddu.pocketaide.weather

import java.net.URI
import java.net.URISyntaxException

object WeatherDeepLink {
    const val URL = "pocketaide://weather"
    private val SCHEMES = setOf("pocketaide", "pocketaide-dev")
    private const val HOST = "weather"

    fun matches(uri: String): Boolean {
        val parsed = try {
            URI(uri)
        } catch (e: URISyntaxException) {
            return false
        }
        val scheme = parsed.scheme?.lowercase() ?: return false
        return scheme in SCHEMES && parsed.host?.lowercase() == HOST
    }
}
