package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.weather.WeatherSummary
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.Duration
import java.time.Instant

class WidgetHeaderLayoutTest {
    private val now = Instant.parse("2026-11-02T00:00:00Z")
    private val width = WidgetHeaderLayout.sectionWidth(360f)

    private fun event(title: String, hours: Long) =
        UpcomingEvent(title, now.plus(Duration.ofHours(hours)), now.plus(Duration.ofHours(hours + 1)), isAllDay = false)

    private fun entry(weather: WidgetWeatherState, calendar: WidgetCalendarState) = PocketAideWidgetEntry(
        date = now,
        affirmation = WidgetAffirmationState.Empty,
        calendar = calendar,
        weather = weather,
        notifications = WidgetNotificationState.Error,
        scratchpad = WidgetScratchpadState.Error,
    )

    private fun calendar(vararg events: UpcomingEvent, more: Int = 0) =
        WidgetCalendarState.Loaded(UpcomingEventsSummary(events.toList(), more))

    private val weather = WidgetWeatherState.Loaded(WeatherSummary(12, "맑음", 15, 4, 10), "서울")

    @Test
    fun sectionWidthSplitsWhatIsLeftAfterPaddingAndRule() {
        assertEquals(151.5f, width, 0.01f)
    }

    @Test
    fun ruleFollowsTheTallerSection() {
        val oneEvent = entry(weather, calendar(event("러닝", 2)))
        val threeEvents = entry(weather, calendar(event("러닝", 2), event("회의", 4), event("저녁", 6), more = 2))
        val short = WidgetHeaderLayout.ruleHeight(oneEvent, width, 1f)
        val tall = WidgetHeaderLayout.ruleHeight(threeEvents, width, 1f)
        val weatherHeight = WidgetHeaderLayout.height(WidgetHeaderLayout.weatherBlocks(weather), width, 1f)
        val calendarHeight = WidgetHeaderLayout.height(
            WidgetHeaderLayout.calendarBlocks(threeEvents.calendar, now),
            width,
            1f,
        )
        assertTrue(tall > short)
        assertEquals(maxOf(weatherHeight, calendarHeight), tall, 0.001f)
    }

    @Test
    fun ruleStaysWellBelowAFullWidgetHeight() {
        val busy = entry(
            WidgetWeatherState.NeedsLocation,
            calendar(event("아주 긴 제목의 일정이 두 줄로 접히는 경우", 2), event("회의", 4), event("저녁", 6), more = 5),
        )
        assertTrue(WidgetHeaderLayout.ruleHeight(busy, width, 1f) < 140f)
    }

    @Test
    fun largerFontScaleMakesTheRuleTaller() {
        val e = entry(weather, calendar(event("러닝", 2)))
        assertTrue(WidgetHeaderLayout.ruleHeight(e, width, 1.3f) > WidgetHeaderLayout.ruleHeight(e, width, 1f))
    }

    @Test
    fun longTextWrapsUpToItsMaxLines() {
        val long = "서울시청 광장에서 열리는 아주 긴 이름의 저녁 러닝 모임"
        assertEquals(1, WidgetHeaderLayout.lineCount("러닝", 13f, 1f, width, 2))
        assertEquals(2, WidgetHeaderLayout.lineCount(long, 13f, 1f, width, 2))
        assertEquals(1, WidgetHeaderLayout.lineCount(long, 13f, 1f, width, 1))
    }

    @Test
    fun hangulLinesAreTallerThanLatinOnly() {
        assertTrue(WidgetHeaderLayout.lineHeight("날씨", 11f, 1f) > WidgetHeaderLayout.lineHeight("12°", 11f, 1f))
    }
}
