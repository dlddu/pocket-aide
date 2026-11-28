package com.dlddu.pocketaide.chat

import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import com.dlddu.pocketaide.designsystem.DesignTokens
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

class ChatTasksTest {
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

    // One task as the backend now serves it: the agreed task fields, the turn
    // count, and the open turn the panel reads "something is waiting" from.
    private fun task(
        id: String,
        status: String,
        description: String = "무엇을 why",
        criteria: String = "머지되면 끝",
        question: String = "",
        turns: Long = 0,
    ) = """{"id":"$id","title":"작업 $id","status":"$status",""" +
        """"description":"$description","completion_criteria":"$criteria","turns_total":$turns""" +
        (if (question.isBlank()) "}" else ""","waiting":{"turn_id":7,"request":"이거 해 줘","question":"$question"}}""")

    private fun tasksJson(
        tasks: String = "",
        sessionId: String = "cx1",
        counts: String = "{}",
        truncated: Boolean = false,
    ) = """{"session_id":"$sessionId","tasks":[$tasks],"counts":$counts,"truncated":$truncated}"""

    // The wide layout is decided by the window's own available width, so the
    // boundary is exactly the token and nothing either side of it drifts.
    @Test
    fun wideLayoutStartsAtTheBreakpoint() {
        assertEquals(600.dp, DesignTokens.Breakpoints.wide)
        assertFalse("a folded phone stays narrow", isWideChatWindow(411.dp))
        assertFalse("a split window stays narrow", isWideChatWindow(599.dp))
        assertTrue("an unfolded foldable goes wide, portrait included", isWideChatWindow(600.dp))
        assertTrue(isWideChatWindow(841.dp))
    }

