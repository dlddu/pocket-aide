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
    private val visible = mutableStateOf(true)
    private val memory = ChatScrollMemory()
    private var conversation = ASKED
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

    @Test
    fun theReadingPositionSurvivesLeavingTheTabAndComingBack() {
        startHistory()
        dragToTheTop()
        val readingAt = firstMessage().top

        leaveAndComeBack()

        val movedBy = abs((firstMessage().top - readingAt).value)
        assertTrue("돌아왔더니 읽던 자리가 ${movedBy}dp 움직였다", movedBy <= EPSILON.value)
    }

    @Test
    fun aVisitThatEndedAtTheNewestMessageComesBackToTheNewestMessage() {
        startHistory()

        leaveAndComeBack()

        assertEndAtTheFold()
    }

    @Test
    fun comingBackAfterTheMessageBeingReadIsGoneLandsAtTheEnd() {
        startHistory()
        dragToTheTop()

        // 읽던 메시지가 없는 이력으로 바뀐다 — 지우기·동기화·새 세션이 모두
        // 이 모양이고, 떠나 있는 동안 일어난다.
        leaveAndComeBack {
            conversation = HISTORY.drop(DROPPED)
            state.value = replying(text)
        }

        assertEndAtTheFold()
    }

    @Test
    fun theListEndButtonOffersTheWayBackAndFollowingResumesFromThere() {
        startHistory()
        compose.onNodeWithTag(LIST_END_TAG).assertDoesNotExist()

        dragToTheTop()
        compose.onNodeWithTag(LIST_END_TAG).assertIsDisplayed()
        compose.onNodeWithTag(LIST_END_TAG).performClick()
        compose.waitForIdle()

        assertEndAtTheFold()
        compose.onNodeWithTag(LIST_END_TAG).assertDoesNotExist()
        append("버튼으로 돌아온 뒤 덧붙은 줄")
        assertEndAtTheFold()
    }

    private fun start(first: String) {
        text = first
        state.value = replying(first)
        compose.setContent {
            Box(Modifier.size(width = VIEWPORT_WIDTH, height = VIEWPORT_HEIGHT)) {
                if (visible.value) Messages(state.value, Modifier.fillMaxSize(), memory)
            }
        }
        compose.waitForIdle()
    }

    private fun startHistory() {
        conversation = HISTORY
        start(SHORT_REPLY)
    }

    // 탭을 떠나면 화면이 컴포지션에서 빠지고(RootScreen 은 탭을 `when` 으로
    // 갈아 끼운다) 목록이 들고 있던 자리도 함께 사라진다 — 돌아오면 목록은
    // 처음부터 다시 만들어진다. `visible` 이 그 떠남과 돌아옴이다.
    private fun leaveAndComeBack(whileAway: () -> Unit = {}) {
        compose.runOnUiThread { visible.value = false }
        compose.waitForIdle()
        compose.runOnUiThread(whileAway)
        compose.waitForIdle()
        compose.runOnUiThread { visible.value = true }
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
        ChatUiState(messages = conversation, pending = ChatPending(content = content), hasLoaded = true)

    private fun viewport() = compose.onNodeWithTag(LIST_TAG).getUnclippedBoundsInRoot()

    private fun bubble(tag: String = PENDING_TAG) = compose.onNodeWithTag(tag).getUnclippedBoundsInRoot()

    private fun assertEndAtTheFold(tag: String = PENDING_TAG) {
        val end = bubble(tag).bottom
        val fold = viewport().bottom
        assertTrue("답의 끝이 화면 아래로 넘어갔다 — 끝 $end, 목록 끝 $fold", end <= fold + EPSILON)
        assertTrue("답의 끝이 화면 아래에 붙지 않았다 — 끝 $end, 목록 끝 $fold", end >= fold - AT_THE_FOLD)
    }

    private fun dragTowardEarlierText() = dragList(STEP)

    // 넘치게 끌어 목록의 처음에 자리를 맞춘다 — 끌리는 거리가 아니라 목록의
    // 끝이 판정의 기준이 되도록.
    private fun dragToTheTop() = dragList(STEP, steps = STEPS * 3)

    private fun firstMessage() = compose.onNodeWithText(FIRST_MESSAGE).getUnclippedBoundsInRoot()

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
        const val LIST_END_TAG = "chat.listend.button"
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

        // 자리 복원은 끝이 아닌 자리에서만 뜻이 있으니, 한 화면에 다 담기지
        // 않는 짧은 메시지 여러 개로 본다. 답은 짧게 둬 끝이 한눈에 보인다.
        const val HISTORY_SIZE = 10
        const val DROPPED = 3
        const val SHORT_REPLY = "짧은 답"
        val HISTORY = (1..HISTORY_SIZE).map {
            ChatMessage(id = it.toLong(), role = ChatMessage.ROLE_USER, content = message(it))
        }
        val FIRST_MESSAGE = message(1)

        fun message(number: Int) = "메시지 $number\n두 번째 줄\n세 번째 줄"

        fun reply(lines: Int = 60) = (1..lines).joinToString("\n") { "답의 $it 번째 줄" }
    }
}
