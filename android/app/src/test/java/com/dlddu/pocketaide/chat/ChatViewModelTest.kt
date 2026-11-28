package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.CoroutineStart
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.Dispatcher
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.util.concurrent.CopyOnWriteArrayList
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger

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

    private fun chatJson(messages: String = "", pending: String = "null", queued: String = "null") =
        """{"conversation":{"id":7,"title":"첫 질문","created_at":1,"updated_at":2},"messages":[$messages],"pending":$pending,"queued":$queued}"""

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

    // PRD-1 AC1 (2026-11-27 개정): after the first read a poll asks only for
    // what came after the newest message it has, repeats the last ETag, and a
    // 304 leaves the screen as it was.
    @Test
    fun pollsReadOnlyNewerMessagesAndHonourNotModified() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"))))
        vm.load()
        assertEquals("/api/chat", server.takeRequest().path)

        server.enqueue(MockResponse().setHeader("ETag", "\"e1\"").setBody(chatJson(message(2, "assistant", "답"))))
        vm.load()
        val incremental = server.takeRequest()
        assertEquals("/api/chat?after=1", incremental.path)
        assertNull(incremental.getHeader("If-None-Match"))
        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })

        server.enqueue(MockResponse().setResponseCode(304).setHeader("ETag", "\"e1\""))
        vm.load()
        val conditional = server.takeRequest()
        assertEquals("/api/chat?after=2", conditional.path)
        assertEquals("\"e1\"", conditional.getHeader("If-None-Match"))
        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })
        assertNull(vm.state.value.error)
    }

    @Test
    fun aDifferentConversationIsReadInFull() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "옛 대화"))))
        vm.load()
        server.takeRequest()

        val other = """{"conversation":{"id":8,"title":"새","created_at":3,"updated_at":3},"messages":[${message(5, "user", "새 대화")}],"pending":null}"""
        server.enqueue(MockResponse().setBody(other))
        server.enqueue(MockResponse().setBody(other))
        vm.load()
        assertEquals("/api/chat?after=1", server.takeRequest().path)
        assertEquals("/api/chat", server.takeRequest().path)
        assertEquals(8L, vm.state.value.conversation?.id)
        assertEquals(listOf("새 대화"), vm.state.value.messages.map { it.content })
    }

    @Test
    fun mergingKeepsEachMessageOnceInIdOrder() {
        val onScreen = listOf(ChatMessage(1, "user", "a"), ChatMessage(4, "user", "d"))
        val polled = listOf(ChatMessage(3, "assistant", "c"), ChatMessage(4, "user", "d"))
        assertEquals(listOf(1L, 3L, 4L), onScreen.merging(polled).map { it.id })
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
        assertTrue(state.canSend)

        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"), """{"content":"쓰는 중","started_at":1}""")))
        vm.load()
        assertEquals("쓰는 중", vm.state.value.pending?.content)
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문") + "," + message(2, "assistant", "답"))))
        vm.load()
        assertNull(vm.state.value.pending)
        assertTrue(vm.state.value.canSend)
    }

    // PRD-1 AC1 (2026-11-09 개정): a message sent while a reply is in progress
    // is held instead of failing, and can be cancelled before it is sent.
    @Test
    fun sendWhileBusyQueuesInsteadOfFailing() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(202).setBody("""{"queued":{"id":"q1","content":"다음","queued_at":5}}"""))

        assertTrue(vm.performSend("다음"))
        assertEquals("/api/chat/messages", server.takeRequest().path)
        assertEquals("다음", vm.state.value.queued?.content)
        assertEquals("q1", vm.state.value.queued?.id)
        assertFalse(vm.state.value.canSend)

        server.enqueue(MockResponse().setResponseCode(204))
        vm.performCancelQueued()

        val request = server.takeRequest()
        // The cancel names the message on screen, so one that arrives after
        // it went out cannot drop whatever took its place.
        assertEquals("/api/chat/queue?id=q1", request.path)
        assertEquals("DELETE", request.method)
        assertNull(vm.state.value.queued)
        assertTrue(vm.state.value.canSend)
    }

    @Test
    fun cancelQueuedSwallowsAlreadyDispatched() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(202).setBody("""{"queued":{"id":"q1","content":"다음","queued_at":5}}"""))
        vm.performSend("다음")

        server.enqueue(MockResponse().setResponseCode(404).setBody("""{"error":"no message queued"}"""))
        vm.performCancelQueued()

        assertNull(vm.state.value.queued)
        assertNull(vm.state.value.error)
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

        server.enqueue(
            MockResponse().setResponseCode(503)
                .setBody("""{"error":"all login slots are in use","code":"auth_pool_exhausted"}"""),
        )
        assertFalse(vm.performSend("질문"))
        assertEquals(ChatError.SLOTS_BUSY, vm.state.value.error)

        server.enqueue(MockResponse().setResponseCode(503).setBody("""{"error":"feature disabled on the session platform"}"""))
        assertFalse(vm.performSend("질문"))
        assertEquals(ChatError.UNREACHABLE, vm.state.value.error)
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

    // PRD-AI채팅.AC1 (2026-11-28 개정): the button reads the way a poll does —
    // what came after the last read, merged into what is on screen — so a tap on
    // a long conversation costs one poll instead of the whole history.
    @Test
    fun syncPullsHistoryAddedInTheBackground() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"))))
        vm.load()
        server.takeRequest()

        // The reply finished server-side while the tab wasn't polling — sync pulls it now.
        server.enqueue(MockResponse().setHeader("ETag", "\"e2\"").setBody(chatJson(message(2, "assistant", "답"))))
        assertTrue(vm.performSync())

        val incremental = server.takeRequest()
        assertEquals("/api/chat?after=1", incremental.path)
        val state = vm.state.value
        assertEquals(listOf("질문", "답"), state.messages.map { it.content })
        assertFalse(state.isSyncing)
        assertTrue(state.canSync)

        // The tag the button's own read answered with is kept, so the next read
        // can be told nothing changed.
        server.enqueue(MockResponse().setResponseCode(304).setHeader("ETag", "\"e2\""))
        assertTrue(vm.performSync())
        val conditional = server.takeRequest()
        assertEquals("/api/chat?after=2", conditional.path)
        assertEquals("\"e2\"", conditional.getHeader("If-None-Match"))
        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })
    }

    // 새 메시지가 0건이어도 답이 시작·끝나거나 대기 메시지가 생기면 응답 본문이
    // 달라져 ETag 도 달라진다 — 200 이 오고, 빈 `messages` 는 화면의 이력을
    // 지우지 않으면서 진행 상태만 갱신한다.
    @Test
    fun syncAppliesProgressChangesThatCarryNoNewMessages() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"))))
        vm.load()
        server.takeRequest()
        assertFalse(vm.state.value.isReplying)

        // A reply started and a message was held behind it, both with nothing
        // new to show yet.
        server.enqueue(
            MockResponse().setHeader("ETag", "\"pending\"").setBody(
                chatJson(
                    pending = """{"content":"쓰는 중","started_at":9}""",
                    queued = """{"id":"q1","content":"다음","queued_at":9}""",
                ),
            ),
        )
        assertTrue(vm.performSync())

        assertEquals("/api/chat?after=1", server.takeRequest().path)
        assertEquals("쓰는 중", vm.state.value.pending?.content)
        assertEquals("q1", vm.state.value.queued?.id)
        assertEquals(listOf("질문"), vm.state.value.messages.map { it.content })
        // An empty delta does not move the cursor backwards either.
        server.enqueue(MockResponse().setResponseCode(304).setHeader("ETag", "\"pending\""))
        assertTrue(vm.performSync())
        assertEquals("/api/chat?after=1", server.takeRequest().path)
    }

    // 304 는 「지난 읽기 뒤로 바뀐 게 없다」는 뜻이므로 아무것도 적용하지
    // 않는다 — 빈 본문을 빈 목록으로 읽어 이력·진행 상태를 덮어쓰면 안 된다.
    @Test
    fun syncHonoursNotModifiedWithoutBlankingWhatIsOnScreen() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"))))
        vm.load()
        server.takeRequest()
        server.enqueue(
            MockResponse().setHeader("ETag", "\"e1\"").setBody(
                chatJson(message(2, "assistant", "답"), pending = """{"content":"더 쓰는 중","started_at":9}"""),
            ),
        )
        vm.load()
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(304).setHeader("ETag", "\"e1\""))
        assertTrue(vm.performSync())

        val conditional = server.takeRequest()
        assertEquals("/api/chat?after=2", conditional.path)
        assertEquals("\"e1\"", conditional.getHeader("If-None-Match"))
        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })
        assertEquals("더 쓰는 중", vm.state.value.pending?.content)
        assertNull(vm.state.value.error)
    }

    // 커서는 그 대화의 것이다 — 버튼이 받은 답이 다른 대화면 그 커서로는
    // 아무것도 말할 수 없으므로 전체를 받는 안전한 경로로 돌아간다.
    @Test
    fun syncFallsBackToAFullReadWhenTheConversationChanged() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "옛 대화"))))
        vm.load()
        server.takeRequest()

        val other = """{"conversation":{"id":8,"title":"새","created_at":3,"updated_at":3},"messages":[${message(5, "user", "새 대화")}],"pending":null,"queued":null}"""
        server.enqueue(MockResponse().setHeader("ETag", "\"other\"").setBody(other))
        server.enqueue(MockResponse().setBody(other))
        assertTrue(vm.performSync())

        assertEquals("/api/chat?after=1", server.takeRequest().path)
        val full = server.takeRequest()
        assertEquals("/api/chat", full.path)
        assertNull(full.getHeader("If-None-Match"))
        assertEquals(8L, vm.state.value.conversation?.id)
        // 옛 대화의 메시지는 새 대화에 섞이지 않는다.
        assertEquals(listOf("새 대화"), vm.state.value.messages.map { it.content })
    }

    // 보낸 메시지는 서버가 준 id 로 화면에 붙지만 커서가 되지는 않는다 — 그
    // 사이 끝난 답이 더 작은 id 를 갖고 있으면 그 답을 건너뛰게 되기 때문이다.
    @Test
    fun syncAfterSendAsksFromTheServerCursorNotTheSentMessage() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(5, "assistant", "지난 답"))))
        vm.load()
        server.takeRequest()

        // 보내는 사이 서버에서는 답(6)이 끝나 있었고, 보낸 메시지는 8 을 받았다.
        server.enqueue(MockResponse().setResponseCode(202).setBody("""{"message":${message(8, "user", "질문")}}"""))
        assertTrue(vm.performSend("질문"))
        server.takeRequest()
        assertEquals(listOf(5L, 8L), vm.state.value.messages.map { it.id })

        server.enqueue(
            MockResponse().setBody(
                chatJson(message(6, "assistant", "못 본 답") + "," + message(8, "user", "질문")),
            ),
        )
        assertTrue(vm.performSync())

        assertEquals("/api/chat?after=5", server.takeRequest().path)
        assertEquals(listOf(5L, 6L, 8L), vm.state.value.messages.map { it.id })
        assertEquals(listOf("지난 답", "못 본 답", "질문"), vm.state.value.messages.map { it.content })
    }

    // 폴링과 버튼이 겹치면 두 읽기가 함께 떠 있다. 늦게 도착한 앞 읽기의 답을
    // 적용하면 화면이 이미 지나온 상태로 되돌아가므로(끝난 답이 「쓰는 중」으로)
    // 그 답은 버린다 — 작업 패널의 taskReadSeq 와 같은 방식이다.
    @Test
    fun aSupersededReadDoesNotPutTheScreenBack() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"), """{"content":"쓰는 중","started_at":1}""")))
        vm.load()
        server.takeRequest()
        assertTrue(vm.state.value.isReplying)

        // 도착 순서가 아니라 **적용 순서**만 보게 한다: 먼저 닿은 읽기(버튼)의
        // 답을 뒤 읽기(폴링)가 적용될 때까지 붙잡아 둔다. 시간에 매달리지
        // 않으므로 러너가 느려도 같은 순서로 돈다.
        val firstArrived = CountDownLatch(1)
        val releaseFirst = CountDownLatch(1)
        val arrivals = AtomicInteger()
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse =
                when (arrivals.incrementAndGet()) {
                    1 -> {
                        firstArrived.countDown()
                        releaseFirst.await(5, TimeUnit.SECONDS)
                        MockResponse().setHeader("ETag", "\"stale\"")
                            .setBody(chatJson(pending = """{"content":"쓰는 중","started_at":1}"""))
                    }
                    2 -> MockResponse().setHeader("ETag", "\"fresh\"").setBody(chatJson(message(2, "assistant", "답")))
                    else -> MockResponse().setResponseCode(304).setHeader("ETag", "\"fresh\"")
                }
        }

        val tapped = launch(start = CoroutineStart.UNDISPATCHED) { vm.performSync() }
        assertTrue("버튼의 읽기가 먼저 서버에 닿아야 한다", firstArrived.await(5, TimeUnit.SECONDS))
        vm.load()

        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })
        assertNull(vm.state.value.pending)

        releaseFirst.countDown()
        tapped.join()

        // 늦게 온 앞 읽기의 답은 닿을 곳이 없다 — 끝난 답이 「쓰는 중」으로
        // 돌아가지 않는다.
        assertEquals(listOf("질문", "답"), vm.state.value.messages.map { it.content })
        assertNull(vm.state.value.pending)
        assertFalse(vm.state.value.isSyncing)

        // 버려진 읽기의 ETag 도 남지 않는다 — 다음 읽기는 적용된 쪽의 것을 쓴다.
        server.takeRequest()
        server.takeRequest()
        vm.load()
        val next = server.takeRequest()
        assertEquals("/api/chat?after=2", next.path)
        assertEquals("\"fresh\"", next.getHeader("If-None-Match"))
    }

    // 버려진 읽기는 **아무 요청도 새로 시작하지 않는다**. 겹친 두 증분 읽기가
    // 모두 「대화가 바뀌었다」는 답을 받으면 둘 다 전체 조회로 폴백하려 하는데,
    // 전체 조회는 긴 대화에서 수 MB 다 — 최신 읽기의 한 번만 나가고, 뒤늦게
    // 풀린 구 읽기는 결과를 버리는 데 그치지 않고 요청 자체를 내지 않아야 한다.
    @Test
    fun aSupersededReadThatLandsOnAnotherConversationStartsNoFullRead() = runTest {
        val vm = ChatViewModel(api())
        // 초기 A 전체 읽기 — 아래 요청 셈과는 별도다.
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "옛 대화"))))
        vm.load()
        assertEquals("/api/chat", server.takeRequest().path)

        val conversationB = """{"conversation":{"id":8,"title":"새","created_at":3,"updated_at":3},"messages":[${message(5, "user", "새 대화")}],"pending":null,"queued":null}"""
        val paths = CopyOnWriteArrayList<String>()
        val firstArrived = CountDownLatch(1)
        val releaseStale = CountDownLatch(1)
        val arrivals = AtomicInteger()
        // 어떤 요청이든 B 로 전환된 답을 준다. 먼저 닿은 읽기(버튼)의 답만
        // 최신 읽기가 B 전체를 끝낸 뒤로 미뤄, 도착이 아니라 **적용 순서**를
        // 고정한다 — 시간에 기대지 않는다.
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse {
                val n = arrivals.incrementAndGet()
                paths += request.path.orEmpty()
                if (n == 1) {
                    firstArrived.countDown()
                    releaseStale.await(5, TimeUnit.SECONDS)
                }
                return MockResponse().setBody(conversationB)
            }
        }

        val stale = launch(start = CoroutineStart.UNDISPATCHED) { vm.performSync() }
        assertTrue("구 읽기가 먼저 서버에 닿아야 한다", firstArrived.await(5, TimeUnit.SECONDS))
        // 최신 읽기: 같은 A 커서로 증분 → B 로 전환 → B 전체 1회.
        vm.load()

        assertEquals(8L, vm.state.value.conversation?.id)
        assertEquals(listOf("새 대화"), vm.state.value.messages.map { it.content })

        // 구 읽기가 끝까지 돌기를 기다린다 — 추가 요청을 냈다면 이 사이에 나간다.
        releaseStale.countDown()
        stale.join()

        assertEquals("두 증분 읽기는 같은 A 커서로 나간다", 2, paths.count { it == "/api/chat?after=1" })
        assertEquals("구 응답이 추가 전체 읽기를 시작했다", 1, paths.count { it == "/api/chat" })
        assertEquals(3, paths.size)
        assertEquals(8L, vm.state.value.conversation?.id)
        assertEquals(listOf("새 대화"), vm.state.value.messages.map { it.content })
        assertNull(vm.state.value.pending)
        assertFalse(vm.state.value.isSyncing)
    }

    @Test
    fun syncFailureSurfacesAnErrorAndCanBeRetried() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(502))
        assertTrue(vm.performSync())
        assertEquals(ChatError.UNREACHABLE, vm.state.value.error)
        assertFalse(vm.state.value.isSyncing)
        assertTrue(vm.state.value.canSync)

        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "재시도"))))
        assertTrue(vm.performSync())
        assertNull(vm.state.value.error)
        assertEquals(listOf("재시도"), vm.state.value.messages.map { it.content })
    }

    @Test
    fun syncIsGuardedAgainstDuplicateDispatchAndAnEmptyConversation() = runTest {
        val vm = ChatViewModel(api())

        // Before the first load, there is nothing to sync yet.
        assertFalse(vm.state.value.canSync)
        assertFalse(vm.performSync())
        assertEquals(0, server.requestCount)

        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()

        // A rapid repeat tap while a sync is already in flight is a no-op.
        assertFalse(ChatUiState(hasLoaded = true, isSyncing = true).canSync)
    }

    @Test
    fun syncStaysAvailableWhileAReplyIsInProgress() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "질문"), """{"content":"쓰는 중","started_at":1}""")))
        vm.load()
        server.takeRequest()
        assertTrue(vm.state.value.isReplying)
        assertTrue(vm.state.value.canSync)

        server.enqueue(MockResponse().setBody(chatJson(message(2, "assistant", "답"))))
        assertTrue(vm.performSync())
        assertNull(vm.state.value.pending)
        assertEquals(2, vm.state.value.messages.size)
    }

    @Test
    fun refreshNoticeWording() {
        assertEquals("스킬이 이미 최신이에요.", ChatViewModel.refreshNotice(ChatPluginRefresh(state = "ready")))
        assertEquals("이 대화에는 새로 받을 스킬이 없어요.", ChatViewModel.refreshNotice(ChatPluginRefresh(state = "unconfigured")))
    }

    // PRD-AI채팅.AC11: the unsent message is the view model's, so a trip to
    // another tab — which takes the screen out of composition — comes back to
    // it. Verbatim is the whole point: a line break the person typed is theirs.
    @Test
    fun theDraftIsKeptVerbatimAndClearedOnlyBySendingIt() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()

        vm.editDraft("첫 줄\n둘째 줄 ")
        assertEquals("첫 줄\n둘째 줄 ", vm.draft.value.text)
        assertEquals(7L, vm.draft.value.conversationId)

        // Re-reading the same conversation is what coming back to the tab does,
        // and it leaves the words alone.
        server.enqueue(MockResponse().setBody(chatJson(message(1, "assistant", "답"))))
        vm.load()
        assertEquals("첫 줄\n둘째 줄 ", vm.draft.value.text)

        server.enqueue(MockResponse().setResponseCode(202).setBody("""{"message":${message(2, "user", "첫 줄")}}"""))
        assertTrue(vm.performSendingDraft())
        assertEquals("/api/chat/messages", server.takeRequest().path)
        assertEquals("", vm.draft.value.text)
    }

    @Test
    fun aFailedSendLeavesTheWordsToRetryAndEmptyingTheFieldSticks() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()

        vm.editDraft("보낼 말")
        server.enqueue(MockResponse().setResponseCode(503))
        assertFalse(vm.performSendingDraft())
        server.takeRequest()
        assertEquals("보낼 말", vm.draft.value.text)

        // Cleared by hand is a draft of nothing, not an absent draft — nothing
        // may put the old words back.
        vm.editDraft("")
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        assertEquals("", vm.draft.value.text)
    }

    @Test
    fun theDraftIsNotCarriedIntoAnotherConversation() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson(message(1, "user", "옛 대화"))))
        vm.load()
        server.takeRequest()
        vm.editDraft("옛 대화에 쓰던 말")

        val other = """{"conversation":{"id":8,"title":"새","created_at":3,"updated_at":3},"messages":[],"pending":null}"""
        server.enqueue(MockResponse().setBody(other))
        server.enqueue(MockResponse().setBody(other))
        vm.load()

        assertEquals(8L, vm.state.value.conversation?.id)
        assertEquals("", vm.draft.value.text)
        assertEquals(8L, vm.draft.value.conversationId)
    }

    /**
     * Words typed before the first read have no conversation yet; the first one
     * read adopts them rather than throwing away what was already written.
     */
    @Test
    fun wordsWrittenBeforeTheConversationWasKnownAreAdoptedByIt() = runTest {
        val vm = ChatViewModel(api())
        vm.editDraft("먼저 쓴 말")
        assertNull(vm.draft.value.conversationId)

        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()

        assertEquals("먼저 쓴 말", vm.draft.value.text)
        assertEquals(7L, vm.draft.value.conversationId)
    }

    // PRD-AI채팅.AC11: holding the words and handing them back is not sending
    // them. Nothing but the person's own tap on 보내기 posts a message.
    @Test
    fun keepingAndRestoringTheDraftSendsNothing() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        vm.editDraft("보내지 않을 말")
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()

        assertEquals("보내지 않을 말", vm.draft.value.text)
        assertEquals(2, server.requestCount)
        repeat(2) {
            val request = server.takeRequest()
            assertEquals("GET", request.method)
        }
    }
}
