package com.dlddu.pocketaide.api

import com.dlddu.pocketaide.auth.TokenRefresher
import com.dlddu.pocketaide.auth.TokenStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json
import kotlinx.serialization.serializer
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.IOException

sealed class ApiException(message: String, cause: Throwable? = null) : Exception(message, cause) {
    class BadStatus(val status: Int, val body: String) : ApiException("HTTP $status: $body")
    class Decoding(cause: Throwable) : ApiException("decode: ${cause.message}", cause)
    class Transport(cause: Throwable) : ApiException("transport: ${cause.message}", cause)
}

class ApiClient(
    baseUrl: String,
    private val tokenStore: TokenStore,
    private val http: OkHttpClient = OkHttpClient(),
    private val clock: () -> Long = System::currentTimeMillis,
) {
    val baseUrl: HttpUrl = baseUrl.trimEnd('/').toHttpUrl()

    @Volatile
    var tokenRefresher: TokenRefresher? = null

    suspend fun health() {
        get("/healthz", HealthResponse.serializer(), authenticated = false)
    }

    suspend fun authConfig(): AuthConfig = get("/api/auth/config", AuthConfig.serializer(), authenticated = false)

    suspend fun me(): MeResponse = get("/api/me", MeResponse.serializer())

    suspend fun registerDeviceToken(token: String) {
        post(
            "/api/device-tokens",
            DeviceTokenRegistration(token = token, platform = DeviceTokenRegistration.ANDROID),
            DeviceTokenRegistration.serializer(),
            EmptyResponse.serializer(),
        )
    }

    suspend fun <T> get(path: String, response: KSerializer<T>, authenticated: Boolean = true): T =
        decode(response, send("GET", path, null, authenticated))

    suspend fun <I, O> post(
        path: String,
        body: I,
        request: KSerializer<I>,
        response: KSerializer<O>,
        authenticated: Boolean = true,
    ): O = decode(response, send("POST", path, encode(request, body), authenticated))

    suspend fun <I, O> patch(
        path: String,
        body: I,
        request: KSerializer<I>,
        response: KSerializer<O>,
        authenticated: Boolean = true,
    ): O = decode(response, send("PATCH", path, encode(request, body), authenticated))

    suspend fun delete(path: String, authenticated: Boolean = true) {
        send("DELETE", path, null, authenticated)
    }

    suspend inline fun <reified T> get(path: String, authenticated: Boolean = true): T =
        get(path, serializer<T>(), authenticated)

    suspend inline fun <reified I, reified O> post(path: String, body: I, authenticated: Boolean = true): O =
        post(path, body, serializer<I>(), serializer<O>(), authenticated)

    suspend inline fun <reified I, reified O> patch(path: String, body: I, authenticated: Boolean = true): O =
        patch(path, body, serializer<I>(), serializer<O>(), authenticated)

    fun url(path: String): HttpUrl {
        val parts = path.split("?", limit = 2)
        val builder = baseUrl.newBuilder()
        parts[0].trim('/').split('/').filter { it.isNotEmpty() }.forEach { builder.addPathSegment(it) }
        if (parts.size == 2) builder.encodedQuery(parts[1])
        return builder.build()
    }

    private fun <I> encode(serializer: KSerializer<I>, body: I): RequestBody =
        json.encodeToString(serializer, body).toRequestBody(JSON_MEDIA_TYPE)

    private fun <T> decode(serializer: KSerializer<T>, raw: String): T {
        if (raw.isBlank() && serializer.descriptor == EmptyResponse.serializer().descriptor) {
            @Suppress("UNCHECKED_CAST")
            return EmptyResponse() as T
        }
        return try {
            json.decodeFromString(serializer, raw)
        } catch (e: SerializationException) {
            throw ApiException.Decoding(e)
        } catch (e: IllegalArgumentException) {
            throw ApiException.Decoding(e)
        }
    }

    private suspend fun send(method: String, path: String, body: RequestBody?, authenticated: Boolean): String {
        if (authenticated) refreshIfExpired()
        return perform(method, path, body, authenticated, retriedAfterRefresh = false)
    }

    private suspend fun perform(
        method: String,
        path: String,
        body: RequestBody?,
        authenticated: Boolean,
        retriedAfterRefresh: Boolean,
    ): String {
        val builder = Request.Builder().url(url(path)).method(method, body)
        if (authenticated) {
            tokenStore.load()?.accessToken?.let { builder.header("Authorization", "Bearer $it") }
        }
        val (status, text) = withContext(Dispatchers.IO) {
            try {
                http.newCall(builder.build()).execute().use { it.code to (it.body?.string() ?: "") }
            } catch (e: IOException) {
                throw ApiException.Transport(e)
            }
        }
        val refresher = tokenRefresher
        if (authenticated && status == 401 && !retriedAfterRefresh && refresher != null &&
            tokenStore.load()?.refreshToken != null
        ) {
            try {
                refresher.refresh()
            } catch (e: Exception) {
                throw ApiException.BadStatus(status, text)
            }
            return perform(method, path, body, authenticated, retriedAfterRefresh = true)
        }
        if (status !in 200..299) throw ApiException.BadStatus(status, text)
        return text
    }

    private suspend fun refreshIfExpired() {
        val refresher = tokenRefresher ?: return
        val bundle = tokenStore.load() ?: return
        if (bundle.refreshToken == null) return
        if (bundle.expiresAtMillis - clock() < REFRESH_LEEWAY_MILLIS) {
            runCatching { refresher.refresh() }
        }
    }

    companion object {
        private const val REFRESH_LEEWAY_MILLIS = 30_000L
        private val JSON_MEDIA_TYPE = "application/json".toMediaType()

        val json = Json {
            ignoreUnknownKeys = true
            encodeDefaults = true
            explicitNulls = false
        }
    }
}
