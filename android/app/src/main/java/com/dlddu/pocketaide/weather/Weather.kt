package com.dlddu.pocketaide.weather

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.NonCancellable
import kotlinx.coroutines.withContext
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.IOException
import java.time.Duration
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.time.format.DateTimeParseException
import java.util.Locale
import kotlin.math.abs
import kotlin.math.floor
import kotlin.math.sign

@Serializable
data class WeatherCoordinate(
    val latitude: Double,
    val longitude: Double,
    val placeName: String? = null,
) {
    companion object {
        fun of(latitude: Double, longitude: Double, placeName: String?): WeatherCoordinate =
            WeatherCoordinate(Weather.rounded(latitude), Weather.rounded(longitude), placeName)
    }
}

data class WeatherSummary(
    val temperature: Int,
    val condition: String,
    val high: Int,
    val low: Int,
    val precipitationChance: Int?,
)

data class WeatherHour(
    val time: String,
    val temperature: Int,
    val condition: String,
    val precipitationChance: Int?,
)

data class WeatherDay(
    val label: String,
    val high: Int,
    val low: Int,
    val condition: String,
    val precipitationChance: Int?,
)

data class WeatherForecast(
    val summary: WeatherSummary,
    val apparentTemperature: Int,
    val hours: List<WeatherHour>,
    val days: List<WeatherDay>,
)

data class WeatherForecastSnapshot(val forecast: WeatherForecast, val fetchedAt: Instant)

sealed class WeatherException(message: String) : Exception(message) {
    class BadStatus(val status: Int) : WeatherException("HTTP $status")
    class MissingDaily : WeatherException("missing daily")
    class Decoding(cause: Throwable) : WeatherException("decode: ${cause.message}")
    class Transport(cause: Throwable) : WeatherException("transport: ${cause.message}")
}

object Weather {
    const val FORECAST_HOURS = 24
    const val FORECAST_DAYS = 7
    val DEFAULT_BASE_URL: HttpUrl = "https://api.open-meteo.com".toHttpUrl()

    private val weekdays = listOf("일", "월", "화", "수", "목", "금", "토")
    private val updatedFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("HH:mm", Locale.KOREA)
    private val conditions: List<Pair<Set<Int>, String>> = listOf(
        setOf(0) to "맑음",
        setOf(1) to "대체로 맑음",
        setOf(2) to "구름 조금",
        setOf(3) to "흐림",
        setOf(45, 48) to "안개",
        setOf(51, 53, 55, 56, 57) to "이슬비",
        setOf(61, 63, 65, 66, 67) to "비",
        setOf(71, 73, 75, 77) to "눈",
        setOf(80, 81, 82) to "소나기",
        setOf(85, 86) to "눈 소나기",
        setOf(95, 96, 99) to "뇌우",
    )

    internal val json = Json { ignoreUnknownKeys = true }

    fun rounded(value: Double): Double = roundHalfAway(value * 100) / 100

    internal fun wholeDegrees(value: Double): Int = roundHalfAway(value).toInt()

    private fun roundHalfAway(value: Double): Double = sign(value) * floor(abs(value) + 0.5)

    private fun coordinate(value: Double): String = String.format(Locale.ROOT, "%.2f", value)

    fun requestUrl(location: WeatherCoordinate, baseUrl: HttpUrl = DEFAULT_BASE_URL): HttpUrl =
        baseUrl.newBuilder()
            .addPathSegments("v1/forecast")
            .addQueryParameter("latitude", coordinate(location.latitude))
            .addQueryParameter("longitude", coordinate(location.longitude))
            .addQueryParameter("current", "temperature_2m,weather_code")
            .addQueryParameter("daily", "temperature_2m_max,temperature_2m_min,precipitation_probability_max")
            .addQueryParameter("forecast_days", "1")
            .addQueryParameter("timezone", "auto")
            .build()

    fun forecastUrl(location: WeatherCoordinate, baseUrl: HttpUrl = DEFAULT_BASE_URL): HttpUrl =
        baseUrl.newBuilder()
            .addPathSegments("v1/forecast")
            .addQueryParameter("latitude", coordinate(location.latitude))
            .addQueryParameter("longitude", coordinate(location.longitude))
            .addQueryParameter("current", "temperature_2m,weather_code,apparent_temperature")
            .addQueryParameter("hourly", "temperature_2m,weather_code,precipitation_probability")
            .addQueryParameter("forecast_hours", FORECAST_HOURS.toString())
            .addQueryParameter("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max")
            .addQueryParameter("forecast_days", FORECAST_DAYS.toString())
            .addQueryParameter("timezone", "auto")
            .build()

