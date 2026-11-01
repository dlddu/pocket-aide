package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class LoginTest : E2eTest() {
    @Test
    fun relaunchWithStoredTokenLandsOnTabShell() {
        val scenario = relaunch(launchSignedIn())
        assertTrue("A stored token should land on the signed-in tab shell", compose.exists(Tags.SHELL))
        assertFalse("SignInButton should not be on screen while signed in", compose.exists(Tags.SIGN_IN))
        scenario.close()
    }
}
