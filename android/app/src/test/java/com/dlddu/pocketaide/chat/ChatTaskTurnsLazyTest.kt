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

// 작업 상세의 「주고받은 내용」은 맨 아래에 접혀 있고, 접혔다는 것은 **읽지
// 않았다**는 뜻이다. 여기서 세는 것은 **실제로 나간 요청**이다: 접힌 동안 0,
// 첫 펼침에 그 작업의 턴 라우트로 1, 바뀐 것이 없으면 다시 펼쳐도 0.
//
// 어느 테스트도 운영 세션으로 나가지 않는다 — MockWebServer 가 받는다. 세션
// 파드의 이력 조회는 그 턴 라우트 뒤에만 있으므로, 요청이 나가지 않았다는 것이
// 곧 이력을 읽지 않았다는 뜻이다.
class ChatTaskTurnsLazyTest {
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

    /** 한 작업의 턴 한 페이지. */
    private fun page(taskId: String, turnId: Long, request: String, nextCursor: Long = 0) = json(
        """{"task_id":"$taskId","task_version":1,"task":{"id":"$taskId"},""" +
            """"turns":[{"id":$turnId,"request":"$request","choices":[],"artifacts":[],""" +
            """"requested_at":"2026-11-19T00:00:00Z"}]""" +
            (if (nextCursor > 0) ""","next_cursor":$nextCursor""" else "") +
            "}",
    )

    private fun task(
        id: String = "t1",
        turnsTotal: Long = 3,
        waiting: ChatTaskWaiting? = null,
        status: String = ChatTaskStatus.BLOCKED,
        choiceTurnId: Long = 0,
    ) = ChatTask(
        id = id,
        title = "작업 $id",
        status = status,
        turnsTotal = turnsTotal,
        waiting = waiting,
        choiceTurnId = choiceTurnId,
    )

    /** 나간 요청의 경로들. 무엇이 몇 번 나갔는지를 이것으로 센다. */
    private fun paths(): List<String> =
        (1..server.requestCount).mapNotNull { server.takeRequest(1, TimeUnit.SECONDS)?.path }

