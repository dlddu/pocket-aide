package com.dlddu.pocketaide.approvals

import com.dlddu.pocketaide.api.ApiClient
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.ZoneId
import java.time.ZonedDateTime

class ApprovalFormattingTest {
    private val zone = ZoneId.of("Asia/Seoul")
    private val nowMillis = ZonedDateTime.of(2026, 10, 27, 9, 41, 12, 0, zone).toInstant().toEpochMilli()
    private val nowSeconds = nowMillis / 1000

    private fun request(
        context: String = "kubectl delete deploy dear-baby-v1",
        status: ApprovalStatus = ApprovalStatus.PENDING,
        expiresAt: Long? = null,
    ) = ApprovalRequest(id = "a1", status = status, context = context, createdAt = nowSeconds - 60, expiresAt = expiresAt)

    @Test
    fun remainingShowsMinutesAndHours() {
        assertEquals("기한 없음", ApprovalFormatting.remaining(null, nowMillis))
        assertEquals("04:05 남음", ApprovalFormatting.remaining(nowSeconds + 245, nowMillis))
        assertEquals("1:00:05 남음", ApprovalFormatting.remaining(nowSeconds + 3605, nowMillis))
        assertEquals("00:00 남음", ApprovalFormatting.remaining(nowSeconds - 10, nowMillis))
        assertEquals("02:00", ApprovalFormatting.countdown(nowSeconds + 120, nowMillis))
    }

    @Test
    fun progressIsClampedFractionLeft() {
        assertEquals(0.5f, ApprovalFormatting.progress(nowSeconds - 100, nowSeconds + 100, nowMillis), 0.0001f)
        assertEquals(0f, ApprovalFormatting.progress(nowSeconds - 100, nowSeconds - 50, nowMillis), 0.0001f)
        assertEquals(0f, ApprovalFormatting.progress(nowSeconds, nowSeconds, nowMillis), 0.0001f)
    }

    @Test
    fun summaryIsFirstNonBlankLine() {
        assertEquals("배포 정리", ApprovalFormatting.summary("\n   \n  배포 정리  \n세부 내용"))
        assertEquals("", ApprovalFormatting.summary("  \n "))
    }

    @Test
    fun confirmMessagePicksObjectParticle() {
        assertEquals(
            "배포 정리를 허락합니다. 확정하면 되돌릴 수 없습니다.",
            ApprovalFormatting.confirmMessage(ApprovalDecision.APPROVE, "배포 정리"),
        )
        assertEquals(
            "삭제 명령을 거절합니다. 확정하면 되돌릴 수 없습니다.",
            ApprovalFormatting.confirmMessage(ApprovalDecision.REJECT, "삭제 명령"),
        )
        assertEquals("을(를)", ApprovalFormatting.objectParticle("kubectl"))
        assertEquals("승인할까요?", ApprovalFormatting.confirmTitle(ApprovalDecision.APPROVE))
    }

    @Test
    fun sessionAndReasonLabels() {
        assertNull(ApprovalFormatting.sessionLabel(request()))
        assertEquals("정리 세션", ApprovalFormatting.sessionLabel(request().copy(sessionName = " 정리 세션 ", sessionId = "abc")))
        assertEquals("제목 없음 · 0123abcd", ApprovalFormatting.sessionLabel(request().copy(sessionName = "", sessionId = "0123abcdef")))
        assertNull(ApprovalFormatting.agentReason(request().copy(reason = "  ")))
        assertEquals("이유", ApprovalFormatting.agentReason(request().copy(reason = " 이유 ")))
    }

    @Test
    fun decisionModeAndHistoryMeta() {
        assertEquals("만료 — 결정 없음", ApprovalFormatting.decisionModeLabel(ApprovalStatus.EXPIRED, null))
        assertEquals("자동 거절", ApprovalFormatting.decisionModeLabel(ApprovalStatus.REJECTED, "auto"))
        assertEquals("자동 승인", ApprovalFormatting.decisionModeLabel(ApprovalStatus.APPROVED, "auto_approve"))
        assertEquals("수동 결정", ApprovalFormatting.decisionModeLabel(ApprovalStatus.APPROVED, "manual"))
        assertEquals(
            "수동 결정 · 폐기된 키로 생성됨",
            ApprovalFormatting.historyMeta(request(status = ApprovalStatus.APPROVED).copy(decisionMode = "manual", keyRevoked = true)),
        )
    }

