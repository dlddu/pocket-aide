import Foundation

public struct WeatherHour: Equatable, Sendable {
    public let time: String
    public let temperature: Int
    public let condition: String
    public let precipitationChance: Int?

    public init(time: String, temperature: Int, condition: String, precipitationChance: Int?) {
        self.time = time
        self.temperature = temperature
        self.condition = condition
        self.precipitationChance = precipitationChance
    }
}

public struct WeatherDay: Equatable, Sendable {
    public let label: String
    public let high: Int
    public let low: Int
    public let condition: String
    public let precipitationChance: Int?

    public init(label: String, high: Int, low: Int, condition: String, precipitationChance: Int?) {
        self.label = label
        self.high = high
        self.low = low
        self.condition = condition
        self.precipitationChance = precipitationChance
    }
}

public struct WeatherForecast: Equatable, Sendable {
    public let summary: WeatherSummary
    public let apparentTemperature: Int
    public let hours: [WeatherHour]
    public let days: [WeatherDay]

    public init(summary: WeatherSummary, apparentTemperature: Int, hours: [WeatherHour], days: [WeatherDay]) {
        self.summary = summary
        self.apparentTemperature = apparentTemperature
        self.hours = hours
        self.days = days
    }
}

public struct WeatherForecastSnapshot: Equatable, Sendable {
    public let forecast: WeatherForecast
    public let fetchedAt: Date

    public init(forecast: WeatherForecast, fetchedAt: Date) {
        self.forecast = forecast
        self.fetchedAt = fetchedAt
    }
}

public extension Weather {
    static let forecastHours = 24
    static let forecastDays = 7

    static func forecastURL(for location: WeatherCoordinate) -> URL {
        var components = URLComponents()
        components.scheme = "https"
        components.host = "api.open-meteo.com"
        components.path = "/v1/forecast"
        components.queryItems = [
            URLQueryItem(name: "latitude", value: String(format: "%.2f", location.latitude)),
            URLQueryItem(name: "longitude", value: String(format: "%.2f", location.longitude)),
            URLQueryItem(name: "current", value: "temperature_2m,weather_code,apparent_temperature"),
            URLQueryItem(name: "hourly", value: "temperature_2m,weather_code,precipitation_probability"),
            URLQueryItem(name: "forecast_hours", value: String(forecastHours)),
            URLQueryItem(
                name: "daily",
                value: "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max"
            ),
            URLQueryItem(name: "forecast_days", value: String(forecastDays)),
            URLQueryItem(name: "timezone", value: "auto"),
        ]
        guard let url = components.url else {
            preconditionFailure("weather forecast URL")
        }
        return url
    }

    static func forecast(from data: Data) throws -> WeatherForecast {
        let summary = try Weather.summary(from: data)
        let response = try JSONDecoder().decode(DetailedForecastResponse.self, from: data)
        let hours = response.hourly.rows().prefix(forecastHours)
        let days = response.daily.rows().prefix(forecastDays)
        guard !hours.isEmpty, !days.isEmpty else {
            throw WeatherError.missingDaily
        }
        return WeatherForecast(
            summary: summary,
            apparentTemperature: wholeDegrees(response.current.apparentTemperature),
            hours: Array(hours),
            days: Array(days)
        )
    }

    static func apparentLabel(_ forecast: WeatherForecast) -> String {
        "체감 \(forecast.apparentTemperature)°"
    }

    static func chanceLabel(_ chance: Int?) -> String {
        chance.map { "\($0)%" } ?? "—"
    }

    static func updatedLabel(_ date: Date, timeZone: TimeZone = .current) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.timeZone = timeZone
        formatter.dateFormat = "HH:mm"
        return "\(formatter.string(from: date)) 갱신"
    }

    internal static func wholeDegrees(_ value: Double) -> Int {
        Int(value.rounded())
    }

    internal static func hourLabel(_ time: String) -> String {
        let parts = time.split(separator: "T")
        guard parts.count == 2, let hour = Int(parts[1].prefix(2)) else { return time }
        return "\(hour)시"
    }

    internal static func dayLabel(_ day: String, isToday: Bool) -> String {
        if isToday { return "오늘" }
        let parts = day.split(separator: "-").compactMap { Int($0) }
        guard parts.count == 3 else { return day }
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "UTC") ?? .current
        let components = DateComponents(year: parts[0], month: parts[1], day: parts[2])
        guard let date = calendar.date(from: components) else { return day }
        let weekdays = ["일", "월", "화", "수", "목", "금", "토"]
        let weekday = weekdays[calendar.component(.weekday, from: date) - 1]
        return "\(parts[1])/\(parts[2]) (\(weekday))"
    }
}

public struct WeatherForecastCache {
    public static let lifetime: TimeInterval = 30 * 60
    private static let key = "weather.forecast"
    private let defaults: UserDefaults

