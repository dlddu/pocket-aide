package com.dlddu.pocketaide.approvals

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ApprovalNotificationActionTest {
    private fun item(status: ApprovalStatus) =
        ApprovalRequest(id = "a", status = status, context = "dear-baby 정리\n세부", createdAt = 1)

    @Test
    fun actionIdentifiersMatchIos() {
        assertEquals("APPROVAL_REQUEST", ApprovalNotificationAction.CATEGORY)
        assertEquals(ApprovalDecision.APPROVE, ApprovalNotificationAction.decision("APPROVAL_APPROVE"))
        assertEquals(ApprovalDecision.REJECT, ApprovalNotificationAction.decision("APPROVAL_REJECT"))
        assertNull(ApprovalNotificationAction.decision("OTHER"))
        assertNull(ApprovalNotificationAction.decision(null))
        assertEquals("APPROVAL_APPROVE", ApprovalNotificationAction.action(ApprovalDecision.APPROVE))
        assertEquals("승인", ApprovalNotificationAction.title(ApprovalDecision.APPROVE))
        assertEquals("거절", ApprovalNotificationAction.title(ApprovalDecision.REJECT))
    }

    @Test
    fun resultNoticeDescribesOutcome() {
        assertEquals(
            "승인됨 — dear-baby 정리",
            ApprovalNotificationAction.resultNotice(Result.success(ApprovalDecisionOutcome.Decided(item(ApprovalStatus.APPROVED)))),
        )
        assertEquals(
            "이미 처리된 요청입니다 — 거절됨",
            ApprovalNotificationAction.resultNotice(Result.success(ApprovalDecisionOutcome.Conflict(item(ApprovalStatus.REJECTED)))),
        )
        assertEquals(
            ApprovalFormatting.FAILED_DECISION_NOTICE,
            ApprovalNotificationAction.resultNotice(Result.failure(IllegalStateException("offline"))),
        )
    }
}
