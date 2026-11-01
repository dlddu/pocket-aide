package com.dlddu.pocketaide.sessions

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

class SessionsViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L
    private var clockMillis = now

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun api(): ApiClient {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ApiClient(server.url("/").toString(), store, clock = { now })
    }

    private fun sessionJson(id: String, state: String = "active") =
        """{"id":"$id","name":"세션 $id","workloadType":"claude-code","state":"$state","createdAt":"2026-10-01T00:00:00Z","lastAccess":"2026-10-01T00:00:00Z"}"""

    private val configJson =
        """{"workloadTypes":[{"type":"claude-code","label":"Claude Code","takesModel":true,"input":"prompt","defaultModel":"opus","models":["opus"]},{"type":"shell","label":"Shell","takesModel":false,"input":"terminal"}]}"""

    @Test
    fun loadFetchesConfigThenSessions() = runTest {
        val vm = SessionsViewModel(api())
        server.enqueue(MockResponse().setBody(configJson))
        server.enqueue(MockResponse().setBody("""{"sessions":[${sessionJson("a")}]}"""))

        vm.reload()

        assertEquals("/api/sessions/config", server.takeRequest().path)
        assertEquals("/api/sessions", server.takeRequest().path)
        val state = vm.state.value
        assertTrue(state.hasLoaded)
        assertEquals(listOf("a"), state.sessions.map { it.id })
        assertEquals("Claude Code", state.workloadType(state.sessions.single())?.displayName)
    }

    @Test
    fun loadFailureMapsError() = runTest {
        val vm = SessionsViewModel(api())
        server.enqueue(MockResponse().setResponseCode(502))
        server.enqueue(MockResponse().setResponseCode(503))

        vm.reload()

        val state = vm.state.value
        assertTrue(state.configFailed)
        assertEquals(AgentSessionError.UNREACHABLE, state.error)
        assertTrue(state.hasLoaded)
    }

    @Test
    fun createPostsTrimmedPayloadAndPrepends() = runTest {
        val vm = SessionsViewModel(api())
        server.enqueue(MockResponse().setBody(configJson))
        server.enqueue(MockResponse().setBody("""{"sessions":[${sessionJson("a")}]}"""))
        vm.reload()
        server.takeRequest()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(201).setBody(sessionJson("b")))
        val workload = vm.state.value.config!!.type("claude-code")!!

        val created = vm.performCreate("  새 일  ", workload, " opus ")

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/sessions", post.path)
        assertEquals("""{"name":"새 일","workloadType":"claude-code","model":"opus"}""", post.body.readUtf8())
        assertEquals("b", created?.id)
        assertEquals(listOf("b", "a"), vm.state.value.sessions.map { it.id })
        assertFalse(vm.state.value.isCreating)
    }

    @Test
    fun createWithoutModelOmitsField() = runTest {
        val vm = SessionsViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(sessionJson("c")))
        val shell = AgentSessionWorkloadType(type = "shell", input = AgentSessionInput.TERMINAL)

        vm.performCreate("터미널", shell, "ignored")

        assertEquals("""{"name":"터미널","workloadType":"shell"}""", server.takeRequest().body.readUtf8())
    }

    @Test
    fun createFailureSetsCreateError() = runTest {
        val vm = SessionsViewModel(api())
        server.enqueue(MockResponse().setResponseCode(500))

        assertNull(vm.performCreate("x", AgentSessionWorkloadType(type = "claude-code"), null))

        assertEquals(AgentSessionError.UNKNOWN, vm.state.value.createError)
        assertEquals(
            "세션을 띄우지 못했어요. 세션 플랫폼에 「claude-code」 타입 설정(이미지·자격증명)이 있는지 확인해 주세요.",
            createFailureMessage(AgentSessionError.UNKNOWN, AgentSessionWorkloadType(type = "claude-code")),
        )
    }

    private fun detail(state: String = "active", type: AgentSessionWorkloadType? = AgentSessionWorkloadType(type = "claude-code")): SessionDetailViewModel {
        val session = ApiClient.json.decodeFromString(AgentSession.serializer(), sessionJson("a", state))
        return SessionDetailViewModel(session, type, api(), clock = { clockMillis })
    }

    private fun readJson(payload: String, next: Long, state: String = "active"): String {
        val escaped = ApiClient.json.encodeToString(kotlinx.serialization.serializer<String>(), payload)
        return """{"session":${sessionJson("a", state)},"payload":$escaped,"nextOffset":$next}"""
    }

    @Test
    fun readOutputAccumulatesAndParsesTurns() = runTest {
        val vm = detail()
        val s = AgentSessionTranscript.SENTINEL
        server.enqueue(MockResponse().setBody(readJson("${s}session-platform input seq=1\nhi\n", 30)))

        vm.readOutput()

        val read = server.takeRequest()
        assertEquals("/api/sessions/a/read", read.path)
        assertEquals("""{"offset":0}""", read.body.readUtf8())
        val state = vm.state.value
        assertTrue(state.hasLoadedOutput)
        assertEquals(30L, state.output.nextOffset)
        assertEquals(1, state.transcript.turns.size)
        assertEquals(AgentSessionResponsePhase.STREAMING, state.phase)

        server.enqueue(MockResponse().setBody(readJson("${s}session-platform end seq=1 status=ok\n", 60)))
        vm.readOutput()

        assertEquals("""{"offset":30}""", server.takeRequest().body.readUtf8())
        assertEquals(AgentSessionTurnStatus.Ok, vm.state.value.transcript.turns.single().status)
        assertEquals(AgentSessionResponsePhase.NONE, vm.state.value.phase)
    }

    @Test
    fun readOutputRereadsWhenOffsetShrinks() = runTest {
        val vm = detail()
        server.enqueue(MockResponse().setBody(readJson("old", 100)))
        vm.readOutput()
        server.takeRequest()
        server.enqueue(MockResponse().setBody(readJson("", 5)))
        server.enqueue(MockResponse().setBody(readJson("fresh", 5)))

        vm.readOutput()

        assertEquals("""{"offset":100}""", server.takeRequest().body.readUtf8())
        assertEquals("""{"offset":0}""", server.takeRequest().body.readUtf8())
        assertEquals("fresh", vm.state.value.displayText)
    }

    @Test
    fun terminalOutputIsCleaned() = runTest {
        val vm = detail(type = AgentSessionWorkloadType(type = "shell", input = AgentSessionInput.TERMINAL))
        server.enqueue(MockResponse().setBody(readJson("\u001B[32m\$ ls\u001B[0m\r\n", 10)))

        vm.readOutput()

        assertEquals("\$ ls\n", vm.state.value.displayText)
        assertFalse(vm.state.value.transcript.hasTurns)
    }

    @Test
    fun sendQueuesPhaseAndTerminalAppendsNewline() = runTest {
        val vm = detail()
        server.enqueue(MockResponse().setBody("""{"session":${sessionJson("a")}}"""))

        assertTrue(vm.performSend("do it"))

        val write = server.takeRequest()
        assertEquals("/api/sessions/a/write", write.path)
        assertEquals("""{"payload":"do it"}""", write.body.readUtf8())
        assertEquals(AgentSessionResponsePhase.QUEUED, vm.state.value.phase)

        val terminal = detail(type = AgentSessionWorkloadType(type = "shell", input = AgentSessionInput.TERMINAL))
        server.enqueue(MockResponse().setBody("{}"))
        assertTrue(terminal.performSend("ls"))
        assertEquals("""{"payload":"ls\n"}""", server.takeRequest().body.readUtf8())
        assertEquals(AgentSessionResponsePhase.NONE, terminal.state.value.phase)
    }

    @Test
    fun sendRejectedWhileSnapshotOrTooLarge() = runTest {
        val snapshot = detail(state = "snapshot")
        assertFalse(snapshot.performSend("x"))

        val vm = detail()
        assertFalse(vm.performSend("a".repeat(AgentSessionPrompt.MAX_BYTES + 1)))
        assertEquals(AgentSessionError.TOO_LARGE, vm.state.value.error)
        assertEquals(0, server.requestCount)
    }

    @Test
    fun sendBusyMapsError() = runTest {
        val vm = detail()
        server.enqueue(MockResponse().setResponseCode(429))

        assertFalse(vm.performSend("x"))

        assertEquals(AgentSessionError.QUEUE_FULL, vm.state.value.error)
        assertFalse(vm.state.value.isSending)
    }

    @Test
    fun archiveDisabledMapsTo503() = runTest {
        val vm = detail()
        server.enqueue(MockResponse().setResponseCode(503))

        vm.performArchive()

        assertEquals("/api/sessions/a/snapshot", server.takeRequest().path)
        assertEquals(AgentSessionError.DISABLED, vm.state.value.error)
        assertFalse(vm.state.value.isArchiving)
    }

    @Test
    fun restoreSwitchesSnapshot() = runTest {
        val vm = detail(state = "snapshot")
        server.enqueue(MockResponse().setBody(sessionJson("a", "active")))

        vm.performRestore()

        assertEquals("/api/sessions/a/switch", server.takeRequest().path)
        assertEquals(AgentSessionState.ACTIVE, vm.state.value.session.state)
    }

    @Test
    fun deleteTreatsNotFoundAsDeleted() = runTest {
        val vm = detail()
        server.enqueue(MockResponse().setResponseCode(404))

        vm.performDelete()

        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/sessions/a", delete.path)
        assertTrue(vm.state.value.deleted)
    }

    @Test
    fun tickRefreshesStateWhenStaleAndSkipsReadForSnapshot() = runTest {
        val vm = detail(type = null)
        server.enqueue(MockResponse().setBody(configJson))
        server.enqueue(MockResponse().setBody(sessionJson("a", "snapshot")))

        vm.tick()

        assertEquals("/api/sessions/config", server.takeRequest().path)
        assertEquals("/api/sessions/a", server.takeRequest().path)
        assertEquals("Claude Code", vm.state.value.type?.displayName)
        assertEquals(AgentSessionState.SNAPSHOT, vm.state.value.session.state)
        assertEquals(2, server.requestCount)

        clockMillis += 1_000
        vm.tick()
        assertEquals(2, server.requestCount)
    }
}
