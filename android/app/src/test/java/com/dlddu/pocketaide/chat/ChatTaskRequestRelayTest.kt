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

// 요청 등록이 「모니터에게 전달됐는가」를 말하는 길. 세 가지는 끝까지 분리된다 —
// 저장됨 · 모니터가 받음 · 실제로 끝남 — 그리고 마지막 것은 여기서 말할 수 없다.
//
// 저장 경로는 MockWebServer 로만 두드린다: 실제 운영 세션으로는 아무것도 가지
// 않는다.
class ChatTaskRequestRelayTest {
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

    private fun storedJson(relay: String?) =
        """{"turn":{"id":1,"request":"네","requested_at":"2026-11-20T00:00:00Z",""" +
            """"choices":[],"artifacts":[],"origin":"app_request"},"replayed":false""" +
            (relay?.let { ""","relay":$it""" } ?: "") + "}"

    // 백엔드가 전달 결과를 실으면 앱 모델이 그대로 받는다. 실지 않으면 null 이고,
    // null 은 「알릴 것이 없었다」이며 실패가 아니다.
    @Test
    fun theHandoverIsReadWhenItIsThereAndAbsenceIsNotAFailure() = runTest {
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("""{"state":"accepted","attempts":1}""")))
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson(null)))

        val client = api()
        val handed = client.storeChatTaskRequest("t1", "req-1", "네", 0)
        assertEquals(ChatTaskRequestRelay.ACCEPTED, handed.relay?.state)
        assertTrue(handed.relay!!.accepted)
        assertFalse("첫 전달은 재전달이 아니다", handed.relay!!.redelivered)

        val ordinary = client.storeChatTaskRequest("t2", "req-2", "네", 0)
        assertNull("알릴 것이 없으면 전달 결과가 아예 없다", ordinary.relay)
        assertEquals(1L, ordinary.turn.id)
    }

    // 전달 결과의 세 문장. 어느 것도 「처리됐다」고 말하지 않는다 — 그것은 턴이
    // 끝났을 때만 참이고, 저장 응답은 그것을 알 수 없다.
    @Test
    fun theNoteSaysStoredAndHandedOverButNeverDone() {
        val accepted = chatTaskRequestNote(false, ChatTaskRequestRelay("accepted", 1))
        assertTrue(accepted.contains("요청 저장됨"))
        assertTrue(accepted.contains("모니터에 전달했어요"))
        assertTrue("전달과 처리 완료를 구별해야 한다", accepted.contains("처리됐다는 뜻은 아니에요"))

        val again = chatTaskRequestNote(false, ChatTaskRequestRelay("accepted", 2))
        assertTrue("두 번째 전달은 그렇게 말한다", again.contains("다시 전달했어요"))
        assertTrue(again.contains("회신이 유실"))

        val pending = chatTaskRequestNote(false, ChatTaskRequestRelay("pending", 1))
        assertTrue(pending.contains("아직 닿지 못했어요"))
        assertTrue("닿지 못한 것은 유실이 아니라 「다음 확인 때」다", pending.contains("다음 확인 때"))
        assertFalse("실패로 읽히는 말을 쓰지 않는다", pending.contains("실패"))

        // 알릴 것이 없었던 경우. 전에 쓰던 문장과 같은 뜻이고, 전달의 실패로
        // 읽히지 않는다.
        val none = chatTaskRequestNote(false, null)
        assertTrue(none.contains("다음 확인 때"))
        assertFalse(none.contains("전달"))

        // 재전송은 저장 쪽의 사실이고 전달 쪽과 섞이지 않는다.
        val replayed = chatTaskRequestNote(true, ChatTaskRequestRelay("accepted", 1))
        assertTrue(replayed.startsWith("이미 저장된 요청이에요"))
        assertTrue(replayed.contains("모니터에 전달했어요"))
    }

    // 모르는 낱말은 세 문장 중 어느 것도 주장하지 않는다. 백엔드가 이미 걷어내지만,
    // 앱 쪽 분기도 빈틈 없이 닫혀 있어야 한다.
    @Test
    fun aHandoverStateThisAppHasNoSentenceForClaimsNothing() {
        val relay = ChatTaskRequestRelay("half-delivered", 1)
        assertFalse(relay.accepted)
        assertFalse(relay.pending)
        val note = chatTaskRequestNote(false, relay)
        assertFalse(note.contains("전달"))
        assertTrue(note.contains("다음 확인 때"))
    }

    // 등록 전 안내는 그 작업이 무엇을 기다리는지에 따라 다르다 — 사람을 기다리던
    // 작업만 「바로 알려요」이고, 나머지는 오지 않을 회차를 약속하지 않는다.
    @Test
    fun theHintPromisesAnImmediateTellOnlyForTheTasksThatWaitOnAPerson() {
        for (status in listOf(ChatTaskStatus.REVIEW, ChatTaskStatus.BLOCKED)) {
            val hint = chatTaskRequestHint(status)
            assertTrue(status, hint.contains("바로 알려요"))
            assertTrue(status, hint.contains("대화로는 보내지 않아요"))
        }
        for (status in listOf(
            ChatTaskStatus.RUNNING, ChatTaskStatus.PLANNED,
            ChatTaskStatus.COMPLETED, ChatTaskStatus.UNKNOWN,
        )) {
            val hint = chatTaskRequestHint(status)
            assertFalse(status, hint.contains("바로 알려요"))
            assertTrue(status, hint.contains("다음 확인 때"))
        }
    }
}
