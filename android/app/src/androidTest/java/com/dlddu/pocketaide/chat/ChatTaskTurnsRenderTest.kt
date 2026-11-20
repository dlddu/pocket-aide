package com.dlddu.pocketaide.chat

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.test.DeviceConfigurationOverride
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.ForcedSize
import androidx.compose.ui.test.assertCountEquals
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.assertTextEquals
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onAllNodesWithTag
import androidx.compose.ui.test.onFirst
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.performClick
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

// 상세 화면이 **합의한 턴 레코드를 실제로 그리는지** 화면에서 관찰한다. 작업
// 패널과 이력 라우트는 옛 원장의 모양이라 지시·세션 원문·부모 설명·산출물을
// 담을 자리가 없었다 — 그것만으로는 「작업이 있다」까지만 그릴 수 있었다.
//
// 네 가지를 본다: (1) 각 부분이 자기 라벨로 **구분되어** 보이는지, (2) 세션의
// 원문과 부모의 설명이 섞이지 않는지, (3) 기록되지 않은 부분이 빈칸으로
// 보이지 않는지, (4) 옮겨온 작업의 옛 이력이 턴과 **구분 보존**되는지.
@OptIn(ExperimentalTestApi::class)
@RunWith(AndroidJUnit4::class)
class ChatTaskTurnsRenderTest {
    @get:Rule
    val compose = createComposeRule()

    private var turnLoads = 0
    private var turnMores = 0
    private var picked: ChatTaskChoice? = null

    private val task = ChatTask(
        id = "t1",
        title = "옮겨 온 작업",
        status = ChatTaskStatus.BLOCKED,
        stage = "review",
    )

    private val finished = ChatTaskTurn(
        id = 1,
        request = "이거 먼저 해 줘",
        instruction = "저장소를 읽고 패치를 만들어라",
        response = "세션이 스스로 쓴 원문 답",
        explanation = "부모가 사람에게 쓴 설명",
        artifacts = listOf("pull/42"),
        requestedAt = "2026-11-19T00:30:00Z",
        startedAt = "2026-11-19T00:31:00Z",
        endedAt = "2026-11-19T00:40:00Z",
        origin = ChatTaskTurn.ORIGIN_LEDGER,
    )

    private val asking = ChatTaskTurn(
        id = 2,
        request = "리뷰 반영도 해 줘",
        question = "어디까지 반영할까요?",
        choices = listOf(ChatTaskTurnChoice("전부", "코멘트 전부 반영해 줘")),
        requestedAt = "2026-11-19T01:00:00Z",
        origin = ChatTaskTurn.ORIGIN_APP,
    )

