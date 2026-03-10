// Protocol-based abstraction for the network layer.
// Concrete implementations (URLSession-backed, Mock) conform to this protocol
// so that callers in the Presentation/Domain layers are testable without a live server.

import Foundation

// MARK: - APIClientError

/// Errors that an APIClient implementation may throw.
public enum APIClientError: Error, Equatable {
    /// The HTTP response contained a non-2xx status code.
    case httpError(statusCode: Int, message: String)

    /// The response body could not be decoded into the expected type.
    case decodingError(String)

    /// The request could not be constructed or the transport failed.
    case networkError(String)

    /// A required auth token is missing.
    case unauthorized
}

// MARK: - APIClientProtocol

/// Defines the contract for making authenticated HTTP requests to the PocketAide backend.
///
/// All methods are `async throws` so that callers can use structured concurrency.
/// Conforming types are expected to attach a `Bearer` Authorization header when
/// a non-nil `token` is provided.
public protocol APIClientProtocol: AnyObject, Sendable {

    /// Performs a GET /health request and returns the parsed response.
    func fetchHealth() async throws -> HealthResponse

    /// Performs an arbitrary GET request, decoding the response body into `ResponseType`.
    ///
    /// - Parameters:
    ///   - path: URL path component, e.g. `/api/messages`.
    ///   - token: Optional JWT Bearer token. Attached as `Authorization: Bearer <token>`.
    /// - Returns: Decoded value of `ResponseType`.
    func get<ResponseType: Decodable>(
        path: String,
        token: String?
    ) async throws -> ResponseType

    /// Performs an arbitrary POST request, encoding `body` as JSON and decoding the response.
    ///
    /// - Parameters:
    ///   - path: URL path component.
    ///   - body: Encodable request body. Will be JSON-encoded.
    ///   - token: Optional JWT Bearer token.
    /// - Returns: Decoded value of `ResponseType`.
    func post<RequestBody: Encodable, ResponseType: Decodable>(
        path: String,
        body: RequestBody,
        token: String?
    ) async throws -> ResponseType
}
