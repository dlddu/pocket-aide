// 검증 시나리오: test-affirmations.md#시나리오 3
package com.dlddu.pocketaide.e2e

import androidx.compose.ui.test.longClick
import androidx.compose.ui.test.performTouchInput
import androidx.compose.ui.test.swipeLeft
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
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
        val keptId = seed(api, kept)
        val viaSheetId = seed(api, viaSheet)
        val swipedId = seed(api, swiped)
        try {
            var scenario = launchSignedIn()
            compose.openTab("Affirmations")

            val viaSheetTag = "affirmations.row.$viaSheetId"
            compose.scrollTo(viaSheetTag)
            compose.waitForTag(viaSheetTag)
            compose.node(viaSheetTag).performTouchInput { longClick() }
            compose.tap("sheet.delete.button")
            compose.waitForTagGone(viaSheetTag)
            assertTrue("Sheet-deleting should reach the backend", E2e.eventually { !stored(api, viaSheetId) })

            val swipedTag = "affirmations.row.$swipedId"
            compose.scrollTo(swipedTag)
            compose.waitForTag(swipedTag)
            compose.node(swipedTag).performTouchInput { swipeLeft() }
            compose.waitForTagGone(swipedTag)
            assertTrue("Swipe-deleting should reach the backend", E2e.eventually { !stored(api, swipedId) })
            assertTrue("Untouched sentences stay stored", stored(api, keptId))

            compose.tap("affirmations.add.button")
            compose.waitForText("sheet.title", "새 다짐")
            assertFalse("The create-mode sheet has no delete button", compose.exists("sheet.delete.button"))
            compose.tap("sheet.cancel.button")
            compose.waitForTagGone("sheet.text.field")

            scenario = relaunch(scenario)
            compose.openTab("Affirmations")
            compose.waitForTag("affirmations.list.header")
            assertRemoved(listOf(swiped, viaSheet), keptId)
            scenario.close()
        } finally {
            api.call("DELETE", "/api/affirmations/$keptId")
        }
    }

    private fun assertRemoved(removed: List<String>, keptId: Long, rotations: Int = 10) {
        compose.scrollTo("affirmations.row.$keptId")
        compose.waitForTag("affirmations.row.$keptId")
        for (text in removed) {
            assertFalse("'$text' must not be listed or shown in the hero card", compose.hasAnyText(text))
        }
        compose.waitForTag("affirmations.hero.rotate")
        repeat(rotations) { turn ->
            compose.tap("affirmations.hero.rotate")
            for (text in removed) {
                assertFalse("'$text' must not come back on hero rotation ${turn + 1}", compose.hasAnyText(text))
            }
        }
    }

    private fun seed(api: BackendApi, text: String): Long {
        val reply = api.call("POST", "/api/affirmations", JSONObject().put("text", text).put("priority", "normal"))
        assertNotNull("Seeding '$text' should return the created sentence", reply.json)
        return reply.json!!.getLong("id")
    }

    private fun stored(api: BackendApi, id: Long): Boolean = api.affirmations().any { it.getLong("id") == id }
}
