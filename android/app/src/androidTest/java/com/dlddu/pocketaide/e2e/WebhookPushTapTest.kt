// 검증 시나리오: test-github-monitor.md#시나리오 8
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookPushTapTest : E2eTest() {
    private fun event(repo: String, number: Int, token: String) = WebhookEvent(
        repo = repo,
        pullRequest = number,
        commit = Webhooks.commit(number, token),
        branch = "e2e/push-$number",
        workflow = "Push tap",
        conclusion = "success",
    )

    @Test
    fun pushTapHighlightsItemWithoutAcknowledging() {
        val token = Webhooks.runToken()
        val closed = event("dlddu/push-e2e-$token", 81, token)
        val foreground = event("dlddu/push-e2e-$token", 82, token)
        val api = BackendApi.signIn()
        val device = Fcm.registerDevice(api)

        val scenario = PrMonitor.open(compose)
        val before = PrMonitor.unreadCount(compose) ?: 0
        scenario.close()

        Webhooks.send(closed)
        deliverAndTap(closed, device)
        assertHighlighted(closed)
        assertStillUnread(closed, before + 1)
        waitForHighlightToClear(closed)

        compose.openTab("Affirmations")
        compose.waitForTagGone("prmonitor.screen")
        Webhooks.send(foreground)
        deliverAndTap(foreground, device)
        assertHighlighted(foreground)
        assertStillUnread(foreground, before + 2)
        waitForHighlightToClear(foreground)
    }

    private fun deliverAndTap(event: WebhookEvent, device: String) {
        val push = Fcm.titled(device, event.pushTitle)
        assertNotNull("The push for ${event.repo} #${event.pullRequest} should be sent to the Android device", push)
        assertTrue("The push should carry the history row id", Fcm.data(push!!)["event_id"]?.toLongOrNull() != null)
        Shade.clearAppNotifications()
        Fcm.deliver(push)
        Shade.tap(event.pushTitle)
        compose.waitUntil(20_000) { compose.isSelected("tab.More") && compose.exists("prmonitor.screen") }
    }

    private fun assertHighlighted(event: WebhookEvent) {
        assertTrue(
            "The pushed event's group card should be highlighted",
            E2e.eventually(10_000) { PrMonitor.reveal(compose, event.groupKey) && compose.exists(PrMonitor.groupTag(event.groupKey) + ".arrival") },
        )
    }

    private fun assertStillUnread(event: WebhookEvent, expected: Int) {
        assertEquals("A push tap must leave the item unacknowledged", 1, PrMonitor.itemAcknowledgeButtons(compose, event.groupKey).size)
        assertEquals("A push tap must not lower the unacknowledged badge", expected, PrMonitor.waitForUnread(compose, expected))
    }

    private fun waitForHighlightToClear(event: WebhookEvent) {
        compose.waitForTagGone(PrMonitor.groupTag(event.groupKey) + ".arrival", timeoutMillis = 15_000)
    }
}
