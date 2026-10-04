import XCTest
@testable import PocketAideAPI

final class WeatherTests: XCTestCase {
    private func response(temperature: String = "23.2", code: Int = 1, precipitation: String = "[82]") -> Data {
        Data("""
        {"latitude":37.55,"longitude":127.0,"timezone":"Asia/Seoul",
         "current":{"time":"2026-10-04T14:45","interval":900,"temperature_2m":\(temperature),"weather_code":\(code)},
         "daily":{"time":["2026-10-04"],"temperature_2m_max":[23.2],"temperature_2m_min":[13.7],
                  "precipitation_probability_max":\(precipitation)}}
        """.utf8)
    }

    func testDecodesCurrentTemperatureConditionAndDailyRange() throws {
        let summary = try Weather.summary(from: response())
        XCTAssertEqual(summary, WeatherSummary(
            temperature: 23,
            condition: "대체로 맑음",
            high: 23,
            low: 14,
            precipitationChance: 82
        ))
        XCTAssertEqual(Weather.temperatureLabel(summary), "23°")
        XCTAssertEqual(Weather.detailLabel(summary), "최고 23 · 최저 14 · 강수 82%")
    }

    func testMissingPrecipitationLeavesChanceOut() throws {
        let summary = try Weather.summary(from: response(precipitation: "[null]"))
        XCTAssertNil(summary.precipitationChance)
        XCTAssertEqual(Weather.detailLabel(summary), "최고 23 · 최저 14")
    }

    func testNegativeTemperatureRoundsToNearest() throws {
        let summary = try Weather.summary(from: response(temperature: "-3.6"))
        XCTAssertEqual(Weather.temperatureLabel(summary), "-4°")
    }

    func testMissingDailyValuesFail() {
        let data = Data("""
        {"current":{"temperature_2m":1.0,"weather_code":0},
         "daily":{"temperature_2m_max":[null],"temperature_2m_min":[null]}}
        """.utf8)
        XCTAssertThrowsError(try Weather.summary(from: data)) { error in
            XCTAssertEqual(error as? WeatherError, .missingDaily)
        }
    }

    func testConditionCodesMapToKoreanLabels() {
        XCTAssertEqual(Weather.condition(code: 0), "맑음")
        XCTAssertEqual(Weather.condition(code: 3), "흐림")
        XCTAssertEqual(Weather.condition(code: 48), "안개")
        XCTAssertEqual(Weather.condition(code: 63), "비")
        XCTAssertEqual(Weather.condition(code: 75), "눈")
        XCTAssertEqual(Weather.condition(code: 81), "소나기")
        XCTAssertEqual(Weather.condition(code: 95), "뇌우")
        XCTAssertEqual(Weather.condition(code: 42), "날씨 정보 없음")
    }

    func testRequestUsesRoundedCoordinates() throws {
        let location = WeatherCoordinate(latitude: 37.5665, longitude: 126.9784, placeName: "서울")
        XCTAssertEqual(location.latitude, 37.57)
        XCTAssertEqual(location.longitude, 126.98)
        let components = try XCTUnwrap(URLComponents(url: Weather.requestURL(for: location), resolvingAgainstBaseURL: false))
        XCTAssertEqual(components.host, "api.open-meteo.com")
        let query = Dictionary(uniqueKeysWithValues: (components.queryItems ?? []).map { ($0.name, $0.value ?? "") })
        XCTAssertEqual(query["latitude"], "37.57")
        XCTAssertEqual(query["longitude"], "126.98")
        XCTAssertEqual(query["current"], "temperature_2m,weather_code")
        XCTAssertEqual(query["daily"], "temperature_2m_max,temperature_2m_min,precipitation_probability_max")
    }

    func testLocationStoreRoundTripsAndClears() throws {
        let defaults = try XCTUnwrap(UserDefaults(suiteName: "WeatherTests"))
        defaults.removePersistentDomain(forName: "WeatherTests")
        let store = WeatherLocationStore(defaults: defaults)
        XCTAssertNil(store.load())
        let location = WeatherCoordinate(latitude: 35.1796, longitude: 129.0756, placeName: "부산")
        store.save(location)
        XCTAssertEqual(store.load(), location)
        store.clear()
        XCTAssertNil(store.load())
    }
}