    private fun start(
        turns: ChatTaskTurnsUiState,
        choices: List<ChatTaskChoice> = emptyList(),
    ) {
        compose.setContent {
            // 상세 창은 스크롤한다. 짧은 창에서는 타임라인이 화면 밖에 남아
            // `assertIsDisplayed` 와 클릭이 닿지 못하므로, 창을 전체가 들어갈
            // 만큼 높게 준다 — 보이는지 묻는 판정이 레이아웃 탓에 깨지지 않게.
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(DpSize(900.dp, 3200.dp))) {
                // remember: 기억하지 않으면 재구성마다 새 상태가 생겨 초안이
                // 사라지고, 「고른 문장이 칸에 들어갔는지」를 물을 수 없다.
                var draft by remember { mutableStateOf("") }
                ChatTaskDetailPane(
                    task = task.copy(choices = choices),
                    draft = draft,
                    onDraftChange = { draft = it },
                    onChoicePicked = {
                        picked = it
                        draft = it.requestText
                    },
                    request = ChatTaskRequestUiState(),
                    onStore = {},
                    history = ChatTaskHistoryUiState(taskId = "t1", hasLoaded = true),
                    onLoadHistory = {},
                    onLoadMoreHistory = {},
                    turns = turns,
                    onLoadTurns = { turnLoads++ },
                    onLoadMoreTurns = { turnMores++ },
                    onBack = {},
                )
            }
        }
    }

    // (1)(2) 각 부분이 자기 라벨 아래 구분되어 보이고, 세션의 원문과 부모의
    // 설명이 서로의 자리에 들어가지 않는다.
    @Test
    fun everyPartOfATurnIsShownUnderItsOwnLabel() {
        start(
            ChatTaskTurnsUiState(
                taskId = "t1",
                task = ChatTaskTurnTask(id = "t1", title = "옮겨 온 작업", description = null, completionCriteria = null),
                turns = listOf(finished, asking),
                ledgerHistory = true,
                hasLoaded = true,
            ),
        )
        compose.onNodeWithTag("chat.tasks.detail.turns").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.turn.1.request").assertTextEquals("이거 먼저 해 줘")
        compose.onNodeWithTag("chat.tasks.detail.turn.1.instruction")
            .assertTextEquals("저장소를 읽고 패치를 만들어라")
        compose.onNodeWithTag("chat.tasks.detail.turn.1.response")
            .assertTextEquals("세션이 스스로 쓴 원문 답")
        compose.onNodeWithTag("chat.tasks.detail.turn.1.explanation")
            .assertTextEquals("부모가 사람에게 쓴 설명")
        compose.onNodeWithTag("chat.tasks.detail.turn.1.artifact.0").assertTextEquals("pull/42")
        // 끝난 턴은 끝난 시각으로 보인다.
        compose.onNodeWithTag("chat.tasks.detail.turn.1.when").assertTextEquals("2026-11-19T00:40:00Z")
        // 옮겨온 턴은 「이전 기록」으로 표시돼, 여기서 쓴 것처럼 보이지 않는다.
        compose.onNodeWithTag("chat.tasks.detail.turn.1.carried").assertIsDisplayed()

        // 질문과 선택지는 그 턴 아래에 있고, 아직 답을 기다린다고 말한다.
        compose.onNodeWithTag("chat.tasks.detail.turn.2.question").assertTextEquals("어디까지 반영할까요?")
        compose.onNodeWithTag("chat.tasks.detail.turn.2.choice.0").assertTextEquals("전부")
        compose.onNodeWithTag("chat.tasks.detail.turn.2.open").assertIsDisplayed()
    }

    // (3) 기록되지 않은 부분은 빈칸으로 그리지 않는다. 턴이 요청만 들고 있는
    // 것은 평범한 경우이고 결손이 아니다.
    @Test
    fun aPartTheRecordsDoNotHaveIsNotDrawnAsABlank() {
        start(
            ChatTaskTurnsUiState(
                taskId = "t1",
                turns = listOf(ChatTaskTurn(id = 5, request = "방금 보낸 말", requestedAt = "2026-11-19T02:00:00Z")),
                hasLoaded = true,
            ),
        )
        compose.onNodeWithTag("chat.tasks.detail.turn.5.request").assertTextEquals("방금 보낸 말")
        compose.onAllNodesWithTag("chat.tasks.detail.turn.5.response").assertCountEquals(0)
        compose.onAllNodesWithTag("chat.tasks.detail.turn.5.explanation").assertCountEquals(0)
        compose.onAllNodesWithTag("chat.tasks.detail.turn.5.instruction").assertCountEquals(0)
        // 시각은 요청이 도착한 때로 보인다.
        compose.onNodeWithTag("chat.tasks.detail.turn.5.when").assertTextEquals("2026-11-19T02:00:00Z")
    }

    // 턴이 0개인 작업은 깨지지 않고 「없다」고 말한다 — 새로 만든 작업의
    // 평범한 모습이다.
    @Test
    fun aTaskWithNoTurnsSaysSo() {
        start(ChatTaskTurnsUiState(taskId = "t1", hasLoaded = true))
        compose.onNodeWithTag("chat.tasks.detail.turns.empty").assertIsDisplayed()
        assertEquals("첫 페이지는 창이 열릴 때 읽는다", 1, turnLoads)
    }

    // 작업의 무엇을·왜와 완료 기준이 보이고, 기록이 없으면 없다고 말한다.
    @Test
    fun theAgreedTaskFieldsAreShownAndNullSaysSo() {
        start(
            ChatTaskTurnsUiState(
                taskId = "t1",
                task = ChatTaskTurnTask(
                    id = "t1", title = "새 작업",
                    description = "무엇을 why", completionCriteria = "머지되면 끝",
                ),
                hasLoaded = true,
            ),
        )
        compose.onNodeWithTag("chat.tasks.detail.description").assertTextEquals("무엇을 why")
        compose.onNodeWithTag("chat.tasks.detail.criteria").assertTextEquals("머지되면 끝")
    }

    // 옮겨온 작업은 이 둘을 기록한 적이 없다. null 은 빈칸이 아니라 「적혀 있지
    // 않다」로 보여야 한다 — 빈칸은 사람이 답으로 읽는다.
    @Test
    fun aFieldTheRecordsNeverHadSaysSoRatherThanShowingABlank() {
        start(ChatTaskTurnsUiState(taskId = "t1", task = ChatTaskTurnTask(id = "t1"), hasLoaded = true))
        compose.onNodeWithTag("chat.tasks.detail.description").assertTextEquals("원장에 적혀 있지 않아요.")
        compose.onNodeWithTag("chat.tasks.detail.criteria").assertTextEquals("원장에 적혀 있지 않아요.")
    }

    // (4) 옛 원장 이력은 턴과 **다른 영역**으로 남고, 옮겨온 작업에서는 제목이
    // 그것이 이전 기록임을 말한다 — 턴에 섞이지 않는다. 영역 자체는 전과 같이
    // 늘 있다(없는 작업은 그 영역의 빈 상태가 스스로 말한다).
    @Test
    fun theLedgerHistoryIsKeptSeparateFromTheTurns() {
        start(
            ChatTaskTurnsUiState(
                taskId = "t1", turns = listOf(finished), ledgerHistory = true, hasLoaded = true,
            ),
        )
        compose.onNodeWithTag("chat.tasks.detail.turns").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.history").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.history.heading").assertTextEquals("이전 원장 이력")
    }

    // 커토버 뒤에 생긴 작업의 이력은 그 턴들이다. 옛 원장 이력 영역은 전과
    // 같이 있지만 「이전 기록」이라고 말하지 않는다 — 없는 이전 기록을 있는
    // 척하지 않는다.
    @Test
    fun aTaskCreatedAfterTheCutoverDoesNotClaimAnEarlierRecord() {
        start(ChatTaskTurnsUiState(taskId = "t1", turns = listOf(asking), hasLoaded = true))
        compose.onNodeWithTag("chat.tasks.detail.turns").assertIsDisplayed()
        compose.onNodeWithTag("chat.tasks.detail.history.heading").assertTextEquals("작업 이력")
    }

    // 페이지가 더 있으면 더 읽을 수 있고, 누르면 다음 페이지를 요청한다.
    @Test
    fun moreTurnsCanBeRead() {
        start(
            ChatTaskTurnsUiState(
                taskId = "t1", turns = listOf(finished), nextCursor = 1, hasLoaded = true,
            ),
        )
        val before = turnMores
        compose.onNodeWithTag("chat.tasks.detail.turns.more").performClick()
        assertEquals(before + 1, turnMores)
    }

    // 추천 선택을 누르면 **요청 칸이 채워질 뿐** 저장도 실행도 하지 않는다 —
    // 보내는 말은 사람이 읽은 그 말이다.
    @Test
    fun pickingAChoiceOnlyFillsTheRequestBox() {
        val offered = ChatTaskChoice(id = "t2p0", label = "전부", requestText = "코멘트 전부 반영해 줘", basisVersion = 3)
        start(
            ChatTaskTurnsUiState(taskId = "t1", turns = listOf(asking), hasLoaded = true),
            choices = listOf(offered),
        )
        compose.onAllNodesWithTag("chat.tasks.detail.choice").onFirst().performClick()
        assertEquals("고른 문장이 그대로 초안이 된다", offered.requestText, picked?.requestText)
        // 저장도 실행도 하지 않는다: 요청 칸이 그 문장으로 채워질 뿐이다.
        compose.onNodeWithTag("chat.tasks.detail.field").assertTextEquals(offered.requestText)
    }
}
