// APIClientTests.swift
// PocketAideTests
//
// Unit tests for the URLSession-based APIClient.
// These tests are in the TDD Red phase — they will fail until the production
// code under test is implemented.
//
// Targets production types:
//   - PocketAide.APIClient
//   - PocketAide.APIError
//   - PocketAide.HTTPMethod
//   - PocketAide.APIRequest

import XCTest
@testable import PocketAide

// MARK: - Helpers

/// A `URLProtocol` subclass that intercepts network requests and returns
/// pre-configured stub responses without hitting the network.
final class MockURLProtocol: URLProtocol {
    /// Per-test stub handler. Set this before each test.
    static var requestHandler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = MockURLProtocol.requestHandler else {
            client?.urlProtocolDidFinishLoading(self)
            return
        }
        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}

/// Returns a `URLSession` wired up to `MockURLProtocol` so no real network
/// traffic is generated during tests.
private func makeMockSession() -> URLSession {
    let config = URLSessionConfiguration.ephemeral
    config.protocolClasses = [MockURLProtocol.self]
    return URLSession(configuration: config)
}

/// Builds a minimal `HTTPURLResponse` with the given status code.
private func makeHTTPResponse(url: URL, statusCode: Int) -> HTTPURLResponse {
    HTTPURLResponse(
        url: url,
        statusCode: statusCode,
        httpVersion: nil,
        headerFields: ["Content-Type": "application/json"]
    )!
}

// MARK: - APIClientTests

final class APIClientTests: XCTestCase {

    // MARK: Properties

    private var sut: APIClient!
    private let baseURL = URL(string: "https://api.example.com")!

    // MARK: Lifecycle

    override func setUp() {
        super.setUp()
        MockURLProtocol.requestHandler = nil
        // APIClient is initialised with an injected URLSession so tests can
        // substitute the mock session without swizzling.
        sut = APIClient(baseURL: baseURL, session: makeMockSession())
    }

    override func tearDown() {
        MockURLProtocol.requestHandler = nil
        sut = nil
        super.tearDown()
    }

    // MARK: - Happy Path

    // MARK: GET /health

    /// The client must decode a valid `{"status":"ok"}` response.
    func test_request_decodesSuccessResponse() async throws {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let responseBody = """
        {"status":"ok"}
        """.data(using: .utf8)!

        MockURLProtocol.requestHandler = { _ in
            (makeHTTPResponse(url: expectedURL, statusCode: 200), responseBody)
        }

        // Act
        let result: HealthResponse = try await sut.request(
            path: endpoint,
            method: .get,
            token: nil
        )

        // Assert
        XCTAssertEqual(result.status, "ok")
    }

    /// The client must send a `GET` request to the correct URL.
    func test_request_sendsGetToCorrectURL() async throws {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        var capturedRequest: URLRequest?

        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let body = """{"status":"ok"}""".data(using: .utf8)!
            return (makeHTTPResponse(url: expectedURL, statusCode: 200), body)
        }

        // Act
        let _: HealthResponse = try await sut.request(
            path: endpoint,
            method: .get,
            token: nil
        )

