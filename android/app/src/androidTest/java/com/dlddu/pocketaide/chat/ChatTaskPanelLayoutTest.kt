package com.dlddu.pocketaide.chat

import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.test.DeviceConfigurationOverride
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.ForcedSize
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.getUnclippedBoundsInRoot
import androidx.compose.ui.test.junit4.ComposeContentTestRule
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onAllNodesWithTag
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import kotlin.math.abs

// 창 폭만으로 레이아웃이 갈리는지 화면 좌표로 관찰한다 — 기기 종류나 방향을 묻지
// 않는다. 좁은 창은 대화만, 넓은 창은 대화 왼쪽·작업 패널 오른쪽 1:1.
//
// 폭은 `ForcedSize` 로 준다. `Modifier.size` 는 부모(=에뮬레이터 화면) 제약에
// 눌려 600dp 를 넘지 못하므로 넓은 창을 흉내낼 수 없다. 판정은 화면 절대값이
// 아니라 **창 자신의 폭에 대한 비율**로 해, ForcedSize 가 내용을 축소해 담아도
// 그대로 성립한다. 분기 경계값 자체는 ChatTasksTest 가 따로 본다.
@OptIn(ExperimentalTestApi::class)
@RunWith(AndroidJUnit4::class)
class ChatTaskPanelLayoutTest {
    @get:Rule
    val compose = createComposeRule()

    private val tasks = mutableStateOf(
        ChatTasksUiState(
            sessionId = "cx1",
            tasks = listOf(ChatTask(id = "t1", title = "막힌 작업", status = ChatTaskStatus.BLOCKED, stage = "review")),
            counts = mapOf(ChatTaskStatus.BLOCKED to 1),
            hasLoaded = true,
        ),
    )

