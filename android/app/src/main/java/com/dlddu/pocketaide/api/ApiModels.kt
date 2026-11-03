package com.dlddu.pocketaide.api

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class AuthConfig(
    val issuer: String,
    @SerialName("client_id") val clientId: String,
    @SerialName("redirect_uri") val redirectUri: String,
    val audience: String,
)

@Serializable
data class MeResponse(val id: Int, val sub: String)

@Serializable
data class DeviceTokenRegistration(val token: String, val platform: String) {
    companion object {
        const val ANDROID = "android"
    }
}

@Serializable
class EmptyResponse

@Serializable
internal data class HealthResponse(val status: String)
