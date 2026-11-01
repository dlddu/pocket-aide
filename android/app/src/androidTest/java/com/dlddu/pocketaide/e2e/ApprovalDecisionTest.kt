package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalDecisionTest : E2eTest() {
    @Test
    fun approvalNeedsConfirmationAndReachesTheCaller() {
        val token = E2e.uniqueToken()
        val externalId = "decision-$token"
        val api = BackendApi.signIn()
        val caller = api.issueApprovalKey("decision-$token")
        val id = caller.create(externalId, context = "prod 배포 $token\n변경 3건", timeoutSeconds = 1800)
        try {
            val scenario = launchSignedIn()
            compose.openTab("Approval")
            compose.waitForTag("approvals.title")
            compose.tap("approvals.row.$externalId")
            compose.waitForTag("approval.detail.title")
            compose.waitForText("approval.detail.status", "대기 중")

            compose.tap("approval.detail.approve")
            compose.tap("approvals.confirm.cancel", timeoutMillis = 5_000)
            compose.waitForTagGone("approvals.confirm.cancel", timeoutMillis = 5_000)
            compose.waitForText("approval.detail.status", "대기 중")
            assertEquals("Cancelling should not reach the server", "PENDING", api.approvalStatus(id))

            compose.tap("approval.detail.approve")
            compose.tap("approvals.confirm.ok", timeoutMillis = 5_000)
            compose.waitForText("approval.detail.status", "승인됨")
            compose.waitForText("approval.detail.mode", "수동 결정")
            assertTrue("The detail should show when it was processed", compose.exists("approval.detail.processedAt"))
            assertFalse("A processed request should have no decision buttons", compose.exists("approval.detail.approve"))
            assertFalse("A processed request should have no decision buttons", compose.exists("approval.detail.reject"))

            assertEquals("The backend should record the approval", "APPROVED", api.approvalStatus(id))
            val seen = caller.view(id)
            assertEquals("The caller should receive APPROVED", "APPROVED", seen?.optString("status"))
            assertTrue("The caller should see processedAt: $seen", seen?.has("processedAt") == true)

            compose.tap("approvals.back")
            compose.waitForTag("approvals.title")
            compose.waitForTagGone("approvals.row.$externalId")
            scenario.close()
        } finally {
            api.call("POST", "/api/approvals/$id/reject")
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
