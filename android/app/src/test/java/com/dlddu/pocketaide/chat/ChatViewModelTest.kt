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

class ChatViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): ChatViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ChatViewModel(ApiClient(server.url("/").toString(), store, clock = { now }))
    }

    private fun sessionJson(id: Long, title: String = "주제 $id") =
        """{"id":$id,"title":"$title","created_at":1,"updated_at":${10 + id}}"""

    private fun messageJson(id: Long, role: String, content: String) =
        """{"id":$id,"role":"$role","content":"$content","created_at":2}"""

    private fun sse(vararg blocks: String) = MockResponse()
        .setHeader("Content-Type", "text/event-stream")
        .setBody(blocks.joinToString("") { it + "\n\n" })

    @Test
    fun firstLoadResumesTheLatestSession() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${sessionJson(2)},${sessionJson(1)}]}"""))
        server.enqueue(
            MockResponse().setBody(
                """{"session":${sessionJson(2)},"messages":[${messageJson(1, "user", "질문")},${messageJson(2, "assistant", "답")}]}""",
            ),
        )

        vm.reload()

        assertEquals("/api/chat/sessions", server.takeRequest().path)
        assertEquals("/api/chat/sessions/2", server.takeRequest().path)
        assertEquals(2L, vm.state.value.current?.id)
        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })
        assertTrue(vm.state.value.hasLoaded)
    }

    @Test
    fun emptyAccountStartsAFreshConversation() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":null}"""))

        vm.reload()

        assertNull(vm.state.value.current)
        assertTrue(vm.state.value.messages.isEmpty())
        assertEquals(1, server.requestCount)
    }

    @Test
    fun firstSendCreatesTheSessionThenStreamsTheReply() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(201).setBody(sessionJson(5, "")))
        server.enqueue(
            sse(
                """data: {"delta":"안"}""",
                """data: {"delta":"녕"}""",
                """event: done""" + "\n" + """data: {"message":${messageJson(9, "assistant", "안녕")}}""",
            ),
        )
        server.enqueue(MockResponse().setBody("""{"items":[${sessionJson(5, "인사")}]}"""))

        vm.performSend("  인사해줘  ")

        val create = server.takeRequest()
        assertEquals("POST", create.method)
        assertEquals("/api/chat/sessions", create.path)
        val send = server.takeRequest()
        assertEquals("/api/chat/sessions/5/messages", send.path)
        assertEquals("""{"content":"인사해줘"}""", send.body.readUtf8())
        assertEquals("Bearer access", send.getHeader("Authorization"))
        assertEquals(listOf("인사해줘", "안녕"), vm.state.value.messages.map { it.content })
        assertEquals(listOf(ChatRole.USER, ChatRole.ASSISTANT), vm.state.value.messages.map { it.role })
        assertFalse(vm.state.value.isSending)
        assertEquals("인사", vm.state.value.current?.title)
        assertNull(vm.state.value.errorMessage)
    }

    @Test
    fun localMessageIdsNeverCollideAcrossTurns() = runTest {
        val vm = viewModel()
        vm.openForTest(ChatSession(1))
        repeat(2) { turn ->
            server.enqueue(sse("""data: {"delta":"답$turn"}""", """event: done""" + "\n" + """data: {"message":null}"""))
            server.enqueue(MockResponse().setBody("""{"items":[]}"""))
            vm.performSend("질문$turn")
        }

        val ids = vm.state.value.messages.map { it.id }
        assertEquals(4, ids.size)
        assertEquals(4, ids.toSet().size)
        assertEquals(listOf("질문0", "답0", "질문1", "답1"), vm.state.value.messages.map { it.content })
    }

    @Test
    fun notConfiguredShowsGuidanceAndKeepsTheDraftMessageOnly() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(201).setBody(sessionJson(3, "")))
        server.enqueue(MockResponse().setResponseCode(503).setBody("chat is not configured"))
        server.enqueue(MockResponse().setBody("""{"items":[]}"""))

        vm.performSend("안녕")

        assertEquals(ChatCopy.UNCONFIGURED, vm.state.value.errorMessage)
        assertFalse(vm.state.value.isSending)
        assertEquals(listOf(ChatRole.USER), vm.state.value.messages.map { it.role })
    }

    @Test
    fun streamErrorKeepsThePartialReplyAndReportsFailure() = runTest {
        val vm = viewModel()
        vm.openForTest(ChatSession(4))
        server.enqueue(sse("""data: {"delta":"부분"}""", """event: error""" + "\n" + """data: {"message":"reply failed"}"""))
        server.enqueue(MockResponse().setBody("""{"items":[]}"""))

        vm.performSend("질문")

        assertEquals(ChatCopy.REPLY_FAILED, vm.state.value.errorMessage)
        assertEquals(listOf("질문", "부분"), vm.state.value.messages.map { it.content })
        assertEquals(2, vm.state.value.messages.map { it.id }.toSet().size)
    }

    @Test
    fun blankInputIsIgnored() = runTest {
        val vm = viewModel()

        vm.performSend("   ")

        assertEquals(0, server.requestCount)
        assertTrue(vm.state.value.messages.isEmpty())
    }

    @Test
    fun openingASessionRestoresItsHistoryAndClosesTheList() = runTest {
        val vm = viewModel()
        vm.showHistory()
        server.enqueue(
            MockResponse().setBody("""{"session":${sessionJson(1)},"messages":[${messageJson(1, "user", "옛 질문")}]}"""),
        )

        vm.performOpen(ChatSession(1))

        assertEquals(listOf("옛 질문"), vm.state.value.messages.map { it.content })
        assertFalse(vm.state.value.showsHistory)
        assertEquals(1L, vm.state.value.current?.id)
    }

    @Test
    fun newConversationClearsTheCurrentSession() = runTest {
        val vm = viewModel()
        vm.openForTest(ChatSession(1))

        vm.startNewConversation()

        assertNull(vm.state.value.current)
        assertTrue(vm.state.value.messages.isEmpty())
    }

    private suspend fun ChatViewModel.openForTest(session: ChatSession) {
        server.enqueue(MockResponse().setBody("""{"session":${sessionJson(session.id)},"messages":[]}"""))
        performOpen(session)
        server.takeRequest()
    }
}
