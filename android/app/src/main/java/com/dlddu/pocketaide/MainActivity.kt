package com.dlddu.pocketaide

import android.Manifest
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.browser.customtabs.CustomTabsIntent
import com.dlddu.pocketaide.auth.AuthBrowser
import com.dlddu.pocketaide.designsystem.PocketAideTheme
import com.dlddu.pocketaide.push.PushDeepLink
import com.dlddu.pocketaide.ui.RootScreen

class MainActivity : ComponentActivity() {
    private val container: AppContainer
        get() = (application as PocketAideApplication).container

    private val notificationPermission =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { }

    val authBrowser = AuthBrowser { url ->
        CustomTabsIntent.Builder().setShowTitle(true).build().launchUrl(this, Uri.parse(url))
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        handle(intent)
        setContent {
            PocketAideTheme {
                RootScreen(
                    authBrowser = authBrowser,
                    onRequestNotificationPermission = ::requestNotificationPermission,
                )
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handle(intent)
    }

    override fun onPause() {
        super.onPause()
        container.redirectBus.onHostPaused()
    }

    override fun onResume() {
        super.onResume()
        container.redirectBus.onHostResumed()
    }

    private fun requestNotificationPermission() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU && !container.pushRegistrar.notificationsPermitted()) {
            notificationPermission.launch(Manifest.permission.POST_NOTIFICATIONS)
        }
    }

    private fun handle(intent: Intent?) {
        intent ?: return
        val data = intent.data
        if (data != null && data.scheme in REDIRECT_SCHEMES && data.host == REDIRECT_HOST) {
            container.redirectBus.deliver(data.toString())
            return
        }
        val extras = intent.extras ?: return
        val payload = PushDeepLink.KEYS.associateWith { key -> extras.get(key)?.toString() }
        PushDeepLink.from(payload)?.let(container.deepLinks::receive)
    }

    private companion object {
        val REDIRECT_SCHEMES = setOf("pocketaide", "pocketaide-dev")
        const val REDIRECT_HOST = "callback"
    }
}
