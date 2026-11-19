package com.dlddu.pocketaide.chat

import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.test.DeviceConfigurationOverride
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.ForcedSize
import androidx.compose.ui.test.assertCountEquals
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.assertTextContains
import androidx.compose.ui.test.getUnclippedBoundsInRoot
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onAllNodesWithTag
import androidx.compose.ui.test.onAllNodesWithText
import androidx.compose.ui.test.onFirst
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performTextInput
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import kotlin.math.abs

// 작업 패널을 실제로 눌러 본다: 칩과 완료 숨김으로 목록을 좁히고, 한 작업으로
// 들어가 보낼 문장을 고르거나 직접 쓰고, 실패한 요청을 다시 보낸다.
//
// 어느 테스트도 네트워크로 나가지 않는다 — 저장은 세어 보는 콜백에서 멈춘다.
// 운영 세션·원장에는 아무 요청도 가지 않는다.
@OptIn(ExperimentalTestApi::class)
@RunWith(AndroidJUnit4::class)
class ChatTaskPanelInteractionTest {
    @get:Rule
    val compose = createComposeRule()

    private var stores = 0
    private var historyLoads = 0
    private var historyMores = 0

    private val blocked = ChatTask(
        id = "t-blocked",
        title = "막힌 작업",
        status = ChatTaskStatus.BLOCKED,
        stage = "review",
        reason = "CI 가 막혔다",
    )

    private val ledger = ChatTasksUiState(
        sessionId = "cx1",
        tasks = listOf(
            blocked,
            ChatTask(id = "t-review", title = "검토 대기", status = ChatTaskStatus.REVIEW, stage = "review"),
            ChatTask(id = "t-running", title = "도는 작업", status = ChatTaskStatus.RUNNING),
            ChatTask(id = "t-done1", title = "끝난 작업 1", status = ChatTaskStatus.COMPLETED),
            ChatTask(id = "t-done2", title = "끝난 작업 2", status = ChatTaskStatus.COMPLETED),
        ),
        counts = mapOf(
            ChatTaskStatus.BLOCKED to 1,
            ChatTaskStatus.REVIEW to 1,
            ChatTaskStatus.RUNNING to 1,
            ChatTaskStatus.COMPLETED to 2,
        ),
        hasLoaded = true,
    )

