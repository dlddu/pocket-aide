package com.dlddu.pocketaide.widget

import org.junit.Assert.assertEquals
import org.junit.Test
import java.time.Duration
import java.time.Instant
import java.time.ZoneId
import java.time.ZonedDateTime

class UpcomingEventsTest {
    private val zone: ZoneId = ZoneId.of("Asia/Seoul")
    private val now: Instant = ZonedDateTime.of(2026, 11, 1, 10, 0, 0, 0, zone).toInstant()

    private fun event(title: String, startMinutes: Long, durationMinutes: Long = 60, allDay: Boolean = false): UpcomingEvent {
        val start = now.plus(Duration.ofMinutes(startMinutes))
        return UpcomingEvent(title, start, start.plus(Duration.ofMinutes(durationMinutes)), allDay)
    }

    @Test
    fun showsNearestThreeInStartOrder() {
        val events = listOf(event("d", 240), event("b", 60), event("a", 30), event("c", 120))
        val summary = UpcomingEvents.summarize(events, now)
        assertEquals(listOf("a", "b", "c"), summary.events.map { it.title })
        assertEquals(1, summary.moreCount)
    }

    @Test
    fun deletedNearestLetsNextMoveUp() {
        val events = listOf(event("b", 60), event("c", 120), event("d", 240))
        val summary = UpcomingEvents.summarize(events, now)
        assertEquals(listOf("b", "c", "d"), summary.events.map { it.title })
        assertEquals(0, summary.moreCount)
    }

    @Test
    fun endedEventsAreDroppedAndOngoingKept() {
        val events = listOf(event("ended", -120, 60), event("ongoing", -30, 60), event("later", 60))
        val summary = UpcomingEvents.summarize(events, now)
        assertEquals(listOf("ongoing", "later"), summary.events.map { it.title })
    }

    @Test
    fun eventsBeyondLookaheadAreExcluded() {
        val beyond = UpcomingEvents.LOOKAHEAD.toMinutes() + 1
        val summary = UpcomingEvents.summarize(listOf(event("near", 60), event("far", beyond)), now)
        assertEquals(listOf("near"), summary.events.map { it.title })
    }

    @Test
    fun sameStartSortsByTitle() {
        val summary = UpcomingEvents.summarize(listOf(event("b", 60), event("a", 60)), now)
        assertEquals(listOf("a", "b"), summary.events.map { it.title })
    }

    @Test
    fun labels() {
        assertEquals("11:00", UpcomingEvents.startLabel(event("today", 60), now, zone))
        assertEquals("11:00 — 12:30", UpcomingEvents.rangeLabel(event("today", 60, 90), now, zone))
        assertEquals("09:30", UpcomingEvents.startLabel(event("ongoing", -30), now, zone))
        assertEquals("내일 09:00", UpcomingEvents.startLabel(event("tomorrow", 23 * 60), now, zone))
        assertEquals("11/3 10:00", UpcomingEvents.startLabel(event("later", 48 * 60), now, zone))
        val allDay = UpcomingEvent(
            "holiday",
            ZonedDateTime.of(2026, 11, 2, 0, 0, 0, 0, zone).toInstant(),
            ZonedDateTime.of(2026, 11, 3, 0, 0, 0, 0, zone).toInstant(),
            isAllDay = true,
        )
        assertEquals("내일 종일", UpcomingEvents.startLabel(allDay, now, zone))
        assertEquals("내일 종일", UpcomingEvents.rangeLabel(allDay, now, zone))
    }

    @Test
    fun allDayInstantIsLocalMidnightOfUtcDate() {
        val utcMidnight = ZonedDateTime.of(2026, 11, 2, 0, 0, 0, 0, ZoneId.of("UTC")).toInstant().toEpochMilli()
        val newYork = ZoneId.of("America/New_York")
        assertEquals(
            ZonedDateTime.of(2026, 11, 2, 0, 0, 0, 0, newYork).toInstant(),
            UpcomingEvents.allDayInstant(utcMidnight, newYork),
        )
        assertEquals(
            ZonedDateTime.of(2026, 11, 2, 0, 0, 0, 0, zone).toInstant(),
            UpcomingEvents.allDayInstant(utcMidnight, zone),
        )
    }

    @Test
    fun calendarSnapshotReportsPermissionOrSummary() {
        assertEquals(WidgetCalendarState.NeedsPermission, CalendarSnapshot(authorized = false, events = emptyList()).state(now))
        val state = CalendarSnapshot(authorized = true, events = listOf(event("a", 30))).state(now)
        assertEquals(WidgetCalendarState.Loaded(UpcomingEventsSummary(listOf(event("a", 30)), 0)), state)
    }
}
