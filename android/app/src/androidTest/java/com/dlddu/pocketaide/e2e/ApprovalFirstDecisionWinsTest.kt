// 검증 시나리오: test-approval-gate.md#시나리오 10
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalFirstDecisionWinsTest : E2eTest() {
    @Test
    fun lateDecisionIsRefusedAndExplained() {
        val token = E2e.uniqueToken()
        val externalId = "race-$token"
        val api = BackendApi.signIn()
        val caller = api.issueApprovalKey("race-$token")
        val id = caller.create(externalId, context = "동시 결정 $token", timeoutSeconds = 1800)
        try {
            val scenario = launchSignedIn()
            ApprovalUi.openTab(compose)
            ApprovalUi.waitForRow(compose, externalId)
            compose.tap("approvals.row.$externalId")
            compose.waitForTag("approval.detail.title")
            compose.waitForText("approval.detail.status", "대기 중")

            val first = api.call("POST", "/api/approvals/$id/approve")
            assertEquals("The other device's approval should land first: ${first.body}", 200, first.status)

            compose.tap("approval.detail.reject")
            compose.tap("approvals.confirm.ok", timeoutMillis = 5_000)
            compose.waitForText("approval.detail.notice", "이미 처리된 요청입니다 — 승인됨")
            compose.waitForText("approval.detail.status", "승인됨")
            assertFalse("A processed request should have no decision buttons", compose.exists("approval.detail.reject"))
            assertFalse("A processed request should have no decision buttons", compose.exists("approval.detail.approve"))

            val again = api.call("POST", "/api/approvals/$id/reject")
            assertEquals("Deciding a processed request again should be refused: ${again.body}", 409, again.status)
            assertEquals("The first decision should stand", "APPROVED", api.approvalStatus(id))
            scenario.close()
        } finally {
            api.call("POST", "/api/approvals/$id/reject")
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
