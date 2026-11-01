package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.notifications.ExternalNotification
import com.dlddu.pocketaide.weather.WeatherSummary
import java.time.Instant

sealed interface WidgetAffirmationState {
    data class Loaded(val affirmation: Affirmation) : WidgetAffirmationState
    data object Empty : WidgetAffirmationState
    data object NeedsLogin : WidgetAffirmationState
    data object Error : WidgetAffirmationState
}

sealed interface WidgetCalendarState {
    data class Loaded(val summary: UpcomingEventsSummary) : WidgetCalendarState
    data object NeedsPermission : WidgetCalendarState
}

sealed interface WidgetWeatherState {
    data class Loaded(val summary: WeatherSummary, val place: String?) : WidgetWeatherState
    data object NeedsLocation : WidgetWeatherState
    data object Error : WidgetWeatherState
}

data class WidgetNotificationsSummary(
    val latest: ExternalNotification?,
    val moreCount: Int,
)

sealed interface WidgetNotificationState {
    data class Loaded(val summary: WidgetNotificationsSummary) : WidgetNotificationState
    data object NeedsLogin : WidgetNotificationState
    data object Error : WidgetNotificationState
}

sealed interface WidgetScratchpadState {
    data class Loaded(val unclassified: Int) : WidgetScratchpadState
    data object NeedsLogin : WidgetScratchpadState
    data object Error : WidgetScratchpadState
}

data class PocketAideWidgetEntry(
    val date: Instant,
    val affirmation: WidgetAffirmationState,
    val calendar: WidgetCalendarState,
    val weather: WidgetWeatherState,
    val notifications: WidgetNotificationState,
    val scratchpad: WidgetScratchpadState,
)
