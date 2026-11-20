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

// The agreed records as they arrive in the app. The old ledger shapes could not
// carry a description, a completion criterion, or anything a turn holds, so
// these are the assertions that the detail pane can show the work rather than
// only that work exists.
class ChatTaskTurnsTest {
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

    private fun json(body: String) = MockResponse().setResponseCode(200)
        .setHeader("Content-Type", "application/json").setBody(body)

    // A task carried over from the ledger, with two turns: the first finished,
    // the second asking a question with a choice.
    private val migrated = """
        {"task_id":"t1","task_version":3,
         "task":{"id":"t1","title":"옮겨 온 작업","description":null,
                 "completion_criteria":null,"status":"blocked","version":3,
                 "session_id":"worker-now","created_at":null,"updated_at":null},
         "ledger_history":true,
         "turns":[
           {"id":1,"request":"이거 먼저 해 줘","instruction":"저장소를 읽고 패치를 만들어라",
            "response":"세션이 스스로 쓴 원문 답","explanation":"부모가 사람에게 쓴 설명",
            "question":null,"choices":[],"artifacts":["pull/42"],
            "requested_at":"2026-11-19T00:30:00Z","started_at":"2026-11-19T00:31:00Z",
            "ended_at":"2026-11-19T00:40:00Z","origin":"user_request"},
           {"id":2,"request":"리뷰 반영도 해 줘","instruction":null,"response":null,
            "explanation":null,"question":"어디까지 반영할까요?",
            "choices":[{"label":"전부","request_text":"코멘트 전부 반영해 줘"}],
            "artifacts":[],"requested_at":"2026-11-19T01:00:00Z",
            "started_at":null,"ended_at":null,"origin":"app_request"}]}
    """.trimIndent()

    // A task created after the cutover: the agreed fields it did record, and no
    // turns yet.
    private val fresh = """
        {"task_id":"t2","task_version":1,
         "task":{"id":"t2","title":"새 작업","description":"무엇을 why",
                 "completion_criteria":"머지되면 끝","status":"planned","version":1,
                 "session_id":"worker-1","created_at":"2026-11-19T02:00:00Z",
                 "updated_at":"2026-11-19T02:00:00Z"},
         "turns":[]}
    """.trimIndent()

    @Test
    fun everyAgreedFieldOfATurnReachesTheApp() = runTest {
        server.enqueue(json(migrated))
        val page = api().chatTaskTurns("t1", 0, 0)

        assertEquals("t1", page.taskId)
        assertEquals(3L, page.taskVersion)
        assertEquals("옮겨 온 작업", page.task.title)
        // A carried-over task never recorded these: null arrives as null, so
        // the pane says nothing is recorded instead of showing a blank.
        assertNull("description was invented", page.task.description)
        assertNull("completion criteria was invented", page.task.completionCriteria)

        assertEquals(2, page.turns.size)
        val one = page.turns[0]
        assertEquals("이거 먼저 해 줘", one.request)
        assertEquals("저장소를 읽고 패치를 만들어라", one.instruction)
        // The session's own words and the parent's summary stay distinct: a
        // pane that showed one as the other would put words in a session's
        // mouth.
        assertEquals("세션이 스스로 쓴 원문 답", one.response)
        assertEquals("부모가 사람에게 쓴 설명", one.explanation)
        assertEquals(listOf("pull/42"), one.artifacts)
        assertEquals("2026-11-19T00:31:00Z", one.startedAt)
        assertEquals("2026-11-19T00:40:00Z", one.endedAt)
        assertFalse("a finished turn is not open", one.isOpen)
        assertTrue("a migrated turn is marked as carried over", one.isCarriedOver)

        val two = page.turns[1]
        assertEquals("어디까지 반영할까요?", two.question)
        assertEquals(1, two.choices.size)
        assertEquals("코멘트 전부 반영해 줘", two.choices[0].requestText)
        assertTrue("an unanswered turn is still open", two.isOpen)
        assertFalse("a turn written here is not carried over", two.isCarriedOver)
        assertEquals(ChatTaskTurn.ORIGIN_APP, two.origin)
    }

    @Test
    fun aFreshTaskHasItsFieldsAndNoInventedHistory() = runTest {
        server.enqueue(json(fresh))
        val page = api().chatTaskTurns("t2", 0, 0)
        assertEquals("무엇을 why", page.task.description)
        assertEquals("머지되면 끝", page.task.completionCriteria)
        assertTrue("a fresh task has no turns yet", page.turns.isEmpty())
        // No ledger history to offer, and the pane must not imply one is
        // missing.
        assertFalse("a task created after the cutover claimed a history", page.ledgerHistory)
    }

    @Test
    fun theCursorAndPageSizeGoToThePlatform() = runTest {
        server.enqueue(json("""{"task_id":"t1","task_version":1,"task":{"id":"t1"},"turns":[],"next_cursor":7}"""))
        val page = api().chatTaskTurns("t1", 4, 2)
        assertEquals(7L, page.nextCursor)
        val asked = server.takeRequest().path ?: ""
        assertTrue(asked, asked.contains("after_turn_id=4"))
        assertTrue(asked, asked.contains("limit=2"))
        assertTrue(asked, asked.startsWith("/api/chat/tasks/t1/turns"))
    }

    // Only the latest turn's choices are current, which is why a pick is sent
    // as the next turn's request rather than against a turn already answered.
    @Test
    fun onlyTheLatestTurnOffersChoices() {
        val state = ChatTaskTurnsUiState(
            taskId = "t1",
            turns = listOf(
                ChatTaskTurn(id = 1, choices = listOf(ChatTaskTurnChoice("낡은", "낡은 문장"))),
                ChatTaskTurn(id = 2, choices = listOf(ChatTaskTurnChoice("지금", "지금 문장"))),
            ),
            hasLoaded = true,
        )
        assertEquals(1, state.offered.size)
        assertEquals("지금 문장", state.offered[0].requestText)
    }

    @Test
    fun aTaskWithNoTurnsIsEmptyRatherThanBroken() {
        val loaded = ChatTaskTurnsUiState(taskId = "t1", hasLoaded = true)
        assertTrue(loaded.isEmpty)
        assertFalse("nothing more to page", loaded.hasMore)
        // Before the first answer arrives it is not "empty", it is unknown.
        assertFalse(ChatTaskTurnsUiState(taskId = "t1").isEmpty)
    }

    @Test
    fun hasMoreFollowsTheCursor() {
        assertTrue(ChatTaskTurnsUiState(taskId = "t1", nextCursor = 9, hasLoaded = true).hasMore)
        assertFalse(ChatTaskTurnsUiState(taskId = "t1", nextCursor = 0, hasLoaded = true).hasMore)
    }

    // The platform's four answers are four different things, and none of them
    // is "this task has no work".
    @Test
    fun platformAnswersAreNotFlattenedIntoEmpty() = runTest {
        for ((status, want) in listOf(
            404 to ChatTaskHistoryError.NOT_FOUND,
            409 to ChatTaskHistoryError.UNAVAILABLE,
            501 to ChatTaskHistoryError.UNSUPPORTED,
            500 to ChatTaskHistoryError.UNREACHABLE,
        )) {
            server.enqueue(MockResponse().setResponseCode(status).setBody("""{"error":"x"}"""))
            val failed = runCatching { api().chatTaskTurns("t1", 0, 0) }.exceptionOrNull()
            assertTrue("status $status produced no error", failed is Exception)
            assertEquals("status $status", want, ChatTaskHistoryError.from(failed as Exception))
        }
    }
}
