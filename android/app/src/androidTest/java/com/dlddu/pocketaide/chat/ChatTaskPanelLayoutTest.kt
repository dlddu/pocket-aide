package com.dlddu.pocketaide.chat

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.getUnclippedBoundsInRoot
import androidx.compose.ui.test.junit4.ComposeContentTestRule
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onAllNodesWithTag
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import kotlin.math.abs

// 실제 창 폭만으로 레이아웃이 갈리는지 화면 좌표로 관찰한다 — 기기 종류나 방향을
// 묻지 않는다. 좁은 창은 대화만, 넓은 창은 대화 왼쪽·작업 패널 오른쪽 1:1.
// 분기 경계값 자체는 ChatTasksTest 가 따로 본다.
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

    /** The real layout under a window of exactly this width, panel open. */
    private fun start(width: Dp, panelOpen: Boolean = true) {
        compose.setContent {
            Box(Modifier.fillMaxSize()) {
                Box(Modifier.size(width, 800.dp)) {
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
                        // The panel's reads are the ViewModel's; the layout is
                        // what this test observes.
                        pollTasks = {},
                    )
                }
            }
        }
        compose.waitForIdle()
    }

    @Test
    fun aNarrowWindowShowsTheConversationOnly() {
        start(599.dp)
        compose.onNodeWithTag("chat.messages.pane").assertIsDisplayed()
        compose.onAllNodesWithTagCount("chat.tasks.panel", 0)
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
        // 1:1 — 구분선 1dp 를 빼면 두 폭이 같다. 반올림 오차까지 허용한다.
        val leftWidth = (left.right - left.left).value
        val rightWidth = (right.right - right.left).value
        assertTrue(
            "두 창의 폭이 1:1 이어야 한다 (대화 $leftWidth, 패널 $rightWidth)",
            abs(leftWidth - rightWidth) <= 2f,
        )
    }

    // 넓은 창에서도 접으면 패널은 그려지지 않는다 — 가려 두는 것이 아니라 올리지
    // 않는다. 대화는 접힌 동안 창 전체를 쓴다.
    @Test
    fun collapsingThePanelInAWideWindowRemovesIt() {
        start(800.dp, panelOpen = false)
        compose.onNodeWithTag("chat.messages.pane").assertIsDisplayed()
        compose.onAllNodesWithTagCount("chat.tasks.panel", 0)
        val pane = compose.onNodeWithTag("chat.messages.pane").getUnclippedBoundsInRoot()
        val paneWidth = (pane.right - pane.left).value
        assertTrue("접힌 뒤 대화가 창 전체를 써야 한다 ($paneWidth)", paneWidth >= 790f)
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
        compose.onNodeWithTag("chat.tasks.item.status.blocked").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.count.blocked").assertIsDisplayed()
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
}

private fun ComposeContentTestRule.onAllNodesWithTagCount(tag: String, want: Int) {
    val found = onAllNodesWithTag(tag).fetchSemanticsNodes().size
    assertEquals("`$tag` 가 $want 개여야 한다", want, found)
}
