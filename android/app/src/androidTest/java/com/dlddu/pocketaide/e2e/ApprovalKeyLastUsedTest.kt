package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalKeyLastUsedTest : E2eTest() {
    private val lastUsedToday = Regex("마지막 사용 \\d{2}:\\d{2}$")

    @Test
    fun callerRequestAndPollUpdateTheKeysLastUse() {
        val token = E2e.uniqueToken()
        val name = "infra-ci-$token"
        val api = BackendApi.signIn()
        val caller = api.issueApprovalKey(name)
        val created = mutableListOf<String>()
        try {
            val scenario = launchSignedIn()
            ApprovalUi.openKeys(compose)
            compose.waitForTag("approval.keys.row.$name")
            assertTrue("Before any call the key should be unused", compose.textOf("approval.keys.meta.$name").endsWith("마지막 사용 없음"))

            val id = caller.create("deploy-1-$token", context = "prod 배포 deploy-1", timeoutSeconds = 600)
            created += id
            val polled = caller.view(id)
            assertNotNull("The caller should read its request back", polled)
            assertEquals("PENDING", polled!!.optString("status"))
            assertTrue("A pending request should have no processedAt", polled.isNull("processedAt"))
            assertFalse("The caller view should carry no user identity", polled.has("sub"))
            assertFalse("The caller view should carry no user identity", polled.has("email"))

            ApprovalUi.back(compose)
            ApprovalUi.openKeys(compose)
            compose.waitForTag("approval.keys.row.$name")
            assertTrue(
                "The key list should show today's last use after the caller's calls: ${compose.textOf("approval.keys.meta.$name")}",
                E2e.eventually(15_000) { lastUsedToday.containsMatchIn(compose.textOf("approval.keys.meta.$name")) },
            )
            val stored = ApprovalKeysApi.key(api, name)
            assertTrue("The server should have stamped the last use", stored != null && !stored.isNull("last_used_at"))
            scenario.close()
        } finally {
            created.forEach { api.call("POST", "/api/approvals/$it/reject") }
            api.call("POST", "/api/approval-keys/${caller.id}/revoke")
        }
    }
}
