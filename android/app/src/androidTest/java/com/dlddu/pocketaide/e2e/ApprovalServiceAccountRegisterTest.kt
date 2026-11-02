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
class ApprovalServiceAccountRegisterTest : E2eTest() {
    @Test
    fun registeredServiceAccountIsListedWithoutASecretAndCannotBeTakenTwice() {
        val token = E2e.uniqueToken().lowercase()
        val name = "homelab-k3s-mcp-$token"
        val duplicateName = "homelab-k3s-mcp-again-$token"
        val malformedName = "homelab-k3s-mcp-bad-$token"
        val namespace = "e2e-$token"
        val account = "$namespace/homelab-k3s-mcp"
        val owner = BackendApi.signIn()
        try {
            val scenario = launchSignedIn()
            ApprovalUi.openKeys(compose)

            ApprovalUi.registerServiceAccount(compose, name, namespace, "homelab-k3s-mcp")
            compose.waitForTagGone("approval.keys.sa.name", timeoutMillis = 15_000)
            assertFalse("Registering a service account should not show a key", compose.exists("approval.keys.issued.title"))
            compose.waitForTag("approval.keys.row.$name")
            val meta = compose.textOf("approval.keys.meta.$name")
            assertTrue("The row should name the service account instead of a key: $meta", meta.startsWith("SA $account · 등록 "))
            assertTrue("A fresh caller should have no last use yet: $meta", meta.endsWith("마지막 사용 없음"))
            assertTrue("The service account caller should be revocable like a key", compose.exists("approval.keys.revoke.$name"))

            val stored = ApprovalKeysApi.key(owner, name)
            assertNotNull("The server should list $name", stored)
            assertEquals("The server should store a service account caller", "k8s_sa", stored!!.optString("kind"))
            assertEquals("The server should store the service account", account, stored.optString("service_account"))
            assertEquals("A service account caller has no key", "", stored.optString("key_prefix"))

            ApprovalUi.registerServiceAccount(compose, duplicateName, namespace, "homelab-k3s-mcp")
            compose.waitForText("approval.keys.sa.notice", "이미 등록된 서비스 어카운트입니다")
            compose.tap("approval.keys.sa.cancel")
            compose.waitForTagGone("approval.keys.sa.name", timeoutMillis = 10_000)
            assertFalse("The refused caller should not be listed", compose.exists("approval.keys.row.$duplicateName"))

            ApprovalUi.registerServiceAccount(compose, malformedName, namespace, "Homelab_MCP")
            compose.waitForTag("approval.keys.sa.notice", timeoutMillis = 15_000)
            compose.tap("approval.keys.sa.cancel")
            compose.waitForTagGone("approval.keys.sa.name", timeoutMillis = 10_000)
            assertFalse("The malformed caller should not be listed", compose.exists("approval.keys.row.$malformedName"))

            val other = BackendApi.signIn(subject = "e2e-approval-sa-other-$token")
            val taken = other.call(
                "POST", "/api/approval-keys",
                JSONObject().put("name", "mcp-$token").put("kind", "k8s_sa").put("service_account", account),
            )
            assertEquals("Another user should not register the same service account: ${taken.body}", 409, taken.status)
            assertTrue(taken.body, taken.body.contains("service account already registered"))
            assertEquals("Only one caller should hold $account", 1, ApprovalKeysApi.keys(owner).count { it.optString("service_account") == account })
            scenario.close()
        } finally {
            ApprovalKeysApi.revokeAll(owner, name)
            ApprovalKeysApi.revokeAll(owner, duplicateName)
            ApprovalKeysApi.revokeAll(owner, malformedName)
        }
    }
}
