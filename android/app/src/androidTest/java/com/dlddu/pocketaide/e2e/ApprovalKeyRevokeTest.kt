package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalKeyRevokeTest : E2eTest() {
    @Test
    fun revokingOneKeyBlocksOnlyThatKeyAndKeepsItsPendingRequest() {
        val token = E2e.uniqueToken()
        val revokedName = "infra-ci-$token"
        val keptName = "staging-ci-$token"
        val api = BackendApi.signIn()
        val revokedCaller = api.issueApprovalKey(revokedName)
        val keptCaller = api.issueApprovalKey(keptName)
        val created = mutableListOf<String>()
        try {
            val pendingExternalId = "pending-$token"
            val pending = revokedCaller.create(pendingExternalId, context = "요청 $pendingExternalId", timeoutSeconds = 1800)
            created += pending
            assertEquals("Precondition: 자동 승인 should be stored", 200, ApprovalKeysApi.setAutoResponse(api, "AUTO_APPROVE").status)

            val scenario = launchSignedIn()
            ApprovalUi.openKeys(compose)
            val revoke = "approval.keys.revoke.$revokedName"
            compose.waitForTag(revoke)

            compose.tapScrolled(revoke)
            compose.tap("approvals.confirm.cancel", timeoutMillis = 5_000)
            compose.waitForTagGone("approvals.confirm.cancel", timeoutMillis = 5_000)
            assertTrue("Cancelling should leave the key active", compose.exists(revoke))
            assertTrue("Cancelling should not revoke", ApprovalKeysApi.key(api, revokedName)?.isNull("revoked_at") == true)

            compose.tapScrolled(revoke)
            compose.tap("approvals.confirm.ok", timeoutMillis = 5_000)
            compose.waitForTag("approval.keys.revoked.$revokedName")
            assertTrue("The revoked key should stay listed with its revoke date", compose.textOf("approval.keys.revoked.$revokedName").startsWith("폐기됨 "))
            assertFalse("A revoked key should offer no way to re-enable it", compose.exists(revoke))
            assertTrue("The other key should stay active", compose.exists("approval.keys.revoke.$keptName"))

            val blocked = BackendApi.perform(
                "POST", "/api/external/approvals", "Bearer ${revokedCaller.key}",
                JSONObject().put("externalId", "after-$token").put("context", "c"),
            )
            assertEquals("The revoked key should not create requests: ${blocked.body}", 401, blocked.status)
            val blockedPoll = BackendApi.perform("GET", "/api/external/approvals/$pending", "Bearer ${revokedCaller.key}", null)
            assertEquals("The revoked key should not read requests: ${blockedPoll.body}", 401, blockedPoll.status)
            val keptReply = keptCaller.createReply("kept-$token", "요청 kept-$token", timeoutSeconds = 1800)
            assertEquals("The other key should still create requests: ${keptReply.body}", 201, keptReply.status)
            assertEquals("In 자동 승인 the other key's request should be APPROVED at creation", "APPROVED", keptReply.json?.optString("status"))
            created += requireNotNull(keptReply.json).getString("id")

            ApprovalUi.back(compose)
            ApprovalUi.waitForRow(compose, pendingExternalId)
            val row = "approvals.row.$pendingExternalId"
            assertTrue("The pending request should carry the revoked-key warning: ${compose.mergedText(row)}", compose.mergedText(row).contains("폐기된 키로 생성됨"))
            assertEquals("Revoking should not decide the pending request", "PENDING", api.approvalStatus(pending))

            compose.tap(row)
            compose.waitForTag("approval.detail.revoked", timeoutMillis = 15_000)
            compose.waitForText("approval.detail.status", "대기 중")
            compose.tap("approval.detail.reject")
            compose.tap("approvals.confirm.ok", timeoutMillis = 5_000)
            compose.waitForText("approval.detail.status", "거절됨")
            assertEquals("The server should store R as REJECTED", "REJECTED", api.approvalStatus(pending))
            ApprovalUi.back(compose)

            ApprovalUi.openHistory(compose)
            val history = "approvals.history.row.$pendingExternalId"
            compose.waitForTag(history)
            val label = compose.mergedText(history)
            assertTrue("R's history entry should keep the key name $revokedName: $label", label.contains(revokedName))
            assertTrue("R's history entry should note the revoked key: $label", label.contains("폐기된 키로 생성됨"))
            scenario.close()
        } finally {
            ApprovalKeysApi.setAutoResponse(api, "NONE")
            created.forEach { api.call("POST", "/api/approvals/$it/reject") }
            api.call("POST", "/api/approval-keys/${revokedCaller.id}/revoke")
            api.call("POST", "/api/approval-keys/${keptCaller.id}/revoke")
        }
    }
}
