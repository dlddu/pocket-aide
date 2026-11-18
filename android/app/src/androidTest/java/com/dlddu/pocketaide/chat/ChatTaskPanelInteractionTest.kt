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
    ) {
        compose.setContent {
            // 넓은 창이면서 상태 칩이 한 줄에 다 들어오는 폭 — 칩이 가로
            // 스크롤 밖으로 밀려나면 누를 수 없어 테스트가 레이아웃 탓에 깨진다.
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(DpSize(1200.dp, 900.dp))) {
                var filter by remember { mutableStateOf(ChatTaskFilter()) }
                var openId by remember { mutableStateOf<String?>(null) }
                var request by remember { mutableStateOf("") }
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
                    onRequestDraftChange = { request = it },
                    onStoreRequest = { stores++ },
                )
            }
        }
        compose.waitForIdle()
    }

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

    // 추천을 누르면 보내지 않고 요청 칸을 채운다. 보내는 문장은 사용자가 먼저
    // 읽는다 — 붙어 나갈 줄도 화면에 있다.
    @Test
    fun aSuggestionFillsTheRequestWithoutSendingIt() {
        start()
        openBlocked()

        compose.onAllNodesWithTag("chat.tasks.detail.suggestion").onFirst().performClick()
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.field")
            .assertTextContains("「CI 가 막혔다」", substring = true)
        assertEquals("고르는 것은 저장하는 것이 아니다", 0, stores)
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

    private fun assertRatio(what: String, got: Float, want: Float) {
        org.junit.Assert.assertTrue("$what 가 창 폭의 $want 배여야 한다 (실측 $got)", abs(got - want) <= 0.02f)
    }
}
