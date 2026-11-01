package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.affirmations.RotationSelector
import com.dlddu.pocketaide.affirmations.SeededRng
import com.dlddu.pocketaide.notifications.ExternalNotificationsPage
import java.time.Duration
import java.time.Instant

object WidgetTimeline {
    val REFRESH_INTERVAL: Duration = Duration.ofMinutes(30)

    fun slotStart(at: Instant): Instant {
        val seconds = at.epochSecond
        return Instant.ofEpochSecond(seconds - Math.floorMod(seconds, REFRESH_INTERVAL.seconds))
    }

    fun rotationSeed(at: Instant): ULong = slotStart(at).epochSecond.toULong()

    fun pick(pool: List<Affirmation>, at: Instant, selector: RotationSelector = RotationSelector()): WidgetAffirmationState {
        if (pool.isEmpty()) return WidgetAffirmationState.Empty
        val picked = selector.pick(pool, SeededRng(rotationSeed(at))) ?: pool.first()
        return WidgetAffirmationState.Loaded(picked)
    }
}

object WidgetNotifications {
    fun summarize(page: ExternalNotificationsPage): WidgetNotificationsSummary {
        val latest = page.items.filter { it.isUnread }.maxByOrNull { it.id }
            ?: return WidgetNotificationsSummary(latest = null, moreCount = 0)
        return WidgetNotificationsSummary(latest = latest, moreCount = maxOf(0, page.unreadCount - 1))
    }
}
