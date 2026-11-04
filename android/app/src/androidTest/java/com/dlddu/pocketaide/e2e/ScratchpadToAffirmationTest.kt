// 검증 시나리오: test-affirmations.md#시나리오 4
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ScratchpadToAffirmationTest : E2eTest() {
    @Test
    fun scratchpadMemoMovesIntoAffirmationsWithNormalPriority() {
        val token = E2e.uniqueToken()
        val sentence = "작게 시작해도 괜찮다 $token"
        val keptMemo = "나중에 정리할 생각 $token"
        val api = BackendApi.signIn()
        var scenario = launchSignedIn()
        compose.openMore("Scratchpad")

        addMemo(keptMemo)
        addMemo(sentence)
        val sentenceId = memoId(api, sentence)
        val keptId = memoId(api, keptMemo)

        compose.tap("scratchpad.row.$sentenceId.move.affirmation")
        compose.waitForText("sheet.title", "우선순위 설정")
        compose.tap("sheet.cancel.button")
        compose.waitForTagGone("sheet.text.field")
        compose.waitForTagGone("scratchpad.row.$sentenceId")
        assertTrue("The moved memo should leave the scratchpad", E2e.eventually { api.scratchpad().none { it.optString("text") == sentence } })
        assertTrue("The memo that was not moved should stay", api.scratchpad().any { it.optString("text") == keptMemo })
        compose.waitForTag("scratchpad.row.$keptId")

        val moved = api.affirmations().firstOrNull { it.optString("text") == sentence }
        assertTrue("The memo text should be stored as a sentence", moved != null)
        assertEquals("The moved sentence should default to 보통", "normal", moved!!.optString("priority"))
        assertTrue("The memo that was not moved must not become a sentence", api.affirmations().none { it.optString("text") == keptMemo })
        val affirmationId = moved.getLong("id")

        scenario = relaunch(scenario)
        compose.openMore("Scratchpad")
        compose.waitForTag("scratchpad.row.$keptId")
        assertFalse("The moved memo must not come back to the scratchpad", compose.exists("scratchpad.row.$sentenceId"))
        compose.openTab("Affirmations")
        compose.waitForTag("affirmations.list.header")
        compose.scrollTo("affirmations.row.$affirmationId")
        compose.waitForTag("affirmations.row.$affirmationId")
        assertTrue("The moved sentence should be listed", compose.hasAnyText(sentence))
        assertFalse("The memo that was not moved must not be listed as a sentence", compose.hasAnyText(keptMemo))
        scenario.close()

        api.call("DELETE", "/api/affirmations/$affirmationId")
        api.call("DELETE", "/api/scratchpad/$keptId")
    }

    private fun addMemo(text: String) {
        compose.tap("scratchpad.add.button")
        compose.type("scratchpad.sheet.text.field", text)
        compose.tap("scratchpad.sheet.save.button")
        compose.waitForTagGone("scratchpad.sheet.text.field")
        compose.waitUntil(15_000) { compose.hasAnyText(text) }
    }

    private fun memoId(api: BackendApi, text: String): Long {
        assertTrue("'$text' should be stored", E2e.eventually { api.scratchpad().any { it.optString("text") == text } })
        return api.scratchpad().first { it.optString("text") == text }.getLong("id")
    }
}
