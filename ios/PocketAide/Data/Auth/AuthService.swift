// AuthService.swift
// PocketAide

import Foundation

/// 서버와 통신하여 로그인/회원가입을 수행하는 데이터 레이어 서비스.
final class AuthService {

    // MARK: - Response Types

    private struct LoginResponse: Decodable {
        let token: String
    }

    private struct RegisterResponse: Decodable {
        let id: Int
        let email: String
    }

    // MARK: - Request Types

    private struct AuthRequest: Encodable {
        let email: String
        let password: String
    }

    // MARK: - Public Interface

    /// 서버에 로그인 요청을 보내고 JWT 토큰을 반환합니다.
    ///
    /// - Parameters:
    ///   - email: 사용자 이메일
    ///   - password: 비밀번호
    ///   - serverURL: 서버 주소 (예: "http://localhost:8080")
    /// - Returns: JWT 토큰 문자열
    /// - Throws: ``APIError``
    func login(email: String, password: String, serverURL: String) async throws -> String {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }

        let client = APIClient(baseURL: baseURL)
        let body = AuthRequest(email: email, password: password)
        let response: LoginResponse = try await client.request(
            path: "/auth/login",
            method: .post,
            body: body
        )
        return response.token
    }

    /// 서버에 회원가입 요청을 보냅니다.
    ///
    /// - Parameters:
    ///   - email: 사용자 이메일
    ///   - password: 비밀번호
    ///   - serverURL: 서버 주소
    /// - Throws: ``APIError``
    func register(email: String, password: String, serverURL: String) async throws {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }

        let client = APIClient(baseURL: baseURL)
        let body = AuthRequest(email: email, password: password)
        let _: RegisterResponse = try await client.request(
            path: "/auth/register",
            method: .post,
            body: body
        )
    }
}
