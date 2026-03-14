// APIClient.swift
// PocketAide

import Foundation

/// URLSession 기반의 HTTP API 클라이언트.
///
/// 테스트 시에는 `session` 파라미터에 `MockURLProtocol`을 등록한
/// `URLSession`을 주입하여 네트워크 호출을 차단할 수 있습니다.
public final class APIClient {

    // MARK: Properties

    private let baseURL: URL
    private let session: URLSession
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    // MARK: Init

    public init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
        self.encoder = JSONEncoder()
        self.decoder = JSONDecoder()
    }

    // MARK: Request

    /// 주어진 경로와 메서드로 HTTP 요청을 수행하고 응답을 디코딩합니다.
    ///
    /// - Parameters:
    ///   - path: 베이스 URL에 붙일 경로 (e.g. `"/health"`).
    ///   - method: HTTP 메서드.
    ///   - body: 요청 본문으로 직렬화할 `Encodable` 값. 기본값 `nil`.
    ///   - token: Bearer 인증 토큰. 있으면 `Authorization` 헤더에 추가됩니다.
    /// - Returns: 응답 본문을 `T`로 디코딩한 값.
    /// - Throws: ``APIError``
    public func request<T: Decodable>(
        path: String,
        method: HTTPMethod,
        body: (some Encodable)? = nil,
        token: String? = nil
    ) async throws -> T {
        let data = try await performRequest(path: path, method: method, body: body, token: token)

        // JSON 디코딩
        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw APIError.decodingError(error)
        }
    }

    /// 주어진 경로와 메서드로 HTTP 요청을 수행합니다. 응답 본문을 디코딩하지 않습니다.
    ///
    /// - Parameters:
    ///   - path: 베이스 URL에 붙일 경로 (e.g. `"/memos/1"`).
    ///   - method: HTTP 메서드.
    ///   - body: 요청 본문으로 직렬화할 `Encodable` 값. 기본값 `nil`.
    ///   - token: Bearer 인증 토큰. 있으면 `Authorization` 헤더에 추가됩니다.
    /// - Throws: ``APIError``
    public func requestVoid(
        path: String,
        method: HTTPMethod,
        body: (some Encodable)? = nil,
        token: String? = nil
    ) async throws {
        _ = try await performRequest(path: path, method: method, body: body, token: token)
    }

    // MARK: Private

    private func performRequest(
        path: String,
        method: HTTPMethod,
        body: (some Encodable)?,
        token: String?
    ) async throws -> Data {
        // URL 구성
        let url: URL
        if path.isEmpty {
            url = baseURL
        } else {
            url = baseURL.appendingPathComponent(path)
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method.rawValue
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body {
            urlRequest.httpBody = try encoder.encode(body)
        }

        // 네트워크 호출
        let data: Data
        let response: URLResponse

        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.networkError(error)
        }

        // HTTP 상태 코드 처리
        if let httpResponse = response as? HTTPURLResponse {
            switch httpResponse.statusCode {
            case 200..<300:
                break
            case 401:
                throw APIError.unauthorized
            case 404:
                throw APIError.notFound
            default:
                throw APIError.serverError(httpResponse.statusCode)
            }
        }

        return data
    }
}
