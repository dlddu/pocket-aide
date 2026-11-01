package com.dlddu.pocketaide.weather

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.runTest
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertThrows
import org.junit.Before
import org.junit.Test
import java.time.Duration
import java.time.Instant
import java.time.ZoneId
import java.util.concurrent.TimeUnit

class WeatherForecastTest {
    private val server = MockWebServer()
    private val location = WeatherCoordinate.of(latitude = 37.5665, longitude = 126.9784, placeName = "서울")

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    @Test
    fun forecastSummaryMatchesWidgetSummary() {
        val body = WeatherFixtures.forecastResponse()
        val forecast = Weather.forecast(body)
        assertEquals(Weather.summary(body), forecast.summary)
        assertEquals(
            WeatherSummary(temperature = 14, condition = "맑음", high = 19, low = 13, precipitationChance = 96),
            forecast.summary,
        )
        assertEquals("체감 12°", Weather.apparentLabel(forecast))
    }

    @Test
    fun hourlyHasTwentyFourRowsFromNow() {
        val forecast = Weather.forecast(WeatherFixtures.forecastResponse())
        assertEquals(24, forecast.hours.size)
        assertEquals(WeatherHour(time = "21시", temperature = 14, condition = "비", precipitationChance = 40), forecast.hours.first())
        assertEquals("0시", forecast.hours[3].time)
        assertEquals("20시", forecast.hours.last().time)
    }

    @Test
    fun dailyHasSevenRowsStartingToday() {
        val forecast = Weather.forecast(WeatherFixtures.forecastResponse())
        assertEquals(7, forecast.days.size)
        assertEquals(
            WeatherDay(label = "오늘", high = 19, low = 13, condition = "구름 조금", precipitationChance = 96),
            forecast.days.first(),
        )
        assertEquals("10/6 (화)", forecast.days[1].label)
        assertEquals("10/11 (일)", forecast.days.last().label)
    }

    @Test
    fun extraRowsAreTrimmedAndEmptyForecastFails() {
        val forecast = Weather.forecast(WeatherFixtures.forecastResponse(hours = 30, days = 9))
        assertEquals(24, forecast.hours.size)
        assertEquals(7, forecast.days.size)
        assertThrows(WeatherException::class.java) { Weather.forecast(WeatherFixtures.forecastResponse(hours = 0)) }
    }

    @Test
    fun forecastRequestAsksForHourlyAndWeekly() {
        val url = Weather.forecastUrl(location)
        assertEquals("api.open-meteo.com", url.host)
        assertEquals("37.57", url.queryParameter("latitude"))
        assertEquals("126.98", url.queryParameter("longitude"))
        assertEquals("temperature_2m,weather_code,apparent_temperature", url.queryParameter("current"))
        assertEquals("temperature_2m,weather_code,precipitation_probability", url.queryParameter("hourly"))
        assertEquals("24", url.queryParameter("forecast_hours"))
        assertEquals("7", url.queryParameter("forecast_days"))
        assertEquals("auto", url.queryParameter("timezone"))
    }

    @Test
    fun cacheServesSameLocationWithinThirtyMinutes() {
        val cache = WeatherForecastCache(InMemoryWeatherStorage())
        val savedAt = Instant.ofEpochSecond(1_791_000_000)
        assertNull(cache.load(location, savedAt))

        cache.save(WeatherFixtures.forecastResponse(), location, savedAt)
        val hit = cache.load(location, savedAt.plus(Duration.ofMinutes(29)))
        assertNotNull(hit)
        assertEquals(savedAt, hit!!.fetchedAt)
        assertEquals(Weather.forecast(WeatherFixtures.forecastResponse()), hit.forecast)

        assertNull(cache.load(location, savedAt.plus(Duration.ofMinutes(30))))
        assertNull(cache.load(location, savedAt.minusSeconds(1)))
        val elsewhere = WeatherCoordinate.of(latitude = 35.1796, longitude = 129.0756, placeName = "부산")
        assertNull(cache.load(elsewhere, savedAt))

        cache.clear()
        assertNull(cache.load(location, savedAt))
    }

    @Test
    fun clientFetchesForecastAndCachesPayload() = runTest {
        val storage = InMemoryWeatherStorage()
        val cache = WeatherForecastCache(storage)
        val fetchedAt = Instant.ofEpochSecond(1_791_203_400)
        val client = WeatherClient(OkHttpClient(), cache, server.url("/"), clock = { fetchedAt })
        server.enqueue(MockResponse().setBody(WeatherFixtures.forecastResponse()))

        val snapshot = client.fetchForecast(location)

        val request = server.takeRequest()
        assertEquals("/v1/forecast", request.requestUrl?.encodedPath)
        assertEquals("37.57", request.requestUrl?.queryParameter("latitude"))
        assertEquals(fetchedAt, snapshot.fetchedAt)
        assertEquals(Weather.forecast(WeatherFixtures.forecastResponse()), snapshot.forecast)
        assertNotNull(cache.load(location, fetchedAt))
    }

    @Test
    fun clientReportsBadStatus() = runTest {
        val client = WeatherClient(OkHttpClient(), WeatherForecastCache(InMemoryWeatherStorage()), server.url("/"))
        server.enqueue(MockResponse().setResponseCode(503))

        val error = runCatching { client.fetchForecast(location) }.exceptionOrNull()

        assertEquals(503, (error as WeatherException.BadStatus).status)
    }

    @Test
    fun fetchToCompletionOutlivesACancelledCaller() = runTest {
        val cache = WeatherForecastCache(InMemoryWeatherStorage())
        val fetchedAt = Instant.ofEpochSecond(1_791_203_400)
        val client = WeatherClient(OkHttpClient(), cache, server.url("/"), clock = { fetchedAt })
        server.enqueue(
            MockResponse().setBody(WeatherFixtures.forecastResponse()).setBodyDelay(300, TimeUnit.MILLISECONDS),
        )

        val caller = launch(Dispatchers.Default) { client.fetchForecastToCompletion(location) }
        server.takeRequest()
        caller.cancel()
        caller.join()

        assertNotNull(cache.load(location, fetchedAt))
    }

    @Test
    fun updatedLabelShowsFetchTime() {
        val date = Instant.ofEpochSecond(1_791_203_400)
        assertEquals("21:30 갱신", Weather.updatedLabel(date, ZoneId.of("Asia/Seoul")))
        assertEquals("40%", Weather.chanceLabel(40))
        assertEquals("—", Weather.chanceLabel(null))
    }
}
