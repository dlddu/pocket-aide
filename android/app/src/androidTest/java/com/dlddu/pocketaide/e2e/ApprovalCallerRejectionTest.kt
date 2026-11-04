// 검증 시나리오: test-approval-gate.md#시나리오 3
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalCallerRejectionTest {
    private val createPath = "/api/external/approvals"

    @Test
    fun malformedOrUnauthenticatedCallerRequestsAreRejected() {
        val token = E2e.uniqueToken()
        val owner = BackendApi.signIn(subject = "e2e-approval-$token")
        val infra = owner.issueApprovalKey("infra-ci-$token")
        val staging = owner.issueApprovalKey("staging-ci-$token")
        try {
            val deploy = { JSONObject().put("externalId", "deploy-1").put("context", "프로덕션 배포 승인") }

            val first = infra.createReply("deploy-1", "프로덕션 배포 승인")
            assertEquals("infra-ci should create deploy-1 first: ${first.body}", 201, first.status)
            assertEquals("deploy-1 should be user A's only pending request", 1, pendingCount(owner))

            val missingContext = BackendApi.perform("POST", createPath, "Bearer ${infra.key}", JSONObject().put("externalId", "deploy-2"))
            assertEquals("Step 1: a request without context should be rejected: ${missingContext.body}", 400, missingContext.status)
            val legacyUser = BackendApi.perform(
                "POST",
                createPath,
                "Bearer ${infra.key}",
                JSONObject().put("externalId", "deploy-2").put("context", "프로덕션 배포 승인").put("userId", "ckabc"),
            )
            assertEquals("Step 2: the old gatekeeper userId field should be rejected: ${legacyUser.body}", 400, legacyUser.status)
            assertEquals("Steps 1-2 should not create a request", 1, pendingCount(owner))

            val duplicate = BackendApi.perform("POST", createPath, "Bearer ${infra.key}", deploy())
            assertEquals("Step 3: the same key should not reuse deploy-1: ${duplicate.body}", 409, duplicate.status)

            val otherKey = BackendApi.perform("POST", createPath, "Bearer ${staging.key}", deploy())
            assertEquals("Step 4: another key may use deploy-1: ${otherKey.body}", 201, otherKey.status)
            assertEquals("Step 4 should add exactly one request", 2, pendingCount(owner))

            val body = { JSONObject().put("externalId", "deploy-3").put("context", "c") }
            val noHeader = BackendApi.perform("POST", createPath, null, body())
            assertEquals("Step 5: a request without Authorization should be rejected: ${noHeader.body}", 401, noHeader.status)
            val wrongKey = BackendApi.perform("POST", createPath, "Bearer pak_wrong", body())
            assertEquals("Step 5: a wrong key should be rejected: ${wrongKey.body}", 401, wrongKey.status)

            val userToken = owner.call("POST", createPath, body())
            assertEquals("Step 6: user A's sign-in token should not open the caller route: ${userToken.body}", 401, userToken.status)

            val appRoute = BackendApi.perform("GET", "/api/approvals", "Bearer ${infra.key}", null)
            assertEquals("Step 7: a caller key should not open the app's pending list: ${appRoute.body}", 401, appRoute.status)
            assertEquals("Steps 5-7 should not create a request", 2, pendingCount(owner))
        } finally {
            owner.call("GET", "/api/approvals").items().forEach { owner.call("POST", "/api/approvals/${it.getString("id")}/reject") }
            owner.call("POST", "/api/approval-keys/${infra.id}/revoke")
            owner.call("POST", "/api/approval-keys/${staging.id}/revoke")
        }
    }

    private fun pendingCount(owner: BackendApi): Int {
        val reply = owner.call("GET", "/api/approvals")
        assertEquals("User A's pending list should load: ${reply.body}", 200, reply.status)
        return reply.items().size
    }
}
