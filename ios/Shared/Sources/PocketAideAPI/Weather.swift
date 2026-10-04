import Foundation

public struct WeatherCoordinate: Codable, Equatable, Sendable {
    public let latitude: Double
    public let longitude: Double
    public let placeName: String?

    public init(latitude: Double, longitude: Double, placeName: String?) {
        self.latitude = Weather.rounded(latitude)
        self.longitude = Weather.rounded(longitude)
        self.placeName = placeName
    }
}

public struct WeatherSummary: Equatable, Sendable {
    public let temperature: Int
    public let condition: String
    public let high: Int
    public let low: Int
    public let precipitationChance: Int?

    public init(temperature: Int, condition: String, high: Int, low: Int, precipitationChance: Int?) {
        self.temperature = temperature
        self.condition = condition
        self.high = high
        self.low = low
        self.precipitationChance = precipitationChance
    }
}

public enum WeatherError: Error, Equatable {
    case badStatus(Int)
    case missingDaily
}

public enum Weather {
    public static let appGroup = "group.com.dlddu.PocketAide"

    public static func rounded(_ value: Double) -> Double {
        (value * 100).rounded() / 100
    }

    public static func requestURL(for location: WeatherCoordinate) -> URL {
        var components = URLComponents()
        components.scheme = "https"
        components.host = "api.open-meteo.com"
        components.path = "/v1/forecast"
        components.queryItems = [
            URLQueryItem(name: "latitude", value: String(format: "%.2f", location.latitude)),
            URLQueryItem(name: "longitude", value: String(format: "%.2f", location.longitude)),
            URLQueryItem(name: "current", value: "temperature_2m,weather_code"),
            URLQueryItem(name: "daily", value: "temperature_2m_max,temperature_2m_min,precipitation_probability_max"),
            URLQueryItem(name: "forecast_days", value: "1"),
            URLQueryItem(name: "timezone", value: "auto"),
        ]
        guard let url = components.url else {
            preconditionFailure("weather request URL")
        }
        return url
    }

    public static func summary(from data: Data) throws -> WeatherSummary {
        let response = try JSONDecoder().decode(ForecastResponse.self, from: data)
        guard let high = response.daily.temperatureMax.first ?? nil,
              let low = response.daily.temperatureMin.first ?? nil else {
            throw WeatherError.missingDaily
        }
        let chance = response.daily.precipitationProbabilityMax?.first ?? nil
        return WeatherSummary(
            temperature: whole(response.current.temperature),
            condition: condition(code: response.current.weatherCode),
            high: whole(high),
            low: whole(low),
            precipitationChance: chance.map(whole)
        )
    }

    private static let conditions: [(codes: Set<Int>, label: String)] = [
        ([0], "맑음"),
        ([1], "대체로 맑음"),
        ([2], "구름 조금"),
        ([3], "흐림"),
        ([45, 48], "안개"),
        ([51, 53, 55, 56, 57], "이슬비"),
        ([61, 63, 65, 66, 67], "비"),
        ([71, 73, 75, 77], "눈"),
        ([80, 81, 82], "소나기"),
        ([85, 86], "눈 소나기"),
        ([95, 96, 99], "뇌우"),
    ]

    public static func condition(code: Int) -> String {
        conditions.first { $0.codes.contains(code) }?.label ?? "날씨 정보 없음"
    }

    public static func temperatureLabel(_ summary: WeatherSummary) -> String {
        "\(summary.temperature)°"
    }

    public static func detailLabel(_ summary: WeatherSummary) -> String {
        var parts = ["최고 \(summary.high)", "최저 \(summary.low)"]
        if let chance = summary.precipitationChance {
            parts.append("강수 \(chance)%")
        }
        return parts.joined(separator: " · ")
    }

    private static func whole(_ value: Double) -> Int {
        Int(value.rounded())
    }
}

public struct WeatherLocationStore {
    private static let key = "weather.location"
    private let defaults: UserDefaults

    public init(defaults: UserDefaults? = UserDefaults(suiteName: Weather.appGroup)) {
        self.defaults = defaults ?? .standard
    }

    public func load() -> WeatherCoordinate? {
        guard let data = defaults.data(forKey: Self.key) else { return nil }
        return try? JSONDecoder().decode(WeatherCoordinate.self, from: data)
    }

    public func save(_ location: WeatherCoordinate) {
        guard let data = try? JSONEncoder().encode(location) else { return }
        defaults.set(data, forKey: Self.key)
    }

    public func clear() {
        defaults.removeObject(forKey: Self.key)
    }
}

public enum WeatherClient {
    public static func fetch(_ location: WeatherCoordinate, session: URLSession = .shared) async throws -> WeatherSummary {
        let (data, response) = try await session.data(from: Weather.requestURL(for: location))
        let status = (response as? HTTPURLResponse)?.statusCode ?? 0
        guard status == 200 else { throw WeatherError.badStatus(status) }
        return try Weather.summary(from: data)
    }
}

private struct ForecastResponse: Decodable {
    struct Current: Decodable {
        let temperature: Double
        let weatherCode: Int

        enum CodingKeys: String, CodingKey {
            case temperature = "temperature_2m"
            case weatherCode = "weather_code"
        }
    }

    struct Daily: Decodable {
        let temperatureMax: [Double?]
        let temperatureMin: [Double?]
        let precipitationProbabilityMax: [Double?]?

        enum CodingKeys: String, CodingKey {
            case temperatureMax = "temperature_2m_max"
            case temperatureMin = "temperature_2m_min"
            case precipitationProbabilityMax = "precipitation_probability_max"
        }
    }

    let current: Current
    let daily: Daily
}
