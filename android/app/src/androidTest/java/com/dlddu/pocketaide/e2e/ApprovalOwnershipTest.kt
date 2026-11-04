// 검증 시나리오: test-approval-gate.md#시나리오 4
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalOwnershipTest : E2eTest() {
    @Test
    fun requestIsVisibleOnlyToTheKeyOwner() {
        val token = E2e.uniqueToken()
        val otherSubject = "e2e-approval-b-${token.lowercase()}"
        val owner = BackendApi.signIn()
        val other = BackendApi.signIn(subject = otherSubject)
        val infra = owner.issueApprovalKey("infra-ci-$token")
        val foreign = other.issueApprovalKey("b-ci-$token")
        val externalId = "r-$token"
        val requestId = infra.create(externalId, context = "요청 $externalId", timeoutSeconds = 1800)
        try {
            var scenario = launchSignedIn()
            ApprovalUi.openTab(compose)
            ApprovalUi.waitForRow(compose, externalId)

            try {
                scenario = AppAccount.switchUser(compose, scenario, otherSubject)
                ApprovalUi.openTab(compose)
                compose.waitForTag("approvals.empty.state", timeoutMillis = 20_000)
                assertFalse("Step 1: R should not be in user B's pending list", compose.exists("approvals.row.$externalId"))
                assertEquals("Step 1: user B should own no pending request", emptyList<String>(), pendingIds(other))
                assertTrue("Step 1: R should stay in user A's pending list", pendingIds(owner).contains(requestId))
                scenario.close()
            } finally {
                AppAccount.restoreDefaultUser(compose)
            }

            listOf("GET" to "/api/approvals/$requestId", "POST" to "/api/approvals/$requestId/approve", "POST" to "/api/approvals/$requestId/reject")
                .forEach { (method, path) ->
                    val reply = other.call(method, path)
                    assertEquals("Step 2: user B's $method $path should not find R: ${reply.body}", 404, reply.status)
                }

            val anonymous = BackendApi.perform("GET", "/api/approvals", null, null)
            assertEquals("Step 3: the pending list should need sign-in: ${anonymous.body}", 401, anonymous.status)

            val foreignPoll = BackendApi.perform("GET", "/api/external/approvals/$requestId", "Bearer ${foreign.key}", null)
            assertEquals("Step 4: b-ci should not learn that R exists: ${foreignPoll.body}", 404, foreignPoll.status)

            val ownerPoll = BackendApi.perform("GET", "/api/external/approvals/$requestId", "Bearer ${infra.key}", null)
            assertEquals("infra-ci should still read R: ${ownerPoll.body}", 200, ownerPoll.status)
            assertEquals("Step 2: R should stay PENDING after user B's attempts", "PENDING", ownerPoll.json?.optString("status"))
        } finally {
            owner.call("POST", "/api/approvals/$requestId/reject")
            owner.call("POST", "/api/approval-keys/${infra.id}/revoke")
            other.call("POST", "/api/approval-keys/${foreign.id}/revoke")
        }
    }

    private fun pendingIds(api: BackendApi): List<String> {
        val reply = api.call("GET", "/api/approvals")
        assertEquals("The pending list should load: ${reply.body}", 200, reply.status)
        return reply.items().map { it.getString("id") }
    }
}
