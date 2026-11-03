// 검증 시나리오: test-github-monitor.md#시나리오 15
package com.dlddu.pocketaide.e2e

import android.content.Intent
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import androidx.test.uiautomator.UiDevice
import com.dlddu.pocketaide.MainActivity
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookHistoryExternalLinkTest : E2eTest() {
    private val instrumentation get() = InstrumentationRegistry.getInstrumentation()
    private val device: UiDevice get() = UiDevice.getInstance(instrumentation)
    private val runLink = Regex("prmonitor\\.row\\.\\d+\\.link\\.run")

    private fun linkTag(label: String, group: String): String? = when (label) {
        "PR" -> PrMonitor.groupTag(group) + ".link.pr"
        else -> compose.tagsUnder(PrMonitor.groupTag(group)).firstOrNull(runLink::matches)
    }

    private fun openExternally(label: String, group: String) {
        val tag = linkTag(label, group)
        assertTrue("The unacknowledged item should offer the $label link", tag != null && compose.exists(tag))
        PrMonitor.scrollIntoView(compose, tag!!)
        compose.tap(tag)
        assertTrue(
            "Tapping $label should open the GitHub page in the browser",
            E2e.eventually(60_000) { device.currentPackageName.let { it != null && it != instrumentation.targetContext.packageName } },
        )
        Thread.sleep(3_000)
        val context = instrumentation.targetContext
        context.startActivity(
            Intent(context, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_REORDER_TO_FRONT),
        )
        assertTrue(
            "The app should come back to the front",
            E2e.eventually(30_000) { device.currentPackageName == context.packageName },
        )
        compose.waitForTag("prmonitor.screen")
    }

    private fun assertStillUnacknowledged(group: String, unread: Int) {
        assertTrue("The group should keep its one unacknowledged item", PrMonitor.waitForLabel(compose, group, "미확인 1"))
        assertEquals("The item should not read as acknowledged", 0, PrMonitor.acknowledgedItems(compose, group))
        assertEquals("The item should still offer 확인", 1, PrMonitor.itemAcknowledgeButtons(compose, group).size)
        assertEquals("The group should still offer 모두 확인", 1, PrMonitor.acknowledgeAllButtons(compose, group))
        assertEquals("The unacknowledged badge should not move", unread, PrMonitor.unreadCount(compose))
    }

    @Test
    fun openingExternalLinksLeavesTheItemUnacknowledged() {
        val token = Webhooks.runToken()
        val event = WebhookEvent(
            repo = "dlddu/e2e-link-$token",
            pullRequest = 1,
            commit = Webhooks.commit(1, token),
            branch = "feature/1",
            workflow = "Build",
            conclusion = "success",
        )
        Webhooks.send(event)
        val group = event.groupKey

        PrMonitor.openUntil(compose, group, "미확인 1")
        val before = PrMonitor.unreadCount(compose)
        assertNotNull("The header should show the unacknowledged badge", before)
        for (label in listOf("PR", "런")) {
            openExternally(label, group)
            assertStillUnacknowledged(group, before!!)
        }

        AppSession.closeAll()
        val scenario = PrMonitor.openUntil(compose, group, "미확인 1")
        assertStillUnacknowledged(group, before!!)
        scenario.close()
    }
}
