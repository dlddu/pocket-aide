// 검증 시나리오: test-approval-gate.md#시나리오 1
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class ApprovalKeyIssueTest : E2eTest() {
    @Test
    fun issuedKeyShowsItsSecretOnlyOnceAndStaysWithItsOwner() {
        val token = E2e.uniqueToken()
        val name = "infra-ci-$token"
        val owner = BackendApi.signIn()
        try {
            var scenario = launchSignedIn()
            ApprovalUi.openKeys(compose)

            ApprovalUi.issue(compose, name)
            compose.waitForTag("approval.keys.issued.title")
            val secret = compose.textOf("approval.keys.issued.secret")
            assertTrue("The secret should be shown right after issuing: $secret", secret.startsWith("pak_"))
            compose.tap("approval.keys.issued.copy")
            compose.waitUntil(5_000) { compose.mergedText("approval.keys.issued.copy") == "복사됨" }
            compose.tap("approval.keys.issued.done")
            compose.waitForTagGone("approval.keys.issued.title", timeoutMillis = 10_000)

            compose.waitForTag("approval.keys.row.$name")
            val meta = compose.textOf("approval.keys.meta.$name")
            assertTrue("The list should show only the start of the key: $meta", meta.startsWith(secret.take(12) + "…"))
            assertTrue("A fresh key should have no last use yet: $meta", meta.endsWith("마지막 사용 없음"))
            assertTrue("The secret should not be on screen after the sheet closed", ApprovalUi.secretIsNowhere(compose, secret))

            ApprovalUi.back(compose)
            ApprovalUi.openKeys(compose)
            compose.waitForTag("approval.keys.row.$name")
            assertTrue("The secret should not be shown again on return", ApprovalUi.secretIsNowhere(compose, secret))

            ApprovalUi.issue(compose, name)
            compose.waitForTag("approval.keys.issue.notice")
            assertFalse("The refused issue should not show a secret", compose.exists("approval.keys.issued.title"))
            compose.tap("approval.keys.issue.cancel")
            compose.waitForTagGone("approval.keys.issue.name")

            val stored = ApprovalKeysApi.key(owner, name)
            assertNotNull("The server should list $name", stored)
            assertFalse("The key list should never carry the secret", stored!!.has("key"))
            assertEquals("Only one $name should exist", 1, ApprovalKeysApi.keys(owner).count { it.optString("name") == name })

            val other = "e2e-approval-keys-other-$token"
            try {
                scenario = AppAccount.switchUser(compose, scenario, other)
                ApprovalUi.openKeys(compose)
                compose.waitForTag("approval.keys.empty", timeoutMillis = 20_000)
                assertFalse("$name should not be in the other user's list", compose.exists("approval.keys.row.$name"))
                val second = BackendApi.signIn(subject = other)
                assertNull("The other user should own no $name", ApprovalKeysApi.key(second, name))
                scenario.close()
            } finally {
                AppAccount.restoreDefaultUser(compose)
            }
        } finally {
            ApprovalKeysApi.revokeAll(owner, name)
        }
    }
}
