package com.dlddu.pocketaide.chat

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.test.DeviceConfigurationOverride
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.ForcedSize
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.assertTextContains
import androidx.compose.ui.test.assertTextEquals
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onAllNodesWithTag
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performScrollTo
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

// 검토 자료가 **실제 진입점의 화면에 그려지는지** 관찰한다. 진입점은 확장 창
// 작업 패널의 작업 상세(AC10)이고, 이 시험은 그 상세 자체를 띄운다.
//
// 세 가지를 가른다.
//
//  (1) **양성** — 기록에 명시된 두 대안이 각자의 장점·단점을 갖고, 서로 다른
//      내용으로 **비교 가능하게** 보이고, 각 줄이 **출처 턴**을 말한다. 결정과
//      미결 질문은 서로 다른 칸에 있다.
//  (2) **음성** — 일반 설명과 선택지 버튼만 있는 작업에서는 장단점·결정이
//      「기록 없음」으로 보인다. 버튼 문구가 장점으로 올라오지 않는다.
//  (3) **짧은 후속 뒤의 회귀** — 예전 턴이 올린 PR 들과 자료가 마지막 턴의
//      「수정해 주세요」 뒤에도 화면에 남고, 출처 턴도 남는다.
//
// 그리고 #425 의 약속을 함께 고정한다: 검토 자료가 화면에 있는 동안에도
// 「주고받은 내용」은 **접힌 채**이고, 펼칠 때 비로소 이력을 읽는다.
@OptIn(ExperimentalTestApi::class)
@RunWith(AndroidJUnit4::class)
class ChatTaskMaterialsRenderTest {
    @get:Rule
    val compose = createComposeRule()

    private var materialLoads = 0
    private var turnLoads = 0
    private var opened: ChatTaskArtifact? = null

    private val task = ChatTask(
        id = "t1",
        title = "검토 대상",
        status = ChatTaskStatus.REVIEW,
        turnsTotal = 3,
    )

    // 기록에 명시적으로 적혀 있던 자료. 두 대안은 **양쪽 모두** 다른 내용이고,
    // 둘째 대안은 단점이 적히지 않아 그 한 칸만 미기록이다.
    private val recorded = ChatTaskMaterials(
        taskId = "t1",
        sessionId = "worker-1",
        artifacts = listOf(
            ChatTaskArtifact(
                ref = "https://github.com/dlddu/pocket-aide/pull/430",
                url = "https://github.com/dlddu/pocket-aide/pull/430",
                repo = "dlddu/pocket-aide",
                number = 430,
                turnId = 2,
            ),
            ChatTaskArtifact(
                ref = "https://github.com/dlddu/pocket-aide/pull/425",
                url = "https://github.com/dlddu/pocket-aide/pull/425",
                title = "주고받은 내용을 펼칠 때 읽는다",
                repo = "dlddu/pocket-aide",
                number = 425,
                turnId = 1,
            ),
            ChatTaskArtifact(ref = "pull/42", turnId = 1),
        ),
        tradeoffs = listOf(
            ChatTaskTradeoff(
                alternative = "백엔드에서 접기",
                pros = listOf("모바일이 이력을 받지 않는다"),
                cons = listOf("서버 메모리를 쓴다"),
                turnId = 1,
                source = "response",
            ),
            ChatTaskTradeoff(
                alternative = "앱에서 접기",
                pros = listOf("서버 상태가 없다"),
                cons = emptyList(),
                turnId = 1,
                source = "response",
            ),
        ),
        decisions = listOf(
            ChatTaskDecision(text = "접기는 백엔드에서 한다", turnId = 1, source = "explanation"),
        ),
        open = listOf(
            ChatTaskOpenPoint(
                turnId = 3,
                question = "어느 쪽을 고칠까요?",
                choices = listOf(ChatTaskChoice("앱만", "앱만 고쳐 주세요")),
            ),
        ),
        tradeoffsRecorded = true,
        decisionsRecorded = true,
        foldedThrough = 3,
    )

