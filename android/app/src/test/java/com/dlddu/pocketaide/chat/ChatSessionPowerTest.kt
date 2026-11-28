package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

// PRD-1 AC12: the chat tab shows whether the session answering this
// conversation is on, and turns that same session back on when it is not.
class ChatSessionPowerTest {
    private val server = MockWebServer()
    private val now = 1_000_000L
    private var clock = now

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun api(): ApiClient {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ApiClient(server.url("/").toString(), store, clock = { now })
    }

    private fun viewModel() = ChatViewModel(api(), now = { clock })

    private fun powerJson(
        power: String,
        state: String? = null,
        lastAccess: String? = null,
        checkpoint: String? = null,
        restored: Boolean = false,
    ): String = buildList {
        add("\"power\":\"$power\"")
        state?.let { add("\"state\":\"$it\"") }
        lastAccess?.let { add("\"last_access\":\"$it\"") }
        checkpoint?.let { add("\"checkpoint\":$it") }
        if (restored) add("\"restored\":true")
    }.joinToString(",", "{", "}")

    @Test
    fun showsALiveSession() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("on", state = "active", lastAccess = "2026-11-28T00:00:00Z")))

        vm.loadSessionPower()

        assertEquals("/api/chat/session", server.takeRequest().path)
        val power = vm.state.value.sessionPower
        assertEquals(ChatSessionPowerState.ON, power?.powerState)
        assertFalse(vm.state.value.canPowerOn)
        assertNull(vm.state.value.powerError)
    }

    // PRD-11 AC5: an archived session is the one state that offers the button,
    // and it carries the archive time and size shown beside it.
    @Test
    fun anArchivedSessionOffersTheButton() = runTest {
        val vm = viewModel()
        server.enqueue(
            MockResponse().setBody(
                powerJson(
                    "off",
                    state = "snapshot",
                    checkpoint = """{"createdAt":"2026-11-28T00:00:00Z","sizeBytes":2048}""",
                ),
            ),
        )

        vm.loadSessionPower()
        server.takeRequest()

        assertEquals(ChatSessionPowerState.OFF, vm.state.value.sessionPower?.powerState)
        assertTrue(vm.state.value.canPowerOn)
        val detail = vm.state.value.sessionPower!!.detail()
        assertTrue(detail, detail.contains("아카이브"))
        assertTrue(detail, detail.contains("이어져요"))
    }

    // The first message is what opens a session (AC4), so before it there is
    // nothing to turn on — and no button.
    @Test
    fun noSessionYetOffersNoButton() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("none")))

        vm.loadSessionPower()
        server.takeRequest()

        assertEquals(ChatSessionPowerState.ABSENT, vm.state.value.sessionPower?.powerState)
        assertFalse(vm.state.value.canPowerOn)
        assertEquals("첫 메시지를 보내면 세션이 열려요", vm.state.value.sessionPower?.detail())
    }

    // A state neither side can place stays unknown, with the platform's own
    // word kept and no button — a guess is what this must not do.
    @Test
    fun anUnrecognisedStateStaysUnknown() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("hibernating", state = "hibernating")))

        vm.loadSessionPower()
        server.takeRequest()

        assertEquals(ChatSessionPowerState.UNKNOWN, vm.state.value.sessionPower?.powerState)
        assertFalse(vm.state.value.canPowerOn)
        assertTrue(vm.state.value.sessionPower!!.detail().contains("hibernating"))
    }

    // PRD-1 AC12: the turn-on is one request with no session id — the backend
    // restores the one answering this user — and what it says is what the
    // screen shows.
    @Test
    fun powerOnRestoresTheSessionAndSaysSo() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("off", state = "snapshot")))
        vm.loadSessionPower()
        server.takeRequest()

        server.enqueue(MockResponse().setBody(powerJson("on", state = "active", restored = true)))
        vm.performPowerOn()

        val request = server.takeRequest()
        assertEquals("/api/chat/session/power", request.path)
        assertEquals("POST", request.method)
        assertEquals(ChatSessionPowerState.ON, vm.state.value.sessionPower?.powerState)
        assertEquals("세션을 켰어요. 이어서 보내면 같은 대화로 이어져요.", vm.state.value.notice)
        assertFalse(vm.state.value.isPoweringOn)
        assertFalse(vm.state.value.canPowerOn)
    }

    // A session that was already on is not reported as "turned on", and a
    // restore the platform took without a state coming back is reported as
    // accepted — never as on.
    @Test
    fun powerOnDoesNotClaimMoreThanHappened() = runTest {
        val alreadyOn = ChatSessionPower(power = "on", state = "active")
        assertTrue(alreadyOn.powerOnNotice.contains("이미 켜져"))
        val accepted = ChatSessionPower(power = "unknown", restored = true)
        assertTrue(accepted.powerOnNotice.contains("접수"))
        assertFalse(accepted.powerOnNotice.contains("켰어요"))
    }

    // A failed read keeps the last state on screen and says it failed: a chip
    // blanked by a network blip would read as "no session", a different fact.
    @Test
    fun aFailedReadKeepsTheLastStateAndSaysItFailed() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("off", state = "snapshot")))
        vm.loadSessionPower()
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"session platform unreachable"}"""))
        vm.loadSessionPower(force = true)
        server.takeRequest()

        assertEquals(ChatPowerError.UNREACHABLE, vm.state.value.powerError)
        assertEquals(ChatSessionPowerState.OFF, vm.state.value.sessionPower?.powerState)
    }

    // PRD-1 AC12 · PRD-11 AC11: the refusals the tab explains differently are
    // told apart by the backend's status and code.
    @Test
    fun refusalsAreToldApart() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("off", state = "snapshot")))
        vm.loadSessionPower()
        server.takeRequest()

        for ((response, expected) in listOf(
            MockResponse().setResponseCode(409).setBody("""{"code":"chat_power_busy"}""") to ChatPowerError.BUSY,
            MockResponse().setResponseCode(409).setBody("""{"code":"chat_power_no_session"}""") to ChatPowerError.NO_SESSION,
            MockResponse().setResponseCode(404).setBody("""{"code":"chat_power_session_gone"}""") to ChatPowerError.GONE,
            MockResponse().setResponseCode(503).setBody("""{"code":"auth_pool_exhausted"}""") to ChatPowerError.SLOTS_BUSY,
            MockResponse().setResponseCode(503).setBody("""{"error":"session platform error"}""") to ChatPowerError.UNREACHABLE,
        )) {
            server.enqueue(response)
            vm.performPowerOn()
            server.takeRequest()
            assertEquals(expected, vm.state.value.powerError)
            // The state the screen was showing is still the archived session,
            // so the button is still there to try again.
            assertTrue(vm.state.value.canPowerOn)
        }
    }

    // The state changes on the platform's own schedule, not the
    // conversation's, so the poll reads it once a minute rather than once a
    // poll — and a send, which restores an archived session, makes it due again.
    @Test
    fun theStateIsReadOnItsOwnSlowerSchedule() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(powerJson("on", state = "active")))
        vm.loadSessionPower()
        assertEquals("/api/chat/session", server.takeRequest().path)

        vm.loadSessionPower()
        assertEquals("a second read inside the window asked again", 1, server.requestCount)

        clock += ChatViewModel.POWER_REFRESH_MILLIS
        server.enqueue(MockResponse().setBody(powerJson("off", state = "snapshot")))
        vm.loadSessionPower()
        assertEquals("/api/chat/session", server.takeRequest().path)
        assertEquals(ChatSessionPowerState.OFF, vm.state.value.sessionPower?.powerState)
    }
}
