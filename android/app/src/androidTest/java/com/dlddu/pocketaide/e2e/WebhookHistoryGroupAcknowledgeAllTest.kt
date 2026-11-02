package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookHistoryGroupAcknowledgeAllTest : E2eTest() {
    private fun event(repo: String, pullRequest: Int, workflow: String, conclusion: String, token: String) = WebhookEvent(
        repo = repo,
        pullRequest = pullRequest,
        commit = Webhooks.commit(pullRequest, token),
        branch = "feature/$pullRequest",
        workflow = workflow,
        conclusion = conclusion,
    )

    private fun assertFullyAcknowledged(group: String) {
        assertTrue("Group $group should stay listed once acknowledged", PrMonitor.reveal(compose, group))
        assertTrue("Group $group should read 모두 확인", E2e.eventually(10_000) { PrMonitor.badge(compose, group) == "모두 확인" })
        val labels = PrMonitor.labels(compose, group)
        assertFalse("Group $group should have no unacknowledged count: $labels", labels.any { it.startsWith("미확인") })
        assertEquals("A fully acknowledged group should not offer 모두 확인", 0, PrMonitor.acknowledgeAllButtons(compose, group))
        assertEquals("A fully acknowledged group should have no item left to acknowledge", 0, PrMonitor.itemAcknowledgeButtons(compose, group).size)
    }

    private fun arrange(groupA: String, groupC: String) {
        val scenario = PrMonitor.openUntil(compose, groupA, "CI 3건")
        PrMonitor.tapFirstItemAcknowledge(compose, groupA)
        assertTrue("Acknowledging one item should leave group A with two unacknowledged", PrMonitor.waitForLabel(compose, groupA, "미확인 2"))
        assertTrue("Group C should be listed", PrMonitor.reveal(compose, groupC))
        PrMonitor.tapFirstItemAcknowledge(compose, groupC)
        assertTrue(
            "Acknowledging group C's only item should leave nothing to acknowledge in it",
            E2e.eventually(10_000) { PrMonitor.reveal(compose, groupC) && PrMonitor.itemAcknowledgeButtons(compose, groupC).isEmpty() },
        )
        Thread.sleep(3_000)
        scenario.close()
    }

    @Test
    fun acknowledgeAllClearsOnlyTheUnacknowledgedItemsOfThatGroup() {
        val token = Webhooks.runToken()
        val repo = "dlddu/e2e-ackall-$token"
        val sent = listOf(
            event(repo, 3, "Build", "success", token),
            event(repo, 2, "Build", "success", token),
            event(repo, 1, "Build", "success", token),
            event(repo, 1, "Lint", "success", token),
            event(repo, 1, "Deploy", "failure", token),
        )
        Webhooks.send(sent)
        val groupC = sent[0].groupKey
        val groupB = sent[1].groupKey
        val groupA = sent[2].groupKey
        arrange(groupA, groupC)

        var scenario = PrMonitor.openUntil(compose, groupA, "미확인 2")
        assertEquals("Group A should start with one acknowledged item", 1, PrMonitor.acknowledgedItems(compose, groupA))
        assertEquals("Group A should offer 모두 확인", 1, PrMonitor.acknowledgeAllButtons(compose, groupA))
        assertTrue("Group B should be listed", PrMonitor.reveal(compose, groupB))
        assertTrue("Group B should have one unacknowledged item", PrMonitor.hasLabel(compose, groupB, "미확인 1"))
        assertEquals("Group B should offer 모두 확인", 1, PrMonitor.acknowledgeAllButtons(compose, groupB))
        assertFullyAcknowledged(groupC)

        scenario.close()
        scenario = PrMonitor.openUntil(compose, groupA, "미확인 2")
        val before = PrMonitor.unreadCount(compose)
        assertNotNull("The header should show the unacknowledged badge", before)
        PrMonitor.tapAcknowledgeAll(compose, groupA)
        assertEquals("모두 확인 should take two off the unacknowledged badge", before!! - 2, PrMonitor.waitForUnread(compose, before - 2))
        assertTrue("Group B should stay listed", PrMonitor.reveal(compose, groupB))
        assertTrue("Group B should keep its unacknowledged item", PrMonitor.hasLabel(compose, groupB, "미확인 1"))
        assertEquals("Group B should still offer 모두 확인", 1, PrMonitor.acknowledgeAllButtons(compose, groupB))
        assertFullyAcknowledged(groupA)

        scenario.close()
        scenario = PrMonitor.openUntil(compose, groupB, "미확인 1")
        assertEquals("The acknowledgements should persist across a relaunch", before - 2, PrMonitor.unreadCount(compose))
        assertFullyAcknowledged(groupA)
        scenario.close()
    }
}
