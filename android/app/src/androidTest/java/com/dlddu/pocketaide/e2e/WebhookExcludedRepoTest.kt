// 검증 시나리오: test-github-monitor.md#시나리오 7
package com.dlddu.pocketaide.e2e

import androidx.compose.ui.semantics.SemanticsProperties
import androidx.compose.ui.semantics.getOrNull
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookExcludedRepoTest : E2eTest() {
    private val excludedRow = Regex("prmonitor\\.excluded\\.row\\.\\d+")

    private fun event(repo: String, pullRequest: Int, token: String) = WebhookEvent(
        repo = repo,
        pullRequest = pullRequest,
        commit = Webhooks.commit(pullRequest, token),
        branch = "feature/$pullRequest",
        workflow = "Build",
        conclusion = "success",
    )

    private fun openSheet() {
        compose.tap("prmonitor.excluded.button")
        compose.waitForTag("prmonitor.excluded.input", timeoutMillis = 10_000)
    }

    private fun closeSheet() {
        compose.tap("prmonitor.sheet.done", timeoutMillis = 5_000)
        compose.waitForTagGone("prmonitor.excluded.input", timeoutMillis = 10_000)
    }

    private fun submit(text: String) {
        compose.type("prmonitor.excluded.input", text)
        compose.tap("prmonitor.excluded.add.button")
    }

    private fun rows(repo: String): List<String> =
        compose.onAllNodes(hasTagMatching(excludedRow), useUnmergedTree = true).fetchSemanticsNodes()
            .mapNotNull { it.config.getOrNull(SemanticsProperties.TestTag) }
            .filter { compose.textsUnder(it).contains(repo) }

    private fun exclude(repo: String, token: String) {
        val scenario = PrMonitor.open(compose)
        openSheet()
        submit(repo)
        assertTrue("$repo should be listed as excluded", E2e.eventually(10_000) { rows(repo).size == 1 })

        val malformed = "e2e-excl-$token"
        submit(malformed)
        assertTrue("A name that is not owner/name should be refused", E2e.eventually(10_000) { compose.anyTextContains("HTTP 400") })
        assertEquals("The refused name should not be listed", 0, rows(malformed).size)

        submit(repo)
        assertTrue("An already excluded repository should be refused", E2e.eventually(10_000) { compose.anyTextContains("HTTP 409") })
        assertEquals("$repo should be listed once", 1, rows(repo).size)

        closeSheet()
        scenario.close()
    }

    private fun include(repo: String) {
        val scenario = PrMonitor.open(compose)
        openSheet()
        assertTrue("$repo should still be listed as excluded", E2e.eventually(10_000) { rows(repo).size == 1 })
        val row = rows(repo).single()
        compose.tap("$row.delete")
        assertTrue("$repo should leave the excluded list", E2e.eventually(10_000) { rows(repo).isEmpty() })
        closeSheet()
        scenario.close()
    }

    @Test
    fun excludedRepositoryRecordsNothingUntilItIsIncludedAgain() {
        val token = Webhooks.runToken()
        val excluded = "dlddu/e2e-excl-a-$token"
        val other = "dlddu/e2e-excl-b-$token"
        val api = BackendApi.signIn()
        exclude(excluded, token)

        val dropped = event(excluded, 1, token)
        val kept = event(other, 1, token)
        Webhooks.send(listOf(dropped, kept))
        var scenario = PrMonitor.openUntil(compose, kept.groupKey, "CI 1건")
        assertEquals("A workflow finishing in the excluded repository should leave no history", emptySet<String>(), PrMonitor.groupsOf(compose, excluded))
        assertEquals("The other repository should be recorded", setOf(kept.groupKey), PrMonitor.groupsOf(compose, other))
        assertTrue("The server should store nothing for the excluded repository", Webhooks.items(api, excluded).isEmpty())
        scenario.close()

        include(excluded)
        val restored = event(excluded, 2, token)
        Webhooks.send(restored)
        scenario = PrMonitor.openUntil(compose, restored.groupKey, "CI 1건")
        PrMonitor.scrollToTop(compose)
        assertEquals(
            "Only the workflow that finished after the repository was included again should be recorded",
            setOf(restored.groupKey),
            PrMonitor.groupsOf(compose, excluded),
        )
        assertTrue("The other repository should stay recorded", PrMonitor.reveal(compose, kept.groupKey))
        assertEquals("The server should hold only the later event for the excluded repository", 1, Webhooks.items(api, excluded).size)
        scenario.close()
    }
}