    public init(defaults: UserDefaults? = UserDefaults(suiteName: Weather.appGroup)) {
        self.defaults = defaults ?? .standard
    }

    public func save(_ payload: Data, for location: WeatherCoordinate, at date: Date) {
        let entry = Entry(
            latitude: location.latitude,
            longitude: location.longitude,
            fetchedAt: date,
            payload: payload
        )
        guard let data = try? JSONEncoder().encode(entry) else { return }
        defaults.set(data, forKey: Self.key)
    }

    public func load(for location: WeatherCoordinate, now: Date) -> WeatherForecastSnapshot? {
        guard let data = defaults.data(forKey: Self.key),
              let entry = try? JSONDecoder().decode(Entry.self, from: data),
              entry.latitude == location.latitude,
              entry.longitude == location.longitude else { return nil }
        let age = now.timeIntervalSince(entry.fetchedAt)
        guard age >= 0, age < Self.lifetime,
              let forecast = try? Weather.forecast(from: entry.payload) else { return nil }
        return WeatherForecastSnapshot(forecast: forecast, fetchedAt: entry.fetchedAt)
    }

    public func clear() {
        defaults.removeObject(forKey: Self.key)
    }

    private struct Entry: Codable {
        let latitude: Double
        let longitude: Double
        let fetchedAt: Date
        let payload: Data
    }
}

public extension WeatherClient {
    static func fetchForecast(
        _ location: WeatherCoordinate,
        session: URLSession = .shared
    ) async throws -> WeatherForecastSnapshot {
        let (data, response) = try await session.data(from: Weather.forecastURL(for: location))
        let status = (response as? HTTPURLResponse)?.statusCode ?? 0
        guard status == 200 else { throw WeatherError.badStatus(status) }
        let forecast = try Weather.forecast(from: data)
        let now = Date()
        WeatherForecastCache().save(data, for: location, at: now)
        return WeatherForecastSnapshot(forecast: forecast, fetchedAt: now)
    }

    static func fetchForecastToCompletion(
        _ location: WeatherCoordinate,
        session: URLSession = .shared
    ) async throws -> WeatherForecastSnapshot {
        try await Task { try await fetchForecast(location, session: session) }.value
    }
}

private struct DetailedForecastResponse: Decodable {
    struct Current: Decodable {
        let apparentTemperature: Double

        enum CodingKeys: String, CodingKey {
            case apparentTemperature = "apparent_temperature"
        }
    }

    struct Hourly: Decodable {
        let time: [String]
        let temperature: [Double?]
        let weatherCode: [Int?]
        let precipitationProbability: [Double?]?

        enum CodingKeys: String, CodingKey {
            case time
            case temperature = "temperature_2m"
            case weatherCode = "weather_code"
            case precipitationProbability = "precipitation_probability"
        }

        func rows() -> [WeatherHour] {
            time.indices.compactMap { index -> WeatherHour? in
                guard index < temperature.count, index < weatherCode.count,
                      let degrees = temperature[index], let code = weatherCode[index] else { return nil }
                let chance = precipitationProbability.flatMap { index < $0.count ? $0[index] : nil }
                return WeatherHour(
                    time: Weather.hourLabel(time[index]),
                    temperature: Weather.wholeDegrees(degrees),
                    condition: Weather.condition(code: code),
                    precipitationChance: chance.map(Weather.wholeDegrees)
                )
            }
        }
    }

    struct Daily: Decodable {
        let time: [String]
        let weatherCode: [Int?]
        let temperatureMax: [Double?]
        let temperatureMin: [Double?]
        let precipitationProbabilityMax: [Double?]?

        enum CodingKeys: String, CodingKey {
            case time
            case weatherCode = "weather_code"
            case temperatureMax = "temperature_2m_max"
            case temperatureMin = "temperature_2m_min"
            case precipitationProbabilityMax = "precipitation_probability_max"
        }

        func rows() -> [WeatherDay] {
            time.indices.compactMap { index -> WeatherDay? in
                guard index < weatherCode.count, index < temperatureMax.count, index < temperatureMin.count,
                      let code = weatherCode[index], let high = temperatureMax[index],
                      let low = temperatureMin[index] else { return nil }
                let chance = precipitationProbabilityMax.flatMap { index < $0.count ? $0[index] : nil }
                return WeatherDay(
                    label: Weather.dayLabel(time[index], isToday: index == 0),
                    high: Weather.wholeDegrees(high),
                    low: Weather.wholeDegrees(low),
                    condition: Weather.condition(code: code),
                    precipitationChance: chance.map(Weather.wholeDegrees)
                )
            }
        }
    }

    let current: Current
    let hourly: Hourly
    let daily: Daily
}
