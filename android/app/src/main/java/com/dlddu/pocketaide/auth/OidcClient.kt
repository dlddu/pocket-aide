package com.dlddu.pocketaide.auth

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.AuthConfig
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.FormBody
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.IOException
import java.net.URI
import java.net.URLDecoder

sealed class OidcException(message: String, cause: Throwable? = null) : Exception(message, cause) {
    class Discovery(cause: Throwable) : OidcException("discovery: ${cause.message}", cause)
    class NoCode : OidcException("no code in callback")
    class StateMismatch : OidcException("state mismatch in callback")
    class TokenExchange(val status: Int, val body: String) : OidcException("token $status: $body")
    class Cancelled : OidcException("cancelled")
    class NoRefreshToken : OidcException("no refresh token stored")
}

fun interface AuthBrowser {
    fun open(url: String)
}

class OidcClient(
    private val api: ApiClient,
    private val tokenStore: TokenStore,
    private val redirects: OidcRedirectBus,
    private val http: OkHttpClient = OkHttpClient(),
    private val clock: () -> Long = System::currentTimeMillis,
) : TokenRefresher {
    private val refreshLock = Mutex()

    suspend fun signIn(browser: AuthBrowser): TokenBundle {
        val cfg = api.authConfig()
        val discovery = fetchDiscovery(cfg.issuer)
        val verifier = Pkce.codeVerifier()
        val state = Pkce.state()
        val authUrl = authorizationUrl(discovery.authorizationEndpoint, cfg, state, Pkce.codeChallenge(verifier))

        val callback = redirects.begin()
        browser.open(authUrl)
        val params = queryParameters(callback.await())
        if (params["state"] != state) throw OidcException.StateMismatch()
        val code = params["code"] ?: throw OidcException.NoCode()

        val bundle = exchange(
            discovery.tokenEndpoint,
            mapOf(
                "grant_type" to "authorization_code",
                "code" to code,
                "redirect_uri" to cfg.redirectUri,
                "client_id" to cfg.clientId,
                "code_verifier" to verifier,
            ),
            fallbackRefreshToken = null,
        )
        tokenStore.save(bundle)
        return bundle
    }

    fun signOut() {
        tokenStore.clear()
    }

    override suspend fun refresh(): TokenBundle {
        val tokenBeforeWait = tokenStore.load()?.accessToken
        return refreshLock.withLock {
            val current = tokenStore.load()
            if (current != null && current.accessToken != tokenBeforeWait) return@withLock current
            val refreshToken = current?.refreshToken ?: throw OidcException.NoRefreshToken()
            val cfg = api.authConfig()
            val discovery = fetchDiscovery(cfg.issuer)
            val bundle = try {
                exchange(
                    discovery.tokenEndpoint,
                    mapOf(
                        "grant_type" to "refresh_token",
                        "refresh_token" to refreshToken,
                        "client_id" to cfg.clientId,
                    ),
                    fallbackRefreshToken = refreshToken,
                )
            } catch (e: OidcException.TokenExchange) {
                if (e.status in 400..499) tokenStore.clear()
                throw e
            }
            tokenStore.save(bundle)
            bundle
        }
    }

    private suspend fun exchange(
        tokenEndpoint: String,
        form: Map<String, String>,
        fallbackRefreshToken: String?,
    ): TokenBundle {
        val body = FormBody.Builder().apply { form.forEach { (k, v) -> add(k, v) } }.build()
        val request = Request.Builder().url(tokenEndpoint).post(body).build()
        val (status, text) = withContext(Dispatchers.IO) {
            try {
                http.newCall(request).execute().use { it.code to (it.body?.string() ?: "") }
            } catch (e: IOException) {
                throw OidcException.TokenExchange(-1, e.message ?: "transport error")
            }
        }
        if (status !in 200..299) throw OidcException.TokenExchange(status, text)
        val parsed = json.decodeFromString(TokenResponse.serializer(), text)
        return TokenBundle(
            accessToken = parsed.accessToken,
            refreshToken = parsed.refreshToken ?: fallbackRefreshToken,
            expiresAtMillis = clock() + (parsed.expiresIn ?: DEFAULT_EXPIRES_IN_SECONDS) * 1000,
        )
    }

    private suspend fun fetchDiscovery(issuer: String): Discovery = withContext(Dispatchers.IO) {
        try {
            val url = issuer.trimEnd('/') + "/.well-known/openid-configuration"
            http.newCall(Request.Builder().url(url).build()).execute().use { response ->
                if (!response.isSuccessful) throw IOException("HTTP ${response.code}")
                json.decodeFromString(Discovery.serializer(), response.body?.string() ?: "")
            }
        } catch (e: Exception) {
            throw OidcException.Discovery(e)
        }
    }

    @Serializable
    private data class Discovery(
        val issuer: String,
        @SerialName("authorization_endpoint") val authorizationEndpoint: String,
        @SerialName("token_endpoint") val tokenEndpoint: String,
    )

    @Serializable
    private data class TokenResponse(
        @SerialName("access_token") val accessToken: String,
        @SerialName("refresh_token") val refreshToken: String? = null,
        @SerialName("expires_in") val expiresIn: Long? = null,
    )

    companion object {
        private const val DEFAULT_EXPIRES_IN_SECONDS = 3600L
        private val json = Json { ignoreUnknownKeys = true }

        fun authorizationUrl(endpoint: String, cfg: AuthConfig, state: String, challenge: String): String =
            endpoint.toHttpUrl().newBuilder()
                .addQueryParameter("response_type", "code")
                .addQueryParameter("client_id", cfg.clientId)
                .addQueryParameter("redirect_uri", cfg.redirectUri)
                .addQueryParameter("scope", "openid profile email offline_access")
                .addQueryParameter("state", state)
                .addQueryParameter("code_challenge", challenge)
                .addQueryParameter("code_challenge_method", "S256")
                .build()
                .toString()

        fun queryParameters(uri: String): Map<String, String> {
            val query = URI(uri).rawQuery ?: return emptyMap()
            return query.split('&').filter { it.isNotEmpty() }.associate { pair ->
                val (k, v) = (pair.split('=', limit = 2) + "").take(2)
                URLDecoder.decode(k, "UTF-8") to URLDecoder.decode(v, "UTF-8")
            }
        }
    }
}