    /** 실제 가용 폭이 width 인 창에서의 진짜 레이아웃. */
    private fun start(width: Dp, panelOpen: Boolean = true) {
        compose.setContent {
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(DpSize(width, 800.dp))) {
                ChatPanes(
                    state = ChatUiState(hasLoaded = true),
                    tasks = tasks.value,
                    draft = "",
                    onDraftChange = {},
                    onSend = {},
                    onRefresh = {},
                    onDismissNotice = {},
                    listState = rememberLazyListState(),
                    panelOpen = panelOpen,
                    onTogglePanel = {},
                    // 패널의 조회는 뷰모델 몫이다. 이 테스트가 보는 것은 레이아웃이다.
                    pollTasks = {},
                    taskListState = rememberLazyListState(),
                    // 보는 방식과 고른 작업은 상호작용 테스트가 본다. 여기서는
                    // 목록이 그려진 기본 상태의 레이아웃만 본다.
                    filter = ChatTaskFilter(),
                    onFilterChange = {},
                    openTask = null,
                    onOpenTask = {},
                    onCloseTask = {},
                    taskRequest = ChatTaskRequestUiState(),
                    requestDraft = "",
                    onRequestDraftChange = {},
                    onStoreRequest = {},
                )
            }
        }
        compose.waitForIdle()
    }

    /** 창(= 화면 분기를 재는 그 상자) 자신의 폭. 비율 판정의 기준이다. */
    private fun windowWidth(): Float = compose.widthOf("chat.screen")

    @Test
    fun aNarrowWindowShowsTheConversationOnly() {
        start(599.dp)
        compose.onNodeWithTag("chat.messages.pane").assertIsDisplayed()
        compose.onAllNodesWithTagCount("chat.tasks.panel", 0)
        // 대화가 창 전체를 쓴다.
        assertRatio("좁은 창의 대화", compose.widthOf("chat.messages.pane") / windowWidth(), 1f)
    }

    @Test
    fun aWideWindowSplitsTheConversationAndThePanelOneToOne() {
        start(800.dp)
        compose.onNodeWithTag("chat.messages.pane").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.panel").assertIsDisplayed()

        val left = compose.onNodeWithTag("chat.messages.pane").getUnclippedBoundsInRoot()
        val right = compose.onNodeWithTag("chat.tasks.panel").getUnclippedBoundsInRoot()
        // 대화가 왼쪽, 패널이 오른쪽이다.
        assertTrue("대화가 패널 왼쪽에 있어야 한다 (${left.left} < ${right.left})", left.left < right.left)
        // 1:1 — 각 창이 창 폭의 절반이다(구분선 1dp 는 반올림 오차 안에 든다).
        val window = windowWidth()
        assertRatio("대화", (left.right - left.left).value / window, 0.5f)
        assertRatio("작업 패널", (right.right - right.left).value / window, 0.5f)
    }

    // 넓은 창에서도 접으면 패널은 그려지지 않는다 — 가려 두는 것이 아니라 올리지
    // 않는다. 대화는 접힌 동안 창 전체를 쓴다.
    @Test
    fun collapsingThePanelInAWideWindowRemovesIt() {
        start(800.dp, panelOpen = false)
        compose.onNodeWithTag("chat.messages.pane").assertIsDisplayed()
        compose.onAllNodesWithTagCount("chat.tasks.panel", 0)
        assertRatio("접힌 뒤 대화", compose.widthOf("chat.messages.pane") / windowWidth(), 1f)
    }

    // 좁은 창에는 열 두 번째 창이 없으므로 토글 자체를 두지 않는다.
    @Test
    fun aNarrowWindowHasNoPanelToggle() {
        start(599.dp)
        compose.onAllNodesWithTagCount("chat.tasks.toggle", 0)
    }

    @Test
    fun aWideWindowOffersThePanelToggle() {
        start(800.dp)
        compose.onNodeWithTag("chat.tasks.toggle").assertIsDisplayed()
    }

    @Test
    fun theBlockedTaskShowsItsStatusInTheWideLayout() {
        start(800.dp)
        // 작업 카드는 눌러서 상세로 들어가는 버튼이 되었고, 그래서 자손의
        // 의미를 자기에게 합친다(MergeDescendants) — 배지는 합쳐진 트리에 자기
        // 노드가 없으므로 합치지 않은 트리에서 본다.
        compose.onNodeWithTag("chat.tasks.item.status.blocked", useUnmergedTree = true).assertIsDisplayed()
        // 상태별 개수는 그 상태로 좁히는 칩이 들고 있다.
        compose.onNodeWithTag("chat.tasks.filter.blocked").assertIsDisplayed()
    }

    @Test
    fun anUnreadableLedgerSaysSoInsteadOfLookingEmpty() {
        tasks.value = ChatTasksUiState(hasLoaded = true, error = ChatTaskError.UNAVAILABLE)
        start(800.dp)
        compose.onNodeWithTag("chat.tasks.error").assertIsDisplayed()
        compose.onAllNodesWithTagCount("chat.tasks.empty", 0)
        compose.onAllNodesWithTagCount("chat.tasks.list", 0)
    }

    @Test
    fun anEmptyLedgerSaysThereIsNoWorkYet() {
        tasks.value = ChatTasksUiState(sessionId = "cx1", hasLoaded = true)
        start(800.dp)
        compose.onNodeWithTag("chat.tasks.empty").assertIsDisplayed()
        compose.onAllNodesWithTagCount("chat.tasks.error", 0)
    }

    @Test
    fun aTruncatedListSaysItIsPartial() {
        tasks.value = tasks.value.copy(truncated = true)
        start(800.dp)
        compose.onNodeWithTag("chat.tasks.truncated").assertIsDisplayed()
    }

    private fun assertRatio(what: String, got: Float, want: Float) {
        assertTrue("$what 가 창 폭의 ${want} 배여야 한다 (실측 $got)", abs(got - want) <= 0.02f)
    }
}

private fun ComposeContentTestRule.widthOf(tag: String): Float {
    val b = onNodeWithTag(tag).getUnclippedBoundsInRoot()
    return (b.right - b.left).value
}

private fun ComposeContentTestRule.onAllNodesWithTagCount(tag: String, want: Int) {
    val found = onAllNodesWithTag(tag).fetchSemanticsNodes().size
    assertEquals("`$tag` 가 $want 개여야 한다", want, found)
}
