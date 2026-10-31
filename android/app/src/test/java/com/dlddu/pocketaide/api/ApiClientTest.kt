package com.dlddu.pocketaide.api

import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import com.dlddu.pocketaide.auth.TokenRefresher
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Before
import org.junit.Test

class ApiClientTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Serializable
    private data class Draft(
        val title: String,
        val memo: String = "",
        @SerialName("due_date") val dueDate: String? = null,
        val done: Boolean = false,
    )

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun client(bundle: TokenBundle? = TokenBundle("access", "refresh", now + 3_600_000)) =
        InMemoryTokenStore(bundle).let { store ->
            store to ApiClient(server.url("/").toString(), store, clock = { now })
        }

    @Test
    fun authenticatedRequestCarriesBearerToken() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setBody("""{"id":7,"sub":"u7"}"""))

        assertEquals(MeResponse(7, "u7"), api.me())
        assertEquals("Bearer access", server.takeRequest().getHeader("Authorization"))
    }

    @Test
    fun unauthenticatedRequestHasNoAuthorizationHeader() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setBody("""{"status":"ok"}"""))

        api.health()

        assertNull(server.takeRequest().getHeader("Authorization"))
    }

    @Test
    fun encodesDefaultsAndOmitsNulls() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setBody("{}"))

        api.post<Draft, EmptyResponse>("/api/todos/personal", Draft(title = "a"))

        assertEquals("""{"title":"a","memo":"","done":false}""", server.takeRequest().body.readUtf8())
    }

    @Test
    fun emptyBodyDecodesAsEmptyResponse() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setResponseCode(204))

        api.delete("/api/scratchpad/3")

        assertEquals("/api/scratchpad/3", server.takeRequest().path)
    }

    @Test
    fun preservesQueryString() = runTest {
        val (_, api) = client()
        val url = api.url("/api/approvals?limit=20&before=5")

        assertEquals("/api/approvals", url.encodedPath)
        assertEquals("limit=20&before=5", url.encodedQuery)
    }

    @Test
    fun retriesOnceAfterRefreshOn401() = runTest {
        val (store, api) = client()
        api.tokenRefresher = TokenRefresher {
            TokenBundle("fresh", "refresh2", now + 3_600_000).also(store::save)
        }
        server.enqueue(MockResponse().setResponseCode(401))
        server.enqueue(MockResponse().setBody("""{"id":1,"sub":"s"}"""))

        api.me()

        assertEquals("Bearer access", server.takeRequest().getHeader("Authorization"))
        assertEquals("Bearer fresh", server.takeRequest().getHeader("Authorization"))
    }

    @Test
    fun refreshesProactivelyWhenTokenIsAboutToExpire() = runTest {
        val (store, api) = client(TokenBundle("stale", "refresh", now + 10_000))
        var refreshed = 0
        api.tokenRefresher = TokenRefresher {
            refreshed++
            TokenBundle("fresh", "refresh", now + 3_600_000).also(store::save)
        }
        server.enqueue(MockResponse().setBody("""{"id":1,"sub":"s"}"""))

        api.me()

        assertEquals(1, refreshed)
        assertEquals("Bearer fresh", server.takeRequest().getHeader("Authorization"))
    }

    @Test
    fun badStatusSurfacesStatusAndBody() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setResponseCode(500).setBody("boom"))

        try {
            api.me()
            fail("expected BadStatus")
        } catch (e: ApiException.BadStatus) {
            assertEquals(500, e.status)
            assertEquals("boom", e.body)
        }
    }

    @Test
    fun malformedJsonIsDecodingError() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setBody("not json"))

        try {
            api.me()
            fail("expected Decoding")
        } catch (e: ApiException) {
            assertTrue(e is ApiException.Decoding)
        }
    }

    @Test
    fun registersAndroidDeviceToken() = runTest {
        val (_, api) = client()
        server.enqueue(MockResponse().setResponseCode(201).setBody("""{"ok":true}"""))

        api.registerDeviceToken("fcm:token")

        val request = server.takeRequest()
        assertEquals("/api/device-tokens", request.path)
        assertEquals("""{"token":"fcm:token","platform":"android"}""", request.body.readUtf8())
    }
}
