package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.CoroutineStart
import kotlinx.coroutines.launch
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
import java.util.concurrent.TimeUnit

// 작업 패널에서 한 작업을 골라 이 대화로 후속 요청을 보내는 길. 보는 방식(칩
// 필터·완료 숨김)은 원장을 바꾸지 않고, 보내는 것은 사용자가 읽은 그 문장이며,
// 원장에 적혀 있지 않은 사유나 다음 행동을 앱이 지어내지 않는지를 본다.
//
// 보내는 경로는 MockWebServer 로만 두드린다 — 실제 운영 세션으로는 아무것도
// 가지 않는다.
class ChatTaskFollowUpTest {
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
        id: String = "t-42",
        status: String = ChatTaskStatus.REVIEW,
        title: String = "승인 상태 붙이기",
        stage: String = "review",
        reason: String = "",
        next: String = "",
    ) = ChatTask(id = id, title = title, status = status, stage = stage, reason = reason, nextAction = next)

    private fun sentJson(content: String = "요청") =
        """{"message":{"id":9,"role":"user","content":"$content","created_at":1}}"""

    private fun chatJson() =
        """{"conversation":{"id":7,"title":"대화","created_at":1,"updated_at":2},"messages":[],"pending":null}"""

    // ── 보는 방식: 칩 필터와 완료 숨김 ───────────────────────────────────────

    private val ledger = listOf(
        task("t1", ChatTaskStatus.BLOCKED),
        task("t2", ChatTaskStatus.REVIEW),
        task("t3", ChatTaskStatus.RUNNING),
        task("t4", ChatTaskStatus.COMPLETED),
        task("t5", ChatTaskStatus.COMPLETED),
    )

    // 두 선택은 함께 걸린다 — 칩이 상태를 좁히고 완료 숨김이 그 안에서 또
    // 덜어낸다. 어느 쪽도 다른 쪽을 대신 풀어 주지 않는다.
    @Test
    fun theChipAndTheHideCombineInsteadOfReplacingEachOther() {
        assertEquals(
            listOf("t1", "t2", "t3", "t4", "t5"),
            ChatTaskFilter().select(ledger).map { it.id },
        )
        assertEquals(
            listOf("t2"),
            ChatTaskFilter(status = ChatTaskStatus.REVIEW).select(ledger).map { it.id },
        )
        assertEquals(
            listOf("t1", "t2", "t3"),
            ChatTaskFilter(hideCompleted = true).select(ledger).map { it.id },
        )
        // 검토 칩 + 완료 숨김: 검토는 완료가 아니므로 숨김이 더 덜어낼 것이 없다.
        assertEquals(
            listOf("t2"),
            ChatTaskFilter(ChatTaskStatus.REVIEW, hideCompleted = true).select(ledger).map { it.id },
        )
        // 완료 칩 + 완료 숨김: 서로 어긋나는 선택이고, 결과는 빈 목록이다.
        assertTrue(ChatTaskFilter(ChatTaskStatus.COMPLETED, hideCompleted = true).select(ledger).isEmpty())
    }

    // 고른 칩을 다시 누르면 전체로 돌아간다 — 완료 숨김은 그대로 남는다.
    @Test
    fun tappingTheChosenChipAgainGoesBackToAllAndLeavesTheHideAlone() {
        val chosen = ChatTaskFilter(hideCompleted = true).toggled(ChatTaskStatus.BLOCKED)
        assertEquals(ChatTaskStatus.BLOCKED, chosen.status)
        assertTrue(chosen.hideCompleted)

        val back = chosen.toggled(ChatTaskStatus.BLOCKED)
        assertNull(back.status)
        assertTrue("완료 숨김은 칩과 다른 선택이므로 함께 풀리지 않는다", back.hideCompleted)

        val other = chosen.toggled(ChatTaskStatus.REVIEW)
        assertEquals(ChatTaskStatus.REVIEW, other.status)
    }

    // 빈 목록의 세 가지 이유는 서로를 대신 말하지 않는다. 특히 숨겨서 비었을
    // 때는 일이 사라진 것이 아니라고 말해야 한다.
    @Test
    fun anEmptyListSaysWhichOfTheThreeReasonsItIs() {
        assertEquals(ChatTaskEmptiness.LEDGER, ChatTaskFilter().emptiness(emptyList()))
        assertNull("볼 것이 있으면 빈 이유는 없다", ChatTaskFilter().emptiness(ledger))

        // 원장에 없는 상태를 골랐다 — 숨김과는 무관하다.
        assertEquals(
            ChatTaskEmptiness.FILTERED,
            ChatTaskFilter(ChatTaskStatus.PLANNED).emptiness(ledger),
        )
        // 완료를 숨겨서 비었다. 완료 칩까지 함께 걸린 경우도 같은 이유다 —
        // 행은 원장에 그대로 있다.
        assertEquals(
            ChatTaskEmptiness.COMPLETED_HIDDEN,
            ChatTaskFilter(ChatTaskStatus.COMPLETED, hideCompleted = true).emptiness(ledger),
        )
        assertEquals(
            ChatTaskEmptiness.COMPLETED_HIDDEN,
            ChatTaskFilter(hideCompleted = true).emptiness(ledger.filter { it.status == ChatTaskStatus.COMPLETED }),
        )
        // 없는 상태 + 완료 숨김은 숨김 탓이 아니다.
        assertEquals(
            ChatTaskEmptiness.FILTERED,
            ChatTaskFilter(ChatTaskStatus.PLANNED, hideCompleted = true).emptiness(ledger),
        )
    }

    // ── 추천: 원장에 적힌 것만 ──────────────────────────────────────────────

    // 적힌 사유·다음 행동은 그대로 인용하고, 적혀 있지 않으면 지어내지 않고
    // 물어본다.
    @Test
    fun suggestionsQuoteTheLedgerAndAskWhenItIsSilent() {
        val recorded = chatTaskSuggestions(
            task(status = ChatTaskStatus.BLOCKED, reason = "CI 가 막혔다", next = "로그를 본다"),
        )
        assertTrue(
            "적힌 차단 사유를 그대로 인용해야 한다",
            recorded.any { it.request.contains("「CI 가 막혔다」") },
        )
        assertTrue(
            "적힌 다음 행동을 그대로 인용해야 한다",
            recorded.any { it.request.contains("「로그를 본다」") && it.label == "적힌 다음 행동 진행" },
        )

        val silent = chatTaskSuggestions(task(status = ChatTaskStatus.BLOCKED))
        assertTrue(
            "사유가 없으면 없다고 말하고 물어봐야 한다",
            silent.any { it.request.contains("차단 사유가 적혀 있지 않아") },
        )
        assertTrue(
            "다음 행동이 없으면 없다고 말하고 물어봐야 한다",
            silent.any { it.request.contains("다음 행동이 적혀 있지 않아") },
        )
        for (suggestion in silent) {
            assertFalse(
                "빈 인용부호는 원장에 없는 말을 적어 넣은 흔적이다: ${suggestion.request}",
                suggestion.request.contains("「」"),
            )
        }
    }

    // 끝난 일에는 다음 행동을 지어 주지 않는다 — 원장이 적어 둔 경우에만 있다.
    @Test
    fun finishedWorkGetsANextStepOnlyWhenTheLedgerRecordedOne() {
        val done = chatTaskSuggestions(task(status = ChatTaskStatus.COMPLETED))
        assertTrue(done.none { it.label.contains("다음 행동") })
        assertTrue(done.any { it.label == "결과 요약 받기" })

        val doneWithNext = chatTaskSuggestions(task(status = ChatTaskStatus.COMPLETED, next = "회고를 적는다"))
        assertTrue(doneWithNext.any { it.request.contains("「회고를 적는다」") })
    }

    // 앱이 모르는 상태에는 아무 상태도 가정하지 않고 무엇인지 묻는다.
    @Test
    fun anUnknownStatusIsAskedAboutRatherThanGuessed() {
        val unknown = chatTaskSuggestions(task(status = "deploying"))
        assertTrue(unknown.any { it.request.contains("상태 값을 앱이 알지 못해") })
        assertTrue(unknown.none { it.request.contains("진행해 줘") })
    }

    // 어떤 상태든 보낼 문장이 있고, 비어 있는 요청은 없다.
    @Test
    fun everyStatusHasSomethingToSendAndNoneOfItIsBlank() {
        for (status in ChatTaskStatus.order + listOf("deploying")) {
            val suggestions = chatTaskSuggestions(task(status = status))
            assertTrue("$status 에 보낼 문장이 없다", suggestions.isNotEmpty())
            for (suggestion in suggestions) {
                assertTrue("$status 의 라벨이 비었다", suggestion.label.isNotBlank())
                assertTrue("$status 의 요청문이 비었다", suggestion.request.isNotBlank())
            }
        }
    }

    // ── 보내는 것: 화면에서 읽은 그 문장, 그 작업 ID ────────────────────────

    // 자유 입력은 손대지 않고(앞뒤 공백만 떼고), 그 앞에 붙는 줄은 원장의 id·
    // 제목·상태를 그대로 적는다 — 답이 어느 작업에 대한 것인지가 메시지 안에
    // 남는다.
    @Test
    fun theMessageNamesTheExactTaskAndCarriesTheTypedWordsUntouched() {
        val message = chatTaskFollowUpMessage(
            task(id = "t-42", title = "승인 상태 붙이기", status = ChatTaskStatus.REVIEW, stage = "review"),
            "  이 작업 그냥 진행해 줘  ",
        )
        assertTrue("작업 id 가 메시지에 그대로 있어야 한다", message.contains("t-42"))
        assertTrue(message.contains("「승인 상태 붙이기」"))
        assertTrue("패널이 보여 준 상태를 적는다", message.contains("원장 상태 검토"))
        assertTrue(message.contains("단계 review"))
        assertTrue(message.endsWith("\n이 작업 그냥 진행해 줘"))
        assertFalse("자유 입력을 바꿔 보내지 않는다", message.contains("  이 작업"))

        // 화면이 인용해 보여 주는 줄이 실제로 붙어 나가는 줄과 같다 — 사용자가
        // 읽지 않은 말이 메시지에 들어가지 않는다.
        val quoted = chatTaskFollowUpMessage(task(), "").trim()
        assertTrue(chatTaskFollowUpMessage(task(), "무엇이든").startsWith(quoted))
    }

    // 제목이 비어 있는 행은 id 로 불린다 — 빈 「」 로 보내지 않는다.
    @Test
    fun aTitlelessRowIsNamedByItsId() {
        val message = chatTaskFollowUpMessage(task(id = "t-7", title = "  "), "확인해 줘")
        assertTrue(message.contains("「t-7」"))
        assertFalse(message.contains("「」"))
    }

    // 단계가 적혀 있지 않으면 단계 칸을 지어내지 않는다.
    @Test
    fun anUnrecordedStageIsLeftOutRatherThanInvented() {
        val message = chatTaskFollowUpMessage(task(stage = ""), "확인해 줘")
        assertFalse(message.contains("단계"))
        assertTrue(message.contains("원장 상태 검토"))
    }

    // 보낸 요청은 이 대화의 메시지로 올라가고, 다음 답을 기다리는 상태가 된다.
    @Test
    fun aFollowUpGoesOutAsAMessageInThisConversation() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(202).setBody(sentJson()))

        assertTrue(vm.performTaskFollowUp(task(id = "t-42"), "  진행해 줘  "))

        val request = server.takeRequest()
        assertEquals("/api/chat/messages", request.path)
        val body = request.body.readUtf8()
        assertTrue("보낸 본문이 그 작업 id 를 담아야 한다", body.contains("t-42"))
        assertTrue(body.contains("진행해 줘"))

        val followUp = vm.followUp.value
        assertEquals("t-42", followUp.taskId)
        assertTrue(followUp.accepted)
        assertFalse(followUp.isSending)
        assertNull(followUp.error)
        // 왼쪽 대화에 올라가고 답을 기다린다.
        assertEquals(1, vm.state.value.messages.size)
        assertTrue(vm.state.value.isReplying)
        // 대화 자신의 오류 칸은 후속 요청이 건드리지 않는다.
        assertNull(vm.state.value.error)
    }

    // 같은 요청을 두 번 누른다고 두 번 가지 않는다 — 첫 요청이 가는 동안의
    // 두 번째 탭은 보내지 않고 거절된다.
    @Test
    fun aSecondTapWhileTheFirstIsInFlightSendsNothingTwice() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setResponseCode(202).setBody(sentJson()).setBodyDelay(300, TimeUnit.MILLISECONDS),
        )

        val first = launch(start = CoroutineStart.UNDISPATCHED) {
            assertTrue(vm.performTaskFollowUp(task(id = "t-42"), "진행해 줘"))
        }
        assertTrue("첫 요청이 가는 중이어야 한다", vm.followUp.value.isSending)

        assertFalse(
            "가는 중인 요청과 같은 것을 또 보내지 않는다",
            vm.performTaskFollowUp(task(id = "t-42"), "진행해 줘"),
        )

        first.join()
        assertEquals("요청은 한 번만 간다", 1, server.requestCount)
        assertTrue(vm.followUp.value.accepted)
        assertEquals(1, vm.state.value.messages.size)
    }

    // 실패는 보낸 것으로 보이지 않는다: 이유를 말하고, 대화에는 아무것도
    // 올리지 않고, 화면이 초안을 지울 근거(참)를 주지 않는다.
    @Test
    fun aRefusedFollowUpSaysSoAndKeepsTheDraftAlive() = runTest {
        for ((status, want) in listOf(
            409 to ChatError.BUSY,
            502 to ChatError.UNREACHABLE,
            500 to ChatError.FAILED,
        )) {
            val vm = ChatViewModel(api())
            server.enqueue(MockResponse().setResponseCode(status).setBody("""{"error":"no"}"""))

            // false 는 화면이 초안을 지우지 않는다는 뜻이다 — 지우기는 참일
            // 때만 하는 콜백에 묶여 있다.
            assertFalse(vm.performTaskFollowUp(task(id = "t-42"), "진행해 줘"))
            server.takeRequest()

            val followUp = vm.followUp.value
            assertEquals("status $status", want, followUp.error)
            assertFalse("실패가 보낸 것으로 보여선 안 된다", followUp.accepted)
            assertFalse(followUp.isSending)
            assertEquals("t-42", followUp.taskId)
            assertTrue("실패했으면 대화에 올라간 것이 없다", vm.state.value.messages.isEmpty())
        }
    }

    // 다시 보낼 수 있다 — 실패한 뒤의 두 번째 시도는 막히지 않는다.
    @Test
    fun aFailedFollowUpCanBeRetried() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        assertFalse(vm.performTaskFollowUp(task(id = "t-42"), "진행해 줘"))
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(202).setBody(sentJson()))
        assertTrue(vm.performTaskFollowUp(task(id = "t-42"), "진행해 줘"))
        server.takeRequest()

        assertTrue(vm.followUp.value.accepted)
        assertNull(vm.followUp.value.error)
        assertEquals(1, vm.state.value.messages.size)
    }

    // 빈 요청과 너무 긴 요청은 서버까지 가지 않는다. 길이 판정은 실제로 보내는
    // 문장(붙는 줄 포함) 기준이다.
    @Test
    fun anEmptyOrOversizedRequestNeverLeavesTheApp() = runTest {
        val vm = ChatViewModel(api())

        assertFalse(vm.performTaskFollowUp(task(), "   "))
        assertNull(vm.followUp.value.error)

        assertFalse(vm.performTaskFollowUp(task(id = "t-42"), "가".repeat(CHAT_MAX_MESSAGE_BYTES)))
        assertEquals(ChatError.TOO_LARGE, vm.followUp.value.error)
        assertEquals("t-42", vm.followUp.value.taskId)
        assertEquals("보내지 않았다", 0, server.requestCount)
    }

    // 결과를 보여 준 뒤에는 치운다 — 다음 작업의 화면에 지난 결과가 남지 않는다.
    @Test
    fun theOutcomeIsClearedOnceShown() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        vm.performTaskFollowUp(task(id = "t-42"), "진행해 줘")
        server.takeRequest()
        assertEquals(ChatError.UNREACHABLE, vm.followUp.value.error)

        vm.dismissFollowUp()

        assertNull(vm.followUp.value.error)
        assertFalse(vm.followUp.value.accepted)
    }
}
