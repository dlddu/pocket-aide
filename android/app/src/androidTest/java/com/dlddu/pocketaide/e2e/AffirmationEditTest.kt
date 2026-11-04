// 검증 시나리오: test-affirmations.md#시나리오 2
package com.dlddu.pocketaide.e2e

import androidx.compose.ui.semantics.SemanticsActions
import androidx.compose.ui.test.performSemanticsAction
import androidx.compose.ui.test.performTextReplacement
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class AffirmationEditTest : E2eTest() {
    @Test
    fun editChangesOnlyTheTargetAndCancelChangesNothing() {
        val token = E2e.uniqueToken()
        val original = "물 자주 마시기 $token"
        val edited = "물 한 잔 먼저 $token"
        val untouched = "일찍 잠들기 $token"
        val discarded = "버려질 문장 $token"
        val api = BackendApi.signIn()
        val untouchedId = seed(api, untouched)
        val originalId = seed(api, original)
        try {
            var scenario = launchSignedIn()
            compose.openTab("Affirmations")
            assertEquals("A seeded sentence starts as 보통", "normal", stored(api, originalId).optString("priority"))

            openEditSheet(originalId)
            compose.tap("filter.pill.low")
            compose.waitForTag("sheet.text.field").performTextReplacement(edited)
            compose.tap("sheet.save.button")
            compose.waitForTagGone("sheet.text.field")
            assertTrue("The edit should reach the backend", E2e.eventually { stored(api, originalId).optString("text") == edited })
            assertEquals("The priority should become 낮음", "low", stored(api, originalId).optString("priority"))

            openEditSheet(untouchedId)
            compose.tap("filter.pill.high")
            compose.waitForTag("sheet.text.field").performTextReplacement(discarded)
            compose.tap("sheet.cancel.button")
            compose.waitForTagGone("sheet.text.field")
            assertEquals("A cancelled edit must keep the sentence", untouched, stored(api, untouchedId).optString("text"))
            assertEquals("A cancelled edit must keep the priority", "normal", stored(api, untouchedId).optString("priority"))
            assertTrue("A cancelled edit must keep the sentence listed", compose.hasAnyText(untouched))
            assertFalse("A cancelled edit must not show the typed sentence", compose.hasAnyText(discarded))

            scenario = relaunch(scenario)
            compose.openTab("Affirmations")
            compose.waitForTag("affirmations.list.header")
            compose.scrollTo("affirmations.row.$originalId")
            compose.waitForTag("affirmations.row.$originalId")
            assertTrue("The edited sentence should be listed after a relaunch", compose.hasAnyText(edited))
            assertFalse("The old sentence should be gone", compose.hasAnyText(original))
            assertTrue("The untouched sentence should stay listed", compose.hasAnyText(untouched))
            assertFalse("The discarded sentence must not be listed", compose.hasAnyText(discarded))
            scenario.close()
        } finally {
            api.call("DELETE", "/api/affirmations/$originalId")
            api.call("DELETE", "/api/affirmations/$untouchedId")
        }
    }

    private fun openEditSheet(id: Long) {
        val rowTag = "affirmations.row.$id"
        compose.scrollTo(rowTag)
        compose.waitForTag(rowTag)
        compose.node(rowTag).performSemanticsAction(SemanticsActions.OnLongClick)
        compose.waitForTag("sheet.text.field")
    }

    private fun seed(api: BackendApi, text: String): Long {
        val reply = api.call("POST", "/api/affirmations", JSONObject().put("text", text).put("priority", "normal"))
        assertNotNull("Seeding '$text' should return the created sentence", reply.json)
        return reply.json!!.getLong("id")
    }

    private fun stored(api: BackendApi, id: Long): JSONObject =
        api.affirmations().first { it.getLong("id") == id }
}