        // Assert
        XCTAssertEqual(capturedRequest?.httpMethod, "GET")
        XCTAssertEqual(capturedRequest?.url?.path, expectedURL.path)
    }

    /// When a JWT token is provided the client must attach it as a Bearer
    /// `Authorization` header.
    func test_request_attachesBearerTokenWhenProvided() async throws {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let token = "test.jwt.token"
        var capturedRequest: URLRequest?

        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let body = """{"status":"ok"}""".data(using: .utf8)!
            return (makeHTTPResponse(url: expectedURL, statusCode: 200), body)
        }

        // Act
        let _: HealthResponse = try await sut.request(
            path: endpoint,
            method: .get,
            token: token
        )

        // Assert
        XCTAssertEqual(
            capturedRequest?.value(forHTTPHeaderField: "Authorization"),
            "Bearer \(token)"
        )
    }

    /// When no token is supplied the `Authorization` header must be absent.
    func test_request_omitsAuthorizationHeaderWhenNoToken() async throws {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        var capturedRequest: URLRequest?

        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let body = """{"status":"ok"}""".data(using: .utf8)!
            return (makeHTTPResponse(url: expectedURL, statusCode: 200), body)
        }

        // Act
        let _: HealthResponse = try await sut.request(
            path: endpoint,
            method: .get,
            token: nil
        )

        // Assert
        XCTAssertNil(capturedRequest?.value(forHTTPHeaderField: "Authorization"))
    }

    /// POST requests must include the serialised body.
    func test_request_sendsPostWithBody() async throws {
        // Arrange
        let endpoint = "/echo"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let payload = EchoPayload(message: "hello")
        var capturedRequest: URLRequest?

        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let body = """{"message":"hello"}""".data(using: .utf8)!
            return (makeHTTPResponse(url: expectedURL, statusCode: 200), body)
        }

        // Act
        let _: EchoPayload = try await sut.request(
            path: endpoint,
            method: .post,
            body: payload,
            token: nil
        )

        // Assert
        XCTAssertEqual(capturedRequest?.httpMethod, "POST")
        XCTAssertNotNil(capturedRequest?.httpBody)
    }

    // MARK: - Error Handling

    /// A 401 response must throw `APIError.unauthorized`.
    func test_request_throwsUnauthorizedOn401() async {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let body = """{"message":"Unauthorized","code":401}""".data(using: .utf8)!

        MockURLProtocol.requestHandler = { _ in
            (makeHTTPResponse(url: expectedURL, statusCode: 401), body)
        }

        // Act & Assert
        do {
            let _: HealthResponse = try await sut.request(
                path: endpoint,
                method: .get,
                token: nil
            )
            XCTFail("Expected APIError.unauthorized to be thrown")
        } catch let error as APIError {
            XCTAssertEqual(error, .unauthorized)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    /// A 404 response must throw `APIError.notFound`.
    func test_request_throwsNotFoundOn404() async {
        // Arrange
        let endpoint = "/nonexistent"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let body = """{"message":"Not Found","code":404}""".data(using: .utf8)!

        MockURLProtocol.requestHandler = { _ in
            (makeHTTPResponse(url: expectedURL, statusCode: 404), body)
        }

        // Act & Assert
        do {
            let _: HealthResponse = try await sut.request(
                path: endpoint,
                method: .get,
                token: nil
            )
            XCTFail("Expected APIError.notFound to be thrown")
        } catch let error as APIError {
            XCTAssertEqual(error, .notFound)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    /// A 500 response must throw `APIError.serverError` carrying the status
    /// code.
    func test_request_throwsServerErrorOn500() async {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let body = """{"message":"Internal Server Error","code":500}""".data(using: .utf8)!

        MockURLProtocol.requestHandler = { _ in
            (makeHTTPResponse(url: expectedURL, statusCode: 500), body)
        }

        // Act & Assert
        do {
            let _: HealthResponse = try await sut.request(
                path: endpoint,
                method: .get,
                token: nil
            )
            XCTFail("Expected APIError.serverError to be thrown")
        } catch let error as APIError {
            if case .serverError(let code) = error {
                XCTAssertEqual(code, 500)
            } else {
                XCTFail("Expected .serverError(500), got \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    /// A network failure (e.g. no connectivity) must surface as
    /// `APIError.networkError`.
    func test_request_throwsNetworkErrorOnConnectionFailure() async {
        // Arrange
        let endpoint = "/health"
        MockURLProtocol.requestHandler = { _ in
            throw URLError(.notConnectedToInternet)
        }

        // Act & Assert
        do {
            let _: HealthResponse = try await sut.request(
                path: endpoint,
                method: .get,
                token: nil
            )
            XCTFail("Expected APIError.networkError to be thrown")
        } catch let error as APIError {
            if case .networkError = error {
                // pass
            } else {
                XCTFail("Expected .networkError, got \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    /// Malformed JSON in the response body must throw `APIError.decodingError`.
    func test_request_throwsDecodingErrorOnMalformedJSON() async {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        let body = "not-json".data(using: .utf8)!

        MockURLProtocol.requestHandler = { _ in
            (makeHTTPResponse(url: expectedURL, statusCode: 200), body)
        }

        // Act & Assert
        do {
            let _: HealthResponse = try await sut.request(
                path: endpoint,
                method: .get,
                token: nil
            )
            XCTFail("Expected APIError.decodingError to be thrown")
        } catch let error as APIError {
            if case .decodingError = error {
                // pass
            } else {
                XCTFail("Expected .decodingError, got \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    // MARK: - Edge Cases

    /// An empty path component should still result in a valid request to the
    /// base URL.
    func test_request_handlesEmptyPath() async throws {
        // Arrange
        let endpoint = ""
        var capturedURL: URL?

        MockURLProtocol.requestHandler = { request in
            capturedURL = request.url
            let body = """{"status":"ok"}""".data(using: .utf8)!
            return (makeHTTPResponse(url: self.baseURL, statusCode: 200), body)
        }

        // Act
        let _: HealthResponse = try await sut.request(
            path: endpoint,
            method: .get,
            token: nil
        )

        // Assert
        XCTAssertNotNil(capturedURL)
    }

    /// The `Content-Type: application/json` header must always be set on
    /// outgoing requests.
    func test_request_alwaysSetsContentTypeHeader() async throws {
        // Arrange
        let endpoint = "/health"
        let expectedURL = baseURL.appendingPathComponent(endpoint)
        var capturedRequest: URLRequest?

        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let body = """{"status":"ok"}""".data(using: .utf8)!
            return (makeHTTPResponse(url: expectedURL, statusCode: 200), body)
        }

        // Act
        let _: HealthResponse = try await sut.request(
            path: endpoint,
            method: .get,
            token: nil
        )

        // Assert
        XCTAssertEqual(
            capturedRequest?.value(forHTTPHeaderField: "Content-Type"),
            "application/json"
        )
    }
}

// MARK: - Test Doubles (Decodable stubs used by tests only)

/// Matches the backend `GET /health` response shape.
private struct HealthResponse: Decodable {
    let status: String
}

/// Arbitrary request/response body used for POST tests.
private struct EchoPayload: Codable {
    let message: String
}
