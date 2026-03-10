// Concrete URLSession-based implementation of APIClientProtocol.
// On Linux this file compiles but URLSession behaviour depends on the platform's
// Foundation implementation (swift-corelibs-foundation).

import Foundation

// MARK: - APIClient

public final class APIClient: APIClientProtocol {

    // MARK: - Properties

    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    // MARK: - Init

    /// Creates an `APIClient` targeting the given base URL.
    ///
    /// - Parameters:
    ///   - baseURL: Root URL of the backend, e.g. `http://localhost:8080`.
    ///   - session: `URLSession` instance to use. Defaults to `.shared`.
    public init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
        self.decoder = JSONDecoder()
        self.encoder = JSONEncoder()
    }

    // MARK: - APIClientProtocol

    public func fetchHealth() async throws -> HealthResponse {
        return try await get(path: "/health", token: nil)
    }

    public func get<ResponseType: Decodable>(
        path: String,
        token: String?
    ) async throws -> ResponseType {
        let request = try buildRequest(method: "GET", path: path, body: Optional<EmptyBody>.none, token: token)
        return try await perform(request: request)
    }

    public func post<RequestBody: Encodable, ResponseType: Decodable>(
        path: String,
        body: RequestBody,
        token: String?
    ) async throws -> ResponseType {
        let request = try buildRequest(method: "POST", path: path, body: body, token: token)
        return try await perform(request: request)
    }

    // MARK: - Private helpers

    private struct EmptyBody: Encodable {}

    private func buildRequest<Body: Encodable>(
        method: String,
        path: String,
        body: Body?,
        token: String?
    ) throws -> URLRequest {
        guard let url = URL(string: path, relativeTo: baseURL) else {
            throw APIClientError.networkError("Invalid path: \(path)")
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        if let token = token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body = body, !(body is EmptyBody) {
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try encoder.encode(body)
        }

        return request
    }

    private func perform<ResponseType: Decodable>(request: URLRequest) async throws -> ResponseType {
        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw APIClientError.networkError(error.localizedDescription)
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIClientError.networkError("Non-HTTP response received")
        }

        switch httpResponse.statusCode {
        case 200..<300:
            do {
                return try decoder.decode(ResponseType.self, from: data)
            } catch {
                throw APIClientError.decodingError(error.localizedDescription)
            }

        case 401:
            // Attempt to decode structured error body; fall back to generic message.
            let message = (try? decoder.decode(APIErrorResponse.self, from: data))?.message ?? "Unauthorized"
            throw APIClientError.httpError(statusCode: 401, message: message)

        default:
            let message = (try? decoder.decode(APIErrorResponse.self, from: data))?.message
                ?? HTTPURLResponse.localizedString(forStatusCode: httpResponse.statusCode)
            throw APIClientError.httpError(statusCode: httpResponse.statusCode, message: message)
        }
    }
}
