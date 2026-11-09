// 검증 시나리오: test-approval-gate.md#시나리오 21
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
class ExternalNotificationReadActionTest {
    @Test
    fun readActionMarksTheNotificationReadAndClearsIt() {
        val token = E2e.uniqueToken()
        val api = BackendApi.signIn()
        val keyName = "notify-${token.lowercase()}"
        val caller = api.issueApprovalKey(keyName)
        try {
            val device = Fcm.registerDevice(api)
            val title = "읽음 버튼 $token"
            val sent = BackendApi.perform(
                "POST",
                "/api/external/notifications",
                "Bearer ${caller.key}",
                JSONObject().put("title", title).put("body", "본문 $token").put("requester", "e2e 읽음 버튼"),
            )
            assertEquals("Step 10: the caller's notification should be sent: ${sent.body}", 200, sent.status)

            val message = Fcm.titled(device, title)
            assertNotNull("Step 10: a push should reach the device", message)
            assertFalse("Step 10: a plain notification push should be data-only so the app adds 읽음", message!!.has("notification"))
            assertEquals("Step 10: the push should carry the external notification category", "EXTERNAL_NOTIFICATION", Fcm.data(message)["category"])
            val id = requireNotNull(Fcm.data(message)["notification_id"]) { "Step 10: the push should carry the record id" }

            Shade.clearAppNotifications()
            Fcm.deliver(message)
            Shade.tapAction(title, "읽음")

            assertTrue(
                "Step 10: 읽음 should mark the notification read on the server",
                E2e.eventually(20_000) { api.call("GET", "/api/external-notifications/$id").json?.has("read_at") == true },
            )
            assertTrue("Step 10: 읽음 should clear the notification from the shade", Shade.awaitNoAppNotifications())
            Shade.close()
        } finally {
            Shade.clearAppNotifications()
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
