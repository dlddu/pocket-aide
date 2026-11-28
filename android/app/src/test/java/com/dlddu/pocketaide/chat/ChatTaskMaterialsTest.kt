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

/**
 * The review materials as the app reads them, and what the pane is allowed to
 * conclude from them.
 *
 * The case these exist for is a task whose pull requests were produced several
 * turns ago and whose newest turn is a short follow-up: the materials must
 * still carry them, the decided and the still-open must stay apart, and the one
 * thing the records do not keep must read as unrecorded rather than as absent.
 */
class ChatTaskMaterialsTest {
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

    // Two pull requests from turn 1, one pick sent on turn 2, and turn 3 asking
    // the next question. One pull request has a recorded title and one does
    // not; one artifact is a bare identifier with no address at all.
    private val reviewed = """
        {"task_id":"t1","session_id":"worker-1",
         "artifacts":[
           {"ref":"https://github.com/dlddu/pocket-aide/pull/426",
            "url":"https://github.com/dlddu/pocket-aide/pull/426",
            "title":"","repo":"dlddu/pocket-aide","number":426,"turn_id":1},
           {"ref":"https://github.com/dlddu/pocket-aide/pull/425",
            "url":"https://github.com/dlddu/pocket-aide/pull/425",
            "title":"채팅: 주고받은 내용을 펼칠 때 읽는다",
            "repo":"dlddu/pocket-aide","number":425,"turn_id":1},
           {"ref":"pull/42","turn_id":1}
         ],
         "decisions":[
           {"turn_id":2,"offered_on_turn_id":1,"label":"둘 다 머지",
            "text":"둘 다 머지해 주세요"}
         ],
         "open":[
           {"turn_id":3,"question":"어느 쪽을 고칠까요?",
            "choices":[{"label":"앱만","request_text":"앱만 고쳐 주세요"}]}
         ],
         "tradeoffs_recorded":false,"folded_through":3,"truncated":false}
    """.trimIndent()

    @Test
    fun `the pull requests of earlier turns arrive with the task's materials`() = runTest {
        server.enqueue(json(reviewed))
        val materials = api().chatTaskMaterials("t1")

        assertEquals("/api/chat/tasks/t1/materials", server.takeRequest().path)
        assertEquals("t1", materials.taskId)
        assertEquals("worker-1", materials.sessionId)
        assertEquals(3, materials.artifacts.size)
        assertEquals(3L, materials.foldedThrough)
        assertTrue(materials.hasAny)
        // Produced on turn 1 even though turn 3 is the latest — the whole point.
        assertTrue(materials.artifacts.all { it.turnId == 1L })
    }

    @Test
    fun `a recorded title is shown and a missing one falls back to the identifier`() = runTest {
        server.enqueue(json(reviewed))
        val artifacts = api().chatTaskMaterials("t1").artifacts

        val titled = artifacts.first { it.number == 425 }
        assertTrue(titled.titled)
        assertEquals("채팅: 주고받은 내용을 펼칠 때 읽는다", titled.label)
        assertTrue(titled.openable)

        // No row named this one, so the line is the identifier rather than a
        // title made up from the number.
        val untitled = artifacts.first { it.number == 426 }
        assertFalse(untitled.titled)
        assertEquals("dlddu/pocket-aide#426", untitled.label)
        assertTrue(untitled.openable)
    }

    @Test
    fun `a ref with no address is not a tap and names no repository`() = runTest {
        server.enqueue(json(reviewed))
        val bare = api().chatTaskMaterials("t1").artifacts.first { it.ref == "pull/42" }

        assertFalse(bare.openable)
        assertFalse(bare.titled)
        assertEquals("", bare.repo)
        assertEquals(0, bare.number)
        // The line is the recorded identifier, verbatim.
        assertEquals("pull/42", bare.label)
    }

    @Test
    fun `decided and still-open are different lists, and trade-offs are unrecorded`() = runTest {
        server.enqueue(json(reviewed))
        val materials = api().chatTaskMaterials("t1")

        assertEquals(1, materials.decisions.size)
        val decided = materials.decisions.single()
        assertEquals("둘 다 머지", decided.label)
        assertEquals("둘 다 머지해 주세요", decided.text)
        assertEquals(2L, decided.turnId)
        assertEquals(1L, decided.offeredOnTurnId)

        assertEquals(1, materials.open.size)
        val open = materials.open.single()
        assertEquals(3L, open.turnId)
        assertEquals("어느 쪽을 고칠까요?", open.question)
        // The sentence to send travels verbatim, so picking it sends the words
        // that were read.
        assertEquals("앱만 고쳐 주세요", open.choices.single().requestText)

        // The decided turn is not the open turn: a question that was answered is
        // not offered again as something to decide.
        assertFalse(materials.open.any { it.turnId == decided.turnId })
        // The records keep no advantages and drawbacks, and the answer says so
        // rather than leaving a silence that reads as "there were none".
        assertFalse(materials.tradeoffsRecorded)
    }

    @Test
    fun `a task with nothing recorded reads as empty rather than as unread`() = runTest {
        server.enqueue(
            json(
                """{"task_id":"t9","session_id":"worker-1","artifacts":[],
                   "decisions":[],"open":[],"tradeoffs_recorded":false,
                   "folded_through":0,"truncated":false}""",
            ),
        )
        val materials = api().chatTaskMaterials("t9")

        assertFalse(materials.hasAny)
        assertTrue(materials.artifacts.isEmpty())
        assertTrue(materials.decisions.isEmpty())
        assertTrue(materials.open.isEmpty())
    }

    // The four refusals stay four different sentences, and none of them says
    // the task produced nothing.
    @Test
    fun `each refusal keeps its own sentence`() {
        assertEquals(
            ChatTaskMaterialsError.NOT_FOUND,
            ChatTaskMaterialsError.from(
                com.dlddu.pocketaide.api.ApiException.BadStatus(404, ""),
            ),
        )
        assertEquals(
            ChatTaskMaterialsError.UNAVAILABLE,
            ChatTaskMaterialsError.from(
                com.dlddu.pocketaide.api.ApiException.BadStatus(409, ""),
            ),
        )
        assertEquals(
            ChatTaskMaterialsError.UNSUPPORTED,
            ChatTaskMaterialsError.from(
                com.dlddu.pocketaide.api.ApiException.BadStatus(501, ""),
            ),
        )
        assertEquals(
            ChatTaskMaterialsError.UNREACHABLE,
            ChatTaskMaterialsError.from(
                com.dlddu.pocketaide.api.ApiException.BadStatus(500, ""),
            ),
        )
        // Four different sentences, because the person's next move differs in
        // each. Two of them reading the same would make the difference moot.
        val sentences = ChatTaskMaterialsError.entries.map { it.message }
        assertEquals(sentences.size, sentences.toSet().size)
    }

    // A field a newer backend adds must not stop an older app from reading the
    // ones it knows, and a backend that does not send this route's newer fields
    // at all must still decode.
    @Test
    fun `unknown and missing fields both decode`() = runTest {
        server.enqueue(
            json(
                """{"task_id":"t1","artifacts":[{"ref":"pull/1","turn_id":2,
                   "something_new":{"a":1}}],"brand_new":"x"}""",
            ),
        )
        val materials = api().chatTaskMaterials("t1")

        assertEquals("t1", materials.taskId)
        assertEquals(1, materials.artifacts.size)
        assertEquals("pull/1", materials.artifacts.single().ref)
        assertFalse(materials.artifacts.single().openable)
        assertTrue(materials.decisions.isEmpty())
        assertFalse(materials.tradeoffsRecorded)
        assertNull(materials.open.firstOrNull())
    }
}
