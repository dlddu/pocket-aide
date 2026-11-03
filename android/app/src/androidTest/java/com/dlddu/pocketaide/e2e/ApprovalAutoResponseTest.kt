// 검증 시나리오: test-approval-gate.md#시나리오 13
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalAutoResponseTest : E2eTest() {
    private fun header(): String = compose.mergedText("approvals.autoResponse.header")

    private fun banner(): String = compose.textOf("approvals.autoResponse.banner.text")

    private fun choose(mode: String, user: BackendApi, step: String) {
        compose.tap("approvals.settings.button")
        compose.waitForTag("approval.keys.title")
        val option = "approval.settings.mode.$mode"
        compose.waitForTag(option)
        compose.waitUntil(15_000) { compose.isEnabled(option) }
        compose.tapScrolled(option)
        compose.waitUntil(15_000) { compose.isSelected(option) }
        assertTrue("$step: the server should store $mode", E2e.eventually { ApprovalKeysApi.autoResponse(user) == mode })
        ApprovalUi.back(compose)
    }

    private fun create(externalId: String, caller: ApprovalCaller, expecting: String, step: String): String {
        val reply = caller.createReply(externalId, "자동 응답 요청 $externalId", timeoutSeconds = 600)
        assertEquals("$step: creating $externalId should succeed: ${reply.body}", 201, reply.status)
        assertEquals("$step: $externalId should be $expecting at creation", expecting, reply.json?.optString("status"))
        val id = requireNotNull(reply.json).getString("id")
        assertEquals("$step: the caller should read $expecting", expecting, caller.view(id)?.optString("status"))
        return id
    }

    private fun assertHistory(externalId: String, label: String, step: String) {
        ApprovalUi.openHistory(compose)
        val row = "approvals.history.row.$externalId"
        compose.waitForTag(row)
        assertTrue("$step: $externalId should read $label: ${compose.mergedText(row)}", compose.mergedText(row).contains(label))
        ApprovalUi.back(compose)
    }

    @Test
    fun autoModeAppliesOnlyToLaterRequestsWithoutPushAndShowsABanner() {
        val token = E2e.uniqueToken()
        val subject = "e2e-approval-auto-${token.lowercase()}"
        val user = BackendApi.signIn(subject = subject)
        val caller = user.issueApprovalKey("infra-ci-${token.lowercase()}")
        val device = Fcm.registerDevice(user)
        var pendingId: String? = null
        try {
            var scenario = launchSignedIn()
            scenario = AppAccount.switchUser(compose, scenario, subject)
            ApprovalUi.openTab(compose)
            compose.waitForTag("approvals.empty.mode", timeoutMillis = 15_000)
            assertTrue("An empty 승인 tab should read 자동 응답 모드 없음", compose.textOf("approvals.empty.mode").contains("자동 응답 모드 없음"))
            assertTrue("The header should read 자동 응답 없음: ${header()}", header().contains("없음"))

            val pending = "p-$token"
            val pendingReply = caller.createReply(pending, "기존 대기 요청 $pending", timeoutSeconds = 600)
            assertEquals("Creating P should succeed: ${pendingReply.body}", 201, pendingReply.status)
            pendingId = requireNotNull(pendingReply.json).getString("id")
            assertNotNull("Precondition: with no auto mode P should be pushed", Fcm.approval(device, pendingId))

            choose("AUTO_APPROVE", user, "Step 1")
            compose.waitForTag("approvals.autoResponse.banner.text", timeoutMillis = 15_000)
            assertTrue("Step 2: the banner should name 자동 승인: ${banner()}", banner().contains("자동 승인 모드가 켜져 있습니다."))
            ApprovalUi.waitForRow(compose, pending)
            assertTrue("Step 2: the header should read 자동 승인: ${header()}", header().contains("자동 승인"))

            val approved = create("q-$token", caller, "APPROVED", "Step 3")
            assertNull("Step 3: no push should be sent for Q", Fcm.approval(device, approved, timeoutMillis = 5_000))
            assertEquals("Step 2: P must not be decided by the new mode", "PENDING", user.approvalStatus(pendingId))
            assertHistory("q-$token", "자동 승인", "Step 4")

            compose.tap("approvals.autoResponse.off", timeoutMillis = 15_000)
            compose.waitForTagGone("approvals.autoResponse.banner.text")
            compose.waitUntil(15_000) { header().contains("없음") }
            assertTrue("Step 5: 끄기 should store 없음", E2e.eventually { ApprovalKeysApi.autoResponse(user) == "NONE" })

            choose("AUTO_REJECT", user, "Step 1 (자동 거절)")
            compose.waitForTag("approvals.autoResponse.banner.text", timeoutMillis = 15_000)
            assertTrue("Step 2 (자동 거절): ${banner()}", banner().contains("자동 거절 모드가 켜져 있습니다."))
            val rejected = create("r-$token", caller, "REJECTED", "Step 3 (자동 거절)")
            assertNull("Step 3 (자동 거절): no push should be sent", Fcm.approval(device, rejected, timeoutMillis = 5_000))
            assertHistory("r-$token", "자동 거절", "Step 4 (자동 거절)")
            assertEquals("P should stay PENDING through both modes", "PENDING", user.approvalStatus(pendingId))
            scenario.close()
        } finally {
            ApprovalKeysApi.setAutoResponse(user, "NONE")
            pendingId?.let { user.call("POST", "/api/approvals/$it/reject") }
            user.call("POST", "/api/approval-keys/${caller.id}/revoke")
            AppAccount.restoreDefaultUser(compose)
        }
    }
}