    fun summary(body: String): WeatherSummary {
        val response = decode(SummaryResponse.serializer(), body)
        val high = response.daily.temperatureMax.firstOrNull() ?: throw WeatherException.MissingDaily()
        val low = response.daily.temperatureMin.firstOrNull() ?: throw WeatherException.MissingDaily()
        val chance = response.daily.precipitationProbabilityMax?.firstOrNull()
        return WeatherSummary(
            temperature = wholeDegrees(response.current.temperature),
            condition = condition(response.current.weatherCode),
            high = wholeDegrees(high),
            low = wholeDegrees(low),
            precipitationChance = chance?.let { wholeDegrees(it) },
        )
    }

    fun forecast(body: String): WeatherForecast {
        val summary = summary(body)
        val response = decode(DetailedResponse.serializer(), body)
        val hours = response.hourly.rows().take(FORECAST_HOURS)
        val days = response.daily.rows().take(FORECAST_DAYS)
        if (hours.isEmpty() || days.isEmpty()) throw WeatherException.MissingDaily()
        return WeatherForecast(
            summary = summary,
            apparentTemperature = wholeDegrees(response.current.apparentTemperature),
            hours = hours,
            days = days,
        )
    }

    fun condition(code: Int): String = conditions.firstOrNull { code in it.first }?.second ?: "날씨 정보 없음"

    fun temperatureLabel(summary: WeatherSummary): String = "${summary.temperature}°"

    fun detailLabel(summary: WeatherSummary): String {
        val parts = mutableListOf("최고 ${summary.high}", "최저 ${summary.low}")
        summary.precipitationChance?.let { parts.add("강수 $it%") }
        return parts.joinToString(" · ")
    }

    fun apparentLabel(forecast: WeatherForecast): String = "체감 ${forecast.apparentTemperature}°"

    fun chanceLabel(chance: Int?): String = chance?.let { "$it%" } ?: "—"

    fun updatedLabel(date: Instant, zone: ZoneId = ZoneId.systemDefault()): String =
        "${date.atZone(zone).format(updatedFormatter)} 갱신"

    internal fun hourLabel(time: String): String {
        val parts = time.split("T")
        if (parts.size != 2) return time
        val hour = parts[1].take(2).toIntOrNull() ?: return time
        return "${hour}시"
    }

    internal fun dayLabel(day: String, isToday: Boolean): String {
        if (isToday) return "오늘"
        val date = try {
            LocalDate.parse(day)
        } catch (e: DateTimeParseException) {
            return day
        }
        val weekday = weekdays[date.dayOfWeek.value % 7]
        return "${date.monthValue}/${date.dayOfMonth} ($weekday)"
    }

    private fun <T> decode(serializer: KSerializer<T>, body: String): T =
        try {
            json.decodeFromString(serializer, body)
        } catch (e: SerializationException) {
            throw WeatherException.Decoding(e)
        } catch (e: IllegalArgumentException) {
            throw WeatherException.Decoding(e)
        }
}

@Serializable
private data class SummaryResponse(val current: Current, val daily: Daily) {
    @Serializable
    data class Current(
        @SerialName("temperature_2m") val temperature: Double,
        @SerialName("weather_code") val weatherCode: Int,
    )

    @Serializable
    data class Daily(
        @SerialName("temperature_2m_max") val temperatureMax: List<Double?>,
        @SerialName("temperature_2m_min") val temperatureMin: List<Double?>,
        @SerialName("precipitation_probability_max") val precipitationProbabilityMax: List<Double?>? = null,
    )
}

@Serializable
private data class DetailedResponse(val current: Current, val hourly: Hourly, val daily: Daily) {
    @Serializable
    data class Current(@SerialName("apparent_temperature") val apparentTemperature: Double)

    @Serializable
    data class Hourly(
        val time: List<String>,
        @SerialName("temperature_2m") val temperature: List<Double?>,
        @SerialName("weather_code") val weatherCode: List<Int?>,
        @SerialName("precipitation_probability") val precipitationProbability: List<Double?>? = null,
    ) {
        fun rows(): List<WeatherHour> = time.indices.mapNotNull { index ->
            val degrees = temperature.getOrNull(index) ?: return@mapNotNull null
            val code = weatherCode.getOrNull(index) ?: return@mapNotNull null
            val chance = precipitationProbability?.getOrNull(index)
            WeatherHour(
                time = Weather.hourLabel(time[index]),
                temperature = Weather.wholeDegrees(degrees),
                condition = Weather.condition(code),
                precipitationChance = chance?.let { Weather.wholeDegrees(it) },
            )
        }
    }

