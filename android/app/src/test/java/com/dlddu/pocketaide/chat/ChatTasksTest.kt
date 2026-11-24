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

class ChatTaskMonitorNoteTest {
    @Test
    fun aLivePairSaysWhatTheMonitorIsDoing() {
        val note = monitorNote(ChatTaskMonitor(sessionId = "m1", state = "active", intakeLive = true), "slot", "")
        assertEquals("모니터 active", note)
    }

    @Test
    fun aPublishedListCarriesWhenItWasSaved() {
        val note = monitorNote(
            ChatTaskMonitor(sessionId = "m1", state = "snapshot", intakeLive = false),
            "published",
            "2026-11-22T01:00:00Z",
        )
        // The time the state was saved, and the fact a send will wait, are both
        // on screen — a frozen panel must not read as live.
        assertTrue(note, note.contains("저장 시각 2026-11-22T01:00:00Z"))
        assertTrue(note, note.contains("접수 후 대기"))
    }

    @Test
    fun anUnknownStateStillReadsAsASentence() {
        val note = monitorNote(ChatTaskMonitor(sessionId = "m1"), "slot", "")
        assertTrue(note, note.contains("상태 미확인"))
    }
}
