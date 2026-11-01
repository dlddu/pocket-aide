package com.dlddu.pocketaide.weather

import kotlinx.coroutines.test.runTest
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.time.Duration
import java.time.Instant

class WeatherViewModelTest {
    private val server = MockWebServer()
    private val storage = InMemoryWeatherStorage()
    private val store = WeatherLocationStore(storage)
    private val cache = WeatherForecastCache(storage)
    private val now = Instant.ofEpochSecond(1_791_203_400)
    private val seoul = WeatherCoordinate.of(latitude = 37.5665, longitude = 126.9784, placeName = "서울")

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): WeatherViewModel =
        WeatherViewModel(store, cache, WeatherClient(OkHttpClient(), cache, server.url("/"), clock = { now }), clock = { now })

    @Test
    fun missingLocationAsksForAccessWithoutFetching() = runTest {
        val vm = viewModel()

        vm.performLoad(useCache = true)

        assertEquals(WeatherContent.NeedsLocation, vm.state.value.content)
        assertEquals("날씨", vm.state.value.title)
        assertEquals(0, server.requestCount)
    }

    @Test
    fun loadShowsPlaceAndFreshForecast() = runTest {
        store.save(seoul)
        server.enqueue(MockResponse().setBody(WeatherFixtures.forecastResponse()))
        val vm = viewModel()

        vm.performLoad(useCache = true)

        val state = vm.state.value
        assertEquals("서울", state.title)
        val loaded = state.content as WeatherContent.Loaded
        assertEquals(now, loaded.snapshot.fetchedAt)
        assertEquals(24, loaded.snapshot.forecast.hours.size)
    }

    @Test
    fun failedFetchShowsErrorEvenWithCache() = runTest {
        store.save(seoul)
        cache.save(WeatherFixtures.forecastResponse(), seoul, now.minus(Duration.ofMinutes(5)))
        server.enqueue(MockResponse().setResponseCode(500))
        val vm = viewModel()

        vm.performLoad(useCache = true)

        assertEquals(WeatherContent.Error, vm.state.value.content)
    }

    @Test
    fun locationRefreshReloadsOnlyWhenTheLocationChanged() = runTest {
        store.save(seoul)
        server.enqueue(MockResponse().setBody(WeatherFixtures.forecastResponse()))
        val vm = viewModel()
        vm.performLoad(useCache = true)
        assertEquals(1, server.requestCount)

        vm.performLocationDidRefresh()
        assertEquals(1, server.requestCount)

        val busan = WeatherCoordinate.of(latitude = 35.1796, longitude = 129.0756, placeName = "부산")
        store.save(busan)
        server.enqueue(MockResponse().setBody(WeatherFixtures.forecastResponse()))
        vm.performLocationDidRefresh()

        assertEquals(2, server.requestCount)
        assertEquals("부산", vm.state.value.title)
        assertTrue(vm.state.value.content is WeatherContent.Loaded)
    }

    @Test
    fun clearedLocationReturnsToNeedsLocation() = runTest {
        store.save(seoul)
        server.enqueue(MockResponse().setBody(WeatherFixtures.forecastResponse()))
        val vm = viewModel()
        vm.performLoad(useCache = true)

        store.clear()
        vm.performLocationDidRefresh()

        assertEquals(WeatherContent.NeedsLocation, vm.state.value.content)
        assertNull(vm.state.value.placeName)
    }
}
