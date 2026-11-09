// 검증 시나리오: test-approval-gate.md#시나리오 6
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalPushTest : E2eTest() {
    private fun push(device: String, approvalId: String, step: String): JSONObject {
        val message = Fcm.approval(device, approvalId)
        assertNotNull("$step: a push should be sent for the request", message)
        return message!!
    }

    private fun createPending(caller: ApprovalCaller, externalId: String, context: String): String {
        val reply = caller.createReply(externalId, context)
        assertEquals("Creating $externalId should succeed: ${reply.body}", 201, reply.status)
        return requireNotNull(reply.json).getString("id")
    }

    @Test
    fun pushNamesTheRealKeyAndOpensTheRequestWithoutDeciding() {
        val token = E2e.uniqueToken()
        val api = BackendApi.signIn()
        val keyName = "infra-ci-${token.lowercase()}"
        val caller = api.issueApprovalKey(keyName)
        try {
            val device = Fcm.registerDevice(api)
            val firstContext = "prod 배포 $token — 승인해 주세요"
            val first = createPending(caller, "push-$token-1", firstContext)
            val firstPush = push(device, first, "Step 2")
            assertEquals("Step 2: the title should name the real key, not Deploy Bot", "승인 요청 · $keyName", Fcm.title(firstPush))
            assertEquals("Step 2: the body should be the context", firstContext, Fcm.body(firstPush))
            assertEquals("Step 2: the push should carry the approval category", "APPROVAL_REQUEST", Fcm.data(firstPush)["category"])
            assertFalse("Step 2: an approval push should be data-only so the app builds its actions", firstPush.has("notification"))

            launchSignedIn().close()

            val secondExternalId = "push-$token-2"
            val secondContext = "staging 배포 $token — 두 번째 요청"
            val second = createPending(caller, secondExternalId, secondContext)
            val secondPush = push(device, second, "Step 3")
            assertEquals("Step 3: the push should carry the request id", second, Fcm.data(secondPush)["approval_id"])

            Shade.clearAppNotifications()
            Fcm.deliver(secondPush)
            val title = "승인 요청 · $keyName"
            assertTrue("Step 3: the notification should offer 승인", Shade.hasAction(title, "승인"))
            assertTrue("Step 3: the notification should offer 거절", Shade.hasAction(title, "거절"))
            Shade.tap(title)

            compose.waitForTag("approval.detail.title", timeoutMillis = 20_000)
            val facts = listOf(
                "approval.detail.context" to secondContext,
                "approval.detail.requester" to "Deploy Bot",
                "approval.detail.key" to keyName,
                "approval.detail.externalId" to secondExternalId,
            )
            for ((tag, expected) in facts) {
                compose.waitForTag(tag, timeoutMillis = 10_000)
                val label = compose.textsUnder(tag).joinToString(" ")
                assertTrue("Step 3: $tag should show $expected: $label", label.contains(expected))
            }
            compose.waitForText("approval.detail.status", "대기 중")
            assertTrue("Step 3: the 승인 tab should be selected", compose.isSelected("tab.Approval"))
            assertEquals("Step 3: opening the push must not decide the request", "PENDING", api.approvalStatus(second))
            assertEquals("The first request should stay pending too", "PENDING", api.approvalStatus(first))
            api.call("POST", "/api/approvals/$first/reject")
            api.call("POST", "/api/approvals/$second/reject")
        } finally {
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }

    @Test
    fun pushOverAnOpenDetailSwitchesToTheNewRequest() {
        val token = E2e.uniqueToken()
        val api = BackendApi.signIn()
        val keyName = "push-swap-${token.lowercase()}"
        val caller = api.issueApprovalKey(keyName)
        try {
            val device = Fcm.registerDevice(api)
            val firstExternalId = "swap-$token-1"
            val secondExternalId = "swap-$token-2"
            val first = createPending(caller, firstExternalId, "첫 요청 $token")
            val second = createPending(caller, secondExternalId, "둘째 요청 $token")
            val firstPush = push(device, first, "First request")
            val secondPush = push(device, second, "Second request")
            val title = "승인 요청 · $keyName"

            launchSignedIn()
            Shade.clearAppNotifications()
            Fcm.deliver(firstPush)
            Shade.tap(title)
            compose.waitUntil(20_000) { compose.exists("approval.detail.externalId") && compose.textOf("approval.detail.externalId") == firstExternalId }

            Shade.clearAppNotifications()
            Fcm.deliver(secondPush)
            Shade.tap(title)
            compose.waitUntil(20_000) { compose.exists("approval.detail.externalId") && compose.textOf("approval.detail.externalId") == secondExternalId }
            assertEquals("Opening the pushes must not decide the requests", "PENDING", api.approvalStatus(first))
            api.call("POST", "/api/approvals/$first/reject")
            api.call("POST", "/api/approvals/$second/reject")
        } finally {
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }

    @Test
    fun notificationActionsDecideOnTheServerWithoutOpeningTheApp() {
        val token = E2e.uniqueToken()
        val api = BackendApi.signIn()
        val rejectKey = "action-no-${token.lowercase()}"
        val approveKey = "action-ok-${token.lowercase()}"
        val rejectCaller = api.issueApprovalKey(rejectKey)
        val approveCaller = api.issueApprovalKey(approveKey)
        try {
            val device = Fcm.registerDevice(api)
            val rejectContext = "거절할 요청 $token"
            val approveContext = "승인할 요청 $token"
            val rejected = createPending(rejectCaller, "action-$token-no", rejectContext)
            val approved = createPending(approveCaller, "action-$token-ok", approveContext)
            val rejectPush = push(device, rejected, "Reject request")
            val approvePush = push(device, approved, "Approve request")

            Shade.clearAppNotifications()
            Fcm.deliver(rejectPush)
            Shade.tapAction("승인 요청 · $rejectKey", "거절")
            assertTrue("거절 on the notification should reject on the server", E2e.eventually(20_000) { api.approvalStatus(rejected) == "REJECTED" })
            assertTrue("A decided notification should leave the shade", Shade.awaitNoAppNotifications())
            assertEquals("The caller should read REJECTED", "REJECTED", rejectCaller.view(rejected)?.optString("status"))
            assertEquals("Rejecting one request must leave the other pending", "PENDING", api.approvalStatus(approved))

            Shade.clearAppNotifications()
            Fcm.deliver(approvePush)
            Shade.tapAction("승인 요청 · $approveKey", "승인")
            assertTrue("승인 on the notification should approve on the server", E2e.eventually(20_000) { api.approvalStatus(approved) == "APPROVED" })
            assertTrue("A decided notification should leave the shade", Shade.awaitNoAppNotifications())
            assertEquals("The caller should read APPROVED", "APPROVED", approveCaller.view(approved)?.optString("status"))
            Shade.close()
        } finally {
            Shade.clearAppNotifications()
            api.call("POST", "/api/approval-keys/${rejectCaller.id}/revoke")
            api.call("POST", "/api/approval-keys/${approveCaller.id}/revoke")
        }
    }
}
