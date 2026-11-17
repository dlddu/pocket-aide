package com.dlddu.pocketaide.auth

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.util.TimeZone

class SignOutReasonTest {
    @Test
    fun clientErrorReadsAsRejectionWithOauthError() {
        val reason = SignOutReason.refreshHttpFailure(400, """{"error":"invalid_grant"}""", 0)

        assertEquals("토큰 갱신 거절 — HTTP 400: invalid_grant", reason.detail)
    }

    @Test
    fun nonJsonBodyIsFlattenedAndCapped() {
        val reason = SignOutReason.refreshHttpFailure(502, "<html>\n  bad   gateway\n</html>" + "x".repeat(300), 0)

        assertTrue(reason.detail.startsWith("토큰 갱신 실패 — HTTP 502: <html> bad gateway </html>xxx"))
        assertTrue(reason.detail.endsWith("x…"))
    }

    @Test
    fun unauthorizedNamesRequestAndKeepsEarlierRefreshFailure() {
        val stored = SignOutReason("토큰 갱신 거절 — HTTP 400: invalid_grant", 1_000)

        val reason = SignOutReason.unauthorized(stored, "GET /api/sessions", "invalid token: oidc: token is expired\n", 5_000)

        assertEquals(
            SignOutReason(
                "토큰 갱신 거절 — HTTP 400: invalid_grant\nAPI 응답 — GET /api/sessions → HTTP 401: invalid token: oidc: token is expired",
                5_000,
            ),
            reason,
        )
    }

    @Test
    fun unreadableTokenNamesTheError() {
        val reason = SignOutReason.tokenUnreadable(IllegalStateException("key invalidated"), 0)

        assertEquals("저장된 로그인 토큰을 읽지 못해 지웠어요 — IllegalStateException: key invalidated", reason.detail)
    }

    @Test
    fun messageLeadsWithLocalTime() {
        val reason = SignOutReason("API 응답 — GET /api/me → HTTP 401", 1_763_253_120_000)

        assertEquals(
            "11월 16일 오전 9:32\n\nAPI 응답 — GET /api/me → HTTP 401",
            reason.message(TimeZone.getTimeZone("Asia/Seoul")),
        )
    }
}
