// In-memory mock implementation of APIClientProtocol for use in tests and
// XCUITest / simulator runs without a live backend.

import Foundation

// MARK: - MockAPIClient

/// A test double for `APIClientProtocol` that returns pre-configured responses
/// or throws pre-configured errors without touching the network.
///
/// Usage:
/// ```swift
/// let mock = MockAPIClient()
/// mock.healthResult = .success(HealthResponse(status: "ok"))
/// let result = try await mock.fetchHealth()
/// ```
public final class MockAPIClient: APIClientProtocol {

    // MARK: - Recorded invocations (inspect in tests)

    /// All (method, path) pairs that have been requested.
    public private(set) var requestLog: [(method: String, path: String)] = []

    // MARK: - Stubbed results

    /// Result returned by `fetchHealth()`.
    public var healthResult: Result<HealthResponse, APIClientError> = .success(HealthResponse(status: "ok"))

    /// Generic GET result storage, keyed by path.
    public var getResults: [String: Any] = [:]

    /// Generic POST result storage, keyed by path.
    public var postResults: [String: Any] = [:]

    // MARK: - Init

    public init() {}

    // MARK: - APIClientProtocol

    public func fetchHealth() async throws -> HealthResponse {
        requestLog.append((method: "GET", path: "/health"))
        switch healthResult {
        case .success(let response): return response
        case .failure(let error):   throw error
        }
    }

    public func get<ResponseType: Decodable>(
        path: String,
        token: String?
    ) async throws -> ResponseType {
        requestLog.append((method: "GET", path: path))

        if let result = getResults[path] {
            if let typed = result as? ResponseType {
                return typed
            }
            if let error = result as? APIClientError {
                throw error
            }
        }

        throw APIClientError.networkError("MockAPIClient: no stub registered for GET \(path)")
    }

    public func post<RequestBody: Encodable, ResponseType: Decodable>(
        path: String,
        body: RequestBody,
        token: String?
    ) async throws -> ResponseType {
        requestLog.append((method: "POST", path: path))

        if let result = postResults[path] {
            if let typed = result as? ResponseType {
                return typed
            }
            if let error = result as? APIClientError {
                throw error
            }
        }

        throw APIClientError.networkError("MockAPIClient: no stub registered for POST \(path)")
    }

    // MARK: - Helpers

    /// Removes all recorded requests (useful in tearDown / between subtests).
    public func reset() {
        requestLog.removeAll()
    }
}
