package com.dlddu.pocketaide.chat

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.getUnclippedBoundsInRoot
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performTouchInput
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import kotlin.math.abs

// 이전 페이지를 목록 앞에 붙여도 읽던 자리가 그대로인지를 실제 목록에서 본다
// (PRD-AI채팅.AC1 2026-11-28 개정 · AC9 「앞선 메시지도 제자리에 있다」). 판정은
// 화면 좌표로만 한다 — 읽고 있던 말풍선이 붙이기 전과 같은 자리에 있는지.
@RunWith(AndroidJUnit4::class)
class ChatOlderPagePrependTest {
    @get:Rule
    val compose = createComposeRule()

    private val state = mutableStateOf(firstPage())
    private var loadOlderTaps = 0

    @Test
    fun prependingThePreviousPageLeavesTheReadingPositionAlone() {
        start()
        dragTowardEarlierText()
        // 끝을 따라가던 목록을 손가락으로 올려, 읽는 자리가 목록의 중간에 있게
        // 한다 — 끝에 붙어 있으면 자리가 안 움직인 것이 우연일 수 있다.
        bubbleNode(ANCHOR_ID).assertIsDisplayed()
        val readingAt = anchor().top

        prependOlderPage()

        val movedBy = abs((anchor().top - readingAt).value)
        assertTrue("이전 페이지를 붙이자 읽던 자리가 ${movedBy}dp 움직였다", movedBy <= EPSILON.value)
    }

    @Test
    fun theOlderRowSitsAtTheTopAndAsksForThePreviousPage() {
        start()
        // 과거는 목록의 위에 있으므로 손잡이도 위에 있다 — 끝까지 올려야 보인다.
        dragToTheTop()
        compose.onNodeWithTag("chat.older.more").assertIsDisplayed().performClick()
        compose.waitForIdle()

        assertEquals(1, loadOlderTaps)
    }

    // 불러오는 중·실패(다시 시도)·더 있음은 서로 다른 화면이고, 어느 것도 보던
    // 대화를 화면에서 치우지 않는다.
    @Test
    fun fetchingAndFailingAreDifferentAndNeitherTakesTheConversationAway() {
        start()
        dragToTheTop()

        set(state.value.copy(isLoadingOlder = true))
        compose.onNodeWithTag("chat.older.loading").assertIsDisplayed()
        bubbleNode(FIRST_PAGE_FROM).assertIsDisplayed()

        set(state.value.copy(isLoadingOlder = false, olderError = ChatError.UNREACHABLE))
        compose.onNodeWithTag("chat.older.error").assertIsDisplayed()
        bubbleNode(FIRST_PAGE_FROM).assertIsDisplayed()
        compose.onNodeWithTag("chat.older.retry").performClick()
        compose.waitForIdle()
        assertEquals(1, loadOlderTaps)
    }

    private fun start() {
        compose.setContent {
            Box(Modifier.size(width = VIEWPORT_WIDTH, height = VIEWPORT_HEIGHT)) {
                Messages(
                    state.value,
                    Modifier.fillMaxSize(),
                    onCancelQueued = {},
                    onLoadOlder = { loadOlderTaps++ },
                )
            }
        }
        compose.waitForIdle()
    }

    private fun set(next: ChatUiState) {
        compose.runOnUiThread { state.value = next }
        compose.waitForIdle()
    }

    private fun prependOlderPage() {
        val older = (OLDER_PAGE_FROM until FIRST_PAGE_FROM).map { bubble(it) }
        set(
            state.value.copy(
                messages = older + state.value.messages,
                hasOlder = false,
                olderCursor = 0,
                isLoadingOlder = false,
            ),
        )
    }

    // 읽고 있던 말풍선. 첫 페이지의 한가운데를 고른다 — 목록의 끝도 시작도
    // 아니어서, 자리가 끝에 붙어 있어 우연히 맞는 일이 없다. 말풍선은 자기 글로
    // 집는다: 화면 순서가 바뀌는 것이 이 시험의 대상이므로 순서로 집을 수 없다.
    private fun anchor() = bubbleNode(ANCHOR_ID).getUnclippedBoundsInRoot()

    private fun bubbleNode(id: Long) = compose.onNodeWithText(opening(id), substring = true)

    private fun dragTowardEarlierText() = dragList(STEP, STEPS)

    private fun dragToTheTop() = dragList(STEP, STEPS * 20)

    private fun dragList(step: androidx.compose.ui.unit.Dp, steps: Int) {
        compose.onNodeWithTag(LIST_TAG).performTouchInput {
            down(center)
            repeat(steps) { moveBy(Offset(0f, step.toPx())) }
            moveBy(Offset.Zero, delayMillis = SETTLE_MILLIS)
            up()
        }
        compose.waitForIdle()
    }

    private companion object {
        const val LIST_TAG = "chat.messages"
        const val STEPS = 6
        const val SETTLE_MILLIS = 300L
        val VIEWPORT_WIDTH = 360.dp
        val VIEWPORT_HEIGHT = 320.dp
        val STEP = 40.dp
        val EPSILON = 2.dp

        const val OLDER_PAGE_FROM = 1L
        const val FIRST_PAGE_FROM = 11L
        const val LAST_ID = 20L
        const val ANCHOR_ID = 16L

        // 말풍선마다 자기만의 첫 줄을 둬, 어느 말풍선을 재는지가 글로 정해진다.
        fun opening(id: Long) = "$id 번째 메시지"

        fun bubble(id: Long) = ChatMessage(
            id = id,
            role = ChatMessage.ROLE_USER,
            content = opening(id) + "\n두 줄로 적어 자리를 차지하게 한다",
        )

        fun firstPage() = ChatUiState(
            messages = (FIRST_PAGE_FROM..LAST_ID).map { bubble(it) },
            hasLoaded = true,
            hasOlder = true,
            olderCursor = FIRST_PAGE_FROM,
        )
    }
}
