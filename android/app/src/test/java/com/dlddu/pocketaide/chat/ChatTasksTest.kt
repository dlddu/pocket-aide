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
                    tasks = task("t1", "review", question = "CI 가 막혔다", turns = 2),
                    counts = """{"review":1,"running":0}""",
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
        assertEquals(ChatTaskStatus.REVIEW, task.status)
        // The two fields the ledger page had nowhere to put.
        assertEquals("무엇을 why", task.description)
        assertEquals("머지되면 끝", task.completionCriteria)
        // What is waiting comes from the task's open turn, not a count of rows
        // beside it.
        assertEquals("CI 가 막혔다", task.waiting?.question)
        assertEquals(7L, task.waiting?.turnId)
        assertEquals(1, state.counts["review"])
    }

    /**
     * A backend pod that has not rolled yet still answers with the ledger's
     * older name for the same judgment. The app reads it as the one status it
     * draws, so the task appears under the 검토 chip and is counted there —
     * instead of being filed under a chip this build does not have and so drawn
     * nowhere.
     */
    @Test
    fun anOlderBackendsBlockedIsTheOneReviewStatus() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = task("t-old", "blocked", question = "권한을 열어 줄까요?") + "," +
                        task("t-new", "review") + "," + task("t-run", "running"),
                    counts = """{"blocked":2,"review":3,"running":1,"planned":0,"completed":0,"unknown":0}""",
                ),
            ),
        )

        vm.loadTasks()
        server.takeRequest()

        val state = vm.tasks.value
        assertEquals(
            listOf(ChatTaskStatus.REVIEW, ChatTaskStatus.REVIEW, ChatTaskStatus.RUNNING),
            state.tasks.map { it.status },
        )
        // What the old record was waiting on is still readable — folding the
        // name must not cost the evidence a person acts on.
        assertEquals("권한을 열어 줄까요?", state.tasks.first().waiting?.question)
        // The two counts add up under the one name, and the retired one is gone
        // so no chip can be drawn for it.
        assertEquals(5, state.counts[ChatTaskStatus.REVIEW])
        assertFalse(state.counts.containsKey("blocked"))
        // Every drawn status still has a key, which is what tells "none of this
        // status" from "no such status" apart.
        assertEquals(ChatTaskStatus.order.toSet(), state.counts.keys)
    }

    /**
     * A count under a status this build has no chip for is added to 알 수 없음
     * rather than dropped. Dropping it would make the chips add up to less than
     * the ledger holds, which reads as work having gone away.
     */
    @Test
    fun aCountThisBuildCannotDrawBecomesUnknownRatherThanVanishing() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(tasks = task("t1", "review"), counts = """{"review":1,"deploying":4}"""),
            ),
        )

        vm.loadTasks()
        server.takeRequest()

        val counts = vm.tasks.value.counts
        assertEquals(1, counts[ChatTaskStatus.REVIEW])
        assertEquals(4, counts[ChatTaskStatus.UNKNOWN])
        assertEquals(5, counts.values.sum())
    }

    // 선택지를 올린 그 턴의 세 가지(보낸 요청·설명·질문)가 목록 응답으로
    // 함께 온다. 셋뿐인 것은 의도다 — 세션의 원문 답은 긴 쪽이고, 그것은
    // 사람이 「주고받은 내용」을 펼칠 때 읽는다.
    //
    // 그리고 이 칸이 없는 **구버전 백엔드**의 응답도 그대로 읽힌다: 없으면
    // null 이고, 화면은 전에 보여 주던 것으로 돌아간다(질문을 잃지 않는다).
    @Test
    fun theChoiceTurnsWordsArriveWithTheListAndAnOlderBackendStillParses() = runTest {
        val vm = ChatViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                tasksJson(
                    tasks = """{"id":"t1","title":"작업","status":"review","turns_total":4,""" +
                        """"waiting":{"turn_id":9,"request":"이거 해 줘","question":"어디까지?"},""" +
                        """"choices":[{"label":"전부","request_text":"전부 해 줘"}],"choice_turn_id":9,""" +
                        """"choice_turn":{"request":"리뷰 반영해 줘","explanation":"필수 3건 반영했어요.",""" +
                        """"question":"남은 2건은요?","response":"백엔드가 보내지 않는 긴 원문"},""" +
                        // 같은 응답의 다른 작업: 이 칸이 없는 작업은 null 이다.
                        """"migrated":false},{"id":"t2","title":"선택지 없음","status":"running"}""",
                ),
            ),
        )

        vm.loadTasks()

        val offered = vm.tasks.value.tasks.first { it.id == "t1" }
        assertEquals("리뷰 반영해 줘", offered.choiceTurn?.request)
        assertEquals("필수 3건 반영했어요.", offered.choiceTurn?.explanation)
        assertEquals("남은 2건은요?", offered.choiceTurn?.question)
        assertTrue(offered.choiceTurn!!.hasAny)
        // 그 턴의 열린 요청(waiting)과는 다른 칸이다 — 서로를 대신하지 않는다.
        assertEquals("이거 해 줘", offered.waiting?.request)
        // 선택지를 올리지 않은 작업은 올릴 말도 없다.
        assertNull(vm.tasks.value.tasks.first { it.id == "t2" }.choiceTurn)
    }

    // 세 칸이 모두 비어 있으면 그릴 것이 없다 — 빈 틀을 띄우지 않기 위해
    // 화면이 묻는 바로 그 값이다.
    @Test
    fun aChoiceTurnWithNothingInItHasNothingToShow() {
        assertFalse(ChatTaskChoiceTurn().hasAny)
        assertFalse(ChatTaskChoiceTurn(request = "   ").hasAny)
        assertTrue(ChatTaskChoiceTurn(question = "되나요?").hasAny)
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
        assertEquals("검토", ChatTaskStatus.label(ChatTaskStatus.REVIEW))
        assertEquals("알 수 없음", ChatTaskStatus.label("deploying"))
        assertTrue(ChatTaskStatus.needsAttention(ChatTaskStatus.REVIEW))
        assertFalse(ChatTaskStatus.needsAttention(ChatTaskStatus.RUNNING))
        assertFalse(ChatTaskStatus.needsAttention(ChatTaskStatus.COMPLETED))
    }

    // Display order puts what needs a person first. There is one such status.
    @Test
    fun displayOrderPutsAttentionFirst() {
        assertEquals(
            listOf("review", "running", "planned", "completed", "unknown"),
            ChatTaskStatus.order,
        )
        assertEquals(1, ChatTaskStatus.order.count { ChatTaskStatus.needsAttention(it) })
    }

    /**
     * The fold is exactly one value wide, and it is case- and padding-proof
     * because that is the platform's own column rather than a different status.
     * A status this build does not know stays 알 수 없음: telling somebody a
     * decision is theirs on no evidence is worse than saying so.
     */
    @Test
    fun onlyTheLedgersOtherNameForAJudgmentFolds() {
        assertEquals(ChatTaskStatus.REVIEW, ChatTaskStatus.normalise("blocked"))
        assertEquals(ChatTaskStatus.REVIEW, ChatTaskStatus.normalise("  BLOCKED  "))
        assertEquals(ChatTaskStatus.REVIEW, ChatTaskStatus.normalise("review"))
        assertEquals(ChatTaskStatus.RUNNING, ChatTaskStatus.normalise("running"))
        assertEquals(ChatTaskStatus.UNKNOWN, ChatTaskStatus.normalise("deploying"))
        assertEquals(ChatTaskStatus.UNKNOWN, ChatTaskStatus.normalise(""))
        // 기다리는 것은 물어보는 것이 아니다. 아직 세션에 넘기지 않은 일(승인된
        // 슬롯이 나기를 기다리는 것도 여기다)은 「예정」에, 이미 도는 일은
        // 「진행」에 남는다 — 사람에게 물을 것이 없으므로 「검토」로 올라오면
        // 안 된다.
        assertEquals(ChatTaskStatus.PLANNED, ChatTaskStatus.normalise("planned"))
        assertEquals(ChatTaskStatus.PLANNED, ChatTaskStatus.normaliseFilter("planned"))
        assertEquals(ChatTaskStatus.RUNNING, ChatTaskStatus.normaliseFilter("running"))
    }

    /**
     * 기다리는 줄과 물어보는 줄은 화면에서 섞이지 않는다 — 「예정」·「진행」은
     * 손댈 것으로 강조되지 않고, 검토 칩으로 좁혀도 끌려 들어오지 않는다.
     * 승인된 슬롯 대기가 「검토」에 섞이면 사용자는 결정할 것이 없는 일을
     * 결정하라는 줄로 읽는다.
     */
    @Test
    fun waitingWorkIsNotDrawnAsSomethingToDecide() {
        assertFalse(ChatTaskStatus.needsAttention(ChatTaskStatus.PLANNED))
        assertFalse(ChatTaskStatus.needsAttention(ChatTaskStatus.RUNNING))
        assertTrue(ChatTaskStatus.needsAttention(ChatTaskStatus.REVIEW))

        val ledger = listOf(
            ChatTask(id = "t-slot", status = ChatTaskStatus.normalise("planned")),
            ChatTask(id = "t-run", status = ChatTaskStatus.normalise("running")),
            ChatTask(id = "t-ask", status = ChatTaskStatus.normalise("blocked")),
        )
        assertEquals(
            listOf("t-ask"),
            ChatTaskFilter(ChatTaskStatus.REVIEW).select(ledger).map { it.id },
        )
        assertEquals(
            listOf("t-slot"),
            ChatTaskFilter(ChatTaskStatus.PLANNED).select(ledger).map { it.id },
        )
    }

    /**
     * A phone keeps the chip it was filtering by across a rotation, a fold and
     * a process death, and that value was written by whichever build last ran.
     * A saved 차단 comes back as 검토 — the same work, the chip that is on
     * screen — and a saved status this build has no chip for falls back to 전체
     * rather than narrowing the list with no way to widen it again.
     */
    @Test
    fun aSavedFilterFromAnotherBuildComesBackAsAChipThatIsOnScreen() {
        assertEquals(ChatTaskStatus.REVIEW, ChatTaskStatus.normaliseFilter("blocked"))
        assertEquals(ChatTaskStatus.REVIEW, ChatTaskStatus.normaliseFilter("review"))
        assertEquals(ChatTaskStatus.UNKNOWN, ChatTaskStatus.normaliseFilter("unknown"))
        assertNull(ChatTaskStatus.normaliseFilter(null))
        assertNull("a status with no chip cannot be undone on screen", ChatTaskStatus.normaliseFilter("deploying"))
    }

    /**
     * The saved filter and the folded list agree, so a phone restored with 차단
     * saved shows that work under 검토 instead of an empty list with the rows
     * still there.
     */
    @Test
    fun aRestoredBlockedFilterSelectsTheSameWorkUnderReview() {
        val tasks = listOf(
            ChatTask(id = "t-old", status = ChatTaskStatus.normalise("blocked")),
            ChatTask(id = "t-run", status = ChatTaskStatus.RUNNING),
        )
        val restored = ChatTaskFilter(ChatTaskStatus.normaliseFilter("blocked"))

        assertEquals(ChatTaskStatus.REVIEW, restored.status)
        assertEquals(listOf("t-old"), restored.select(tasks).map { it.id })
        assertNull(restored.emptiness(tasks))
        // Pressing the chip it came back as still goes to 전체, so the restored
        // narrowing has a way out.
        assertNull(restored.toggled(ChatTaskStatus.REVIEW).status)
    }
}
