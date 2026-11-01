package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.scratchpad.WidgetScratchpad

object WidgetCopy {
    const val WEATHER_LABEL = "날씨"
    const val CALENDAR_LABEL = "다음 일정"
    const val AFFIRMATION_LABEL = "오늘의 다짐"
    const val NOTIFICATION_LABEL = "PocketAide 알림"
    const val SCRATCHPAD_LABEL = "임시 공간"

    const val NEEDS_LOGIN = "앱에서 로그인이 필요해요."
    const val RETRY_LATER = "잠시 후 다시 시도할게요."
    const val NEEDS_LOCATION = "앱에서 위치 접근을 허용해 주세요."
    const val NEEDS_CALENDAR = "앱에서 캘린더 접근을 허용해 주세요."
    const val NO_EVENTS = "다가오는 일정이 없어요."
    const val NO_AFFIRMATIONS = "다짐을 앱에 등록해보세요."
    const val NO_NOTIFICATIONS = "확인할 알림이 없어요."

    fun affirmationMessage(state: WidgetAffirmationState): String = when (state) {
        is WidgetAffirmationState.Loaded -> state.affirmation.text
        WidgetAffirmationState.Empty -> NO_AFFIRMATIONS
        WidgetAffirmationState.NeedsLogin -> NEEDS_LOGIN
        WidgetAffirmationState.Error -> RETRY_LATER
    }

    fun weatherLabel(state: WidgetWeatherState): String {
        val place = (state as? WidgetWeatherState.Loaded)?.place
        return if (place.isNullOrEmpty()) WEATHER_LABEL else place
    }

    fun weatherNotice(state: WidgetWeatherState): String? = when (state) {
        is WidgetWeatherState.Loaded -> null
        WidgetWeatherState.NeedsLocation -> NEEDS_LOCATION
        WidgetWeatherState.Error -> RETRY_LATER
    }

    fun calendarNotice(state: WidgetCalendarState): String? = when (state) {
        WidgetCalendarState.NeedsPermission -> NEEDS_CALENDAR
        is WidgetCalendarState.Loaded -> if (state.summary.events.isEmpty()) NO_EVENTS else null
    }

    fun moreLabel(count: Int): String? = if (count > 0) "+ $count 더" else null

    fun notificationNotice(state: WidgetNotificationState): String? = when (state) {
        is WidgetNotificationState.Loaded -> if (state.summary.latest == null) NO_NOTIFICATIONS else null
        WidgetNotificationState.NeedsLogin -> NEEDS_LOGIN
        WidgetNotificationState.Error -> RETRY_LATER
    }

    fun scratchpadText(state: WidgetScratchpadState): String = when (state) {
        is WidgetScratchpadState.Loaded -> WidgetScratchpad.countText(state.unclassified)
        WidgetScratchpadState.NeedsLogin -> NEEDS_LOGIN
        WidgetScratchpadState.Error -> RETRY_LATER
    }

    fun scratchpadHasItems(state: WidgetScratchpadState): Boolean =
        state is WidgetScratchpadState.Loaded && state.unclassified > 0
}
