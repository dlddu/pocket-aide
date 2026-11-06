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
import org.junit.Assert.assertNotNull
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

    private fun api(): ApiClient {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ApiClient(server.url("/").toString(), store, clock = { now })
    }

    private fun message(id: Long, role: String, content: String) =
        """{"id":$id,"role":"$role","content":"$content","created_at":1}"""

    private fun chatJson(messages: String = "", pending: String = "null") =
        """{"conversation":{"id":7,"title":"첫 질문","created_at":1,"updated_at":2},"messages":[$messages],"pending":$pending}"""

    @Test
    fun loadShowsTheConversation() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "안녕") + "," + message(2, "assistant", "네"))))

        vm.load()

        assertEquals("/api/chat", server.takeRequest().path)
        val state = vm.state.value
        assertTrue(state.hasLoaded)
        assertEquals("첫 질문", state.conversation?.title)
        assertEquals(listOf(true, false), state.messages.map { it.isMine })
        assertFalse(state.isReplying)
        assertTrue(state.canRefresh)
    }

    @Test
    fun sendPostsTrimmedContentAndWaitsForTheReply() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(202).setBody("""{"message":${message(1, "user", "질문")}}"""))

        assertTrue(vm.performSend("  질문  "))

        val request = server.takeRequest()
        assertEquals("/api/chat/messages", request.path)
        assertEquals("""{"content":"질문"}""", request.body.readUtf8())
        val state = vm.state.value
        assertEquals(1, state.messages.size)
        assertTrue(state.isReplying)
        assertFalse(state.canSend)

        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"), """{"content":"쓰는 중","started_at":1}""")))
        vm.load()
        assertEquals("쓰는 중", vm.state.value.pending?.content)
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문") + "," + message(2, "assistant", "답"))))
        vm.load()
        assertNull(vm.state.value.pending)
        assertTrue(vm.state.value.canSend)
    }

    @Test
    fun busyAndUnreachableMapToMessages() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"a reply is still in progress"}"""))
        assertFalse(vm.performSend("질문"))
        assertEquals(ChatError.BUSY, vm.state.value.error)

        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"session platform unreachable"}"""))
        assertFalse(vm.performSend("질문"))
        assertEquals(ChatError.UNREACHABLE, vm.state.value.error)
        assertTrue(vm.state.value.messages.isEmpty())
    }

    @Test
    fun refreshPluginsReportsWhatChanged() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "q") + "," + message(2, "assistant", "a"))))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setBody("""{"state":"ready","revision":"abc","added":["p--new"],"updated":[],"removed":["p--old"]}"""))

        vm.performRefreshPlugins()

        assertEquals("/api/chat/refresh-plugins", server.takeRequest().path)
        val notice = vm.state.value.notice
        assertNotNull(notice)
        assertTrue(notice!!.contains("추가 1") && notice.contains("삭제 1"))
        assertFalse(vm.state.value.isRefreshing)

        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"a reply is still in progress"}"""))
        vm.performRefreshPlugins()
        assertEquals(ChatError.BUSY, vm.state.value.error)
    }

    @Test
    fun refreshNoticeWording() {
        assertEquals("스킬이 이미 최신이에요.", ChatViewModel.refreshNotice(ChatPluginRefresh(state = "ready")))
        assertEquals("이 대화에는 새로 받을 스킬이 없어요.", ChatViewModel.refreshNotice(ChatPluginRefresh(state = "unconfigured")))
    }
}
