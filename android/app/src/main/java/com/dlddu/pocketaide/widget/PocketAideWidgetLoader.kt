package com.dlddu.pocketaide.widget

import android.content.Context
import com.dlddu.pocketaide.PocketAideApplication
import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.affirmations.listAffirmations
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.notifications.ExternalNotificationFormatting
import com.dlddu.pocketaide.notifications.externalNotifications
import com.dlddu.pocketaide.scratchpad.listScratchpad
import com.dlddu.pocketaide.weather.SharedPreferencesWeatherStorage
import com.dlddu.pocketaide.weather.WeatherClient
import com.dlddu.pocketaide.weather.WeatherForecastCache
import com.dlddu.pocketaide.weather.WeatherLocationStore
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import java.time.Instant

class PocketAideWidgetLoader(context: Context) {
    private val appContext = context.applicationContext
    private val container = (appContext as PocketAideApplication).container

    suspend fun load(now: Instant = Instant.now()): PocketAideWidgetEntry = coroutineScope {
        val client = client()
        val affirmations = async { fetch(client) { listAffirmations() } }
        val notifications = async { fetch(client) { externalNotifications(limit = ExternalNotificationFormatting.PAGE_SIZE) } }
        val scratchpad = async { fetch(client) { listScratchpad().size } }
        val weather = async { fetchWeather(now) }
        val calendar = async(Dispatchers.IO) { CalendarSnapshot.load(appContext, now) }
        PocketAideWidgetEntry(
            date = now,
            affirmation = affirmationState(affirmations.await(), now),
            calendar = calendar.await().state(now),
            weather = weather.await(),
            notifications = when (val result = notifications.await()) {
                is Fetch.Success -> WidgetNotificationState.Loaded(WidgetNotifications.summarize(result.value))
                Fetch.NeedsLogin -> WidgetNotificationState.NeedsLogin
                Fetch.Error -> WidgetNotificationState.Error
            },
            scratchpad = when (val result = scratchpad.await()) {
                is Fetch.Success -> WidgetScratchpadState.Loaded(result.value)
                Fetch.NeedsLogin -> WidgetScratchpadState.NeedsLogin
                Fetch.Error -> WidgetScratchpadState.Error
            },
        )
    }

    private fun affirmationState(result: Fetch<List<Affirmation>>, now: Instant): WidgetAffirmationState =
        when (result) {
            is Fetch.Success -> WidgetTimeline.pick(result.value, now)
            Fetch.NeedsLogin -> WidgetAffirmationState.NeedsLogin
            Fetch.Error -> WidgetAffirmationState.Error
        }

    private fun client(): Fetch<ApiClient> {
        val token = try {
            container.tokenStore.load()
        } catch (e: Exception) {
            return Fetch.Error
        }
        if (token == null) return Fetch.NeedsLogin
        if (!container.backendConfigured) return Fetch.Error
        return Fetch.Success(container.api)
    }

    private suspend fun <T> fetch(client: Fetch<ApiClient>, request: suspend ApiClient.() -> T): Fetch<T> {
        val api = when (client) {
            is Fetch.Success -> client.value
            Fetch.NeedsLogin -> return Fetch.NeedsLogin
            Fetch.Error -> return Fetch.Error
        }
        return try {
            Fetch.Success(api.request())
        } catch (e: CancellationException) {
            throw e
        } catch (e: ApiException.BadStatus) {
            if (e.status == HTTP_UNAUTHORIZED) Fetch.NeedsLogin else Fetch.Error
        } catch (e: Exception) {
            Fetch.Error
        }
    }

    private suspend fun fetchWeather(now: Instant): WidgetWeatherState {
        val storage = SharedPreferencesWeatherStorage(appContext)
        val location = withContext(Dispatchers.IO) { WeatherLocationStore(storage).load() }
            ?: return WidgetWeatherState.NeedsLocation
        val cache = WeatherForecastCache(storage)
        return try {
            val snapshot = withContext(Dispatchers.IO) { cache.load(location, now) }
                ?: WeatherClient(http, cache).fetchForecast(location)
            WidgetWeatherState.Loaded(snapshot.forecast.summary, location.placeName)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            WidgetWeatherState.Error
        }
    }

    private sealed interface Fetch<out T> {
        data class Success<T>(val value: T) : Fetch<T>
        data object NeedsLogin : Fetch<Nothing>
        data object Error : Fetch<Nothing>
    }

    private companion object {
        const val HTTP_UNAUTHORIZED = 401
        val http: OkHttpClient by lazy { OkHttpClient() }
    }
}
