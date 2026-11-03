// 검증 시나리오: test-affirmations.md#시나리오 3
package com.dlddu.pocketaide.e2e

import androidx.compose.ui.test.longClick
import androidx.compose.ui.test.performTouchInput
import androidx.compose.ui.test.swipeLeft
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class AffirmationDeleteTest : E2eTest() {
    @Test
    fun swipeAndSheetDeleteRemoveSentencesFromListAndRotation() {
        val token = E2e.uniqueToken()
        val kept = "남길 다짐 $token"
        val swiped = "밀어 지울 다짐 $token"
        val viaSheet = "시트로 지울 다짐 $token"
        val api = BackendApi.signIn()
        val ids = listOf(kept, viaSheet, swiped).associateWith { seed(api, it) }
        var scenario = launchSignedIn()
        compose.openTab("Affirmations")

        val viaSheetTag = rowTag(ids.getValue(viaSheet))
        compose.scrollTo(viaSheetTag)
        compose.waitForTag(viaSheetTag).performTouchInput { longClick() }
        compose.waitForText("sheet.title", "우선순위 설정")
        compose.tap("sheet.delete.button")
        compose.waitForTagGone(viaSheetTag)

        val swipedTag = rowTag(ids.getValue(swiped))
        compose.scrollTo(swipedTag)
        compose.waitForTag(swipedTag).performTouchInput { swipeLeft() }
        compose.waitForTagGone(swipedTag)
        assertTrue("The untouched sentence should stay listed", compose.exists(rowTag(ids.getValue(kept))))

        compose.tap("affirmations.add.button")
        compose.waitForText("sheet.title", "새 다짐")
        assertFalse("The create-mode sheet has no delete button", compose.exists("sheet.delete.button"))
        compose.tap("sheet.cancel.button")
        compose.waitForTagGone("sheet.title")

        assertTrue(
            "The backend should drop both deleted sentences",
            E2e.eventually { api.affirmations().none { it.optString("text") in listOf(swiped, viaSheet) } },
        )
        assertTrue("The backend should keep the untouched sentence", api.affirmations().any { it.optString("text") == kept })

        compose.openTab("PrMonitor")
        compose.waitForTag("prmonitor.openprs.button")
        compose.openTab("Affirmations")
        assertRemoved(listOf(swiped, viaSheet), ids.getValue(kept))

        scenario = relaunch(scenario)
        compose.openTab("Affirmations")
        assertRemoved(listOf(swiped, viaSheet), ids.getValue(kept))
        scenario.close()

        api.call("DELETE", "/api/affirmations/${ids.getValue(kept)}")
    }

    private fun seed(api: BackendApi, text: String): Long {
        val reply = api.call("POST", "/api/affirmations", JSONObject().put("text", text).put("priority", "normal"))
        assertTrue("POST /api/affirmations should create '$text': ${reply.status} ${reply.body}", reply.status == 201)
        return reply.json!!.getLong("id")
    }

    private fun rowTag(id: Long) = "affirmations.row.$id"

    private fun assertRemoved(removed: List<String>, keptId: Long, rotations: Int = 10) {
        compose.waitForTag("affirmations.list.header")
        compose.scrollTo(rowTag(keptId))
        compose.waitForTag(rowTag(keptId))
        compose.waitForTag("affirmations.hero.text")
        for (text in removed) {
            assertFalse("'$text' must not be listed or shown in the hero card", compose.hasAnyText(text))
        }
        for (turn in 1..rotations) {
            compose.tap("affirmations.hero.rotate")
            for (text in removed) {
                assertFalse("'$text' must not come back on hero rotation $turn", compose.hasAnyText(text))
            }
        }
    }
}