    // 아무것도 명시되지 않은 작업. 설명은 결론처럼 읽히고 버튼에는 이유가
    // 괄호로 붙어 있지만, 백엔드는 어느 것도 자료로 읽지 않았다.
    private val nothingMarked = ChatTaskMaterials(
        taskId = "t1",
        sessionId = "worker-1",
        open = listOf(
            ChatTaskOpenPoint(
                turnId = 1,
                question = "A 로 할까요?",
                choices = listOf(
                    ChatTaskChoice("A 로 간다 (빠름)", "A 로 해 주세요"),
                    ChatTaskChoice("B 로 간다 (저렴)", "B 로 해 주세요"),
                ),
            ),
        ),
        tradeoffsRecorded = false,
        decisionsRecorded = false,
        foldedThrough = 1,
    )

    /** 실제 상세 창. 창은 영역 전체가 들어갈 만큼 높게 둬서 레이아웃 탓에 판정이 깨지지 않게 한다. */
    private fun start(materials: ChatTaskMaterials, open: Boolean = false) {
        compose.setContent {
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(DpSize(900.dp, 3600.dp))) {
                var draft by remember { mutableStateOf("") }
                var turnsOpen by remember { mutableStateOf(open) }
                ChatTaskDetailPane(
                    task = task,
                    draft = draft,
                    onDraftChange = { draft = it },
                    onChoicePicked = {},
                    request = ChatTaskRequestUiState(),
                    onStore = {},
                    turns = ChatTaskTurnsUiState(),
                    turnsOpen = turnsOpen,
                    onToggleTurns = { turnsOpen = !turnsOpen },
                    onLoadTurns = { turnLoads++ },
                    onRetryTurns = {},
                    onLoadMoreTurns = {},
                    materials = ChatTaskMaterialsUiState(
                        taskId = task.id,
                        materials = materials,
                        hasLoaded = true,
                    ),
                    onLoadMaterials = { materialLoads++ },
                    onRetryMaterials = {},
                    onOpenArtifact = { opened = it },
                    onBack = {},
                )
            }
        }
        compose.waitForIdle()
    }

    // (1) 양성 — 두 대안이 각자의 장점·단점을 갖고 나란히 비교된다.
    @Test
    fun bothRecordedAlternativesShowTheirOwnProsAndCons() {
        start(recorded)

        compose.onNodeWithTag("chat.tasks.detail.materials").performScrollTo()
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.0.alternative")
            .assertTextEquals("백엔드에서 접기")
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.1.alternative")
            .assertTextEquals("앱에서 접기")

        // 각 대안의 두 칸이 자기 내용으로 보인다.
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.0.pros.0")
            .assertTextContains("모바일이 이력을 받지 않는다", substring = true)
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.0.cons.0")
            .assertTextContains("서버 메모리를 쓴다", substring = true)
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.1.pros.0")
            .assertTextContains("서버 상태가 없다", substring = true)

        // 둘째 대안의 단점은 적혀 있지 않았으므로 그 **한 칸만** 미기록이다 —
        // 빈칸으로 두면 「단점이 없다」로 읽힌다.
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.1.cons.missing")
            .assertIsDisplayed()
        assertEquals(
            0,
            compose.onAllNodesWithTag("chat.tasks.detail.materials.tradeoff.0.cons.missing")
                .fetchSemanticsNodes().size,
        )

        // 각 줄이 출처 턴을 말한다.
        compose.onNodeWithTag("chat.tasks.detail.materials.tradeoff.0.source")
            .assertTextContains("턴 1", substring = true)
        // 장단점이 「기록 없음」으로 보이지 않는다.
        assertEquals(
            0,
            compose.onAllNodesWithTag("chat.tasks.detail.materials.noTradeoffs")
                .fetchSemanticsNodes().size,
        )
    }

    // (1) 양성 — 결정과 미결 질문이 서로 다른 칸에 있고, 각각 출처가 있다.
    @Test
    fun theRecordedDecisionAndTheOpenQuestionAreInDifferentPlaces() {
        start(recorded)

        compose.onNodeWithTag("chat.tasks.detail.materials.decision.0.text")
            .performScrollTo().assertTextEquals("접기는 백엔드에서 한다")
        compose.onNodeWithTag("chat.tasks.detail.materials.decision.0.source")
            .assertTextContains("턴 1", substring = true)

        compose.onNodeWithTag("chat.tasks.detail.materials.open.0.question")
            .performScrollTo().assertTextEquals("어느 쪽을 고칠까요?")

        // 결정 칸이 「기록 없음」으로 보이지 않는다.
        assertEquals(
            0,
            compose.onAllNodesWithTag("chat.tasks.detail.materials.noDecisions")
                .fetchSemanticsNodes().size,
        )
    }

    // (3) 짧은 후속 뒤의 회귀 — 마지막 턴이 「수정해 주세요」로 아무것도 만들지
    // 않았어도 예전 턴의 PR 들이 화면에 남고, 제목이 있는 것은 제목으로, 없는
    // 것은 식별자로 보인다.
    @Test
    fun theEarlierPullRequestsStayOnScreenBehindAShortFollowUp() {
        start(recorded)

        compose.onNodeWithTag("chat.tasks.detail.materials.artifact.0.label")
            .performScrollTo().assertTextEquals("dlddu/pocket-aide#430")
        compose.onNodeWithTag("chat.tasks.detail.materials.artifact.1.label")
            .assertTextEquals("주고받은 내용을 펼칠 때 읽는다")
        // 주소가 적혀 있지 않은 산출물은 식별자로 보이고, 눌러도 아무 데로
        // 가지 않는다.
        compose.onNodeWithTag("chat.tasks.detail.materials.artifact.2.label")
            .assertTextEquals("pull/42")
        compose.onNodeWithTag("chat.tasks.detail.materials.artifact.2.ref")
            .assertTextContains("열 수 없어요", substring = true)
        compose.onNodeWithTag("chat.tasks.detail.materials.artifact.2").performClick()
        assertEquals(null, opened)

        // 열 수 있는 줄은 그 줄의 주소로 연다.
        compose.onNodeWithTag("chat.tasks.detail.materials.artifact.0").performClick()
        assertEquals("https://github.com/dlddu/pocket-aide/pull/430", opened?.url)
    }

    // (2) 음성 — 설명이 결론처럼 읽히고 버튼에 이유가 붙어 있어도, 장단점과
    // 결정은 「기록 없음」으로 보인다.
    @Test
    fun proseAndButtonsAreShownAsUnrecordedRatherThanAsMaterial() {
        start(nothingMarked)

        compose.onNodeWithTag("chat.tasks.detail.materials.noTradeoffs")
            .performScrollTo().assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.materials.noDecisions")
            .performScrollTo().assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.materials.noArtifacts")
            .performScrollTo().assertIsDisplayed()

        // 버튼 문구가 장점 줄로 올라오지 않았다.
        assertEquals(
            0,
            compose.onAllNodesWithTag("chat.tasks.detail.materials.tradeoff.0.pros.0")
                .fetchSemanticsNodes().size,
        )
        assertEquals(
            0,
            compose.onAllNodesWithTag("chat.tasks.detail.materials.decision.0.text")
                .fetchSemanticsNodes().size,
        )
        // 아직 묻는 것은 그대로 보인다 — 자료가 없다는 것이 질문이 없다는 뜻은
        // 아니다.
        compose.onNodeWithTag("chat.tasks.detail.materials.open.0.question")
            .performScrollTo().assertTextEquals("A 로 할까요?")
    }

    // #425 의 약속: 검토 자료가 화면에 있어도 이력은 접힌 채 시작하고, 접힌
    // 동안에는 한 줄도 읽지 않는다. 펼칠 때 비로소 읽는다.
    @Test
    fun theHistoryStaysCollapsedWhileTheMaterialsAreOnScreen() {
        start(recorded)

        // 자료는 읽혔고(상세를 열었으므로), 이력은 아직 아니다.
        assertTrue(materialLoads > 0)
        assertEquals(0, turnLoads)
        assertEquals(
            0,
            compose.onAllNodesWithTag("chat.tasks.detail.turns.empty").fetchSemanticsNodes().size,
        )

        compose.onNodeWithTag("chat.tasks.detail.turns.toggle").performScrollTo().performClick()
        compose.waitForIdle()

        // 명시적으로 펼친 뒤에야 이력을 읽는다.
        assertEquals(1, turnLoads)
    }
}
