import Foundation
import PocketAideStorage

public struct AuthConfig: Codable, Equatable, Sendable {
    public let issuer: String
    public let clientId: String
    public let redirectUri: String
    public let audience: String

    enum CodingKeys: String, CodingKey {
        case issuer
        case clientId = "client_id"
        case redirectUri = "redirect_uri"
        case audience
    }
}

public struct MeResponse: Codable, Equatable, Sendable {
    public let id: Int
    public let sub: String
}

public enum APIError: Error, CustomStringConvertible {
    case badStatus(Int, String)
    case decoding(Error)
    case transport(Error)
    case missingBaseURL

    public var description: String {
        switch self {
        case .badStatus(let s, let body): return "HTTP \(s): \(body)"
        case .decoding(let e): return "decode: \(e)"
        case .transport(let e): return "transport: \(e)"
        case .missingBaseURL: return "BackendBaseURL is not configured"
        }
    }
}

public final class APIClient: @unchecked Sendable {
    public let baseURL: URL
    private let session: URLSession
    private let tokenStore: TokenStoring

    public init(baseURL: URL, tokenStore: TokenStoring, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
        self.tokenStore = tokenStore
    }

    public static func fromBundle(_ bundle: Bundle = .main, tokenStore: TokenStoring) throws -> APIClient {
        guard let raw = bundle.object(forInfoDictionaryKey: "BackendBaseURL") as? String,
              !raw.isEmpty,
              let url = URL(string: raw) else {
            throw APIError.missingBaseURL
        }
        return APIClient(baseURL: url, tokenStore: tokenStore)
    }

    public func health() async throws {
        _ = try await get("/healthz", authenticated: false, decodeAs: HealthResponse.self)
    }

    public func authConfig() async throws -> AuthConfig {
        try await get("/api/auth/config", authenticated: false, decodeAs: AuthConfig.self)
    }

    public func me() async throws -> MeResponse {
        try await get("/api/me", authenticated: true, decodeAs: MeResponse.self)
    }

    private struct HealthResponse: Codable { let status: String }

    private func get<T: Decodable>(
        _ path: String,
        authenticated: Bool,
        decodeAs: T.Type
    ) async throws -> T {
        var req = URLRequest(url: baseURL.appendingPathComponent(path))
        req.httpMethod = "GET"
        if authenticated, let token = try tokenStore.load()?.accessToken {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: req)
        } catch {
            throw APIError.transport(error)
        }
        guard let http = response as? HTTPURLResponse else {
            throw APIError.transport(URLError(.badServerResponse))
        }
        guard (200..<300).contains(http.statusCode) else {
            throw APIError.badStatus(http.statusCode, String(data: data, encoding: .utf8) ?? "")
        }
        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            throw APIError.decoding(error)
        }
    }
}
