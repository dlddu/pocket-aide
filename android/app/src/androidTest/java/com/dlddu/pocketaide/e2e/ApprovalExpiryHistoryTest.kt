// 검증 시나리오: test-approval-gate.md#시나리오 11
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalExpiryHistoryTest : E2eTest() {
    @Test
    fun expiredRequestLeavesPendingAndStaysInHistoryAsExpired() {
        val token = E2e.uniqueToken()
        val expiring = "e-$token"
        val waiting = "n-$token"
        val api = BackendApi.signIn()
        val caller = api.issueApprovalKey("expiry-$token")
        var waitingId: String? = null
        try {
            val expiringId = caller.create(expiring, context = "만료 요청 $expiring", timeoutSeconds = 5)
            waitingId = caller.create(waiting, context = "대기 요청 $waiting", timeoutSeconds = 600)
            assertTrue("E should expire on the server", E2e.eventually(20_000) { api.approvalStatus(expiringId) == "EXPIRED" })

            val scenario = launchSignedIn()
            ApprovalUi.openTab(compose)
            ApprovalUi.waitForRow(compose, waiting)
            assertFalse("Step 2: E should be gone from the pending list", compose.exists("approvals.row.$expiring"))

            ApprovalUi.openHistory(compose)
            val row = "approvals.history.row.$expiring"
            compose.waitForTag(row)
            val label = compose.mergedText(row)
            assertTrue("Step 2: E should read 만료: $label", label.contains("만료"))
            assertTrue("Step 2: E's processing mode should be 만료: $label", label.contains("만료 — 결정 없음"))
            assertFalse("Step 2: N should not be in 처리 이력", compose.exists("approvals.history.row.$waiting"))

            val stored = api.call("GET", "/api/approvals/history").items().firstOrNull { it.optString("external_id") == expiring }
            assertEquals("The history API should list E as expired", "expired", stored?.optString("decision_mode"))

            ApprovalUi.back(compose)
            ApprovalUi.waitForRow(compose, waiting)
            scenario.close()
        } finally {
            waitingId?.let { api.call("POST", "/api/approvals/$it/reject") }
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
