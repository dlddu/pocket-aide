package com.dlddu.pocketaide.chat

import android.content.ClipData
import android.content.ClipboardManager
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.test.DeviceConfigurationOverride
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.ForcedSize
import androidx.compose.ui.test.assertTextEquals
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.performClick
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import org.junit.Assert.assertEquals
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

// 작업 상세에서 **제목을 시스템 클립보드로** 가져올 수 있는지, 그리고 가져온
// 것이 그 작업의 제목 **원문 전체**인지 화면에서 관찰한다. 붙여넣는 쪽은 이
// 앱이 아니므로, 「복사했다」는 화면의 말이 아니라 시스템 클립보드에 실제로
// 들어간 글자를 읽어서 판정한다.
//
// 네 가지를 본다: (1) 한 줄에 들어가지 않는 긴 제목도 잘리지 않고 전부
// 들어가는지 — 복사되는 값이 그려진 글자가 아니라 작업 행의 제목이어야만
// 참이다, (2) 적혀 있는 공백·줄바꿈까지 **한 글자도 다르지 않은지**, (3) 누른
// 것이 화면에서 확인되고 다른 작업으로 가면 그 말이 사라지는지, (4) 제목이
// 적혀 있지 않은 작업에는 버튼이 없는지 — 그 자리에 대신 서 있는 id 를
// 제목으로 건네지 않는다.
@OptIn(ExperimentalTestApi::class)
@RunWith(AndroidJUnit4::class)
class ChatTaskTitleCopyTest {
    @get:Rule
    val compose = createComposeRule()

    // 휴대폰 폭에서 한 줄에 들어가지 않는 제목. 긴 제목이 화면에서 어떻게
    // 접히든 복사는 그와 무관해야 하므로, 판정의 기대값은 이 문자열 전체다.
    private val longTitle =
        "작업 상세 화면에서 작업 제목을 시스템 클립보드에 복사할 수 있게 한다 — " +
            "이 제목은 한 줄에 들어가지 않을 만큼 길고, 좁은 칸에서는 여러 줄로 접히거나 " +
            "줄임으로 보일 수 있으며, 그렇게 보이는 모습과 원문은 서로 다르다"

    @Test
    fun 긴_제목도_원문_전체가_클립보드에_들어간다() {
        emptyClipboard()
        startOnAPhone(ChatTask(id = "t1", title = longTitle))

        compose.onNodeWithTag("chat.tasks.detail.copyTitle").performClick()
        compose.waitForIdle()

        assertEquals(longTitle, clipboardText())
    }

    /**
     * 제목에 적힌 공백과 줄바꿈까지 그대로 간다. 다듬어서 넣으면 붙여넣은 제목이
     * 원장의 제목과 글자 단위로 달라지고, 그 차이는 붙여넣기 전에는 보이지
     * 않는다.
     */
    @Test
    fun 제목의_공백과_줄바꿈까지_그대로_복사한다() {
        emptyClipboard()
        val spaced = "  앞뒤에 공백이 있고\n줄바꿈도 든 제목  "
        startOnAPhone(ChatTask(id = "t2", title = spaced))

        compose.onNodeWithTag("chat.tasks.detail.copyTitle").performClick()
        compose.waitForIdle()

        assertEquals(spaced, clipboardText())
    }

    /**
     * 클립보드는 눌러도 화면이 바뀌지 않는 일이라, 라벨이 바뀌는 것이 눌렸다는
     * 유일한 표시다. 그리고 그 표시는 그 작업의 것이다 — 다른 작업으로 넘어간
     * 화면이 「복사됨」을 들고 있으면, 사람은 지금 보는 제목이 클립보드에 있다고
     * 읽는다.
     */
    @Test
    fun 누르면_복사됨으로_바뀌고_다른_작업으로_가면_되돌아온다() {
        emptyClipboard()
        val shown = mutableStateOf(ChatTask(id = "t1", title = "첫 작업"))
        compose.setContent {
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(PHONE)) {
                Pane(shown.value)
            }
        }

        compose.onNodeWithTag("chat.tasks.detail.copyTitle").assertTextEquals("제목 복사")
        compose.onNodeWithTag("chat.tasks.detail.copyTitle").performClick()
        compose.waitForIdle()
        compose.onNodeWithTag("chat.tasks.detail.copyTitle").assertTextEquals("복사됨")
        assertEquals("첫 작업", clipboardText())

