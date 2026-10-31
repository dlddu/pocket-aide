package com.dlddu.pocketaide.auth

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.AuthConfig
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.test.runTest
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.mockwebserver.Dispatcher
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Before
import org.junit.Test
import java.util.concurrent.atomic.AtomicInteger

class OidcClientTest {
    private val server = MockWebServer()
    private val tokenCalls = AtomicInteger()
    private var tokenResponse = MockResponse().setBody("""{"access_token":"new","expires_in":60}""")
    private var lastTokenForm = ""

    @Before
    fun setUp() {
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse {
                val base = server.url("/").toString().trimEnd('/')
                return when (request.path) {
                    "/api/auth/config" -> MockResponse().setBody(
                        """{"issuer":"$base/","client_id":"app","redirect_uri":"$REDIRECT_URI","audience":"api"}""",
                    )
                    "/.well-known/openid-configuration" -> MockResponse().setBody(
                        """{"issuer":"$base","authorization_endpoint":"$base/authorize","token_endpoint":"$base/token"}""",
                    )
                    "/token" -> {
                        tokenCalls.incrementAndGet()
                        lastTokenForm = request.body.readUtf8()
                        tokenResponse
                    }
                    else -> MockResponse().setResponseCode(404)
                }
            }
        }
        server.start()
    }

    @After
    fun tearDown() = server.shutdown()

    private fun oidc(store: TokenStore, bus: OidcRedirectBus = OidcRedirectBus()): OidcClient {
        val api = ApiClient(server.url("/").toString(), store)
        return OidcClient(api, store, bus, clock = { 0L })
    }

    @Test
    fun authorizationUrlCarriesPkceParameters() {
        val cfg = AuthConfig("https://idp", "app", "pocketaide://callback", "api")
        val url = OidcClient.authorizationUrl("https://idp/authorize?prompt=login", cfg, "st", "ch").toHttpUrl()

        assertEquals("login", url.queryParameter("prompt"))
        assertEquals("code", url.queryParameter("response_type"))
        assertEquals("app", url.queryParameter("client_id"))
        assertEquals("pocketaide://callback", url.queryParameter("redirect_uri"))
        assertEquals("openid profile email offline_access", url.queryParameter("scope"))
        assertEquals("st", url.queryParameter("state"))
        assertEquals("ch", url.queryParameter("code_challenge"))
        assertEquals("S256", url.queryParameter("code_challenge_method"))
    }

    @Test
    fun parsesCallbackQuery() {
        val params = OidcClient.queryParameters("pocketaide://callback?code=a%2Bb&state=x%20y")

        assertEquals("a+b", params["code"])
        assertEquals("x y", params["state"])
    }

    @Test
    fun signInExchangesCodeWithVerifier() = runTest {
        val store = InMemoryTokenStore()
        val bus = OidcRedirectBus()
        tokenResponse = MockResponse().setBody("""{"access_token":"a1","refresh_token":"r1","expires_in":120}""")
        val client = oidc(store, bus)

        val bundle = client.signIn { url ->
            val state = url.toHttpUrl().queryParameter("state")
            bus.deliver("pocketaide://callback?code=the-code&state=$state")
        }

        assertEquals(TokenBundle("a1", "r1", 120_000), bundle)
        assertEquals(bundle, store.load())
        val form = lastTokenForm
        assertTrue(form.contains("grant_type=authorization_code"))
        assertTrue(form.contains("code=the-code"))
        assertTrue(form.contains("code_verifier="))
    }

    @Test
    fun signInRejectsMismatchedState() = runTest {
        val bus = OidcRedirectBus()
        val client = oidc(InMemoryTokenStore(), bus)

        try {
            client.signIn { bus.deliver("pocketaide://callback?code=c&state=forged") }
            fail("expected StateMismatch")
        } catch (_: OidcException.StateMismatch) {
        }
    }

    @Test
    fun refreshKeepsOldRefreshTokenWhenServerOmitsIt() = runTest {
        val store = InMemoryTokenStore(TokenBundle("old", "r0", 0))

        val bundle = oidc(store).refresh()

        assertEquals(TokenBundle("new", "r0", 60_000), bundle)
        assertEquals(bundle, store.load())
    }

    @Test
    fun refreshClearsStoreOnClientError() = runTest {
        val store = InMemoryTokenStore(TokenBundle("old", "r0", 0))
        tokenResponse = MockResponse().setResponseCode(400).setBody("invalid_grant")

        try {
            oidc(store).refresh()
            fail("expected TokenExchange")
        } catch (e: OidcException.TokenExchange) {
            assertEquals(400, e.status)
        }
        assertNull(store.load())
    }

    @Test
    fun refreshKeepsStoreOnServerError() = runTest {
        val original = TokenBundle("old", "r0", 0)
        val store = InMemoryTokenStore(original)
        tokenResponse = MockResponse().setResponseCode(503)

        runCatching { oidc(store).refresh() }

        assertEquals(original, store.load())
    }

    @Test
    fun concurrentRefreshesShareOneExchange() = runTest {
        val store = InMemoryTokenStore(TokenBundle("old", "r0", 0))
        val client = oidc(store)

        val results = List(3) { async { client.refresh() } }.awaitAll()

        assertEquals(1, tokenCalls.get())
        results.forEach { assertEquals("new", it.accessToken) }
    }

    private companion object {
        const val REDIRECT_URI = "pocketaide://callback"
    }
}
