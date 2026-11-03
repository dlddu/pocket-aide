// 검증 시나리오: test-scratchpad.md#시나리오 2
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ScratchpadTest : E2eTest() {
    @Test
    fun addedMemoIsStoredAndMovesIntoPersonalTodos() {
        val token = E2e.uniqueToken()
        val memo = "장보기 목록 정리 $token"
        val kept = "나중에 읽을 글 $token"
        val api = BackendApi.signIn()
        var scenario = launchSignedIn()
        compose.openMore("Scratchpad")

        addMemo(kept)
        addMemo(memo)
        val memoId = memoId(api, memo)
        val keptId = memoId(api, kept)
        compose.waitForTag("scratchpad.row.$memoId")
        compose.waitForTag("scratchpad.row.$keptId")

        compose.tap("scratchpad.row.$memoId.move.personal")
        compose.waitForTagGone("scratchpad.row.$memoId")
        assertTrue("The moved memo should leave the scratchpad", E2e.eventually { api.scratchpad().none { it.optString("text") == memo } })
        assertNotNull("The moved memo should become a personal todo", api.todo("personal", memo))
        assertTrue("The untouched memo should stay", api.scratchpad().any { it.optString("text") == kept })

        compose.openMore("Personal")
        compose.type("todos.personal.search.field", token)
        compose.waitUntil(15_000) { compose.tagWithText(Regex("todos\\.personal\\.row\\.\\d+"), memo) != null }
        assertFalse("The untouched memo must not land in 개인", compose.hasAnyText(kept))

        scenario = relaunch(scenario)
        compose.openMore("Scratchpad")
        compose.waitForTag("scratchpad.row.$keptId")
        assertFalse("The moved memo stays out of the scratchpad after a relaunch", compose.exists("scratchpad.row.$memoId"))
        scenario.close()

        api.call("DELETE", "/api/scratchpad/$keptId")
        api.todo("personal", memo)?.let { api.call("DELETE", "/api/todos/personal/${it.getLong("id")}") }
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
