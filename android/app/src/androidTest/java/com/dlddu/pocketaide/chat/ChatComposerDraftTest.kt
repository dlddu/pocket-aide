package com.dlddu.pocketaide.chat

import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.assertTextContains
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performTextClearance
import androidx.compose.ui.test.performTextInput
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

// PRD-AI채팅.AC11 — 쓰다 만 메시지가 탭을 오가도 남는지를 실제 화면으로 본다.
//
// 탭 전환은 RootScreen 이 하는 그대로 분기 하나다 — 저장 상태 홀더가 없으므로
// 채팅 화면은 조립에서 아예 빠졌다 다시 들어온다. 초안이 화면이 아니라 뷰모델에
// 있다는 것이 이 왕복을 견디는 유일한 근거이고, 이 테스트가 그것을 본다.
//
// API 주소는 닿지 않는 로컬 포트다 — 이 시험은 사람이 쓴 말에 대한 것이고
// 서버에는 아무것도 묻지 않는다. 읽기가 실패해도 입력칸은 그대로 그려진다.
// 한글·줄바꿈을 글자 그대로 보관하는지와 복원이 전송을 일으키지 않는지는
// ChatViewModelTest 가 실제로 나간 요청을 세어 본다.
@RunWith(AndroidJUnit4::class)
class ChatComposerDraftTest {
    @get:Rule
    val compose = createComposeRule()

    private val showsChat = mutableStateOf(true)

    private fun start() {
        val api = ApiClient("http://127.0.0.1:1", InMemoryTokenStore(TokenBundle("access", null, Long.MAX_VALUE)))
        compose.setContent { Tabs(api) }
    }

    @Composable
    private fun Tabs(api: ApiClient) {
        if (showsChat.value) {
            ChatScreen(api = api)
        } else {
            Text("다른 탭", Modifier.testTag("other.tab"))
        }
    }

    private fun leaveAndComeBack() {
        compose.runOnIdle { showsChat.value = false }
        compose.onNodeWithTag("other.tab").assertIsDisplayed()
        compose.runOnIdle { showsChat.value = true }
    }

    @Test
    fun theUnsentMessageIsStillThereAfterATripToAnotherTab() {
        start()

        compose.onNodeWithTag("chat.composer.field").performTextInput("쓰다 만 메시지")
        leaveAndComeBack()

        compose.onNodeWithTag("chat.composer.field").assertTextContains("쓰다 만 메시지")
    }

    @Test
    fun aFieldEmptiedByHandComesBackEmpty() {
        start()

        compose.onNodeWithTag("chat.composer.field").performTextInput("지울 말")
        compose.onNodeWithTag("chat.composer.field").performTextClearance()
        leaveAndComeBack()

        // 안내 문구는 칸이 비어 있을 때만 그려지므로, 그것이 보이는 것이 곧
        // 지운 말이 되살아나지 않았다는 뜻이다.
        compose.onNodeWithText("메시지를 입력하세요").assertIsDisplayed()
    }
}
