package com.dlddu.pocketaide.weather

class InMemoryWeatherStorage : WeatherStorage {
    val values = mutableMapOf<String, String>()

    override fun getString(key: String): String? = values[key]

    override fun putString(key: String, value: String) {
        values[key] = value
    }

    override fun remove(key: String) {
        values.remove(key)
    }
}

object WeatherFixtures {
    fun summaryResponse(temperature: String = "23.2", code: Int = 1, precipitation: String = "[82]"): String =
        """
        {"latitude":37.55,"longitude":127.0,"timezone":"Asia/Seoul",
         "current":{"time":"2026-10-04T14:45","interval":900,"temperature_2m":$temperature,"weather_code":$code},
         "daily":{"time":["2026-10-04"],"temperature_2m_max":[23.2],"temperature_2m_min":[13.7],
                  "precipitation_probability_max":$precipitation}}
        """.trimIndent()

    fun forecastResponse(hours: Int = 24, days: Int = 7, firstHourTemperature: String = "14.3"): String {
        val hourTimes = (0 until hours).map {
            val day = if (it + 21 < 24) "05" else "06"
            "\"2026-10-${day}T${"%02d".format((it + 21) % 24)}:00\""
        }
        val hourTemps = (0 until hours).map { if (it == 0) firstHourTemperature else "13.7" }
        val dayTimes = (0 until days).map { "\"2026-10-${"%02d".format(it + 5)}\"" }
        fun repeated(value: String, count: Int) = List(count) { value }.joinToString(",")
        return """
        {"latitude":37.55,"longitude":127.0,"utc_offset_seconds":32400,"timezone":"Asia/Seoul",
         "current":{"time":"2026-10-05T21:30","interval":900,"temperature_2m":14.0,"weather_code":0,
                    "apparent_temperature":12.3},
         "hourly":{"time":[${hourTimes.joinToString(",")}],
                   "temperature_2m":[${hourTemps.joinToString(",")}],
                   "weather_code":[${repeated("61", hours)}],
                   "precipitation_probability":[${repeated("40", hours)}]},
         "daily":{"time":[${dayTimes.joinToString(",")}],
                  "weather_code":[${repeated("2", days)}],
                  "temperature_2m_max":[${repeated("19.3", days)}],
                  "temperature_2m_min":[${repeated("13.1", days)}],
                  "precipitation_probability_max":[${repeated("96", days)}]}}
        """.trimIndent()
    }
}
