package com.dlddu.pocketaide.widget

import java.time.Duration
import java.time.Instant
import java.time.ZoneId
import java.time.ZoneOffset
import java.util.Locale

data class UpcomingEvent(
    val title: String,
    val start: Instant,
    val end: Instant,
    val isAllDay: Boolean,
)

data class UpcomingEventsSummary(
    val events: List<UpcomingEvent>,
    val moreCount: Int,
)

object UpcomingEvents {
    const val LIMIT = 3
    val LOOKAHEAD: Duration = Duration.ofDays(7)

    fun summarize(events: List<UpcomingEvent>, now: Instant): UpcomingEventsSummary {
        val horizon = now.plus(LOOKAHEAD)
        val upcoming = events
            .filter { it.end > now && it.start < horizon }
            .sortedWith(compareBy<UpcomingEvent> { it.start }.thenBy { it.title })
        return UpcomingEventsSummary(
            events = upcoming.take(LIMIT),
            moreCount = maxOf(0, upcoming.size - LIMIT),
        )
    }

    fun startLabel(event: UpcomingEvent, now: Instant, zone: ZoneId = ZoneId.systemDefault()): String {
        val time = if (event.isAllDay) "종일" else clock(event.start, zone)
        val startDay = event.start.atZone(zone).toLocalDate()
        val today = now.atZone(zone).toLocalDate()
        if (event.start <= now || startDay == today) return time
        if (startDay == today.plusDays(1)) return "내일 $time"
        return "${startDay.monthValue}/${startDay.dayOfMonth} $time"
    }

    fun rangeLabel(event: UpcomingEvent, now: Instant, zone: ZoneId = ZoneId.systemDefault()): String {
        val start = startLabel(event, now, zone)
        if (event.isAllDay) return start
        return "$start — ${clock(event.end, zone)}"
    }

    fun allDayInstant(utcMillis: Long, zone: ZoneId): Instant =
        Instant.ofEpochMilli(utcMillis).atZone(ZoneOffset.UTC).toLocalDate().atStartOfDay(zone).toInstant()

    private fun clock(instant: Instant, zone: ZoneId): String {
        val local = instant.atZone(zone)
        return String.format(Locale.ROOT, "%02d:%02d", local.hour, local.minute)
    }
}
