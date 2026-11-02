// 검증 시나리오: 없음 (스모크/인프라)
package com.dlddu.pocketaide.e2e

import androidx.compose.ui.test.hasAnyDescendant
import androidx.compose.ui.test.hasScrollToNodeAction
import androidx.compose.ui.test.hasTestTag
import androidx.compose.ui.test.performScrollToNode
import androidx.compose.ui.test.performTextClearance
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.FixMethodOrder
import org.junit.Test
import org.junit.runner.RunWith
import org.junit.runners.MethodSorters

@RunWith(AndroidJUnit4::class)
@FixMethodOrder(MethodSorters.NAME_ASCENDING)
class PrMonitorTest : E2eTest() {
    private val fixtureGroup = "pr:dlddu/pocket-aide-e2e:9001"
    private val fixtureTitle = "dlddu/pocket-aide-e2e · #9001"
    private val pushTitle = "CI 통과 — dlddu/pocket-aide-e2e #9001"

    @Test
    fun pushTapOpensPrMonitorWithoutAcknowledging() {
        val push = Fcm.titled(Fcm.FIXTURE_DEVICE, pushTitle, timeoutMillis = 120_000)
        assertNotNull("The backend should push the fixture event to the registered Android device", push)
        assertTrue("The push should carry the history row id", Fcm.data(push!!)["event_id"]?.toLongOrNull() != null)

        val scenario = launchSignedIn()
        assertFalse("Launch should land on a different tab than PR 모니터", compose.isSelected("tab.PrMonitor"))
        scenario.close()

        Shade.clearAppNotifications()
        Fcm.deliver(push)
        Shade.tap(pushTitle)
        compose.waitUntil(20_000) { compose.isSelected("tab.PrMonitor") }
        assertTrue("The pushed event's group should be highlighted", PrMonitor.waitForLabel(compose, fixtureGroup, "방금 진입", 15_000))
        assertTrue("The pushed event's row should be listed", PrMonitor.labels(compose, fixtureGroup).contains(fixtureTitle))
        if (PrMonitor.badge(compose, fixtureGroup) != "모두 확인") {
            assertEquals("A push tap must not acknowledge the row", 1, PrMonitor.itemAcknowledgeButtons(compose, fixtureGroup).size)
        }
    }

    @Test
    fun webhookEventReachesHistoryAndCanBeAcknowledged() {
        val api = BackendApi.signIn()
        var scenario = PrMonitor.open(compose)
        var attempts = 0
        while (!PrMonitor.waitForLabel(compose, fixtureGroup, fixtureTitle, 10_000) && attempts < 6) {
            attempts += 1
            PrMonitor.reopenTab(compose)
        }
        assertTrue("History row for the webhook fixture should appear on the PR 모니터 screen", PrMonitor.reveal(compose, fixtureGroup))

        if (PrMonitor.badge(compose, fixtureGroup) != "모두 확인") {
            assertTrue(
                "Fixture conclusion=success should render as CI 통과 on the unacknowledged row",
                PrMonitor.itemStatuses(compose, fixtureGroup).contains("CI 통과"),
            )
            PrMonitor.tapFirstItemAcknowledge(compose, fixtureGroup)
        }
        compose.waitUntil(10_000) { PrMonitor.badge(compose, fixtureGroup) == "모두 확인" }
        assertTrue("The acknowledged group should leave no 확인 button", PrMonitor.itemAcknowledgeButtons(compose, fixtureGroup).isEmpty())

        scenario = relaunch(scenario)
        compose.openTab("PrMonitor")
        assertTrue("Row should still be listed after relaunch", E2e.eventually { PrMonitor.reveal(compose, fixtureGroup) })
        compose.waitUntil(10_000) { PrMonitor.badge(compose, fixtureGroup) == "모두 확인" }
        assertFalse("No unacknowledged row should remain after the persisted acknowledgement", PrMonitor.itemStatuses(compose, fixtureGroup).contains("CI 통과"))
        val stored = Webhooks.items(api, "dlddu/pocket-aide-e2e")
        assertTrue("The fixture row should be acknowledged on the server: $stored", stored.isNotEmpty() && stored.all { !it.isNull("acknowledged_at") })
        scenario.close()
    }

