package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookItemAcknowledgeTest : E2eTest() {
    private class Entry(val id: Long, val acknowledgedAt: Long?)

    private fun event(repo: String, workflow: String, conclusion: String, token: String) = WebhookEvent(
        repo = repo,
        pullRequest = 14,
        commit = Webhooks.commit(14, token),
        branch = "feature/14",
        workflow = workflow,
        conclusion = conclusion,
    )

    private fun entries(api: BackendApi, repo: String): List<Entry> =
        Webhooks.items(api, repo).map { Entry(it.getLong("id"), if (it.isNull("acknowledged_at")) null else it.getLong("acknowledged_at")) }

    private fun assertOneOfTwoAcknowledged(group: String) {
        assertEquals("Only the acknowledged item should read 확인됨", 1, PrMonitor.acknowledgedItems(compose, group))
        assertEquals("The other item of the group should still offer 확인", 1, PrMonitor.itemAcknowledgeButtons(compose, group).size)
        assertEquals("The group should still offer 모두 확인 for its remaining item", 1, PrMonitor.acknowledgeAllButtons(compose, group))
    }

    private fun acknowledgeInApp(group: String) {
        var scenario = PrMonitor.openUntil(compose, group, "미확인 2")
        assertTrue("The group should hold both events", PrMonitor.hasLabel(compose, group, "CI 2건"))
        assertEquals("Both items should offer 확인", 2, PrMonitor.itemAcknowledgeButtons(compose, group).size)
        val before = PrMonitor.unreadCount(compose)
        assertNotNull("The header should show the unacknowledged badge", before)
        PrMonitor.tapFirstItemAcknowledge(compose, group)
        assertTrue("Acknowledging one item should take the group from 미확인 2 to 미확인 1", PrMonitor.waitForLabel(compose, group, "미확인 1"))
        assertEquals("The header badge should drop by one", before!! - 1, PrMonitor.waitForUnread(compose, before - 1))
        assertOneOfTwoAcknowledged(group)
        Thread.sleep(3_000)
        scenario.close()

        scenario = PrMonitor.openUntil(compose, group, "미확인 1")
        assertEquals("The badge should hold across a relaunch", before - 1, PrMonitor.unreadCount(compose))
        assertOneOfTwoAcknowledged(group)
        scenario.close()
    }

    @Test
    fun itemAcknowledgeRecordsOnlyThatItemAndIsScopedAndIdempotent() {
        val token = Webhooks.runToken()
        val repo = "dlddu/e2e-itemack-$token"
        val sent = listOf(event(repo, "Build", "success", token), event(repo, "Lint", "failure", token))
        Webhooks.send(sent)
        acknowledgeInApp(sent[0].groupKey)

        val owner = BackendApi.signIn()
        val listed = entries(owner, repo)
        assertEquals("The user should have both history items", 2, listed.size)
        val acknowledged = listed.firstOrNull { it.acknowledgedAt != null }
        val pending = listed.firstOrNull { it.acknowledgedAt == null }
        assertNotNull("One item should be acknowledged", acknowledged)
        assertNotNull("One item should stay unacknowledged", pending)

        val other = BackendApi.signIn(subject = "e2e-other-$token")
        val foreign = other.call("POST", "/api/notification-history/${pending!!.id}/ack")
        assertEquals("Acknowledging another user's item should be refused: ${foreign.body}", 404, foreign.status)
        assertNull(
            "The refused acknowledgement should leave the owner's item unacknowledged",
            entries(owner, repo).first { it.id == pending.id }.acknowledgedAt,
        )

        Thread.sleep(2_000)
        val again = owner.call("POST", "/api/notification-history/${acknowledged!!.id}/ack")
        assertTrue("Re-acknowledging should succeed: ${again.status} ${again.body}", again.status in 200..299)
        assertEquals(
            "Re-acknowledging should keep the original acknowledgement time",
            acknowledged.acknowledgedAt,
            entries(owner, repo).first { it.id == acknowledged.id }.acknowledgedAt,
        )
    }
}
