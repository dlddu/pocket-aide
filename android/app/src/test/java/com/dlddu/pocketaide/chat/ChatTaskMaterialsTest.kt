package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotEquals
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
         "tradeoffs":[
           {"alternative":"백엔드에서 접기","pros":["모바일이 이력을 받지 않는다","작업당 1회로 끝난다"],
            "cons":["서버 메모리를 쓴다"],"turn_id":1,"source":"response"},
           {"alternative":"앱에서 접기","pros":["서버 상태가 없다"],
            "cons":[],"turn_id":1,"source":"response"}
         ],
         "decisions":[
           {"text":"접기는 백엔드에서 한다","turn_id":1,"source":"explanation"}
         ],
         "open":[
           {"turn_id":3,"question":"어느 쪽을 고칠까요?",
            "choices":[{"label":"앱만","request_text":"앱만 고쳐 주세요"}]}
         ],
         "tradeoffs_recorded":true,"decisions_recorded":true,
         "folded_through":3,"truncated":false}
    """.trimIndent()

    @Test
    fun thePullRequestsOfEarlierTurnsArriveWithTheTasksMaterials() = runTest {
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
    fun aRecordedTitleIsShownAndAMissingOneFallsBackToTheIdentifier() = runTest {
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
    fun aRefWithNoAddressIsNotATapAndNamesNoRepository() = runTest {
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
    fun decidedAndStillOpenAreDifferentLists() = runTest {
        server.enqueue(json(reviewed))
        val materials = api().chatTaskMaterials("t1")

        assertTrue(materials.decisionsRecorded)
        assertEquals(1, materials.decisions.size)
        val decided = materials.decisions.single()
        assertEquals("접기는 백엔드에서 한다", decided.text)
        assertEquals(1L, decided.turnId)
        assertEquals("explanation", decided.source)

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
    }

    // The positive case as the app reads it: two alternatives, each with the
    // sides their author wrote, comparable side by side and each naming the turn
    // it came from.
    @Test
    fun twoAlternativesArriveWithTheirOwnSidesAndTheirSourceTurn() = runTest {
        server.enqueue(json(reviewed))
        val materials = api().chatTaskMaterials("t1")

        assertTrue(materials.tradeoffsRecorded)
        assertEquals(2, materials.tradeoffs.size)
        val backend = materials.tradeoffs.first { it.alternative == "백엔드에서 접기" }
        val app = materials.tradeoffs.first { it.alternative == "앱에서 접기" }

        assertEquals(listOf("모바일이 이력을 받지 않는다", "작업당 1회로 끝난다"), backend.pros)
        assertEquals(listOf("서버 메모리를 쓴다"), backend.cons)
        assertTrue(backend.bothSides)

        // The second one's drawback was never written, so that side stays empty
        // and the screen says 미기록 for it rather than showing a blank that
        // reads as "no drawbacks".
        assertEquals(listOf("서버 상태가 없다"), app.pros)
        assertTrue(app.cons.isEmpty())
        assertFalse(app.bothSides)

        // Different on both sides — that is what makes them worth comparing.
        assertNotEquals(backend.pros, app.pros)
        // And both say where they were recorded.
        for (entry in materials.tradeoffs) {
            assertEquals(1L, entry.turnId)
            assertEquals("response", entry.source)
        }
    }

    // The negative case: a backend that found nothing marked sends empty lists
    // and both flags false, and the app reads that as "not recorded" rather
    // than inventing anything from prose or buttons.
    @Test
    fun nothingMarkedReadsAsUnrecordedRatherThanAsNone() = runTest {
        server.enqueue(
            json(
                """{"task_id":"t1","session_id":"worker-1","artifacts":[],
                   "tradeoffs":[],"decisions":[],
                   "open":[{"turn_id":1,"question":"A 로 할까요?",
                            "choices":[{"label":"A 로 간다 (빠름)","request_text":"A 로 해 주세요"}]}],
                   "tradeoffs_recorded":false,"decisions_recorded":false,
                   "folded_through":1,"truncated":false}""",
            ),
        )
        val materials = api().chatTaskMaterials("t1")

        assertFalse(materials.tradeoffsRecorded)
        assertFalse(materials.decisionsRecorded)
        assertTrue(materials.tradeoffs.isEmpty())
        assertTrue(materials.decisions.isEmpty())
        // The open question and its button are still there — a button is a
        // sentence to send, and it is not read as an advantage.
        assertEquals(1, materials.open.size)
        assertEquals("A 로 해 주세요", materials.open.single().choices.single().requestText)
        // There is still something to review, so the pane is not "empty".
        assertTrue(materials.hasAny)
    }

    @Test
    fun aTaskWithNothingRecordedReadsAsEmptyRatherThanAsUnread() = runTest {
        server.enqueue(
            json(
                """{"task_id":"t9","session_id":"worker-1","artifacts":[],
                   "tradeoffs":[],"decisions":[],"open":[],
                   "tradeoffs_recorded":false,"decisions_recorded":false,
                   "folded_through":0,"truncated":false}""",
            ),
        )
        val materials = api().chatTaskMaterials("t9")

        assertFalse(materials.hasAny)
        assertTrue(materials.artifacts.isEmpty())
        assertTrue(materials.tradeoffs.isEmpty())
        assertTrue(materials.decisions.isEmpty())
        assertTrue(materials.open.isEmpty())
    }

    // The four refusals stay four different sentences, and none of them says
    // the task produced nothing.
    @Test
    fun eachRefusalKeepsItsOwnSentence() {
        assertEquals(
            ChatTaskMaterialsError.NOT_FOUND,
            ChatTaskMaterialsError.from(
                ApiException.BadStatus(404, ""),
            ),
        )
        assertEquals(
            ChatTaskMaterialsError.UNAVAILABLE,
            ChatTaskMaterialsError.from(
                ApiException.BadStatus(409, ""),
            ),
        )
        assertEquals(
            ChatTaskMaterialsError.UNSUPPORTED,
            ChatTaskMaterialsError.from(
                ApiException.BadStatus(501, ""),
            ),
        )
        assertEquals(
            ChatTaskMaterialsError.UNREACHABLE,
            ChatTaskMaterialsError.from(
                ApiException.BadStatus(500, ""),
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
    fun unknownAndMissingFieldsBothDecode() = runTest {
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