        // 같은 칸이 다른 작업의 상세로 바뀌는 것 — 패널이 작업을 바꿀 때 하는
        // 일이고, 재구성만 일으키는 것과는 다른 상황이다.
        compose.runOnUiThread { shown.value = ChatTask(id = "t2", title = "다음 작업") }
        compose.waitForIdle()

        compose.onNodeWithTag("chat.tasks.detail.title").assertTextEquals("다음 작업")
        compose.onNodeWithTag("chat.tasks.detail.copyTitle").assertTextEquals("제목 복사")
        // 작업을 옮긴 것으로 사람이 복사해 둔 것을 치우지는 않는다 — 클립보드는
        // 앱의 것이 아니다.
        assertEquals("첫 작업", clipboardText())
    }

    /**
     * 제목이 적혀 있지 않은 작업. 화면의 제목 자리에는 id 가 서지만 그것은
     * 읽기 위한 대역이고, 「제목 복사」가 id 를 건네면 적히지 않은 것이 적힌
     * 것으로 바뀐다.
     */
    @Test
    fun 제목이_없는_작업에는_복사_버튼이_없다() {
        startOnAPhone(ChatTask(id = "t3", title = ""))

        compose.onNodeWithTag("chat.tasks.detail.title").assertTextEquals("t3")
        compose.onNodeWithTag("chat.tasks.detail.copyTitle").assertDoesNotExist()
    }

    private fun startOnAPhone(task: ChatTask) {
        compose.setContent {
            DeviceConfigurationOverride(DeviceConfigurationOverride.ForcedSize(PHONE)) {
                Pane(task)
            }
        }
    }

    /** 이 시험이 묻지 않는 것(요청 저장·이력·검토 자료·선택지)은 모두 가만히 있는 상세 창. */
    @Composable
    private fun Pane(task: ChatTask) {
        var draft by remember { mutableStateOf("") }
        ChatTaskDetailPane(
            task = task,
            draft = draft,
            onDraftChange = { draft = it },
            onChoicePicked = {},
            request = ChatTaskRequestUiState(),
            onStore = {},
            turns = ChatTaskTurnsUiState(),
            turnsOpen = false,
            onToggleTurns = {},
            onLoadTurns = {},
            onRetryTurns = {},
            onLoadMoreTurns = {},
            // 이 작업의 것으로 **이미 다 읽은** 빈 자료. 읽는 중(스피너)으로
            // 두면 이 시험과 무관한 상태가 화면에 남는다.
            materials = ChatTaskMaterialsUiState(taskId = task.id, hasLoaded = true),
            onLoadMaterials = {},
            onRetryMaterials = {},
            onOpenArtifact = {},
            onBack = {},
        )
    }

    /**
     * 시스템 클립보드에 지금 들어 있는 글자. 클립보드를 읽는 쪽은 포커스를 가진
     * 앱이어야 하므로(Android 10+), 계측 대상 앱의 메인 스레드에서 읽는다.
     */
    private fun clipboardText(): String? {
        val instrumentation = InstrumentationRegistry.getInstrumentation()
        val context = instrumentation.targetContext
        var text: String? = null
        instrumentation.runOnMainSync {
            val clip = context.getSystemService(ClipboardManager::class.java)?.primaryClip
            text = clip?.takeIf { it.itemCount > 0 }?.getItemAt(0)?.coerceToText(context)?.toString()
        }
        return text
    }

    /**
     * 앞선 시험이나 기기에 남아 있던 내용을 치운다 — 치우지 않으면 「복사됐다」는
     * 판정이 이 시험이 넣지 않은 글자로도 통과할 수 있다. 비우는 길로 빈 클립을
     * 넣는 것은 `clearPrimaryClip` 이 minSdk 보다 높은 API 28 부터라서다.
     */
    private fun emptyClipboard() {
        val instrumentation = InstrumentationRegistry.getInstrumentation()
        val context = instrumentation.targetContext
        instrumentation.runOnMainSync {
            context.getSystemService(ClipboardManager::class.java)
                ?.setPrimaryClip(ClipData.newPlainText("", ""))
        }
    }

    private companion object {
        val PHONE = DpSize(411.dp, 731.dp)
    }
}
