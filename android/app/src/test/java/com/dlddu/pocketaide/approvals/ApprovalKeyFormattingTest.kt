package com.dlddu.pocketaide.approvals

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test
import java.time.ZoneId
import java.time.ZonedDateTime

class ApprovalKeyFormattingTest {
    private val zone = ZoneId.of("Asia/Seoul")
    private val nowMillis = ZonedDateTime.of(2026, 10, 27, 9, 41, 0, 0, zone).toInstant().toEpochMilli()
    private val created = ZonedDateTime.of(2026, 10, 20, 8, 0, 0, 0, zone).toEpochSecond()

    @Test
    fun orderedPutsActiveKeysFirstNewestFirst() {
        val keys = listOf(
            ApprovalCallerKey(id = 1, name = "old", createdAt = 1),
            ApprovalCallerKey(id = 2, name = "revoked", createdAt = 5, revokedAt = 6),
            ApprovalCallerKey(id = 3, name = "new", createdAt = 3),
        )
        assertEquals(listOf("new", "old", "revoked"), ApprovalKeyFormatting.ordered(keys).map { it.name })
    }

    @Test
    fun metaForSecretKeyAndServiceAccount() {
        val usedToday = ZonedDateTime.of(2026, 10, 27, 8, 5, 0, 0, zone).toEpochSecond()
        val key = ApprovalCallerKey(id = 1, name = "infra-ci", keyPrefix = "pa_ab12", createdAt = created, lastUsedAt = usedToday)
        assertEquals("pa_ab12… · 발급 10/20 · 마지막 사용 08:05", ApprovalKeyFormatting.meta(key, nowMillis, zone))
        val account = ApprovalCallerKey(id = 2, name = "mcp", serviceAccount = "homelab/mcp", createdAt = created, lastUsedAt = created)
        assertEquals("SA homelab/mcp · 등록 10/20 · 마지막 사용 10/20", ApprovalKeyFormatting.meta(account, nowMillis, zone))
        assertEquals(
            "pa_ab12… · 발급 10/20 · 마지막 사용 없음",
            ApprovalKeyFormatting.meta(key.copy(lastUsedAt = null), nowMillis, zone),
        )
    }

    @Test
    fun revokedLabelAndMessages() {
        val key = ApprovalCallerKey(id = 1, name = "infra-ci", createdAt = created)
        assertNull(ApprovalKeyFormatting.revokedLabel(key, zone))
        assertEquals("폐기됨 10/20", ApprovalKeyFormatting.revokedLabel(key.copy(revokedAt = created), zone))
        assertEquals("infra-ci 키를 폐기할까요?", ApprovalKeyFormatting.revokeTitle(key))
        assertEquals(ApprovalKeyFormatting.REVOKE_MESSAGE, ApprovalKeyFormatting.revokeMessage(key))
        assertEquals(
            ApprovalKeyFormatting.SERVICE_ACCOUNT_REVOKE_MESSAGE,
            ApprovalKeyFormatting.revokeMessage(key.copy(serviceAccount = "a/b")),
        )
    }

    @Test
    fun serviceAccountRequiresDnsLabels() {
        assertEquals("homelab/k3s-mcp", ApprovalKeyFormatting.serviceAccount(" homelab ", "k3s-mcp"))
        assertNull(ApprovalKeyFormatting.serviceAccount("Homelab", "mcp"))
        assertNull(ApprovalKeyFormatting.serviceAccount("homelab", "-mcp"))
        assertNull(ApprovalKeyFormatting.serviceAccount("a".repeat(64), "mcp"))
        assertNull(ApprovalKeyFormatting.trimmedName("   "))
        assertEquals("x", ApprovalKeyFormatting.trimmedName(" x "))
    }
}
