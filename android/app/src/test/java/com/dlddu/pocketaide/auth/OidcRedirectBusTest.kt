package com.dlddu.pocketaide.auth

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Test

class OidcRedirectBusTest {
    @Test
    fun deliversRedirectToPendingSignIn() = runTest {
        val bus = OidcRedirectBus()
        val pending = bus.begin()

        assertTrue(bus.deliver("pocketaide://callback?code=1"))
        assertEquals("pocketaide://callback?code=1", pending.await())
    }

    @Test
    fun ignoresRedirectWithoutPendingSignIn() {
        assertFalse(OidcRedirectBus().deliver("pocketaide://callback?code=1"))
    }

    @Test
    fun returningFromBrowserWithoutRedirectCancels() = runTest {
        val bus = OidcRedirectBus()
        val pending = bus.begin()
        bus.onHostPaused()
        bus.onHostResumed()

        try {
            pending.await()
            fail("expected Cancelled")
        } catch (_: OidcException.Cancelled) {
        }
    }

    @Test
    fun resumeBeforeLeavingDoesNotCancel() {
        val bus = OidcRedirectBus()
        val pending = bus.begin()
        bus.onHostResumed()

        assertFalse(pending.isCompleted)
    }

    @Test
    fun redirectArrivingBeforeResumeWins() = runTest {
        val bus = OidcRedirectBus()
        val pending = bus.begin()
        bus.onHostPaused()
        bus.deliver("pocketaide://callback?code=2")
        bus.onHostResumed()

        assertEquals("pocketaide://callback?code=2", pending.await())
    }
}