    @Test
    fun dayRelativeTimes() {
        assertEquals("09:40", ApprovalFormatting.requestTime(nowSeconds - 60, nowMillis, zone))
        assertEquals("09:40:12", ApprovalFormatting.historyTime(nowSeconds - 60, nowMillis, zone))
        val yesterday = ZonedDateTime.of(2026, 10, 26, 22, 10, 0, 0, zone).toEpochSecond()
        assertEquals("어제 22:10", ApprovalFormatting.requestTime(yesterday, nowMillis, zone))
        val older = ZonedDateTime.of(2026, 10, 24, 18, 2, 0, 0, zone).toEpochSecond()
        assertEquals("10/24 18:02", ApprovalFormatting.historyTime(older, nowMillis, zone))
        assertEquals("—", ApprovalFormatting.clockTime(null, zone))
    }

    @Test
    fun conflictNoticeDependsOnStatus() {
        assertEquals("만료된 요청이라 결정되지 않았습니다", ApprovalFormatting.conflictNotice(request(status = ApprovalStatus.EXPIRED)))
        assertEquals("이미 처리된 요청입니다 — 승인됨", ApprovalFormatting.conflictNotice(request(status = ApprovalStatus.APPROVED)))
    }

    @Test
    fun expiryFollowsDeadlineOnlyWhilePending() {
        assertTrue(request(expiresAt = nowSeconds).isExpired(nowMillis))
        assertFalse(request(expiresAt = nowSeconds + 1).isExpired(nowMillis))
        assertFalse(request().isExpired(nowMillis))
        assertTrue(request(status = ApprovalStatus.EXPIRED).isExpired(nowMillis))
        assertFalse(request(status = ApprovalStatus.APPROVED, expiresAt = nowSeconds - 5).isExpired(nowMillis))
    }

    @Test
    fun newestFirstSortsByCreatedThenId() {
        val items = listOf(
            ApprovalRequest(id = "a", context = "x", createdAt = 1),
            ApprovalRequest(id = "c", context = "x", createdAt = 2),
            ApprovalRequest(id = "b", context = "x", createdAt = 2),
        )
        assertEquals(listOf("c", "b", "a"), ApprovalFormatting.newestFirst(items).map { it.id })
    }

    @Test
    fun autoResponseCopy() {
        assertNull(ApprovalAutoResponse.OFF.bannerTitle)
        assertEquals("자동 승인 모드가 켜져 있습니다.", ApprovalAutoResponse.AUTO_APPROVE.bannerTitle)
        assertEquals("없음 — 알림으로 받기", ApprovalAutoResponse.OFF.optionLabel)
        assertEquals("새 요청은 알림 없이 바로 거절됩니다.", ApprovalAutoResponse.AUTO_REJECT.outcomeNotice)
    }

    @Test
    fun decodesBackendApprovalJson() {
        val raw = """{"id":"apr_1","status":"APPROVED","external_id":"ext-1","context":"c","requester_name":"","session_id":"",""" +
            """"session_name":"","reason":"","key_name":"infra-ci","key_revoked":false,"decision_mode":"manual",""" +
            """"created_at":10,"expires_at":null,"processed_at":20}"""
        val item = ApiClient.json.decodeFromString(ApprovalRequest.serializer(), raw)
        assertEquals(ApprovalStatus.APPROVED, item.status)
        assertEquals("ext-1", item.externalId)
        assertEquals(20L, item.processedAt)
        assertNull(item.expiresAt)
    }

    @Test
    fun cursorIsPercentEncoded() {
        assertEquals("abc%2B%2F%3D", percentEncodeAlphanumerics("abc+/="))
    }
}
