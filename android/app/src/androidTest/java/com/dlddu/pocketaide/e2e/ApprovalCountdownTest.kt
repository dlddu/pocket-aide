// 검증 시나리오: test-approval-gate.md#시나리오 8
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalCountdownTest : E2eTest() {
    @Test
    fun decisionClosesWhenTheCountdownReachesZero() {
        val token = E2e.uniqueToken()
        val externalId = "countdown-$token"
        val api = BackendApi.signIn()
        val caller = api.issueApprovalKey("countdown-$token")
        try {
            val scenario = launchSignedIn()
            val id = caller.create(externalId, context = "카운트다운 $token", timeoutSeconds = 45)
            ApprovalUi.openTab(compose)
            ApprovalUi.waitForRow(compose, externalId)
            compose.tap("approvals.row.$externalId")
            compose.waitForTag("approval.detail.title")

            compose.waitForTag("approval.detail.countdown", timeoutMillis = 5_000)
            val first = compose.textOf("approval.detail.countdown")
            compose.waitUntil(10_000) { compose.textOf("approval.detail.countdown") < first }
            assertTrue("Before the deadline the 승인 button should be enabled", compose.isEnabled("approval.detail.approve"))

            compose.waitForText("approval.detail.status", "만료", timeoutMillis = 60_000)
            assertFalse("At zero the 승인 button should be disabled", compose.isEnabled("approval.detail.approve"))
            assertFalse("At zero the 거절 button should be disabled", compose.isEnabled("approval.detail.reject"))
            compose.waitForTagGone("approval.detail.countdown", timeoutMillis = 5_000)
            assertTrue(
                "The server should also treat the request as expired",
                E2e.eventually(10_000) { api.approvalStatus(id) == "EXPIRED" },
            )
            scenario.close()
        } finally {
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
