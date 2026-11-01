package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.ui.RootTab
import java.net.URI
import java.net.URISyntaxException

object WidgetDeepLink {
    const val AFFIRMATIONS = "pocketaide://affirmations"
    const val CALENDAR = "pocketaide://calendar"
    const val NOTIFICATIONS = "pocketaide://notifications"
    const val SCRATCHPAD = "pocketaide://scratchpad"

    private val SCHEMES = setOf("pocketaide", "pocketaide-dev")
    private const val HOST_CALENDAR = "calendar"
    private val TABS = mapOf(
        "affirmations" to RootTab.Affirmations,
        "notifications" to RootTab.Notifications,
        "scratchpad" to RootTab.Scratchpad,
    )

    fun tab(uri: String): RootTab? = host(uri)?.let(TABS::get)

    fun isCalendar(uri: String): Boolean = host(uri) == HOST_CALENDAR

    private fun host(uri: String): String? {
        val parsed = try {
            URI(uri)
        } catch (e: URISyntaxException) {
            return null
        }
        val scheme = parsed.scheme?.lowercase() ?: return null
        if (scheme !in SCHEMES) return null
        return parsed.host?.lowercase()
    }
}
