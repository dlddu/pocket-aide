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

    private fun task(
        id: String,
        status: String,
        stage: String = "implementation",
        reason: String = "",
        next: String = "",
    ) = """{"id":"$id","title":"작업 $id","status":"$status","stage":"$stage","reason":"$reason","next_action":"$next"}"""

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
                    tasks = task("t1", "blocked", "review", "CI 가 막혔다", "로그를 본다"),
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
        assertEquals("review", task.stage)
        assertEquals("CI 가 막혔다", task.reason)
        assertEquals("로그를 본다", task.nextAction)
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
