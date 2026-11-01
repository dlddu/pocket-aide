package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class AffirmationTest : E2eTest() {
    @Test
    fun addedAffirmationIsListedAndPersists() {
        val text = "오늘도 한 걸음 ${E2e.uniqueToken()}"
        val api = BackendApi.signIn()
        var scenario = launchSignedIn()
        compose.openTab("Affirmations")

        compose.tap("affirmations.add.button")
        compose.waitForText("sheet.title", "새 다짐")
        compose.type("sheet.text.field", text)
        compose.tap("sheet.save.button")
        compose.waitForTagGone("sheet.text.field")

        assertTrue("The backend should store the affirmation", E2e.eventually { api.affirmations().any { it.optString("text") == text } })
        val stored = api.affirmations().first { it.optString("text") == text }
        assertEquals("A new affirmation should default to 보통", "normal", stored.optString("priority"))
        val rowTag = "affirmations.row.${stored.getLong("id")}"
        compose.scrollTo(rowTag)
        compose.waitForTag(rowTag)

        scenario = relaunch(scenario)
        compose.waitForTag("affirmations.list.header")
        compose.scrollTo(rowTag)
        compose.waitForTag(rowTag)
        scenario.close()

        api.call("DELETE", "/api/affirmations/${stored.getLong("id")}")
    }
}
