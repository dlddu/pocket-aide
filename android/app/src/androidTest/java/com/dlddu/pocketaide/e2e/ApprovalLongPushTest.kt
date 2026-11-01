package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalLongPushTest : E2eTest() {
    @Test
    fun longContextPushIsCutAndBothRequestsWait() {
        val token = E2e.uniqueToken()
        val api = BackendApi.signIn()
        val keyName = "long-ci-${token.lowercase()}"
        val caller = api.issueApprovalKey(keyName)
        val created = mutableListOf<String>()
        try {
            val device = Fcm.registerDevice(api)
            val longContext = "배포 로그 $token 줄을 확인해 주세요. ".repeat(400)
            assertTrue("The context should be at least 10KB", longContext.toByteArray().size >= 10 * 1024)
            val longExternalId = "long-$token"
            val started = System.currentTimeMillis()
            val longReply = caller.createReply(longExternalId, longContext)
            assertTrue("Step 1: creating should not wait for the push", System.currentTimeMillis() - started < 5_000)
            assertEquals("Step 1: the long request should be created: ${longReply.body}", 201, longReply.status)
            val long = requireNotNull(longReply.json).getString("id")
            created += long

            val push = Fcm.approval(device, long)
            assertNotNull("Step 1: the long request should still be pushed", push)
            val body = Fcm.body(push!!)
            assertNotNull("Step 1: the push should have a body", body)
            assertTrue("Step 1: the cut body should end with …", body!!.endsWith("…"))
            assertTrue("Step 1: the body should be the front of the context", longContext.startsWith(body.dropLast(1)))
            val size = push.getJSONObject("data").toString().toByteArray().size
            assertTrue("Step 1: the data payload should fit the 4KB FCM limit, was $size bytes", size <= 4096)

            Shade.clearAppNotifications()
            Fcm.deliver(push)
            assertTrue("Step 1: the cut push should be shown", Shade.find("승인 요청 · $keyName") != null)
            Shade.close()
            Shade.clearAppNotifications()

            val shortExternalId = "short-$token"
            val shortReply = caller.createReply(shortExternalId, "짧은 요청 $token")
            assertEquals("The second request should be created: ${shortReply.body}", 201, shortReply.status)
            created += requireNotNull(shortReply.json).getString("id")

            val scenario = launchSignedIn()
            ApprovalUi.openTab(compose)
            ApprovalUi.waitForRow(compose, longExternalId)
            ApprovalUi.waitForRow(compose, shortExternalId)
            assertEquals("Step 3: the long request should still be waiting", "PENDING", api.approvalStatus(long))
            scenario.close()
        } finally {
            created.forEach { api.call("POST", "/api/approvals/$it/reject") }
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
