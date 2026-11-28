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

// 작업 패널에서 한 작업을 골라 그 작업의 원장에 요청을 남기는 길. 보는 방식(칩
// 필터·완료 숨김)은 원장을 바꾸지 않고, 추천은 원장에 적힌 것만 인용하며,
// 저장된 요청은 대화로 가지 않는다.
//
// 저장 경로는 MockWebServer 로만 두드린다 — 실제 운영 세션으로는 아무것도
// 가지 않는다.
class ChatTaskRequestTest {
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
        choiceTurnId: Long = 0,
    ) = ChatTask(id = id, title = title, status = status, choiceTurnId = choiceTurnId)

    /** One work task as the backend serialises it, for the read-side tests. */
    private fun taskJson(id: String, status: String, extra: String = "") =
        """{"id":"$id","title":"작업 $id","status":"$status"$extra}"""

    /** The answer to a store: the turn the request became. */
    private fun storedJson(text: String, turnId: Long = 1, replayed: Boolean = false) =
        """{"turn":{"id":$turnId,"request":"$text","requested_at":"2026-11-20T00:00:00Z",""" +
            """"choices":[],"artifacts":[],"origin":"app_request"},"replayed":$replayed}"""

    private fun chatJson(pending: String = "null") =
        """{"conversation":{"id":7,"title":"대화","created_at":1,"updated_at":2},"messages":[],"pending":$pending}"""

    private fun tasksJson(tasks: String = "", sessionId: String = "cx1", counts: String = "{}") =
        """{"session_id":"$sessionId","tasks":[$tasks],"counts":$counts,"truncated":false}"""

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
        assertEquals(ChatTaskEmptiness.NO_TASKS, ChatTaskFilter().emptiness(emptyList()))
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

    // ── 저장하는 것: 사용자가 쓴 그 문장, 그 작업 ──────────────────────────

    // 자유 입력은 앞뒤 공백만 떼고 그대로 저장된다 — 앱이 문장을 덧붙이거나
    // 바꾸지 않는다. 작업은 경로가 가리키므로 본문에 붙는 머리줄이 없다.
    @Test
    fun theStoredTextIsTheUsersOwnWords() {
        assertEquals("이 작업 그냥 진행해 줘", chatTaskRequestText("  이 작업 그냥 진행해 줘  "))
        assertEquals("", chatTaskRequestText("   "))
    }

    // 요청은 그 작업의 원장 경로로 가고, 응답의 상태는 「저장됨 · 대기」다.
    // 대화에는 아무것도 올라가지 않고 답을 기다리는 상태로도 바뀌지 않는다.
    @Test
    fun aRequestIsStoredOnTheTasksLedgerAndNotSentToTheChat() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(chatJson()))
        vm.load()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘")))

        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "  진행해 줘  ", 7L))

        val request = server.takeRequest()
        assertEquals("/api/chat/tasks/t-42/requests", request.path)
        val body = request.body.readUtf8()
        assertTrue("요청 번호가 본문에 있어야 한다: $body", body.contains("app-1"))
        assertTrue(body.contains("진행해 줘"))
        assertFalse("작성자는 서버가 정한다 — 앱이 보내지 않는다", body.contains("author"))

        val stored = vm.taskRequest.value
        assertEquals("t-42", stored.taskId)
        // 답은 요청 행이 아니라 그 요청이 된 turn 이다. 그 turn 이 아직 끝나지
        // 않았다는 것이 「저장됐고 기다린다」이고, 앱이 따로 상태를 추론하지
        // 않는다 — 폴링할 생애주기가 이제 없다.
        assertEquals("진행해 줘", stored.stored?.request)
        assertEquals(ChatTaskTurn.ORIGIN_APP, stored.stored?.origin)
        assertFalse("끝나지 않은 turn 이다", stored.stored?.settled ?: true)
        assertFalse(stored.isStoring)
        assertFalse(stored.replayed)
        assertNull(stored.error)
        // 대화는 건드리지 않는다.
        assertTrue(vm.state.value.messages.isEmpty())
        assertFalse(vm.state.value.isReplying)
        assertNull(vm.state.value.error)
    }

    // 답이 오는 중에도 저장된다 — 요청은 턴으로 가지 않기 때문이다.
    @Test
    fun aRequestCanBeStoredWhileAReplyIsBeingWritten() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(chatJson(pending = """{"content":"쓰는 중","started_at":1}""")),
        )
        vm.load()
        server.takeRequest()
        assertTrue("답이 오는 중이어야 한다", vm.state.value.isReplying)

        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        assertEquals("/api/chat/tasks/t-42/requests", server.takeRequest().path)
        assertNull(vm.taskRequest.value.error)
    }

    // 같은 요청 번호로 다시 보내면 서버가 이미 저장된 그 요청을 돌려준다 —
    // 두 건이 되지 않고, 화면은 「이미 저장된 요청」이라고 말한다.
    @Test
    fun resendingTheSameRequestKeyStoresOneRequest() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(200).setBody(storedJson("진행해 줘", replayed = true)))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        val second = server.takeRequest()
        assertTrue("같은 번호로 가야 한다", second.body.readUtf8().contains("app-1"))

        val state = vm.taskRequest.value
        assertTrue("이미 저장된 요청임을 말해야 한다", state.replayed)
        assertNull(state.error)
    }

    // 가는 중인 저장을 또 눌러도 한 번만 간다.
    @Test
    fun aSecondTapWhileStoringSendsNothingTwice() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘"))
                .setBodyDelay(300, TimeUnit.MILLISECONDS),
        )
        val first = launch(start = CoroutineStart.UNDISPATCHED) {
            assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        }
        assertTrue("첫 저장이 가는 중이어야 한다", vm.taskRequest.value.isStoring)
        assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        first.join()
        assertEquals("저장은 한 번만 간다", 1, server.requestCount)
    }

    // 실패는 저장된 것으로 보이지 않는다: 이유를 말하고, 화면이 초안을 지울
    // 근거(참)를 주지 않는다 — 같은 번호로 다시 저장할 수 있다.
    @Test
    fun aRefusedRequestSaysWhyAndKeepsTheDraftAlive() = runTest {
        for ((status, body, want) in listOf(
            Triple(404, """{"code":"task_not_found"}""", ChatTaskRequestError.NOT_FOUND),
            Triple(409, """{"code":"task_request_unavailable"}""", ChatTaskRequestError.UNAVAILABLE),
            Triple(409, """{"code":"task_request_conflict"}""", ChatTaskRequestError.CONFLICT),
            Triple(501, """{"code":"task_request_unsupported"}""", ChatTaskRequestError.UNSUPPORTED),
            Triple(400, """{"code":"task_request_invalid"}""", ChatTaskRequestError.INVALID),
            Triple(502, """{"error":"platform"}""", ChatTaskRequestError.UNREACHABLE),
        )) {
            val vm = ChatViewModel(api())
            server.enqueue(MockResponse().setResponseCode(status).setBody(body))

            // false 는 화면이 초안과 요청 번호를 그대로 둔다는 뜻이다.
            assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
            server.takeRequest()

            val state = vm.taskRequest.value
            assertEquals("status $status", want, state.error)
            assertNull("실패가 저장된 것으로 보여선 안 된다", state.stored)
            assertFalse(state.isStoring)
            assertEquals("t-42", state.taskId)
        }
    }

    // 실패한 뒤 같은 번호로 다시 저장할 수 있다.
    @Test
    fun aFailedRequestCanBeRetriedUnderTheSameKey() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L))
        assertTrue(server.takeRequest().body.readUtf8().contains("app-1"))
        assertNull(vm.taskRequest.value.error)
        assertFalse("끝나지 않은 turn 이다", vm.taskRequest.value.stored?.settled ?: true)
    }

    // 빈 요청과 너무 긴 요청은 서버까지 가지 않는다.
    @Test
    fun anEmptyOrOversizedRequestNeverLeavesTheApp() = runTest {
        val vm = ChatViewModel(api())
        assertFalse(vm.performTaskRequest(task(), "app-1", "   ", 0L))
        assertNull(vm.taskRequest.value.error)

        assertFalse(
            vm.performTaskRequest(task(id = "t-42"), "app-1", "가".repeat(ChatViewModel.CHAT_TASK_REQUEST_MAX_BYTES), 0L),
        )
        assertEquals(ChatTaskRequestError.INVALID, vm.taskRequest.value.error)
        assertEquals(0, server.requestCount)
    }

    // 원장이 세는 대기 건수는 작업 목록으로 함께 온다 — 없으면 아예 없다.
    @Test
    fun theOpenTurnArrivesWithTheTaskList() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = taskJson(
                        "t1", "blocked",
                        extra = ""","waiting":{"turn_id":7,"request":"이거 해 줘","question":"어디까지?"},""" +
                            """"turns_total":3""",
                    ),
                    counts = """{"blocked":1}""",
                ),
            ),
        )

        vm.loadTasks()

        // 대기는 작업 옆의 행 집계가 아니라 그 작업의 열린 turn 이다.
        val waiting = vm.tasks.value.tasks.single().waiting
        assertEquals(7L, waiting?.turnId)
        assertEquals("어디까지?", waiting?.question)
        assertFalse("아무도 집지 않았다", waiting?.running ?: true)
        assertEquals(3L, vm.tasks.value.tasks.single().turnsTotal)
    }

    // 끝난 turn 과 아직 열린 turn 은 다르다 — 이것이 「처리됐다」와 「저장됐고
    // 기다린다」를 가르는 유일한 사실이고, 앱은 그 외의 상태를 만들지 않는다.
    @Test
    fun anEndedTurnIsSettledAndAnOpenOneIsNot() {
        assertTrue(ChatTaskTurnRef(id = 1, endedAt = "2026-11-20T01:00:00Z").settled)
        assertFalse(ChatTaskTurnRef(id = 1).settled)
    }

    // 결과를 보여 준 뒤에는 치운다 — 다음 작업의 화면에 지난 결과가 남지 않는다.
    @Test
    fun theOutcomeIsClearedOnceShown() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 0L)
        server.takeRequest()
        assertEquals(ChatTaskRequestError.UNREACHABLE, vm.taskRequest.value.error)

        vm.dismissTaskRequest()

        assertNull(vm.taskRequest.value.error)
        assertNull(vm.taskRequest.value.stored)
    }

    // 고른 선택지는 **그것이 제시된 turn** 과 함께 간다 — 서버가 그 turn 이
    // 아직 작업의 최신인지로 낡은 화면의 제출을 거부한다. (choice_id,
    // basis_version) 짝을 대체했다: 픽을 현재로 만드는 사실은 하나뿐이고, 두
    // 가지로 적으면 서로 어긋날 수 있었다.
    @Test
    fun theChosenChoiceCarriesTheTurnItWasOfferedOn() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 7L))
        val body = server.takeRequest().body.readUtf8()
        assertTrue("픽의 turn 이 본문에 있어야 한다: $body", body.contains("\"choice_turn_id\":7"))
        assertFalse("옛 짝은 더 보내지 않는다: $body", body.contains("choice_id"))
        assertFalse("옛 짝은 더 보내지 않는다: $body", body.contains("choice_basis_version"))
    }

    // 자유 입력은 turn 을 달지 않는다 — 사람이 직접 쓴 말은 어느 화면에서 썼든
    // 낡을 수 없으므로 낡음으로 거부될 일도 없다.
    @Test
    fun freeTextCarriesNoTurn() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("제 말로 쓸게요")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "제 말로 쓸게요", 0L))
        val body = server.takeRequest().body.readUtf8()
        // 직렬화가 기본값도 싣기 때문에 칸은 있고 값이 0 이다. 0 은 플랫폼이
        // 「픽 없음」으로 읽는 값이므로, 없는 것과 같은 뜻이다 — 칸의 부재가
        // 아니라 값이 0 인 것을 단언한다.
        assertTrue("자유 입력은 turn 을 0 으로 보낸다: $body", body.contains("\"choice_turn_id\":0"))
    }

    // 작업이 움직인 뒤 열어 둔 화면이 제출하면 서버가 거부한다. 그때 앱은
    // **초안을 지우지 않고** 다른 거절과 구별되는 말을 한다.
    @Test
    fun aStalePickIsItsOwnRefusalAndKeepsTheDraft() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setResponseCode(409)
                .setBody("""{"error":"그 선택지는 더 이상 유효하지 않아요","code":"task_request_stale_choice"}"""),
        )
        assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "CI 로그를 확인해 줘.", 7L))
        assertEquals(ChatTaskRequestError.STALE_CHOICE, vm.taskRequest.value.error)
        // 저장된 것은 없다 — 저장 성공으로 오표시하지 않는다.
        assertNull(vm.taskRequest.value.stored)
        // 같은 409 라도 요청 번호 충돌과는 다른 말을 한다.
        assertTrue(
            ChatTaskRequestError.STALE_CHOICE.message != ChatTaskRequestError.CONFLICT.message,
        )
    }

    // 화면이 저장 버튼에 매단 그 경로 그대로: 낡은 선택지로 거부되면 저장
    // 콜백은 돌지 않고(초안을 지우는 쪽이 그쪽이다) 출처만 떨어뜨리는 콜백이
    // 돌고, 작업을 **다시 읽어** 지금 유효한 선택지를 가져온다.
    @Test
    fun aStalePickDropsTheProvenanceAndReReadsTheTask() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setResponseCode(409)
                .setBody("""{"error":"그 선택지는 더 이상 유효하지 않아요","code":"task_request_stale_choice"}"""),
        )
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = """{"id":"t-42","title":"작업","status":"blocked",""" +
                        """"choices":[{"label":"지금 유효한 선택","request_text":"지금 기준의 문장"}],""" +
                        """"choice_turn_id":9}""",
                ),
            ),
        )
        var storedRan = false
        var staleRan = false

        vm.runTaskRequest(
            task = task(id = "t-42"),
            requestID = "app-1",
            request = "CI 로그를 확인해 줘.",
            choiceTurnId = 7L,
            onStored = { storedRan = true },
            onChoiceStale = { staleRan = true },
        )

        assertFalse("저장되지 않았으므로 초안을 지우는 콜백은 돌지 않는다", storedRan)
        assertTrue("출처만 떨어뜨리는 콜백이 돈다", staleRan)
        assertEquals(ChatTaskRequestError.STALE_CHOICE, vm.taskRequest.value.error)
        assertEquals("/api/chat/tasks/t-42/requests", server.takeRequest().path)
        assertEquals("거부 뒤에는 작업을 다시 읽는다", "/api/chat/tasks", server.takeRequest().path)
        // 다시 읽은 작업은 지금 유효한 선택지와 그것이 달린 turn 을 들고 온다.
        val reread = vm.tasks.value.tasks.single()
        assertEquals("지금 유효한 선택", reread.choices.single().label)
        assertEquals(9L, reread.choiceTurnId)
    }

    // 저장에 성공하면 초안을 지우는 쪽만 돌고, 출처만 떨어뜨리는 쪽은 돌지
    // 않는다 — 성공과 거부가 같은 모습이 되지 않는다. 목록은 저장 뒤에 다시
    // 읽는다(원장이 움직이므로 — 아래 접수 테스트들).
    @Test
    fun aStoredRequestRunsOnlyTheStoredCallback() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("진행해 줘")))
        server.enqueue(MockResponse().setBody(tasksJson(tasks = taskJson("t-42", "planned"))))
        var storedRan = false
        var staleRan = false

        vm.runTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", 2L, { storedRan = true }, { staleRan = true })

        assertTrue(storedRan)
        assertFalse(staleRan)
        assertEquals("저장 하나와 그 뒤의 재조회 하나", 2, server.requestCount)
        assertEquals("/api/chat/tasks/t-42/requests", server.takeRequest().path)
        assertEquals("/api/chat/tasks", server.takeRequest().path)
    }

    // ── 접수: 검토를 기다리던 작업이 「예정」으로 보인다 ──────────────────────

    // 사용자가 검토 상태의 작업에 요청을 보내면, 플랫폼이 그것을 저장한 같은
    // 커밋에서 그 작업을 「예정」으로 접수한다. 앱이 하는 일은 **다시 읽는
    // 것**이다 — 상태를 스스로 고치지 않으므로 화면에 뜨는 값은 언제나 서버가
    // 답한 값이다.
    @Test
    fun aRequestSentFromReviewShowsThePlannedStatusTheServerAnswers() = runTest {
        val vm = ChatViewModel(api())
        // 먼저 패널에 검토 작업이 하나 있다.
        server.enqueue(MockResponse().setBody(tasksJson(tasks = taskJson("t-42", "review"))))
        vm.loadTasks()
        assertEquals(ChatTaskStatus.REVIEW, vm.tasks.value.tasks.single().status)
        assertEquals("검토", ChatTaskStatus.label(vm.tasks.value.tasks.single().status))

        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("이 선택지로 해 주세요")))
        server.enqueue(
            MockResponse().setBody(
                tasksJson(tasks = taskJson("t-42", "planned"), counts = """{"planned":1}"""),
            ),
        )
        vm.runTaskRequest(task(id = "t-42"), "app-1", "이 선택지로 해 주세요", 0L, {})

        val received = vm.tasks.value.tasks.single()
        assertEquals(ChatTaskStatus.PLANNED, received.status)
        assertEquals("예정", ChatTaskStatus.label(received.status))
        assertEquals(mapOf("planned" to 1), vm.tasks.value.counts)
        // 「예정」은 접수 대기이고 착수가 아니다 — 손댈 것으로 올려 세우지 않는다.
        assertFalse(ChatTaskStatus.needsAttention(received.status))
    }

    // 앱은 상태를 스스로 옮기지 않는다. 저장은 성공했지만 서버가 여전히
    // 「검토」를 답하면 화면도 「검토」다 — 플랫폼이 접수하지 않은 전이를 앱이
    // 지어내지 않는다.
    @Test
    fun theAppDoesNotMoveTheStatusItself() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("한마디 남길게요")))
        server.enqueue(MockResponse().setBody(tasksJson(tasks = taskJson("t-42", "review"))))

        vm.runTaskRequest(task(id = "t-42"), "app-1", "한마디 남길게요", 0L, {})

        assertEquals(ChatTaskStatus.REVIEW, vm.tasks.value.tasks.single().status)
        assertNull(vm.tasks.value.error)
        // 저장 자체는 성공으로 남는다 — 상태가 안 움직인 것과는 별개다.
        assertEquals("한마디 남길게요", vm.taskRequest.value.stored?.request)
    }

    // 늦은 응답이 앞선 상태를 덮지 않는다. 저장하는 사이에 모니터가 그 작업을
    // 인수해 「진행」·「완료」가 되었으면, 저장 뒤의 재조회는 그 값을 그대로
    // 보인다 — 「예정」으로 되돌리지 않는다.
    @Test
    fun aStoreDoesNotPullAnAlreadyTakenTaskBackToPlanned() = runTest {
        for ((status, label) in listOf("running" to "진행", "completed" to "완료")) {
            val vm = ChatViewModel(api())
            server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("이어서 해 주세요")))
            server.enqueue(MockResponse().setBody(tasksJson(tasks = taskJson("t-42", status))))

            vm.runTaskRequest(task(id = "t-42"), "app-1", "이어서 해 주세요", 0L, {})

            val shown = vm.tasks.value.tasks.single()
            assertEquals(status, shown.status)
            assertEquals(label, ChatTaskStatus.label(shown.status))
        }
    }

    // 저장 뒤의 재조회가 실패해도 저장은 저장이다 — 앱이 상태를 지어내지
    // 않으므로 목록은 읽을 수 없다고 말하고, 요청이 저장됐다는 사실은 남는다.
    @Test
    fun aFailedReReadDoesNotInventAStatusNorUnsayTheStore() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("이어서 해 주세요")))
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        var storedRan = false

        vm.runTaskRequest(task(id = "t-42"), "app-1", "이어서 해 주세요", 0L, { storedRan = true })

        assertTrue("저장은 성공했다", storedRan)
        assertEquals("이어서 해 주세요", vm.taskRequest.value.stored?.request)
        assertNull("저장이 실패로 뒤집히지 않는다", vm.taskRequest.value.error)
        // 읽을 수 없는 목록을 「예정」이 된 목록으로 꾸미지 않는다.
        assertTrue(vm.tasks.value.tasks.isEmpty())
        assertEquals(ChatTaskError.UNREACHABLE, vm.tasks.value.error)
    }

    // 선택지로 채운 초안을 사람이 **고치면** 그것은 더 이상 세션의 문장이
    // 아니다 — 출처와 기준이 함께 떨어지고 자유 입력으로 저장된다. 한 글자도
    // 고치지 않았다면 출처는 유지된다.
    @Test
    fun editingAFilledDraftTurnsItIntoTheUsersOwnWords() {
        val sentence = "실패한 CI 로그를 확인해 줘."
        assertTrue(chatTaskPickHolds(sentence, sentence))
        assertFalse(chatTaskPickHolds(sentence + " 그리고 원인도 알려 줘", sentence))
        assertFalse(chatTaskPickHolds(sentence.dropLast(1), sentence))
        // 처음부터 직접 쓴 초안에는 출처가 없다.
        assertFalse(chatTaskPickHolds("내가 쓴 요청", ""))
    }

    // 선택지는 세션이 준 것만 쓴다 — 앱에는 상태별 목록이 없다.
    @Test
    fun theOfferedChoicesComeFromTheSessionOnly() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = """{"id":"t1","title":"작업","status":"blocked",""" +
                        """"choices":[{"label":"CI 로그 확인","request_text":"CI 로그를 확인해 줘."}],""" +
                        """"choice_turn_id":4}""",
                    counts = """{"blocked":1}""",
                ),
            ),
        )

        vm.loadTasks()

        val task = vm.tasks.value.tasks.single()
        assertEquals(1, task.choices.size)
        assertEquals("CI 로그를 확인해 줘.", task.choices[0].requestText)
        assertEquals("선택지는 제시된 turn 을 들고 온다", 4L, task.choiceTurnId)
        // 선택지를 주지 않은 작업은 비어 있고 turn 도 달지 않는다 — 앱이 대신
        // 만들지 않고, 아무것도 제시하지 않은 turn 으로 픽을 보내지도 않는다.
        server.enqueue(MockResponse().setBody(tasksJson(tasks = taskJson("t2", "running"))))
        vm.loadTasks()
        assertTrue(vm.tasks.value.tasks.single().choices.isEmpty())
        assertEquals(0L, vm.tasks.value.tasks.single().choiceTurnId)
    }
}
