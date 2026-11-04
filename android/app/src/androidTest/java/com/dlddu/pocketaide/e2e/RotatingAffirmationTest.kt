// 검증 시나리오: test-affirmations.md#시나리오 6
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class RotatingAffirmationTest : E2eTest() {
    @Test
    fun heroRotatesAcrossReentriesAndOnTheRotateButton() {
        val token = E2e.uniqueToken()
        val sentences = listOf("오늘도 한 걸음 $token", "서두르지 않는다 $token", "끝까지 해 본다 $token")
        val api = BackendApi.signIn()
        api.affirmations().forEach { api.call("DELETE", "/api/affirmations/${it.getLong("id")}") }
        val ids = mutableListOf<Long>()
        try {
            val scenario = launchSignedIn()
            compose.openTab("Affirmations")
            assertEmptyState()
            reenter()
            assertEmptyState()

            sentences.forEach { ids += seed(api, it) }
            reenter()
            val shown = mutableSetOf(settledHero(sentences))
            assertFalse("The empty state gives way to the hero card", compose.exists("affirmations.empty.state"))
            repeat(ROTATION_ATTEMPTS) {
                if (shown.size < 2) {
                    reenter()
                    shown += settledHero(sentences)
                }
            }
            assertTrue("Re-entering the tab should surface at least two different sentences", shown.size >= 2)

            val before = settledHero(sentences)
            var after = before
            repeat(ROTATION_ATTEMPTS) {
                if (after == before) {
                    compose.tap("affirmations.hero.rotate")
                    after = shownHero(sentences)
                }
            }
            assertTrue("The rotate button should replace the hero sentence", after != before)
            scenario.close()
        } finally {
            ids.forEach { api.call("DELETE", "/api/affirmations/$it") }
        }
    }

    private fun reenter() {
        compose.openTab("PrMonitor")
        compose.openTab("Affirmations")
        compose.waitForTag("affirmations.list.header")
    }

    private fun assertEmptyState() {
        compose.waitForTag("affirmations.empty.state")
        assertFalse("With no sentences there is no hero card", compose.exists("affirmations.hero.card"))
        assertFalse("With no sentences there is no rotate button", compose.exists("affirmations.hero.rotate"))
    }

    private fun shownHero(sentences: List<String>): String {
        compose.waitForTag("affirmations.hero.text")
        var shown: String? = null
        compose.waitUntil(10_000) {
            shown = sentences.firstOrNull { compose.textOf("affirmations.hero.text") == it }
            shown != null
        }
        return shown!!
    }

    private fun settledHero(sentences: List<String>): String {
        var current = shownHero(sentences)
        repeat(5) {
            Thread.sleep(1_500)
            val next = shownHero(sentences)
            if (next == current) return current
            current = next
        }
        return current
    }

    private fun seed(api: BackendApi, text: String): Long {
        val reply = api.call("POST", "/api/affirmations", JSONObject().put("text", text).put("priority", "normal"))
        assertNotNull("Seeding '$text' should return the created sentence", reply.json)
        return reply.json!!.getLong("id")
    }

    private companion object {
        const val ROTATION_ATTEMPTS = 12
    }
}
