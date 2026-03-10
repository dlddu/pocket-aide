// Tests for the network layer.
//
// We test three things:
//   1. The `MockAPIClient` test double itself (verifying its stub/record behaviour).
//   2. The `APIClient` request-builder logic via a local `URLProtocol` stub.
//   3. Error-path mapping (non-2xx → `APIClientError.httpError`).
//
// No live server is required; all HTTP traffic is intercepted by
// `MockURLProtocol` which is registered only for test sessions.

import XCTest
@testable import PocketAide

// MARK: - MockURLProtocol

/// Intercepts `URLSession` requests and returns a pre-configured response.
final class MockURLProtocol: URLProtocol {

    /// Set this before each test to control what the "server" returns.
    static var responseHandler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = MockURLProtocol.responseHandler else {
            client?.urlProtocol(self, didFailWithError: URLError(.unknown))
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

// MARK: - Helpers

private func makeSession() -> URLSession {
    let config = URLSessionConfiguration.ephemeral
    config.protocolClasses = [MockURLProtocol.self]
    return URLSession(configuration: config)
}

private func makeResponse(
    url: URL,
    statusCode: Int,
    json: String
) -> (HTTPURLResponse, Data) {
    let response = HTTPURLResponse(
        url: url,
        statusCode: statusCode,
        httpVersion: nil,
        headerFields: ["Content-Type": "application/json"]
    )!
    return (response, Data(json.utf8))
}

// MARK: - MockAPIClientTests

final class MockAPIClientTests: XCTestCase {

    var mock: MockAPIClient!

    override func setUp() {
        super.setUp()
        mock = MockAPIClient()
    }

    override func tearDown() {
        mock = nil
        super.tearDown()
    }

    // MARK: Happy path

    func test_fetchHealth_returnsStubSuccessResponse() async throws {
        // Arrange
        mock.healthResult = .success(HealthResponse(status: "ok"))

        // Act
        let result = try await mock.fetchHealth()

        // Assert
        XCTAssertEqual(result.status, "ok")
    }

    func test_fetchHealth_recordsRequestInLog() async throws {
        // Arrange
        mock.healthResult = .success(HealthResponse(status: "ok"))

        // Act
        _ = try await mock.fetchHealth()

        // Assert
        XCTAssertEqual(mock.requestLog.count, 1)
        XCTAssertEqual(mock.requestLog[0].method, "GET")
        XCTAssertEqual(mock.requestLog[0].path, "/health")
    }

    func test_get_returnsRegisteredStub() async throws {
        // Arrange
        let expected = HealthResponse(status: "stubbed")
        mock.getResults["/health"] = expected

        // Act
        let result: HealthResponse = try await mock.get(path: "/health", token: nil)

        // Assert
        XCTAssertEqual(result, expected)
    }

    func test_post_returnsRegisteredStub() async throws {
        // Arrange
        struct DummyRequest: Encodable { let value: String }
        let expected = HealthResponse(status: "posted")
        mock.postResults["/echo"] = expected

        // Act
        let result: HealthResponse = try await mock.post(
            path: "/echo",
            body: DummyRequest(value: "ping"),
            token: nil
        )

        // Assert
        XCTAssertEqual(result, expected)
    }

    // MARK: Edge cases

    func test_reset_clearsRequestLog() async throws {
        // Arrange
        _ = try await mock.fetchHealth()
        XCTAssertEqual(mock.requestLog.count, 1)

        // Act
        mock.reset()

        // Assert
        XCTAssertEqual(mock.requestLog.count, 0)
    }

    func test_multipleRequests_allAppearInLog() async throws {
        // Arrange
        mock.healthResult = .success(HealthResponse(status: "ok"))

        // Act
        _ = try await mock.fetchHealth()
        _ = try await mock.fetchHealth()
        _ = try await mock.fetchHealth()

        // Assert
        XCTAssertEqual(mock.requestLog.count, 3)
    }

    // MARK: Error cases

    func test_fetchHealth_throwsConfiguredError() async throws {
        // Arrange
        mock.healthResult = .failure(.unauthorized)

        // Act & Assert
        do {
            _ = try await mock.fetchHealth()
            XCTFail("Expected APIClientError.unauthorized to be thrown")
        } catch let error as APIClientError {
            XCTAssertEqual(error, .unauthorized)
        }
    }

    func test_get_throwsWhenNoStubRegistered() async {
        // Act & Assert
        do {
            let _: HealthResponse = try await mock.get(path: "/unknown", token: nil)
            XCTFail("Expected networkError to be thrown")
        } catch let error as APIClientError {
            if case .networkError(let msg) = error {
                XCTAssertTrue(msg.contains("/unknown"), "Error message should reference the missing path")
            } else {
                XCTFail("Expected networkError, got \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    func test_fetchHealth_throwsHTTPErrorWithStatusCode() async throws {
        // Arrange
        mock.healthResult = .failure(.httpError(statusCode: 503, message: "Service Unavailable"))

        // Act & Assert
        do {
            _ = try await mock.fetchHealth()
            XCTFail("Expected httpError to be thrown")
        } catch let error as APIClientError {
            if case .httpError(let code, let message) = error {
                XCTAssertEqual(code, 503)
                XCTAssertEqual(message, "Service Unavailable")
            } else {
                XCTFail("Expected httpError, got \(error)")
            }
        }
    }
}

// MARK: - APIClientRequestBuilderTests

/// Verifies that the concrete `APIClient` builds correct HTTP requests
/// (method, path, headers) using `MockURLProtocol` to capture them.
final class APIClientRequestBuilderTests: XCTestCase {

    var capturedRequest: URLRequest?
    var client: APIClient!
    let baseURL = URL(string: "http://localhost:8080")!

    override func setUp() {
        super.setUp()
        capturedRequest = nil
        client = APIClient(baseURL: baseURL, session: makeSession())
    }

    override func tearDown() {
        MockURLProtocol.responseHandler = nil
        client = nil
        super.tearDown()
    }

    // MARK: Happy path

    func test_fetchHealth_sendsGETToHealthPath() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { [weak self] request in
            self?.capturedRequest = request
            return makeResponse(url: request.url!, statusCode: 200, json: #"{"status":"ok"}"#)
        }

        // Act
        _ = try await client.fetchHealth()

        // Assert
        XCTAssertNotNil(capturedRequest)
        XCTAssertEqual(capturedRequest?.httpMethod, "GET")
        XCTAssertTrue(capturedRequest?.url?.path == "/health")
    }

    func test_fetchHealth_parsesStatusOK() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { request in
            makeResponse(url: request.url!, statusCode: 200, json: #"{"status":"ok"}"#)
        }

        // Act
        let result = try await client.fetchHealth()

        // Assert
        XCTAssertEqual(result.status, "ok")
    }

    func test_get_attachesBearerTokenHeader() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { [weak self] request in
            self?.capturedRequest = request
            return makeResponse(url: request.url!, statusCode: 200, json: #"{"status":"ok"}"#)
        }

        // Act
        let _: HealthResponse = try await client.get(path: "/health", token: "my-jwt-token")

        // Assert
        let authHeader = capturedRequest?.value(forHTTPHeaderField: "Authorization")
        XCTAssertEqual(authHeader, "Bearer my-jwt-token")
    }

    func test_get_omitsAuthorizationHeaderWhenTokenIsNil() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { [weak self] request in
            self?.capturedRequest = request
            return makeResponse(url: request.url!, statusCode: 200, json: #"{"status":"ok"}"#)
        }

        // Act
        let _: HealthResponse = try await client.get(path: "/health", token: nil)

        // Assert
        XCTAssertNil(capturedRequest?.value(forHTTPHeaderField: "Authorization"))
    }

    // MARK: Error cases

    func test_fetchHealth_throwsHTTPErrorOn401() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { request in
            makeResponse(
                url: request.url!,
                statusCode: 401,
                json: #"{"message":"Unauthorized","code":401}"#
            )
        }

        // Act & Assert
        do {
            _ = try await client.fetchHealth()
            XCTFail("Expected httpError(401) to be thrown")
        } catch let error as APIClientError {
            if case .httpError(let code, let message) = error {
                XCTAssertEqual(code, 401)
                XCTAssertEqual(message, "Unauthorized")
            } else {
                XCTFail("Expected httpError, got \(error)")
            }
        }
    }

    func test_fetchHealth_throwsHTTPErrorOn500() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { request in
            makeResponse(
                url: request.url!,
                statusCode: 500,
                json: #"{"message":"Internal Server Error","code":500}"#
            )
        }

        // Act & Assert
        do {
            _ = try await client.fetchHealth()
            XCTFail("Expected httpError(500) to be thrown")
        } catch let error as APIClientError {
            if case .httpError(let code, _) = error {
                XCTAssertEqual(code, 500)
            } else {
                XCTFail("Expected httpError, got \(error)")
            }
        }
    }

    func test_fetchHealth_throwsDecodingErrorOnMalformedJSON() async throws {
        // Arrange
        MockURLProtocol.responseHandler = { request in
            makeResponse(url: request.url!, statusCode: 200, json: "not-json")
        }

        // Act & Assert
        do {
            _ = try await client.fetchHealth()
            XCTFail("Expected decodingError to be thrown")
        } catch let error as APIClientError {
            if case .decodingError = error {
                // expected
            } else {
                XCTFail("Expected decodingError, got \(error)")
            }
        }
    }
}

// MARK: - APIClientErrorEqualityTests

final class APIClientErrorEqualityTests: XCTestCase {

    func test_unauthorized_equalsUnauthorized() {
        XCTAssertEqual(APIClientError.unauthorized, APIClientError.unauthorized)
    }

    func test_httpError_equalWhenSameCodeAndMessage() {
        let a = APIClientError.httpError(statusCode: 404, message: "Not Found")
        let b = APIClientError.httpError(statusCode: 404, message: "Not Found")
        XCTAssertEqual(a, b)
    }

    func test_httpError_notEqualWhenDifferentCode() {
        let a = APIClientError.httpError(statusCode: 400, message: "Bad Request")
        let b = APIClientError.httpError(statusCode: 500, message: "Bad Request")
        XCTAssertNotEqual(a, b)
    }

    func test_decodingError_equalWhenSameMessage() {
        let a = APIClientError.decodingError("key missing")
        let b = APIClientError.decodingError("key missing")
        XCTAssertEqual(a, b)
    }

    func test_networkError_equalWhenSameMessage() {
        let a = APIClientError.networkError("timeout")
        let b = APIClientError.networkError("timeout")
        XCTAssertEqual(a, b)
    }
}