    /**
     * 넓은 창의 실제 화면. 보는 방식·고른 작업·요청 초안은 화면이 들고 있는
     * 그대로 여기서도 들고 있어, 누른 결과가 진짜 다시 그려진다.
     */
    private fun start(
        tasks: ChatTasksUiState = ledger,
        chat: ChatUiState = ChatUiState(hasLoaded = true),
        taskRequest: ChatTaskRequestUiState = ChatTaskRequestUiState(),
        history: ChatTaskHistoryUiState = ChatTaskHistoryUiState(taskId = "t-blocked", hasLoaded = true),
    ) {
        compose.setContent {
            // 넓은 창이면서 상태 칩이 한 줄에 다 들어오는 폭 — 칩이 가로
            // 스크롤 밖으로 밀려나면 누를 수 없어 테스트가 레이아웃 탓에 깨진다.
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(DpSize(1200.dp, 900.dp))) {
                var filter by remember { mutableStateOf(ChatTaskFilter()) }
                var openId by remember { mutableStateOf<String?>(null) }
                var request by remember { mutableStateOf("") }
                // 화면이 쓰는 그 규칙을 그대로 둔다(`chatTaskPickHolds`): 고른
                // 문장을 한 글자라도 고치면 출처와 기준이 함께 떨어진다.
                var pickId by remember { mutableStateOf("") }
                var pickBasis by remember { mutableStateOf(0L) }
                var pickText by remember { mutableStateOf("") }
                ChatPanes(
                    state = chat,
                    tasks = tasks,
                    draft = "",
                    onDraftChange = {},
                    onSend = {},
                    onRefresh = {},
                    onDismissNotice = {},
                    listState = rememberLazyListState(),
                    panelOpen = true,
                    onTogglePanel = {},
                    pollTasks = {},
                    taskListState = rememberLazyListState(),
                    filter = filter,
                    onFilterChange = { filter = it },
                    openTask = tasks.tasks.firstOrNull { it.id == openId },
                    onOpenTask = {
                        openId = it.id
                        // 다른 작업에 쓰던 말이 따라오지 않는다.
                        request = ""
                    },
                    onCloseTask = { openId = null },
                    taskRequest = taskRequest,
                    requestDraft = request,
                    onRequestDraftChange = {
                        request = it
                        if (!chatTaskPickHolds(it, pickText)) {
                            pickId = ""
                            pickBasis = 0L
                            pickText = ""
                        }
                    },
                    onChoicePicked = {
                        request = it.requestText
                        pickId = it.id
                        pickBasis = it.basisVersion
                        pickText = it.requestText
                    },
                    history = history,
                    onLoadHistory = { historyLoads++ },
                    onLoadMoreHistory = { historyMores++ },
                    onStoreRequest = {
                        stores++
                        sentPick = pickId to pickBasis
                    },
                )
            }
        }
        compose.waitForIdle()
    }

    /** 마지막 저장이 함께 보낸 출처와 기준 version. */
    private var sentPick: Pair<String, Long> = "" to 0L

    private fun rows(): Int = compose.onAllNodesWithTag("chat.tasks.item").fetchSemanticsNodes().size

    private fun openBlocked() {
        compose.onNodeWithText("막힌 작업").performClick()
        compose.waitForIdle()
    }

    // ── 목록을 좁히는 두 가지 ───────────────────────────────────────────────

    // 칩과 완료 숨김은 함께 걸린다 — 한쪽을 쓰는 동안 다른 쪽이 풀리지 않는다.
    @Test
    fun aStatusChipAndHidingCompletedNarrowTheListTogether() {
        start()
        assertEquals("원장의 다섯 줄", 5, rows())

        compose.onNodeWithTag("chat.tasks.hideCompleted").performClick()
        compose.waitForIdle()
        assertEquals("완료 두 줄이 빠진다", 3, rows())

        compose.onNodeWithTag("chat.tasks.filter.review").performClick()
        compose.waitForIdle()
        assertEquals("검토 한 줄만 남는다", 1, rows())
        compose.onNodeWithText("검토 대기").assertIsDisplayed()

        // 고른 칩을 다시 누르면 전체로 돌아가되, 완료 숨김은 그대로다.
        compose.onNodeWithTag("chat.tasks.filter.review").performClick()
        compose.waitForIdle()
        assertEquals(3, rows())
    }

    // 숨겨서 비었을 때는 일이 사라진 것이 아니라고 말하고, 한 번에 되돌린다.
    @Test
    fun hidingCompletedWorkSaysItIsStillThereAndOffersTheWayBack() {
        start(
            ledger.copy(
                tasks = ledger.tasks.filter { it.status == ChatTaskStatus.COMPLETED },
                counts = mapOf(ChatTaskStatus.COMPLETED to 2),
            ),
        )
        compose.onNodeWithTag("chat.tasks.hideCompleted").performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.empty.completedHidden").assertIsDisplayed()
        assertEquals(0, rows())

        compose.onNodeWithTag("chat.tasks.empty.undo").performClick()
        compose.waitForIdle()
        assertEquals("되돌리면 그대로 있다", 2, rows())
    }

    // 고른 상태의 줄이 없어서 비었을 때는 숨김 탓으로 돌리지 않는다. 목록이
    // 일부만 온 원장(truncated)에서 실제로 일어나는 경우다.
    @Test
    fun aStatusWithNoRowsSaysSoWithoutBlamingTheHide() {
        start(ledger.copy(counts = ledger.counts + (ChatTaskStatus.PLANNED to 3), truncated = true))
        compose.onNodeWithTag("chat.tasks.truncated").assertIsDisplayed()

        compose.onNodeWithTag("chat.tasks.filter.planned").performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.empty.filtered").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.empty.completedHidden").assertCountEquals(0)

        compose.onNodeWithTag("chat.tasks.empty.undo").performClick()
        compose.waitForIdle()
        assertEquals(5, rows())
    }

    // ── 한 작업으로 들어가기 ────────────────────────────────────────────────

    // 고른 작업의 상세가 같은 오른쪽 칸에 열리고, 대화는 자기 절반을 지킨다.
    @Test
    fun tappingATaskOpensItsDetailInTheSameHalf() {
        start()
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.list").assertCountEquals(0)
        // 어느 작업인지가 화면에 있다 — 보내는 메시지가 그 id 로 부른다.
        compose.onNodeWithTag("chat.tasks.detail.id").assertTextContains("t-blocked", substring = true)
        compose.onNodeWithTag("chat.tasks.detail.reason").assertTextContains("CI 가 막혔다", substring = true)

        val window = widthOf("chat.screen")
        assertRatio("대화", widthOf("chat.messages.pane") / window, 0.5f)
        assertRatio("상세", widthOf("chat.tasks.detail") / window, 0.5f)
    }

    // 뒤로는 작업 목록으로 돌아간다 — 들어온 그 걸음만 되돌린다.
    @Test
    fun backReturnsToTheTaskList() {
        start()
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.back").performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.list").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.detail").assertCountEquals(0)
    }

    // 원장에 없는 칸은 비워 두지도, 채워 넣지도 않는다.
    @Test
    fun anUnrecordedFieldSaysSoRatherThanLookingBlank() {
        start()
        compose.onNodeWithText("도는 작업").performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.reason").assertTextContains("원장에 적혀 있지 않아요.")
        compose.onNodeWithTag("chat.tasks.detail.next").assertTextContains("원장에 적혀 있지 않아요.")
    }

    // ── 보낼 문장 ───────────────────────────────────────────────────────────

    // 세션이 올려 둔 선택지를 누르면 저장하지 않고 요청 칸을 채운다. 앱에는
    // 상태별 목록이 없으므로, 선택지가 없으면 없다고 말하고 자유 입력만 남는다.
    @Test
    fun aStoredChoiceFillsTheRequestWithoutStoringIt() {
        start(
            ledger.copy(
                tasks = ledger.tasks.map {
                    if (it.id == "t-blocked") {
                        it.copy(choices = listOf(ChatTaskChoice(
                            id = "c-ci", label = "CI 로그 확인 요청",
                            requestText = "방금 실패한 CI 로그에서 어디서 멈췄는지 알려 줘.", basisVersion = 3,
                        )))
                    } else {
                        it
                    }
                },
            ),
        )
        openBlocked()

        compose.onAllNodesWithTag("chat.tasks.detail.choice").onFirst().performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.field")
            .assertTextContains("어디서 멈췄는지", substring = true)
        assertEquals("고르는 것은 저장하는 것이 아니다", 0, stores)
    }

    @Test
    fun noOfferedChoicesSaysSoAndKeepsFreeText() {
        start()
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.noChoices").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.detail.choice").assertCountEquals(0)
        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("직접 쓴 요청")
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()
        assertEquals(1, stores)
    }

    // 상세를 열면 이력을 한 번 읽고, 기록이 있으면 시간순으로 보여 준다.
    @Test
    fun theHistoryIsReadOnceAndShownInOrder() {
        start(
            history = ChatTaskHistoryUiState(
                taskId = "t-blocked", hasLoaded = true, nextCursor = 2,
                events = listOf(
                    ChatTaskEvent(seq = 1, taskId = "t-blocked", at = "2026-11-19T00:00:00Z",
                        kind = "task_created", actor = "parent", summary = "작업을 원장에 등록했다"),
                    ChatTaskEvent(seq = 2, taskId = "t-blocked", at = "2026-11-19T00:01:00Z",
                        kind = "request_stored", actor = "user-1", summary = "요청을 남겼다",
                        detail = mapOf("text" to "아주 긴 요청 ".repeat(20))),
                ),
            ),
        )
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.history").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.detail.history.line").assertCountEquals(2)
        compose.onNodeWithText("내 요청 · 2026-11-19T00:01:00Z").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.history.more").assertIsDisplayed()
        assertEquals("상세를 열면 한 번 읽는다", 1, historyLoads)
        // 긴 본문은 눌러서 펼친다.
        compose.onAllNodesWithTag("chat.tasks.detail.history.body").onFirst().performClick()
        compose.waitForIdle()
    }

    // 「이전 기록 더 보기」는 다음 페이지를 청한다 — 커서가 있을 때만 보인다.
    @Test
    fun askingForOlderHistoryReadsTheNextPage() {
        start(history = historyOf(nextCursor = 2))
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.history.more").performClick()
        compose.waitForIdle()
        assertEquals("더 보기는 다음 페이지를 청한다", 1, historyMores)

        // 커서가 없으면(마지막 페이지) 더 보기를 두지 않는다 — 누를 것이 없다.
        start(history = historyOf(nextCursor = 0))
        openBlocked()
        compose.onAllNodesWithTag("chat.tasks.detail.history.more").assertCountEquals(0)
    }

    // 긴 본문은 접혀 있고, 누르면 실제로 더 많이 보인다 — 화면에서 자란다.
    @Test
    fun aLongHistoryBodyGrowsWhenTapped() {
        start(history = historyOf(nextCursor = 0))
        openBlocked()

        val body = compose.onAllNodesWithTag("chat.tasks.detail.history.body").onFirst()
        val collapsed = body.getUnclippedBoundsInRoot().let { it.bottom - it.top }
        body.performClick()
        compose.waitForIdle()
        val expanded = compose.onAllNodesWithTag("chat.tasks.detail.history.body").onFirst()
            .getUnclippedBoundsInRoot().let { it.bottom - it.top }
        assertTrue("펼치면 더 많이 보여야 한다 (접힘 $collapsed → 펼침 $expanded)", expanded > collapsed)
    }

    // 끝난 요청의 이력 줄은 **사람이 읽을 결과**를 담는다 — 상태 낱말만 두지
    // 않는다.
    @Test
    fun aSettledRequestShowsTheResultThePersonWillRead() {
        start(
            history = ChatTaskHistoryUiState(
                taskId = "t-blocked", hasLoaded = true,
                events = listOf(
                    ChatTaskEvent(
                        seq = 9, taskId = "t-blocked", at = "2026-11-19T00:09:00Z",
                        kind = "request_settled", actor = "parent", summary = "요청을 처리했다",
                        detail = mapOf("result" to "PR 올리고 CI 통과까지 확인했다", "outcome" to "processed"),
                    ),
                ),
            ),
        )
        openBlocked()

        compose.onNodeWithText("PR 올리고 CI 통과까지 확인했다", substring = true).assertIsDisplayed()
    }

    // 불러오는 중은 빈 기록도 실패도 아니다.
    @Test
    fun aHistoryBeingReadSaysItIsLoading() {
        start(history = ChatTaskHistoryUiState(taskId = "t-blocked", isLoading = true))
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.history.loading").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.detail.history.empty").assertCountEquals(0)
        compose.onAllNodesWithTag("chat.tasks.detail.history.error").assertCountEquals(0)
    }

    /** 이력 두 줄 — 하나는 등록, 하나는 아주 긴 요청 원문. */
    private fun historyOf(nextCursor: Long) = ChatTaskHistoryUiState(
        taskId = "t-blocked", hasLoaded = true, nextCursor = nextCursor,
        events = listOf(
            ChatTaskEvent(
                seq = 1, taskId = "t-blocked", at = "2026-11-19T00:00:00Z",
                kind = "task_created", actor = "parent", summary = "작업을 원장에 등록했다",
            ),
            ChatTaskEvent(
                seq = 2, taskId = "t-blocked", at = "2026-11-19T00:01:00Z",
                kind = "request_stored", actor = "user-1", summary = "요청을 남겼다",
                detail = mapOf("text" to "아주 긴 요청 ".repeat(20)),
            ),
        ),
    )

    // 기록 없음과 실패는 다른 화면이고, 실패는 다시 시도를 내민다.
    @Test
    fun anEmptyHistoryAndAFailureLookDifferent() {
        start()
        openBlocked()
        compose.onNodeWithTag("chat.tasks.detail.history.empty").assertIsDisplayed()
        compose.onAllNodesWithTag("chat.tasks.detail.history.error").assertCountEquals(0)
    }

    @Test
    fun aFailedHistoryOffersARetry() {
        start(history = ChatTaskHistoryUiState(
            taskId = "t-blocked", hasLoaded = true, error = ChatTaskHistoryError.UNREACHABLE,
        ))
        openBlocked()
        compose.onNodeWithTag("chat.tasks.detail.history.error").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.history.retry").performClick()
        compose.waitForIdle()
        assertEquals("다시 시도는 다시 읽는다", 2, historyLoads)
        compose.onAllNodesWithTag("chat.tasks.detail.history.empty").assertCountEquals(0)
    }

    // 직접 쓴 요청도 그 작업의 원장에 저장된다.
    @Test
    fun aTypedRequestIsStoredForTheTaskOnScreen() {
        start()
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("내가 직접 쓴 요청")
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()

        assertEquals(1, stores)
    }

    // 빈 요청으로는 저장할 수 없다.
    @Test
    fun anEmptyRequestCannotBeStored() {
        start()
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()

        assertEquals(0, stores)
    }

    // 답이 오는 중에도 저장된다 — 요청은 턴으로 가지 않고 원장으로 간다.
    @Test
    fun aRequestIsStorableWhileAReplyIsBeingWritten() {
        start(chat = ChatUiState(hasLoaded = true, pending = ChatPending()))
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("지금 저장해 줘")
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()

        assertEquals("답을 쓰는 중이어도 저장된다", 1, stores)
    }

    // 저장됐다는 말은 처리됐다는 말이 아니다.
    @Test
    fun theStoredNoteSaysWaitingNotDone() {
        start(
            taskRequest = ChatTaskRequestUiState(
                taskId = "t-blocked",
                stored = ChatTaskRequest(id = "app-1", taskId = "t-blocked", text = "요청"),
            ),
        )
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.stored").assertIsDisplayed()
        compose.onNodeWithText("요청 저장됨 · 부모 확인 대기").assertIsDisplayed()
        compose.onAllNodesWithText("처리됨").assertCountEquals(0)
    }

    // 대기 중인 요청 수는 원장이 센 것을 그대로 보여 준다.
    @Test
    fun waitingRequestsAreShownAsTheLedgerCountsThem() {
        start(
            ledger.copy(
                tasks = ledger.tasks.map {
                    if (it.id == "t-blocked") it.copy(requests = ChatTaskRequests(pending = 2, reconciling = 1, lastSeq = 9)) else it
                },
            ),
        )
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.waitingRequests")
            .assertTextContains("확인 대기 3건", substring = true)
    }

    // 실패는 쓴 말을 지우지 않는다 — 이유를 말하고 다시 저장하기를 내민다.
    @Test
    fun aRefusedRequestKeepsTheWordsAndOffersARetry() {
        start(taskRequest = ChatTaskRequestUiState(taskId = "t-blocked", error = ChatTaskRequestError.UNREACHABLE))
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("다시 보낼 요청")
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.error").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.field").assertTextContains("다시 보낼 요청", substring = true)
        compose.onNodeWithText("다시 저장하기").assertIsDisplayed()

        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()
        assertEquals("다시 저장하기는 저장할 수 있다", 1, stores)
    }

    // 저장됐다는 말은 저장한 그 작업 아래에만 뜬다.
    @Test
    fun anotherTasksOutcomeIsNotShownUnderThisTask() {
        start(
            taskRequest = ChatTaskRequestUiState(
                taskId = "t-review",
                stored = ChatTaskRequest(id = "app-9", taskId = "t-review", text = "다른 작업"),
            ),
        )
        openBlocked()

        compose.onAllNodesWithTag("chat.tasks.detail.stored").assertCountEquals(0)
        compose.onNodeWithTag("chat.tasks.detail.hint").assertIsDisplayed()
    }

    // 다른 작업으로 들어가면 요청 칸은 비어 있다 — 앞 작업에 쓰던 말이 다음
    // 작업의 요청으로 나가지 않는다.
    @Test
    fun openingAnotherTaskStartsFromAnEmptyRequest() {
        start()
        openBlocked()
        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("막힌 작업에 쓰던 말")
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.back").performClick()
        compose.waitForIdle()
        compose.onNodeWithText("검토 대기").performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.id").assertTextContains("t-review", substring = true)
        // 빈 칸이라 길잡이 문구가 다시 보이고, 앞 작업에 쓰던 말은 화면에 없다.
        compose.onNodeWithText("직접 쓰거나 위에서 골라요").assertIsDisplayed()
        compose.onAllNodesWithText("막힌 작업에 쓰던 말").assertCountEquals(0)
    }

    private fun widthOf(tag: String): Float {
        val bounds = compose.onNodeWithTag(tag).getUnclippedBoundsInRoot()
        return (bounds.right - bounds.left).value
    }

    // 고른 문장은 그 선택지를 본 작업 version 과 함께 저장으로 간다. 그리고 그
    // 초안을 사람이 고치면 출처가 떨어져 자유 입력으로 간다 — 세션에게 「내
    // 문장이 돌아왔다」고 잘못 말하지 않는다.
    @Test
    fun aPickedSentenceCarriesItsBasisAndAnEditedOneDoesNot() {
        start(withChoice())
        openBlocked()

        compose.onAllNodesWithTag("chat.tasks.detail.choice").onFirst().performClick()
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()
        assertEquals("고른 선택지의 id 와 기준 version 이 함께 간다", "c-ci" to 3L, sentPick)

        // 같은 문장을 고른 뒤 한 글자를 덧붙이면 그것은 사람의 말이다.
        compose.onAllNodesWithTag("chat.tasks.detail.choice").onFirst().performClick()
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("!")
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.store").performClick()
        compose.waitForIdle()
        assertEquals("고친 초안은 출처 없이 간다", "" to 0L, sentPick)
    }

    // 작업이 움직여 선택지가 낡았다는 거절은 다른 실패와 다른 말을 하고, 쓴 말을
    // 지우지 않는다 — 그대로 다시 저장할 수 있다.
    @Test
    fun aStalePickSaysSoAndKeepsTheWords() {
        start(taskRequest = ChatTaskRequestUiState(taskId = "t-blocked", error = ChatTaskRequestError.STALE_CHOICE))
        openBlocked()

        compose.onNodeWithTag("chat.tasks.detail.field").performTextInput("내가 쓴 요청")
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.error").assertTextContains("선택지", substring = true)
        compose.onNodeWithTag("chat.tasks.detail.field").assertTextContains("내가 쓴 요청", substring = true)
        compose.onNodeWithText("다시 저장하기").assertIsDisplayed()
    }

    /** 막힌 작업에 선택지 하나를 올려 둔 원장. */
    private fun withChoice(): ChatTasksUiState = ledger.copy(
        tasks = ledger.tasks.map {
            if (it.id == "t-blocked") {
                it.copy(
                    choices = listOf(
                        ChatTaskChoice(
                            id = "c-ci",
                            label = "CI 로그 확인 요청",
                            requestText = "방금 실패한 CI 로그에서 어디서 멈췄는지 알려 줘.",
                            basisVersion = 3,
                        ),
                    ),
                )
            } else {
                it
            }
        },
    )

    private fun assertRatio(what: String, got: Float, want: Float) {
        assertTrue("$what 가 창 폭의 $want 배여야 한다 (실측 $got)", abs(got - want) <= 0.02f)
    }
}