    // 접힌 동안은 아무 요청도 나가지 않는다. 상세가 보여 주는 지금 정보(묻는
    // 것·선택지·그 turn·완료 기준)는 패널이 이미 폴링하는 작업 목록이 들고
    // 오므로, 접힌 화면을 그리기 위해 턴 라우트를 부를 일이 없다.
    @Test
    fun aClosedExchangeAsksForNothing() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            json(
                """{"session_id":"cx1","tasks":[{"id":"t1","title":"작업","status":"blocked",""" +
                    """"description":"무엇을 why","completion_criteria":"머지되면 끝","turns_total":9,""" +
                    """"waiting":{"turn_id":7,"question":"어디까지?"},""" +
                    """"choices":[{"label":"전부","request_text":"전부 해 줘"}],"choice_turn_id":7}],""" +
                    """"counts":{"blocked":1},"truncated":false}""",
            ),
        )

        vm.loadTasks()

        val row = vm.tasks.value.tasks.single()
        // 상세에 필요한 지금 정보는 목록 한 번으로 이미 손에 있다.
        assertEquals("무엇을 why", row.description)
        assertEquals("머지되면 끝", row.completionCriteria)
        assertEquals("어디까지?", row.waiting?.question)
        assertEquals(1, row.choices.size)
        assertEquals(7L, row.choiceTurnId)
        assertEquals(9L, row.turnsTotal)
        // 그리고 아직 읽은 이력이 없다 — 나간 요청은 목록 하나뿐이다.
        assertEquals(listOf("/api/chat/tasks"), paths())
        assertEquals(ChatTaskTurnsUiState(), vm.taskTurns.value)
    }

    // 첫 펼침에 선택된 작업의 이력만 한 번 읽는다 — 그 작업의 턴 라우트 하나로.
    @Test
    fun theFirstOpenReadsOnlyThatTasksExchangeOnce() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(page("t1", 1, "이거 해 줘"))

        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(task("t1")), more = false)

        assertEquals(listOf("/api/chat/tasks/t1/turns?limit=20"), paths())
        assertEquals("t1", vm.taskTurns.value.taskId)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })
        assertTrue(vm.taskTurns.value.hasLoaded)
        assertNull(vm.taskTurns.value.error)
    }

    // 바뀐 것이 없으면 다시 펼쳐도 읽지 않는다. 「바뀐 것이 없다」는 판단은
    // 패널이 이미 폴링하는 그 작업 행에서 나오므로, 알아보기 위한 요청도 없다.
    @Test
    fun openingItAgainAfterNothingChangedAsksForNothing() = runTest {
        val vm = ChatViewModel(api())
        val row = task("t1")
        server.enqueue(page("t1", 1, "이거 해 줘"))

        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(row), more = false)
        assertEquals(1, server.requestCount)

        // 접었다 다시 펼침 = 같은 작업·같은 신호로 다시 요구하는 것.
        vm.openTaskTurns(row)
        assertEquals("쥐고 있는 것을 다시 읽지 않는다", 1, server.requestCount)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })
    }

    // 이력이 **바뀌면** 펼친 화면에도, 다시 펼친 화면에도 반영된다. 바뀌었다는
    // 신호는 작업 목록이 이미 들고 오는 것들이다 — 턴 수, 열린 turn 과 그것을
    // 집었는지, 상태, 선택지가 제시된 turn.
    @Test
    fun aChangedExchangeIsReadAgain() = runTest {
        val vm = ChatViewModel(api())
        val before = task("t1", turnsTotal = 1)
        server.enqueue(page("t1", 1, "이거 해 줘"))
        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(before), more = false)
        assertEquals(1, server.requestCount)

        // 새 턴이 하나 생겼다.
        server.enqueue(page("t1", 2, "그다음도 해 줘"))
        vm.openTaskTurns(task("t1", turnsTotal = 2))
        assertEquals("턴이 늘면 다시 읽는다", 2, server.requestCount)
        assertEquals(listOf(2L), vm.taskTurns.value.turns.map { it.id })

        // 열린 turn 을 부모가 집었다 — 턴 수는 그대로다.
        server.enqueue(page("t1", 2, "그다음도 해 줘"))
        vm.openTaskTurns(
            task("t1", turnsTotal = 2, waiting = ChatTaskWaiting(turnId = 2, running = true)),
        )
        assertEquals("열린 turn 이 움직이면 다시 읽는다", 3, server.requestCount)

        // 상태가 움직였다.
        server.enqueue(page("t1", 2, "그다음도 해 줘"))
        vm.openTaskTurns(
            task(
                "t1", turnsTotal = 2, status = ChatTaskStatus.REVIEW,
                waiting = ChatTaskWaiting(turnId = 2, running = true),
            ),
        )
        assertEquals("작업이 움직이면 다시 읽는다", 4, server.requestCount)
    }

    // 다시 읽는 동안 읽던 것이 화면에서 사라지지 않는다. 다시 읽기는 **펼친
    // 채로 읽고 있을 때** 일어나는 일이므로(그 작업에 턴이 하나 생겼다), 읽고
    // 있던 긴 이력을 스피너로 비우고 「더 생겼다」를 알리는 것은 읽던 것을
    // 빼앗는 일이다. 새 페이지가 닿으면 그때 바뀐다.
    @Test
    fun aRereadKeepsWhatIsOnScreenUntilTheNewPageArrives() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(page("t1", 1, "이거 해 줘"))
        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(task("t1", turnsTotal = 1)), more = false)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })

        server.enqueue(page("t1", 2, "그다음도 해 줘").setBodyDelay(200, TimeUnit.MILLISECONDS))
        val reread = launch(start = CoroutineStart.UNDISPATCHED) {
            vm.openTaskTurns(task("t1", turnsTotal = 2))
        }
        assertTrue("다시 읽는 중이다", vm.taskTurns.value.isLoading)
        assertEquals(
            "다시 읽는 동안에도 읽던 이력이 그대로 있다",
            listOf(1L),
            vm.taskTurns.value.turns.map { it.id },
        )
        assertFalse("읽은 적 없는 화면으로 돌아가지 않는다", vm.taskTurns.value.isEmpty)
        reread.join()

        // 닿으면 그 페이지가 화면이 된다.
        assertEquals(listOf(2L), vm.taskTurns.value.turns.map { it.id })
        assertEquals(2, server.requestCount)
    }

    // 다른 작업으로 바뀔 때는 쥐고 있던 페이지를 들고 가지 않는다 — 앞 작업의
    // 이력이 한순간이라도 뒤 작업의 제목 아래 있어서는 안 된다.
    @Test
    fun switchingTasksDoesNotCarryThePreviousPageIntoTheNewOne() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(page("t1", 1, "이거 해 줘"))
        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(task("t1")), more = false)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })

        server.enqueue(page("t2", 2, "저거 해 줘").setBodyDelay(200, TimeUnit.MILLISECONDS))
        val next = launch(start = CoroutineStart.UNDISPATCHED) {
            vm.openTaskTurns(task("t2"))
        }
        assertEquals("t2", vm.taskTurns.value.taskId)
        assertTrue("앞 작업의 턴을 들고 가지 않는다", vm.taskTurns.value.turns.isEmpty())
        next.join()
        assertEquals(listOf(2L), vm.taskTurns.value.turns.map { it.id })
    }

    // 접힌 동안 폴링이 돌아도 읽지 않는다: 신호가 바뀌어도 펼쳐져 있지 않으면
    // 아무도 요구하지 않기 때문이다 — 요구하는 길은 화면의 펼침 하나다.
    @Test
    fun pollingBehindAClosedExchangeReadsNothing() = runTest {
        val vm = ChatViewModel(api())
        repeat(3) {
            server.enqueue(
                json(
                    """{"session_id":"cx1","tasks":[{"id":"t1","title":"작업","status":"blocked",""" +
                        """"turns_total":${it + 1}}],"counts":{"blocked":1},"truncated":false}""",
                ),
            )
            vm.loadTasks()
        }
        assertEquals("폴링 세 번은 목록 세 번이고 이력은 0번이다", 3, server.requestCount)
        assertTrue(paths().all { it == "/api/chat/tasks" })
        assertEquals(ChatTaskTurnsUiState(), vm.taskTurns.value)
    }

    // 중복 인플라이트 방지: 같은 작업·같은 신호로 두 번 요구해도 요청은 하나다.
    @Test
    fun aSecondOpenWhileTheFirstIsInFlightAsksOnce() = runTest {
        val vm = ChatViewModel(api())
        val row = task("t1")
        server.enqueue(page("t1", 1, "이거 해 줘").setBodyDelay(200, TimeUnit.MILLISECONDS))

        val first = launch(start = CoroutineStart.UNDISPATCHED) {
            vm.performTaskTurns("t1", ChatTaskHistorySignal.of(row), more = false)
        }
        assertTrue("첫 요구가 떠 있다", vm.taskTurns.value.isLoading)
        // 쥐고 있는 중인 것도 쥐고 있는 것이다 — 두 번째 요구는 나가지 않는다.
        vm.openTaskTurns(row)
        assertFalse(
            vm.performTaskTurns("t1", ChatTaskHistorySignal.of(row), more = false),
        )
        first.join()

        assertEquals(1, server.requestCount)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })
    }

    // 빠른 작업 전환: 앞 작업의 늦은 응답이 뒤 작업의 이력으로 보이지 않는다.
    // 뒤 작업의 요구는 앞의 것을 대신하고, 먼저 나간 응답은 아무 데도 닿지
    // 않는다.
    @Test
    fun aLateAnswerForTheTaskWeLeftIsDropped() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(page("t1", 1, "앞 작업의 말").setBodyDelay(300, TimeUnit.MILLISECONDS))
        server.enqueue(page("t2", 2, "뒤 작업의 말"))

        val slow = launch(start = CoroutineStart.UNDISPATCHED) {
            vm.performTaskTurns("t1", ChatTaskHistorySignal.of(task("t1")), more = false)
        }
        // 아직 t1 이 떠 있는 동안 t2 로 바꿔 펼친다.
        vm.performTaskTurns("t2", ChatTaskHistorySignal.of(task("t2")), more = false)
        assertEquals("t2", vm.taskTurns.value.taskId)
        assertEquals(listOf(2L), vm.taskTurns.value.turns.map { it.id })

        slow.join()

        // 늦게 도착한 t1 의 페이지는 버려진다 — 화면은 t2 의 것을 들고 있다.
        assertEquals("t2", vm.taskTurns.value.taskId)
        assertEquals(listOf(2L), vm.taskTurns.value.turns.map { it.id })
    }

    // 상세를 닫으면 쥐고 있던 이력을 버리고, 떠 있던 응답도 닿을 곳이 없다 —
    // 같은 작업으로 다시 들어가면 접힌 채로 시작하고, 펼쳐야 다시 읽는다.
    @Test
    fun closingThePaneDropsThePageAndTheAnswerStillInTheAir() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(page("t1", 1, "이거 해 줘").setBodyDelay(200, TimeUnit.MILLISECONDS))

        val inFlight = launch(start = CoroutineStart.UNDISPATCHED) {
            vm.performTaskTurns("t1", ChatTaskHistorySignal.of(task("t1")), more = false)
        }
        vm.clearTaskTurns()
        inFlight.join()

        assertEquals("닫은 화면에 늦은 페이지가 들어오지 않는다", ChatTaskTurnsUiState(), vm.taskTurns.value)

        // 같은 작업으로 다시 들어와 펼치면 읽는다.
        server.enqueue(page("t1", 1, "이거 해 줘"))
        vm.openTaskTurns(task("t1"))
        assertEquals(2, server.requestCount)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })
    }

    // 실패는 「이 작업에 주고받은 내용이 없다」가 아니고, 다시 시도는 쥐고
    // 있는 것과 무관하게 다시 읽는다.
    @Test
    fun aFailureIsNotAnEmptyExchangeAndRetryReadsAgain() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"x"}"""))

        val row = task("t1")
        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(row), more = false)
        assertEquals(ChatTaskTurnsError.UNAVAILABLE, vm.taskTurns.value.error)
        assertFalse("실패는 빈 이력이 아니다", vm.taskTurns.value.isEmpty)

        // 바뀐 것이 없어도 재시도는 읽는다 — 실패를 쥐고 있는 것은 쥐고 있는
        // 것이 아니다.
        server.enqueue(page("t1", 1, "이거 해 줘"))
        vm.openTaskTurns(row)
        assertEquals(2, server.requestCount)
        assertNull(vm.taskTurns.value.error)
        assertEquals(listOf(1L), vm.taskTurns.value.turns.map { it.id })
    }

    // 턴이 없는 작업은 빈 이력이고, 그것도 한 번만 읽는다.
    @Test
    fun aTaskWithNoExchangeIsReadOnceAndStaysEmpty() = runTest {
        val vm = ChatViewModel(api())
        val row = task("t1", turnsTotal = 0)
        server.enqueue(json("""{"task_id":"t1","task_version":1,"task":{"id":"t1"},"turns":[]}"""))

        vm.openTaskTurns(row)
        assertTrue(vm.taskTurns.value.isEmpty)
        assertEquals(1, server.requestCount)

        // 턴이 0개라는 답도 답이다 — 다시 펼쳐도 또 묻지 않는다.
        vm.openTaskTurns(row)
        assertEquals(1, server.requestCount)
    }

    // 「이전 내용 더 보기」는 커서를 들고 같은 작업에 한 번 더 나가고, 받은
    // 페이지는 오래된 것부터의 순서를 지켜 앞에 붙는다.
    @Test
    fun readingMoreKeepsTheOldestFirstOrder() = runTest {
        val vm = ChatViewModel(api())
        val row = task("t1")
        server.enqueue(page("t1", 5, "먼저 온 말", nextCursor = 5))
        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(row), more = false)
        assertTrue(vm.taskTurns.value.hasMore)

        server.enqueue(page("t1", 6, "그다음 말"))
        vm.performTaskTurns("t1", vm.taskTurns.value.signal, more = true)

        assertEquals(listOf(5L, 6L), vm.taskTurns.value.turns.map { it.id })
        assertFalse(vm.taskTurns.value.hasMore)
        assertEquals(
            listOf("/api/chat/tasks/t1/turns?limit=20", "/api/chat/tasks/t1/turns?after_turn_id=5&limit=20"),
            paths(),
        )
    }

    // 첫 페이지를 다시 읽으면 커서도 처음으로 돌아간다 — 앞 읽기의 커서가 남아
    // 아무도 쥐고 있지 않은 페이지의 「더 보기」를 내밀지 않는다.
    @Test
    fun rereadingTheFirstPageForgetsTheOldCursor() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(page("t1", 5, "먼저 온 말", nextCursor = 5))
        vm.performTaskTurns("t1", ChatTaskHistorySignal.of(task("t1", turnsTotal = 1)), more = false)
        assertTrue(vm.taskTurns.value.hasMore)

        server.enqueue(page("t1", 6, "새 말"))
        vm.openTaskTurns(task("t1", turnsTotal = 2))

        assertFalse("새 첫 페이지의 커서가 앞의 것을 물려받지 않는다", vm.taskTurns.value.hasMore)
        assertEquals(listOf(6L), vm.taskTurns.value.turns.map { it.id })
    }

    // 쥐고 있는지의 판정은 작업·신호·실패를 함께 본다. 아무것도 읽지 않은
    // 화면은 **어떤** 작업도 쥐고 있지 않다 — 턴이 0개인 작업까지 포함해서.
    @Test
    fun holdingIsPerTaskAndPerSignal() {
        val signal = ChatTaskHistorySignal.of(task("t1", turnsTotal = 0))
        assertFalse("읽은 적 없는 화면은 아무것도 쥐고 있지 않다", ChatTaskTurnsUiState().holds("t1", signal))

        val loaded = ChatTaskTurnsUiState(taskId = "t1", signal = signal, hasLoaded = true)
        assertTrue(loaded.holds("t1", signal))
        assertFalse("다른 작업의 페이지는 이 작업의 것이 아니다", loaded.holds("t2", signal))
        assertFalse(
            "신호가 움직였으면 쥐고 있는 것은 낡은 것이다",
            loaded.holds("t1", ChatTaskHistorySignal.of(task("t1", turnsTotal = 1))),
        )
        assertFalse(
            "실패를 쥐고 있는 것은 쥐고 있는 것이 아니다",
            loaded.copy(error = ChatTaskTurnsError.UNREACHABLE).holds("t1", signal),
        )
    }
}
