package com.dlddu.pocketaide.auth

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class AuthNoticesTest {
    private val store = InMemorySignOutReasonStore()
    private val notices = AuthNotices(store, clock = { 7L })

    @Test
    fun launchWithTokenRemembersSessionWithoutNotice() {
        notices.onLaunch(hasToken = true)

        assertTrue(store.hadSession())
        assertNull(notices.current.value)
    }

    @Test
    fun launchWithoutTokenAfterSessionShowsRecordedReason() {
        val rejected = SignOutReason("토큰 갱신 거절 — HTTP 400: invalid_grant", 3)
        store.save(rejected)
        store.markSession()

        notices.onLaunch(hasToken = false)

        assertEquals(AuthNotice(AuthNotice.Kind.SIGNED_OUT, rejected), notices.current.value)
    }

    @Test
    fun launchWithoutTokenOrRecordSaysCauseIsUnknown() {
        store.markSession()

        notices.onLaunch(hasToken = false)

        assertEquals(AuthNotice(AuthNotice.Kind.SIGNED_OUT, SignOutReason.tokenMissing(7)), notices.current.value)
    }

    @Test
    fun launchWithoutTokenAndNoEarlierSessionStaysQuiet() {
        notices.onLaunch(hasToken = false)

        assertNull(notices.current.value)
    }

    @Test
    fun unauthorizedShowsRequestAndKeepsFirstNotice() {
        notices.onUnauthorized("GET /api/sessions", "invalid token")
        notices.onUnauthorized("GET /api/todos", "invalid token")

        assertEquals(
            AuthNotice(AuthNotice.Kind.UNAUTHORIZED, SignOutReason("API 응답 — GET /api/sessions → HTTP 401: invalid token", 7)),
            notices.current.value,
        )
    }

    @Test
    fun dismissingUnauthorizedKeepsSessionRecord() {
        store.markSession()
        store.save(SignOutReason("토큰 갱신 실패 — HTTP 503", 1))
        notices.onUnauthorized("GET /api/me", "")

        notices.dismiss()

        assertNull(notices.current.value)
        assertTrue(store.hadSession())
        assertEquals(SignOutReason("토큰 갱신 실패 — HTTP 503", 1), store.load())
    }

    @Test
    fun expiredSessionReplacesUnauthorizedAndSurvivesRestartUntilDismissed() {
        store.markSession()
        notices.onUnauthorized("GET /api/me", "invalid token")
        notices.onSessionExpired("GET /api/me", "invalid token")
        val expired = AuthNotice(AuthNotice.Kind.SIGNED_OUT, SignOutReason("API 응답 — GET /api/me → HTTP 401: invalid token", 7))
        assertEquals(expired, notices.current.value)

        val afterRestart = AuthNotices(store, clock = { 9L }).apply { onLaunch(hasToken = false) }
        assertEquals(expired, afterRestart.current.value)

        afterRestart.dismiss()
        assertNull(store.load())
        assertFalse(store.hadSession())
    }

    @Test
    fun signingOutByHandForgetsEverything() {
        store.markSession()
        store.save(SignOutReason("토큰 갱신 실패 — HTTP 503", 1))

        notices.onSignedOutByUser()
        notices.onLaunch(hasToken = false)

        assertNull(notices.current.value)
        assertNull(store.load())
    }
}
