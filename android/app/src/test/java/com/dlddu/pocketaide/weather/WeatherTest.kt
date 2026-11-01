package com.dlddu.pocketaide.weather

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Test

class WeatherTest {
    @Test
    fun decodesCurrentTemperatureConditionAndDailyRange() {
        val summary = Weather.summary(WeatherFixtures.summaryResponse())
        assertEquals(
            WeatherSummary(temperature = 23, condition = "대체로 맑음", high = 23, low = 14, precipitationChance = 82),
            summary,
        )
        assertEquals("23°", Weather.temperatureLabel(summary))
        assertEquals("최고 23 · 최저 14 · 강수 82%", Weather.detailLabel(summary))
    }

    @Test
    fun missingPrecipitationLeavesChanceOut() {
        val summary = Weather.summary(WeatherFixtures.summaryResponse(precipitation = "[null]"))
        assertNull(summary.precipitationChance)
        assertEquals("최고 23 · 최저 14", Weather.detailLabel(summary))
    }

    @Test
    fun negativeTemperatureRoundsToNearest() {
        val summary = Weather.summary(WeatherFixtures.summaryResponse(temperature = "-3.6"))
        assertEquals("-4°", Weather.temperatureLabel(summary))
        assertEquals(-3, Weather.wholeDegrees(-2.5))
        assertEquals(3, Weather.wholeDegrees(2.5))
    }

    @Test
    fun missingDailyValuesFail() {
        val body = """{"current":{"temperature_2m":1.0,"weather_code":0},"daily":{"temperature_2m_max":[null],"temperature_2m_min":[null]}}"""
        assertThrows(WeatherException.MissingDaily::class.java) { Weather.summary(body) }
    }

    @Test
    fun conditionCodesMapToKoreanLabels() {
        assertEquals("맑음", Weather.condition(0))
        assertEquals("흐림", Weather.condition(3))
        assertEquals("안개", Weather.condition(48))
        assertEquals("비", Weather.condition(63))
        assertEquals("눈", Weather.condition(75))
        assertEquals("소나기", Weather.condition(81))
        assertEquals("뇌우", Weather.condition(95))
        assertEquals("날씨 정보 없음", Weather.condition(42))
    }

    @Test
    fun requestUsesRoundedCoordinates() {
        val location = WeatherCoordinate.of(latitude = 37.5665, longitude = 126.9784, placeName = "서울")
        assertEquals(37.57, location.latitude, 0.0)
        assertEquals(126.98, location.longitude, 0.0)
        val url = Weather.requestUrl(location)
        assertEquals("api.open-meteo.com", url.host)
        assertEquals("/v1/forecast", url.encodedPath)
        assertEquals("37.57", url.queryParameter("latitude"))
        assertEquals("126.98", url.queryParameter("longitude"))
        assertEquals("temperature_2m,weather_code", url.queryParameter("current"))
        assertEquals("temperature_2m_max,temperature_2m_min,precipitation_probability_max", url.queryParameter("daily"))
    }

    @Test
    fun locationStoreRoundTripsAndClears() {
        val store = WeatherLocationStore(InMemoryWeatherStorage())
        assertNull(store.load())
        val location = WeatherCoordinate.of(latitude = 35.1796, longitude = 129.0756, placeName = "부산")
        store.save(location)
        assertEquals(location, store.load())
        store.clear()
        assertNull(store.load())
    }

    @Test
    fun locationStoreIgnoresCorruptValues() {
        val storage = InMemoryWeatherStorage()
        storage.putString("weather.location", "not json")
        assertNull(WeatherLocationStore(storage).load())
    }

    @Test
    fun deepLinkMatchesWeatherHostOnly() {
        assertTrue(WeatherDeepLink.matches("pocketaide://weather"))
        assertTrue(WeatherDeepLink.matches("pocketaide-dev://weather"))
        assertFalse(WeatherDeepLink.matches("pocketaide://oidc/callback?code=x"))
        assertFalse(WeatherDeepLink.matches("https://weather"))
        assertFalse(WeatherDeepLink.matches("not a uri"))
    }
}
