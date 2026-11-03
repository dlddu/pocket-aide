package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.affirmations.AffirmationPriority
import com.dlddu.pocketaide.notifications.ExternalNotification
import com.dlddu.pocketaide.notifications.ExternalNotificationsPage
import com.dlddu.pocketaide.ui.RootTab
import com.dlddu.pocketaide.weather.WeatherDeepLink
import com.dlddu.pocketaide.weather.WeatherSummary
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.Instant

class WidgetSectionsTest {
    private fun notification(id: Long, readAt: Long? = null) =
        ExternalNotification(id = id, title = "n$id", body = "b$id", readAt = readAt, createdAt = id)

    @Test
    fun notificationsShowLatestUnreadAndCountTheRest() {
        val page = ExternalNotificationsPage(
            items = listOf(notification(5, readAt = 10), notification(4), notification(3)),
            unreadCount = 2,
        )
        val summary = WidgetNotifications.summarize(page)
        assertEquals(4L, summary.latest?.id)
        assertEquals(1, summary.moreCount)
    }

    @Test
    fun notificationsWithoutUnreadAreEmpty() {
        val page = ExternalNotificationsPage(items = listOf(notification(1, readAt = 2)), unreadCount = 0)
        assertEquals(WidgetNotificationsSummary(latest = null, moreCount = 0), WidgetNotifications.summarize(page))
        val state = WidgetNotificationState.Loaded(WidgetNotifications.summarize(page))
        assertEquals("확인할 알림이 없어요.", WidgetCopy.notificationNotice(state))
    }

    @Test
    fun notificationNotices() {
        val loaded = WidgetNotificationState.Loaded(WidgetNotificationsSummary(notification(1), 0))
        assertNull(WidgetCopy.notificationNotice(loaded))
        assertEquals("앱에서 로그인이 필요해요.", WidgetCopy.notificationNotice(WidgetNotificationState.NeedsLogin))
        assertEquals("잠시 후 다시 시도할게요.", WidgetCopy.notificationNotice(WidgetNotificationState.Error))
    }

    @Test
    fun affirmationMessages() {
        val affirmation = Affirmation(1, "작게 시작해서 매일 1%씩.", AffirmationPriority.HIGH, 0, 0)
        assertEquals(affirmation.text, WidgetCopy.affirmationMessage(WidgetAffirmationState.Loaded(affirmation)))
        assertEquals("다짐을 앱에 등록해보세요.", WidgetCopy.affirmationMessage(WidgetAffirmationState.Empty))
        assertEquals("앱에서 로그인이 필요해요.", WidgetCopy.affirmationMessage(WidgetAffirmationState.NeedsLogin))
        assertEquals("잠시 후 다시 시도할게요.", WidgetCopy.affirmationMessage(WidgetAffirmationState.Error))
    }

    @Test
    fun weatherLabelUsesPlaceWhenKnown() {
        val summary = WeatherSummary(18, "비", 22, 14, 60)
        assertEquals("서울", WidgetCopy.weatherLabel(WidgetWeatherState.Loaded(summary, "서울")))
        assertEquals("날씨", WidgetCopy.weatherLabel(WidgetWeatherState.Loaded(summary, null)))
        assertEquals("날씨", WidgetCopy.weatherLabel(WidgetWeatherState.Loaded(summary, "")))
        assertEquals("날씨", WidgetCopy.weatherLabel(WidgetWeatherState.NeedsLocation))
        assertNull(WidgetCopy.weatherNotice(WidgetWeatherState.Loaded(summary, "서울")))
        assertEquals("앱에서 위치 접근을 허용해 주세요.", WidgetCopy.weatherNotice(WidgetWeatherState.NeedsLocation))
        assertEquals("잠시 후 다시 시도할게요.", WidgetCopy.weatherNotice(WidgetWeatherState.Error))
    }

    @Test
    fun calendarNotices() {
        assertEquals("앱에서 캘린더 접근을 허용해 주세요.", WidgetCopy.calendarNotice(WidgetCalendarState.NeedsPermission))
        val empty = WidgetCalendarState.Loaded(UpcomingEventsSummary(emptyList(), 0))
        assertEquals("다가오는 일정이 없어요.", WidgetCopy.calendarNotice(empty))
        val event = UpcomingEvent("분기 리뷰", Instant.EPOCH, Instant.EPOCH.plusSeconds(60), isAllDay = false)
        assertNull(WidgetCopy.calendarNotice(WidgetCalendarState.Loaded(UpcomingEventsSummary(listOf(event), 0))))
        assertEquals("+ 2 더", WidgetCopy.moreLabel(2))
        assertNull(WidgetCopy.moreLabel(0))
    }

    @Test
    fun scratchpadCountText() {
        assertEquals("미분류 3개", WidgetCopy.scratchpadText(WidgetScratchpadState.Loaded(3)))
        assertTrue(WidgetCopy.scratchpadHasItems(WidgetScratchpadState.Loaded(3)))
        assertEquals("정리할 항목 없음", WidgetCopy.scratchpadText(WidgetScratchpadState.Loaded(0)))
        assertFalse(WidgetCopy.scratchpadHasItems(WidgetScratchpadState.Loaded(0)))
        assertEquals("앱에서 로그인이 필요해요.", WidgetCopy.scratchpadText(WidgetScratchpadState.NeedsLogin))
        assertEquals("잠시 후 다시 시도할게요.", WidgetCopy.scratchpadText(WidgetScratchpadState.Error))
        assertFalse(WidgetCopy.scratchpadHasItems(WidgetScratchpadState.Error))
    }

    @Test
    fun deepLinksRouteToTabsAndCalendar() {
        assertEquals(RootTab.Affirmations, WidgetDeepLink.tab(WidgetDeepLink.AFFIRMATIONS))
        assertEquals(RootTab.Notifications, WidgetDeepLink.tab(WidgetDeepLink.NOTIFICATIONS))
        assertEquals(RootTab.Scratchpad, WidgetDeepLink.tab(WidgetDeepLink.SCRATCHPAD))
        assertNull(WidgetDeepLink.tab(WidgetDeepLink.CALENDAR))
        assertNull(WidgetDeepLink.tab(WeatherDeepLink.URL))
        assertTrue(WidgetDeepLink.isCalendar(WidgetDeepLink.CALENDAR))
        assertFalse(WidgetDeepLink.isCalendar(WidgetDeepLink.SCRATCHPAD))
        assertNull(WidgetDeepLink.tab(OTHER_SCHEME))
        assertNull(WidgetDeepLink.tab("not a uri"))
    }

    private companion object {
        const val OTHER_SCHEME = "https://affirmations"
    }
}