    @Test
    fun openPullRequestsConnectListsRolesCiStatusAndFilters() {
        var scenario = openPullRequestsSheet()
        submitToken(GitHubToken.REVOKED)
        compose.waitForTag("openprs.connect.error")
        assertTrue("Connect error should name the rejected token", compose.textOf("openprs.connect.error").contains("토큰을 거부"))
        assertFalse("A rejected PAT must not be stored", compose.exists("openprs.disconnect.button"))

        scenario.close()
        scenario = openPullRequestsSheet()
        submitToken(GitHubToken.VALID)
        compose.waitForText("openprs.account.login", "@pocket-aide-e2e")

        if (!compose.isSelected("filter.pill.all")) compose.tap("filter.pill.all")
        assertRow(11, "작성자", "성공")
        assertRow(14, "작성자", "상태 없음")
        assertRow(12, "리뷰어", "실패", "octocat")
        assertRow(13, "리뷰어", "진행 중", "hubot")

        compose.waitForTag("openprs.banner.access", timeoutMillis = 5_000)
        assertTrue("The access banner should count the one hidden PR", compose.anyTextContains("PR 1개는"))

        compose.tap("filter.pill.ciFailing")
        compose.waitForTag(row(12), timeoutMillis = 5_000)
        compose.waitForTagGone(row(11), timeoutMillis = 5_000)
        assertFalse("Pending PRs should be filtered out", compose.exists(row(13)))
        compose.tap("filter.pill.mine")
        compose.waitForTag(row(11), timeoutMillis = 5_000)
        compose.waitForTagGone(row(12), timeoutMillis = 5_000)
        compose.tap("filter.pill.all")

        compose.tap("openprs.disconnect.button")
        compose.waitForTag("openprs.token.field", timeoutMillis = 10_000)
        scenario.close()
    }

    @Test
    fun openPullRequestsBannersForRevokedAndRateLimitedTokens() {
        val scenario = openPullRequestsSheet()
        submitToken(GitHubToken.EXPIRES)
        compose.waitForTag("openprs.banner.token")
        assertEquals("토큰 다시 연결", compose.textOf("openprs.banner.action"))
        compose.tap("openprs.banner.action")

        submitToken(GitHubToken.RATE_LIMITED)
        compose.waitForTag("openprs.banner.ratelimit")
        assertEquals("다시 시도", compose.textOf("openprs.banner.action"))

        compose.tap("openprs.disconnect.button")
        compose.waitForTag("openprs.token.field", timeoutMillis = 10_000)
        scenario.close()
    }

    private fun openPullRequestsSheet() = launchSignedIn().also {
        compose.openTab("PrMonitor")
        compose.tap("prmonitor.openprs.button")
        compose.waitUntil(10_000) { compose.exists("openprs.token.field") || compose.exists("openprs.disconnect.button") }
        if (compose.exists("openprs.disconnect.button")) compose.tap("openprs.disconnect.button")
        compose.waitForTag("openprs.token.field", timeoutMillis = 10_000)
    }

    private fun submitToken(token: String) {
        compose.waitForTag("openprs.token.field").performTextClearance()
        compose.type("openprs.token.field", token)
        compose.tap("openprs.connect.button")
    }

    private fun row(number: Int) = "openprs.row.dlddu/pocket-aide-e2e#$number"

    private val rowList = hasScrollToNodeAction() and hasAnyDescendant(hasTagMatching(Regex("openprs\\.row\\..+")))

    private fun assertRow(number: Int, vararg parts: String) {
        compose.waitUntil(15_000) { compose.onAllNodes(rowList).fetchSemanticsNodes().isNotEmpty() }
        runCatching { compose.onAllNodes(rowList)[0].performScrollToNode(hasTestTag(row(number))) }
        compose.waitForTag(row(number))
        val label = compose.mergedText(row(number))
        parts.forEach { assertTrue("PR #$number row should read $it: $label", label.contains(it)) }
    }

    private object GitHubToken {
        const val VALID = "ghp_e2e_valid"
        const val REVOKED = "ghp_e2e_revoked"
        const val EXPIRES = "ghp_e2e_expires"
        const val RATE_LIMITED = "ghp_e2e_ratelimited"
    }
}
