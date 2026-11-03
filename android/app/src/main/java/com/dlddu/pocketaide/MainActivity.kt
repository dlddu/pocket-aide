package com.dlddu.pocketaide

import android.Manifest
import android.content.ActivityNotFoundException
import android.content.ContentUris
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.provider.CalendarContract
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.browser.customtabs.CustomTabsIntent
import com.dlddu.pocketaide.auth.AuthBrowser
import com.dlddu.pocketaide.auth.OidcRedirectBus
import com.dlddu.pocketaide.designsystem.PocketAideTheme
import com.dlddu.pocketaide.push.PushDeepLink
import com.dlddu.pocketaide.ui.RootScreen
import com.dlddu.pocketaide.ui.RootTab
import com.dlddu.pocketaide.weather.SharedPreferencesWeatherStorage
import com.dlddu.pocketaide.weather.WeatherDeepLink
import com.dlddu.pocketaide.weather.WeatherLocationAccess
import com.dlddu.pocketaide.weather.WeatherLocationStore
import com.dlddu.pocketaide.widget.CalendarSnapshot
import com.dlddu.pocketaide.widget.WidgetDeepLink
import com.dlddu.pocketaide.widget.WidgetRefresher
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    private val container: AppContainer
        get() = (application as PocketAideApplication).container

    private val weatherStore by lazy { WeatherLocationStore(SharedPreferencesWeatherStorage(this)) }
    private val weatherAccess by lazy { WeatherLocationAccess(this, weatherStore) }
    private var locating = false

    private val startupPermissions =
        registerForActivityResult(ActivityResultContracts.RequestMultiplePermissions()) { results ->
            if (results[CalendarSnapshot.PERMISSION] == true) WidgetRefresher.reloadAll(this)
            if (WeatherLocationAccess.PERMISSIONS.any { results[it] == true }) refreshWeatherLocation()
        }

    val authBrowser = AuthBrowser { url ->
        CustomTabsIntent.Builder().setShowTitle(true).build().launchUrl(this, Uri.parse(url))
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        handle(intent, restored = savedInstanceState != null)
        setContent {
            PocketAideTheme {
                RootScreen(
                    authBrowser = authBrowser,
                    onRequestStartupPermissions = ::requestStartupPermissions,
                )
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handle(intent, restored = false)
    }

    override fun onStart() {
        super.onStart()
        if (container.tokenStore.load() != null) refreshWeatherLocation()
    }

    override fun onPause() {
        super.onPause()
        container.redirectBus.onHostPaused()
    }

    override fun onResume() {
        super.onResume()
        container.redirectBus.onHostResumed()
    }

    private fun requestStartupPermissions() {
        val missing = buildList {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU && !container.pushRegistrar.notificationsPermitted()) {
                add(Manifest.permission.POST_NOTIFICATIONS)
            }
            if (!CalendarSnapshot.hasPermission(this@MainActivity)) add(CalendarSnapshot.PERMISSION)
            if (!weatherAccess.hasPrecisePermission()) addAll(WeatherLocationAccess.PERMISSIONS)
        }
        if (missing.isNotEmpty()) startupPermissions.launch(missing.toTypedArray())
        if (weatherAccess.hasPermission()) refreshWeatherLocation()
    }

    private fun refreshWeatherLocation() {
        if (locating) return
        locating = true
        container.appScope.launch {
            val before = weatherStore.load()
            try {
                weatherAccess.refresh()
            } finally {
                locating = false
            }
            if (weatherStore.load() != before) WidgetRefresher.reloadAll(this@MainActivity)
        }
    }

    private fun openCalendar() {
        val uri = CalendarContract.CONTENT_URI.buildUpon().appendPath("time").also {
            ContentUris.appendId(it, System.currentTimeMillis())
        }.build()
        try {
            startActivity(Intent(Intent.ACTION_VIEW, uri))
        } catch (e: ActivityNotFoundException) {
            container.deepLinks.openTab(RootTab.Affirmations)
        }
    }

    private fun handle(intent: Intent?, restored: Boolean) {
        intent ?: return
        val data = intent.data
        if (data != null && WeatherDeepLink.matches(data.toString())) {
            container.deepLinks.openWeather()
            return
        }
        if (data != null) {
            val link = data.toString()
            WidgetDeepLink.tab(link)?.let {
                container.deepLinks.openTab(it)
                return
            }
            if (WidgetDeepLink.isCalendar(link)) {
                if (!restored) openCalendar()
                return
            }
        }
        if (data != null && OidcRedirectBus.isRedirect(data.toString())) {
            container.redirectBus.deliver(data.toString())
            return
        }
        val extras = intent.extras ?: return
        val payload = PushDeepLink.KEYS.associateWith { key -> extras.get(key)?.toString() }
        PushDeepLink.from(payload)?.let(container.deepLinks::receive)
    }
}