    @Serializable
    data class Daily(
        val time: List<String>,
        @SerialName("weather_code") val weatherCode: List<Int?>,
        @SerialName("temperature_2m_max") val temperatureMax: List<Double?>,
        @SerialName("temperature_2m_min") val temperatureMin: List<Double?>,
        @SerialName("precipitation_probability_max") val precipitationProbabilityMax: List<Double?>? = null,
    ) {
        fun rows(): List<WeatherDay> = time.indices.mapNotNull { index ->
            val code = weatherCode.getOrNull(index) ?: return@mapNotNull null
            val high = temperatureMax.getOrNull(index) ?: return@mapNotNull null
            val low = temperatureMin.getOrNull(index) ?: return@mapNotNull null
            val chance = precipitationProbabilityMax?.getOrNull(index)
            WeatherDay(
                label = Weather.dayLabel(time[index], isToday = index == 0),
                high = Weather.wholeDegrees(high),
                low = Weather.wholeDegrees(low),
                condition = Weather.condition(code),
                precipitationChance = chance?.let { Weather.wholeDegrees(it) },
            )
        }
    }
}

interface WeatherStorage {
    fun getString(key: String): String?
    fun putString(key: String, value: String)
    fun remove(key: String)
}

class WeatherLocationStore(private val storage: WeatherStorage) {
    fun load(): WeatherCoordinate? {
        val raw = storage.getString(KEY) ?: return null
        return try {
            Weather.json.decodeFromString(WeatherCoordinate.serializer(), raw)
        } catch (e: SerializationException) {
            null
        } catch (e: IllegalArgumentException) {
            null
        }
    }

    fun save(location: WeatherCoordinate) {
        storage.putString(KEY, Weather.json.encodeToString(WeatherCoordinate.serializer(), location))
    }

    fun clear() {
        storage.remove(KEY)
    }

    private companion object {
        const val KEY = "weather.location"
    }
}

class WeatherForecastCache(private val storage: WeatherStorage) {
    fun save(payload: String, location: WeatherCoordinate, at: Instant) {
        val entry = Entry(location.latitude, location.longitude, at.toEpochMilli(), payload)
        storage.putString(KEY, Weather.json.encodeToString(Entry.serializer(), entry))
    }

    fun load(location: WeatherCoordinate, now: Instant): WeatherForecastSnapshot? {
        val raw = storage.getString(KEY) ?: return null
        val entry = try {
            Weather.json.decodeFromString(Entry.serializer(), raw)
        } catch (e: SerializationException) {
            return null
        } catch (e: IllegalArgumentException) {
            return null
        }
        if (entry.latitude != location.latitude || entry.longitude != location.longitude) return null
        val fetchedAt = Instant.ofEpochMilli(entry.fetchedAtMillis)
        val age = Duration.between(fetchedAt, now)
        if (age.isNegative || age >= LIFETIME) return null
        val forecast = try {
            Weather.forecast(entry.payload)
        } catch (e: WeatherException) {
            return null
        }
        return WeatherForecastSnapshot(forecast, fetchedAt)
    }

    fun clear() {
        storage.remove(KEY)
    }

    @Serializable
    private data class Entry(
        val latitude: Double,
        val longitude: Double,
        val fetchedAtMillis: Long,
        val payload: String,
    )

    companion object {
        val LIFETIME: Duration = Duration.ofMinutes(30)
        private const val KEY = "weather.forecast"
    }
}

class WeatherClient(
    private val http: OkHttpClient,
    private val cache: WeatherForecastCache,
    private val baseUrl: HttpUrl = Weather.DEFAULT_BASE_URL,
    private val clock: () -> Instant = Instant::now,
) {
    suspend fun fetchForecast(location: WeatherCoordinate): WeatherForecastSnapshot {
        val request = Request.Builder().url(Weather.forecastUrl(location, baseUrl)).get().build()
        val (status, body) = withContext(Dispatchers.IO) {
            try {
                http.newCall(request).execute().use { it.code to (it.body?.string() ?: "") }
            } catch (e: IOException) {
                throw WeatherException.Transport(e)
            }
        }
        if (status != HTTP_OK) throw WeatherException.BadStatus(status)
        val forecast = Weather.forecast(body)
        val now = clock()
        cache.save(body, location, now)
        return WeatherForecastSnapshot(forecast, now)
    }

    suspend fun fetchForecastToCompletion(location: WeatherCoordinate): WeatherForecastSnapshot =
        withContext(NonCancellable) { fetchForecast(location) }

    private companion object {
        const val HTTP_OK = 200
    }
}