    @Test
    fun loadShowsTheLedgerWithStatusStageAndNote() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = task("t1", "blocked", question = "CI 가 막혔다", turns = 2),
                    counts = """{"blocked":1,"running":0}""",
                ),
            ),
        )

        vm.loadTasks()

        assertEquals("/api/chat/tasks", server.takeRequest().path)
        val state = vm.tasks.value
        assertTrue(state.hasLoaded)
        assertNull(state.error)
        assertFalse(state.isLoading)
        assertFalse(state.isEmpty)
        assertEquals("cx1", state.sessionId)
        val task = state.tasks.single()
        assertEquals(ChatTaskStatus.BLOCKED, task.status)
        // The two fields the ledger page had nowhere to put.
        assertEquals("무엇을 why", task.description)
        assertEquals("머지되면 끝", task.completionCriteria)
        // What is waiting comes from the task's open turn, not a count of rows
        // beside it.
        assertEquals("CI 가 막혔다", task.waiting?.question)
        assertEquals(7L, task.waiting?.turnId)
        assertEquals(1, state.counts["blocked"])
    }

    // 선택지를 올린 그 턴의 세 가지(보낸 요청·설명·질문)가 목록 응답으로
    // 함께 온다. 셋뿐인 것은 의도다 — 세션의 원문 답은 긴 쪽이고, 그것은
    // 사람이 「주고받은 내용」을 펼칠 때 읽는다.
    //
    // 그리고 이 칸이 없는 **구버전 백엔드**의 응답도 그대로 읽힌다: 없으면
    // null 이고, 화면은 전에 보여 주던 것으로 돌아간다(질문을 잃지 않는다).
    @Test
    fun theChoiceTurnsWordsArriveWithTheListAndAnOlderBackendStillParses() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = """{"id":"t1","title":"작업","status":"review","turns_total":4,""" +
                        """"waiting":{"turn_id":9,"request":"이거 해 줘","question":"어디까지?"},""" +
                        """"choices":[{"label":"전부","request_text":"전부 해 줘"}],"choice_turn_id":9,""" +
                        """"choice_turn":{"request":"리뷰 반영해 줘","explanation":"필수 3건 반영했어요.",""" +
                        """"question":"남은 2건은요?","response":"백엔드가 보내지 않는 긴 원문"},""" +
                        // 같은 응답의 다른 작업: 이 칸이 없는 작업은 null 이다.
                        """"migrated":false},{"id":"t2","title":"선택지 없음","status":"running"}""",
                ),
            ),
        )

        vm.loadTasks()

        val offered = vm.tasks.value.tasks.first { it.id == "t1" }
        assertEquals("리뷰 반영해 줘", offered.choiceTurn?.request)
        assertEquals("필수 3건 반영했어요.", offered.choiceTurn?.explanation)
        assertEquals("남은 2건은요?", offered.choiceTurn?.question)
        assertTrue(offered.choiceTurn!!.hasAny)
        // 그 턴의 열린 요청(waiting)과는 다른 칸이다 — 서로를 대신하지 않는다.
        assertEquals("이거 해 줘", offered.waiting?.request)
        // 선택지를 올리지 않은 작업은 올릴 말도 없다.
        assertNull(vm.tasks.value.tasks.first { it.id == "t2" }.choiceTurn)
    }

    // 세 칸이 모두 비어 있으면 그릴 것이 없다 — 빈 틀을 띄우지 않기 위해
    // 화면이 묻는 바로 그 값이다.
    @Test
    fun aChoiceTurnWithNothingInItHasNothingToShow() {
        assertFalse(ChatTaskChoiceTurn().hasAny)
        assertFalse(ChatTaskChoiceTurn(request = "   ").hasAny)
        assertTrue(ChatTaskChoiceTurn(question = "되나요?").hasAny)
    }

    // An empty ledger is its own state: the session really has no tasks.
    @Test
    fun anEmptyLedgerIsTheEmptyState() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(tasksJson()))

        vm.loadTasks()

        val state = vm.tasks.value
        assertTrue(state.isEmpty)
        assertNull(state.error)
        assertFalse(state.truncated)
    }

    // Before the first answer the panel is loading — not empty.
    @Test
    fun beforeTheFirstAnswerThePanelIsLoadingNotEmpty() {
        val state = ChatTasksUiState()
        assertTrue(state.isLoading)
        assertFalse(state.isEmpty)
    }

    // A prefix says so, so a partial list cannot read as the whole ledger.
    @Test
    fun aTruncatedWalkIsCarriedToThePanel() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(tasksJson(tasks = task("t1", "running"), truncated = true)))

        vm.loadTasks()

        assertTrue(vm.tasks.value.truncated)
    }

    // "Cannot read" is never "no tasks": each backend status gets its own
    // sentence, and the stale list is cleared rather than left looking current.
    @Test
    fun unreadableLedgersAreTheirOwnStatesAndClearTheList() = runTest {
        for ((status, want) in listOf(
            501 to ChatTaskError.UNSUPPORTED,
            409 to ChatTaskError.UNAVAILABLE,
            500 to ChatTaskError.UNREACHABLE,
        )) {
            val vm = ChatViewModel(api())
            server.enqueue(MockResponse().setBody(tasksJson(tasks = task("t1", "running"))))
            vm.loadTasks()
            server.takeRequest()
            assertEquals(1, vm.tasks.value.tasks.size)

            server.enqueue(MockResponse().setResponseCode(status).setBody("""{"code":"x","error":"no"}"""))
            vm.loadTasks()
            server.takeRequest()

            val state = vm.tasks.value
            assertEquals("status $status", want, state.error)
            assertTrue("status $status must not leave a list on screen", state.tasks.isEmpty())
            assertFalse("status $status is not the empty state", state.isEmpty)
        }
    }

    // After the conversation moves to another session the panel shows that
    // session's rows and names it — two sessions' tasks are never mixed.
    @Test
    fun aSwitchedSessionReplacesTheListOutright() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(tasksJson(tasks = task("old", "running"), sessionId = "cx1")))
        vm.loadTasks()
        server.takeRequest()
        assertEquals(listOf("old"), vm.tasks.value.tasks.map { it.id })

        server.enqueue(MockResponse().setBody(tasksJson(tasks = task("new", "planned"), sessionId = "cx2")))
        vm.loadTasks()
        server.takeRequest()

        val state = vm.tasks.value
        assertEquals("cx2", state.sessionId)
        assertEquals(listOf("new"), state.tasks.map { it.id })
    }

    // A status the app does not know is shown as unknown rather than guessed.
    @Test
    fun anUnknownStatusIsLabelledNotGuessed() {
        assertEquals("차단", ChatTaskStatus.label(ChatTaskStatus.BLOCKED))
        assertEquals("검토", ChatTaskStatus.label(ChatTaskStatus.REVIEW))
        assertEquals("알 수 없음", ChatTaskStatus.label("deploying"))
        assertTrue(ChatTaskStatus.needsAttention(ChatTaskStatus.BLOCKED))
        assertTrue(ChatTaskStatus.needsAttention(ChatTaskStatus.REVIEW))
        assertFalse(ChatTaskStatus.needsAttention(ChatTaskStatus.RUNNING))
        assertFalse(ChatTaskStatus.needsAttention(ChatTaskStatus.COMPLETED))
    }

    // Display order puts what needs a person first.
    @Test
    fun displayOrderPutsAttentionFirst() {
        assertEquals(
            listOf("blocked", "review", "running", "planned", "completed", "unknown"),
            ChatTaskStatus.order,
        )
    }
}
