// 검증 시나리오: test-github-monitor.md#시나리오 12
package com.dlddu.pocketaide.e2e

import androidx.compose.ui.test.hasText
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookUserScopedHistoryTest : E2eTest() {
    private fun event(repo: String, pullRequest: Int?, seed: Int, workflow: String, token: String) = WebhookEvent(
        repo = repo,
        pullRequest = pullRequest,
        commit = Webhooks.commit(seed, token),
        branch = pullRequest?.let { "feature/$it" } ?: "main",
        workflow = workflow,
        conclusion = "success",
    )

    private fun acknowledgeEverything(api: BackendApi) {
        Webhooks.history(api).filter { it.isNull("acknowledged_at") }.forEach {
            val reply = api.call("POST", "/api/notification-history/${it.getLong("id")}/ack")
            assertTrue("Acknowledging ${it.getLong("id")} should succeed: ${reply.body}", reply.status in 200..299)
        }
        val left = Webhooks.history(api).count { it.isNull("acknowledged_at") }
        assertEquals("Every earlier history item should be acknowledged before the scenario starts", 0, left)
    }

    private fun assertLinks(group: String, withPullRequest: Boolean) {
        val tags = compose.tagsUnder(PrMonitor.groupTag(group))
        for (link in listOf("commit", "run")) {
            val tag = tags.firstOrNull { Regex("prmonitor\\.row\\.\\d+\\.link\\.$link").matches(it) }
            assertTrue("Group $group should offer the $link link", tag != null && compose.isEnabled(tag))
        }
        assertEquals(
            "Group $group should offer a PR link only when the event has a pull request",
            withPullRequest,
            compose.exists(PrMonitor.groupTag(group) + ".link.pr"),
        )
    }

    private fun assertUnreadFirst(unread: List<String>, read: String) {
        PrMonitor.scrollToTop(compose)
        assertTrue(
            "The unacknowledged section should hold exactly the new groups",
            E2e.eventually(10_000) { compose.hasAnyText("미확인 · ${unread.size}개 그룹") },
        )
        unread.forEach { assertTrue("Group $it should be listed", PrMonitor.reveal(compose, it)) }
        val tops = unread.map { PrMonitor.top(compose, PrMonitor.groupTag(it)) }
        assertEquals("Unacknowledged groups should be newest first: ${unread.zip(tops)}", tops.sorted(), tops)
        for (group in unread) {
            assertTrue("Group $group should be emphasised as 미확인 1", PrMonitor.hasLabel(compose, group, "미확인 1"))
            assertEquals("Group $group should list its item with 확인", 1, PrMonitor.itemAcknowledgeButtons(compose, group).size)
        }
        assertTrue("The acknowledged group should be listed", PrMonitor.reveal(compose, read))
        val header = compose.onAllNodes(hasText("확인 완료"), useUnmergedTree = true).fetchSemanticsNodes().firstOrNull()
        assertTrue("The acknowledged section header should be shown", header != null)
        assertTrue("The acknowledged group should sit under 확인 완료", header!!.boundsInRoot.top < PrMonitor.top(compose, PrMonitor.groupTag(read)))
        if (compose.exists(PrMonitor.groupTag(unread.last()))) {
            assertTrue("Unacknowledged groups should sit above 확인 완료", PrMonitor.top(compose, PrMonitor.groupTag(unread.last())) < header.boundsInRoot.top)
        }
        val labels = PrMonitor.labels(compose, read)
        assertEquals("The acknowledged group should read 모두 확인: $labels", "모두 확인", PrMonitor.badge(compose, read))
        assertFalse("The acknowledged group should be collapsed to its header: $labels", labels.contains("Build"))
        assertEquals("The acknowledged group should offer no 모두 확인 button", 0, PrMonitor.acknowledgeAllButtons(compose, read))
    }

    @Test
    fun completedEventsListUnacknowledgedFirstAndStayWithTheirUser() {
        val token = Webhooks.runToken()
        val repo = "dlddu/e2e-userscope-$token"
        val owner = BackendApi.signIn()

        val earlier = event(repo, 1, 1, "Build", token)
        Webhooks.send(earlier)
        Webhooks.awaitItems(owner, repo, 1)
        acknowledgeEverything(owner)

        val pulled = event(repo, 2, 2, "Lint", token)
        val pushed = event(repo, null, 3, "Deploy", token)
        Webhooks.send(listOf(pulled, pushed))
        Webhooks.awaitItems(owner, repo, 3)

        var scenario = PrMonitor.openUntil(compose, pushed.groupKey, "미확인 1")
        assertTrue("Both new events should be in the history", PrMonitor.reveal(compose, pulled.groupKey))
        assertLinks(pushed.groupKey, withPullRequest = false)
        assertLinks(pulled.groupKey, withPullRequest = true)
        assertEquals("The header badge should equal the two unacknowledged items", 2, PrMonitor.waitForUnread(compose, 2))
        assertUnreadFirst(listOf(pushed.groupKey, pulled.groupKey), earlier.groupKey)

        val other = "e2e-userscope-other-$token"
        try {
            scenario = AppAccount.switchUser(compose, scenario, other)
            compose.openTab("PrMonitor")
            compose.waitForTag("prmonitor.empty.state", timeoutMillis = 20_000)
            assertTrue("The first user's history should not show for the other user", PrMonitor.groupsOf(compose, repo).isEmpty())
            assertFalse("The other user should have no unacknowledged badge", compose.exists("prmonitor.unread.badge"))

            val second = BackendApi.signIn(subject = other)
            assertTrue("The other user should own no history rows of $repo", Webhooks.items(second, repo).isEmpty())
            assertEquals("The first user's rows should stay with the first user", 3, Webhooks.items(owner, repo).size)
            scenario.close()
        } finally {
            AppAccount.restoreDefaultUser(compose)
        }
    }
}
