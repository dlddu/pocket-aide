package com.dlddu.pocketaide.chat

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.test.getUnclippedBoundsInRoot
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.performTouchInput
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import kotlin.math.abs

// 실제 목록을 손가락으로 끌어 PRD-AI채팅 AC9 를 관찰한다. 판정은 화면 좌표로만
// 한다 — 말풍선의 끝이 목록의 끝에 닿아 있는지, 읽던 자리가 그대로인지. 구현의
// 끝 판정식은 ChatScrollTest 가 따로 본다.
@RunWith(AndroidJUnit4::class)
class ChatScrollBehaviorTest {
    @get:Rule
    val compose = createComposeRule()

    private val state = mutableStateOf(ChatUiState(hasLoaded = true))
    private var text = ""

    @Test
    fun theEndOfAReplyTallerThanTheViewportStaysInView() {
        start(reply())
        append("방금 덧붙은 줄")

        assertEndAtTheFold()
        assertTrue(
            "말풍선의 시작이 아직 화면 안이다 — 긴 답의 첫 줄에 머물러 있다는 뜻이다 (시작 ${bubble().top})",
            bubble().top < viewport().top,
        )
    }

    @Test
    fun appendingWhileTheUserReadsEarlierTextLeavesThePositionAlone() {
        start(reply())
        dragTowardEarlierText()

        assertTrue(
            "손가락으로 올렸는데도 답의 끝이 화면에 남아 있다 — 끌어 올리지 못했다",
            bubble().bottom > viewport().bottom,
        )
        val readingAt = bubble().top
        append("읽는 중에 덧붙은 줄")

        val movedBy = abs((bubble().top - readingAt).value)
        assertTrue("읽던 자리가 ${movedBy}dp 움직였다", movedBy <= EPSILON.value)
    }

    @Test
    fun comingBackToTheEndStartsFollowingAgain() {
        start(reply())
        dragTowardEarlierText()
        append("올려 읽는 동안 덧붙은 줄")
        dragBackToTheEnd()

        append("끝으로 내려온 뒤 덧붙은 줄")
        assertEndAtTheFold()
    }

    @Test
    fun aBurstOfChunksAndThenTheFinishedReplyBothLandAtTheEnd() {
        start(reply())
        // 폴링이 토큰을 쏟는 동안에는 한 변화가 가라앉을 틈이 없다. 시계를 잡고
        // 두 프레임씩만 주면, 자리 이동이 애니메이션이던 때처럼 매번 다시
        // 시작돼 끝에 닿지 못하는 회귀가 그대로 드러난다.
        compose.mainClock.autoAdvance = false
        repeat(CHUNKS) { chunk ->
            grow("조각 $chunk")
            compose.mainClock.advanceTimeByFrame()
            compose.mainClock.advanceTimeByFrame()
        }
        compose.mainClock.autoAdvance = true
        compose.waitForIdle()
        assertEndAtTheFold()

        finish()
        assertEndAtTheFold(ANSWER_TAG)
    }

    private fun start(first: String) {
        text = first
        state.value = replying(first)
        compose.setContent {
            Box(Modifier.size(width = VIEWPORT_WIDTH, height = VIEWPORT_HEIGHT)) {
                Messages(state.value, Modifier.fillMaxSize())
            }
        }
        compose.waitForIdle()
    }

    private fun append(line: String) {
        grow(line)
        compose.waitForIdle()
    }

    private fun grow(line: String) {
        text += "\n$line"
        val grown = text
        compose.runOnUiThread { state.value = replying(grown) }
    }

    private fun finish() {
        val answered = text
        compose.runOnUiThread {
            state.value = ChatUiState(
                messages = ASKED + ChatMessage(id = 2, role = "assistant", content = answered),
                pending = null,
                hasLoaded = true,
            )
        }
        compose.waitForIdle()
    }

    private fun replying(content: String) =
        ChatUiState(messages = ASKED, pending = ChatPending(content = content), hasLoaded = true)

    private fun viewport() = compose.onNodeWithTag(LIST_TAG).getUnclippedBoundsInRoot()

    private fun bubble(tag: String = PENDING_TAG) = compose.onNodeWithTag(tag).getUnclippedBoundsInRoot()

    private fun assertEndAtTheFold(tag: String = PENDING_TAG) {
        val end = bubble(tag).bottom
        val fold = viewport().bottom
        assertTrue("답의 끝이 화면 아래로 넘어갔다 — 끝 $end, 목록 끝 $fold", end <= fold + EPSILON)
        assertTrue("답의 끝이 화면 아래에 붙지 않았다 — 끝 $end, 목록 끝 $fold", end >= fold - AT_THE_FOLD)
    }

    private fun dragTowardEarlierText() = dragList(STEP)

    // 끝에서는 더 끌리지 않으니, 넘치게 끌어 자리를 끝에 맞춘다.
    private fun dragBackToTheEnd() = dragList(-STEP, steps = STEPS * 3)

    private fun dragList(step: Dp, steps: Int = STEPS) {
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
        const val PENDING_TAG = "chat.message.pending"
        const val ANSWER_TAG = "chat.message.ai"
        const val CHUNKS = 12
        const val STEPS = 6
        const val SETTLE_MILLIS = 300L
        val VIEWPORT_WIDTH = 360.dp
        val VIEWPORT_HEIGHT = 320.dp
        val STEP = 40.dp
        val EPSILON = 2.dp

        // 목록의 아래 여백(Spacing.lg) 만큼은 말풍선 끝과 화면 끝 사이에 남는다.
        val AT_THE_FOLD = 40.dp
        val ASKED = listOf(ChatMessage(id = 1, role = ChatMessage.ROLE_USER, content = "긴 답을 주세요"))

        fun reply(lines: Int = 60) = (1..lines).joinToString("\n") { "답의 $it 번째 줄" }
    }
}
