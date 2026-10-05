import XCTest
@testable import PocketAideAPI

final class WeatherForecastTests: XCTestCase {
    private let location = WeatherCoordinate(latitude: 37.5665, longitude: 126.9784, placeName: "서울")

    private func response(hours: Int = 24, days: Int = 7, firstHourTemperature: String = "14.3") -> Data {
        let hourTimes = (0..<hours).map { "\"2026-10-\($0 + 21 < 24 ? "05" : "06")T\(String(format: "%02d", ($0 + 21) % 24)):00\"" }
        let hourTemps = (0..<hours).map { $0 == 0 ? firstHourTemperature : "13.7" }
        let dayTimes = (0..<days).map { "\"2026-10-\(String(format: "%02d", $0 + 5))\"" }
        return Data("""
        {"latitude":37.55,"longitude":127.0,"utc_offset_seconds":32400,"timezone":"Asia/Seoul",
         "current":{"time":"2026-10-05T21:30","interval":900,"temperature_2m":14.0,"weather_code":0,
                    "apparent_temperature":12.3},
         "hourly":{"time":[\(hourTimes.joined(separator: ","))],
                   "temperature_2m":[\(hourTemps.joined(separator: ","))],
                   "weather_code":[\(Array(repeating: "61", count: hours).joined(separator: ","))],
                   "precipitation_probability":[\(Array(repeating: "40", count: hours).joined(separator: ","))]},
         "daily":{"time":[\(dayTimes.joined(separator: ","))],
                  "weather_code":[\(Array(repeating: "2", count: days).joined(separator: ","))],
                  "temperature_2m_max":[\(Array(repeating: "19.3", count: days).joined(separator: ","))],
                  "temperature_2m_min":[\(Array(repeating: "13.1", count: days).joined(separator: ","))],
                  "precipitation_probability_max":[\(Array(repeating: "96", count: days).joined(separator: ","))]}}
        """.utf8)
    }

    func testForecastSummaryMatchesWidgetSummary() throws {
        let data = response()
        let forecast = try Weather.forecast(from: data)
        XCTAssertEqual(forecast.summary, try Weather.summary(from: data))
        XCTAssertEqual(forecast.summary, WeatherSummary(
            temperature: 14,
            condition: "맑음",
            high: 19,
            low: 13,
            precipitationChance: 96
        ))
        XCTAssertEqual(Weather.apparentLabel(forecast), "체감 12°")
    }

    func testHourlyHasTwentyFourRowsFromNow() throws {
        let forecast = try Weather.forecast(from: response())
        XCTAssertEqual(forecast.hours.count, 24)
        XCTAssertEqual(forecast.hours.first, WeatherHour(
            time: "21시",
            temperature: 14,
            condition: "비",
            precipitationChance: 40
        ))
        XCTAssertEqual(forecast.hours[3].time, "0시")
        XCTAssertEqual(forecast.hours.last?.time, "20시")
    }

    func testDailyHasSevenRowsStartingToday() throws {
        let forecast = try Weather.forecast(from: response())
        XCTAssertEqual(forecast.days.count, 7)
        XCTAssertEqual(forecast.days.first, WeatherDay(
            label: "오늘",
            high: 19,
            low: 13,
            condition: "구름 조금",
            precipitationChance: 96
        ))
        XCTAssertEqual(forecast.days[1].label, "10/6 (화)")
        XCTAssertEqual(forecast.days.last?.label, "10/11 (일)")
    }

    func testExtraRowsAreTrimmedAndEmptyForecastFails() throws {
        let forecast = try Weather.forecast(from: response(hours: 30, days: 9))
        XCTAssertEqual(forecast.hours.count, 24)
        XCTAssertEqual(forecast.days.count, 7)
        XCTAssertThrowsError(try Weather.forecast(from: response(hours: 0)))
    }

    func testForecastRequestAsksForHourlyAndWeekly() throws {
        let url = Weather.forecastURL(for: location)
        let components = try XCTUnwrap(URLComponents(url: url, resolvingAgainstBaseURL: false))
        XCTAssertEqual(components.host, "api.open-meteo.com")
        let query = Dictionary(uniqueKeysWithValues: (components.queryItems ?? []).map { ($0.name, $0.value ?? "") })
        XCTAssertEqual(query["latitude"], "37.57")
        XCTAssertEqual(query["longitude"], "126.98")
        XCTAssertEqual(query["current"], "temperature_2m,weather_code,apparent_temperature")
        XCTAssertEqual(query["hourly"], "temperature_2m,weather_code,precipitation_probability")
        XCTAssertEqual(query["forecast_hours"], "24")
        XCTAssertEqual(query["forecast_days"], "7")
    }

    func testCacheServesSameLocationWithinThirtyMinutes() throws {
        let defaults = try XCTUnwrap(UserDefaults(suiteName: "WeatherForecastTests"))
        defaults.removePersistentDomain(forName: "WeatherForecastTests")
        let cache = WeatherForecastCache(defaults: defaults)
        let savedAt = Date(timeIntervalSince1970: 1_791_000_000)
        XCTAssertNil(cache.load(for: location, now: savedAt))

        cache.save(response(), for: location, at: savedAt)
        let hit = try XCTUnwrap(cache.load(for: location, now: savedAt.addingTimeInterval(29 * 60)))
        XCTAssertEqual(hit.fetchedAt, savedAt)
        XCTAssertEqual(hit.forecast, try Weather.forecast(from: response()))

        XCTAssertNil(cache.load(for: location, now: savedAt.addingTimeInterval(30 * 60)))
        let elsewhere = WeatherCoordinate(latitude: 35.1796, longitude: 129.0756, placeName: "부산")
        XCTAssertNil(cache.load(for: elsewhere, now: savedAt))

        cache.clear()
        XCTAssertNil(cache.load(for: location, now: savedAt))
    }

    func testUpdatedLabelShowsFetchTime() throws {
        let seoul = try XCTUnwrap(TimeZone(identifier: "Asia/Seoul"))
        let date = Date(timeIntervalSince1970: 1_791_203_400)
        XCTAssertEqual(Weather.updatedLabel(date, timeZone: seoul), "21:30 갱신")
        XCTAssertEqual(Weather.chanceLabel(40), "40%")
        XCTAssertEqual(Weather.chanceLabel(nil), "—")
    }
}
