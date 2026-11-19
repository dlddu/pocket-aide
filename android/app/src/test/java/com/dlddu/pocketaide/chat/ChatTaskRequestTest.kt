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
        stage: String = "review",
        reason: String = "",
        next: String = "",
    ) = ChatTask(id = id, title = title, status = status, stage = stage, reason = reason, nextAction = next)

    /** One ledger row as the backend serialises it, for the read-side tests. */
    private fun taskJson(id: String, status: String, requests: String = "") =
        """{"id":"$id","title":"작업 $id","status":"$status","stage":"review"$requests}"""

    private fun storedJson(id: String, taskId: String, text: String, replayed: Boolean = false) =
        """{"request":{"id":"$id","task_id":"$taskId","seq":1,"text":"$text","author":"user-1",""" +
            """"created_at":"2026-11-18T00:00:00Z","status":"pending"},"replayed":$replayed}"""

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
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘")))

        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "  진행해 줘  ", "c-go"))

        val request = server.takeRequest()
        assertEquals("/api/chat/tasks/t-42/requests", request.path)
        val body = request.body.readUtf8()
        assertTrue("요청 번호가 본문에 있어야 한다: $body", body.contains("app-1"))
        assertTrue(body.contains("진행해 줘"))
        assertFalse("작성자는 서버가 정한다 — 앱이 보내지 않는다", body.contains("author"))

        val stored = vm.taskRequest.value
        assertEquals("t-42", stored.taskId)
        assertEquals(ChatTaskRequestStatus.PENDING, stored.stored?.status)
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

        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
        assertEquals("/api/chat/tasks/t-42/requests", server.takeRequest().path)
        assertNull(vm.taskRequest.value.error)
    }

    // 같은 요청 번호로 다시 보내면 서버가 이미 저장된 그 요청을 돌려준다 —
    // 두 건이 되지 않고, 화면은 「이미 저장된 요청」이라고 말한다.
    @Test
    fun resendingTheSameRequestKeyStoresOneRequest() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(200).setBody(storedJson("app-1", "t-42", "진행해 줘", replayed = true)))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
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
            MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘"))
                .setBodyDelay(300, TimeUnit.MILLISECONDS),
        )
        val first = launch(start = CoroutineStart.UNDISPATCHED) {
            assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
        }
        assertTrue("첫 저장이 가는 중이어야 한다", vm.taskRequest.value.isStoring)
        assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
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
            assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
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
        assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", ""))
        assertTrue(server.takeRequest().body.readUtf8().contains("app-1"))
        assertNull(vm.taskRequest.value.error)
        assertEquals(ChatTaskRequestStatus.PENDING, vm.taskRequest.value.stored?.status)
    }

    // 빈 요청과 너무 긴 요청은 서버까지 가지 않는다.
    @Test
    fun anEmptyOrOversizedRequestNeverLeavesTheApp() = runTest {
        val vm = ChatViewModel(api())
        assertFalse(vm.performTaskRequest(task(), "app-1", "   ", ""))
        assertNull(vm.taskRequest.value.error)

        assertFalse(
            vm.performTaskRequest(task(id = "t-42"), "app-1", "가".repeat(ChatViewModel.CHAT_TASK_REQUEST_MAX_BYTES), ""),
        )
        assertEquals(ChatTaskRequestError.INVALID, vm.taskRequest.value.error)
        assertEquals(0, server.requestCount)
    }

    // 원장이 세는 대기 건수는 작업 목록으로 함께 온다 — 없으면 아예 없다.
    @Test
    fun waitingRequestCountsArriveWithTheTaskList() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = taskJson("t1", "blocked", requests = ""","requests":{"pending":2,"claimed":1,"last_seq":7}"""),
                    counts = """{"blocked":1}""",
                ),
            ),
        )

        vm.loadTasks()

        val waiting = vm.tasks.value.tasks.single().requests
        assertEquals(2, waiting?.pending)
        assertEquals(3, waiting?.open)
        assertEquals(7L, waiting?.lastSeq)
    }

    // 상태 문구는 저장과 처리를 구분한다.
    @Test
    fun theStatusWordingSeparatesStoredFromProcessed() {
        assertEquals("저장됨 · 확인 대기", ChatTaskRequestStatus.label(ChatTaskRequestStatus.PENDING))
        assertEquals("확인 중", ChatTaskRequestStatus.label(ChatTaskRequestStatus.CLAIMED))
        assertEquals("확인 중 · 결과 대조", ChatTaskRequestStatus.label(ChatTaskRequestStatus.RECONCILING))
        assertEquals("처리됨", ChatTaskRequestStatus.label(ChatTaskRequestStatus.PROCESSED))
        assertEquals("처리 실패", ChatTaskRequestStatus.label(ChatTaskRequestStatus.FAILED))
        assertTrue(ChatTaskRequest(id = "r", status = ChatTaskRequestStatus.PROCESSED).settled)
        assertFalse(ChatTaskRequest(id = "r", status = ChatTaskRequestStatus.PENDING).settled)
    }

    // 결과를 보여 준 뒤에는 치운다 — 다음 작업의 화면에 지난 결과가 남지 않는다.
    @Test
    fun theOutcomeIsClearedOnceShown() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", "")
        server.takeRequest()
        assertEquals(ChatTaskRequestError.UNREACHABLE, vm.taskRequest.value.error)

        vm.dismissTaskRequest()

        assertNull(vm.taskRequest.value.error)
        assertNull(vm.taskRequest.value.stored)
    }

    // 고른 선택지의 id 가 본문에 함께 간다 — 세션이 자기 문장 중 무엇이
    // 돌아왔는지 알 수 있어야 한다.
    @Test
    fun theChosenChoiceIdTravelsWithTheRequest() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", "c-go"))
        val body = server.takeRequest().body.readUtf8()
        assertTrue("선택지 id 가 본문에 있어야 한다: $body", body.contains("c-go"))
    }

    // 고른 선택지는 **그 선택지를 본 작업 version** 과 함께 간다 — 서버가 그
    // 기준으로 낡은 화면의 제출을 거부할 수 있어야 한다.
    @Test
    fun theChosenChoiceCarriesTheVersionItWasOfferedOn() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(201).setBody(storedJson("app-1", "t-42", "진행해 줘")))
        assertTrue(vm.performTaskRequest(task(id = "t-42"), "app-1", "진행해 줘", "c-go", 7L))
        val body = server.takeRequest().body.readUtf8()
        assertTrue("기준 version 이 본문에 있어야 한다: $body", body.contains("\"choice_basis_version\":7"))
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
        assertFalse(vm.performTaskRequest(task(id = "t-42"), "app-1", "CI 로그를 확인해 줘.", "c-ci", 1L))
        assertEquals(ChatTaskRequestError.STALE_CHOICE, vm.taskRequest.value.error)
        // 저장된 것은 없다 — 저장 성공으로 오표시하지 않는다.
        assertNull(vm.taskRequest.value.stored)
        // 같은 409 라도 요청 번호 충돌과는 다른 말을 한다.
        assertTrue(
            ChatTaskRequestError.STALE_CHOICE.message != ChatTaskRequestError.CONFLICT.message,
        )
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

    // 선택지는 원장이 준 것만 쓴다 — 앱에는 상태별 목록이 없다.
    @Test
    fun theOfferedChoicesComeFromTheLedgerOnly() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = """{"id":"t1","title":"작업","status":"blocked","stage":"review",""" +
                        """"choices":[{"id":"c-ci","label":"CI 로그 확인","request_text":"CI 로그를 확인해 줘.","basis_version":3}]}""",
                    counts = """{"blocked":1}""",
                ),
            ),
        )

        vm.loadTasks()

        val task = vm.tasks.value.tasks.single()
        assertEquals(1, task.choices.size)
        assertEquals("c-ci", task.choices[0].id)
        assertEquals("CI 로그를 확인해 줘.", task.choices[0].requestText)
        assertEquals(3L, task.choices[0].basisVersion)
        // 선택지를 주지 않은 작업은 비어 있다 — 앱이 대신 만들지 않는다.
        server.enqueue(MockResponse().setBody(tasksJson(tasks = taskJson("t2", "running"))))
        vm.loadTasks()
        assertTrue(vm.tasks.value.tasks.single().choices.isEmpty())
    }

    // 이력은 원장이 적어 둔 것만, 도착 순서대로 — 페이지를 이어 읽으면 앞 페이지에
    // 덧붙는다.
    @Test
    fun theHistoryPagesInOrderAndKeepsWhatItAlreadyShowed() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody(eventsJson(firstPage = true)))

        assertTrue(vm.performTaskHistory("t-42", more = false))

        assertEquals("/api/chat/tasks/t-42/events?limit=20", server.takeRequest().path)
        var state = vm.taskHistory.value
        assertEquals(listOf(1L, 2L), state.events.map { it.seq })
        assertEquals("요청을 남겼다", state.events[1].summary)
        assertEquals("이 작업 진행해 줘", state.events[1].text)
        assertTrue(state.hasMore)
        assertFalse(state.isEmpty)

        server.enqueue(MockResponse().setBody(eventsJson(firstPage = false)))
        assertTrue(vm.performTaskHistory("t-42", more = true))
        assertEquals("/api/chat/tasks/t-42/events?after_seq=2&limit=20", server.takeRequest().path)
        state = vm.taskHistory.value
        assertEquals(listOf(1L, 2L, 3L), state.events.map { it.seq })
        assertFalse("마지막 페이지에는 커서가 없다", state.hasMore)
    }

    // 기록이 없는 작업은 빈 상태이고, 실패는 빈 상태로 보이지 않으며 이미 보던
    // 페이지를 지우지 않는다.
    @Test
    fun anEmptyHistoryAndAFailureAreNotTheSameThing() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setBody("""{"events":[]}"""))
        assertTrue(vm.performTaskHistory("t-42", more = false))
        server.takeRequest()
        assertTrue(vm.taskHistory.value.isEmpty)
        assertNull(vm.taskHistory.value.error)

        server.enqueue(MockResponse().setBody(eventsJson(firstPage = true)))
        vm.performTaskHistory("t-42", more = false)
        server.takeRequest()
        val shown = vm.taskHistory.value.events
        server.enqueue(MockResponse().setResponseCode(502).setBody("""{"error":"no"}"""))
        assertFalse(vm.performTaskHistory("t-42", more = true))
        server.takeRequest()
        val failed = vm.taskHistory.value
        assertEquals(ChatTaskHistoryError.UNREACHABLE, failed.error)
        assertEquals("보던 페이지는 남는다", shown.map { it.seq }, failed.events.map { it.seq })
        assertFalse(failed.isEmpty)

        for ((status, want) in listOf(
            404 to ChatTaskHistoryError.NOT_FOUND,
            409 to ChatTaskHistoryError.UNAVAILABLE,
            501 to ChatTaskHistoryError.UNSUPPORTED,
        )) {
            val other = ChatViewModel(api())
            server.enqueue(MockResponse().setResponseCode(status).setBody("""{"code":"x"}"""))
            assertFalse(other.performTaskHistory("t-42", more = false))
            server.takeRequest()
            assertEquals("status $status", want, other.taskHistory.value.error)
        }
    }

    private fun eventsJson(firstPage: Boolean) = if (firstPage) {
        """{"events":[{"seq":1,"task_id":"t-42","at":"2026-11-19T00:00:00Z","kind":"task_created",""" +
            """"actor":"parent","summary":"작업을 원장에 등록했다"},""" +
            """{"seq":2,"task_id":"t-42","at":"2026-11-19T00:01:00Z","kind":"request_stored","actor":"user-1",""" +
            """"summary":"요청을 남겼다","detail":{"text":"이 작업 진행해 줘"},"request_id":"app-1"}],"next_cursor":2}"""
    } else {
        """{"events":[{"seq":3,"task_id":"t-42","at":"2026-11-19T00:02:00Z","kind":"request_settled",""" +
            """"actor":"parent","summary":"요청을 processed 로 닫았다","detail":{"result":"진행했다"}}]}"""
    }
}