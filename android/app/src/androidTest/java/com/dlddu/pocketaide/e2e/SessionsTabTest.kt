// 검증 시나리오: test-session-platform.md#시나리오 1
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class SessionsTabTest : E2eTest() {
    @Test
    fun theSessionsTabSitsInTheBottomBarAndGuidesToTheFirstSession() {
        launchSignedIn()

        // 1단계: 하단 탭 바에 「세션」이 보인다.
        compose.waitForTag("tab.Sessions")
        assertTrue("The sessions tab should be visible in the bottom bar", compose.exists("tab.Sessions"))
        assertFalse(
            "The sessions tab belongs to the bottom bar, not to the 「더 보기」 menu",
            compose.exists("more.Sessions"),
        )

        val approvalLeft = compose.node("tab.Approval").fetchSemanticsNode().boundsInRoot.left
        val sessionsLeft = compose.node("tab.Sessions").fetchSemanticsNode().boundsInRoot.left
        val moreLeft = compose.node("tab.More").fetchSemanticsNode().boundsInRoot.left
        assertTrue("「세션」 should sit right after 「승인」 in the bottom bar", approvalLeft < sessionsLeft)
        assertTrue("「세션」 should sit before 「더 보기」 in the bottom bar", sessionsLeft < moreLeft)

        // 2단계: 「세션」 탭을 누르면 「에이전트 세션」 화면이 열리고 목록 대신 빈 상태 안내가 보인다.
        compose.tap("tab.Sessions")
        compose.waitForTag("sessions.screen")
        assertEquals("The sessions screen should show its title", SESSIONS_TITLE, compose.textOf("sessions.title"))
        compose.waitForTag("sessions.empty.state")
        assertEquals(
            "The empty state should invite the first session",
            EMPTY_STATE,
            compose.textOf("sessions.empty.state"),
        )
        assertFalse(
            "An unreachable platform must not replace the empty state",
            compose.exists("sessions.error.retry"),
        )
    }

    private companion object {
        const val SESSIONS_TITLE = "에이전트 세션"
        const val EMPTY_STATE = "아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요."
    }
}
