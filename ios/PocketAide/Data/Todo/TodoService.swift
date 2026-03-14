// TodoService.swift
// PocketAide

import Foundation

/// 서버와 통신하여 투두 CRUD 및 토글 처리를 수행하는 데이터 레이어 서비스.
final class TodoService {

    // MARK: - Request Types

    struct CreateRequest: Encodable {
        let title: String
        let type: String
    }

    struct UpdateRequest: Encodable {
        let title: String?
    }

    // MARK: - CRUD

    /// 투두 목록을 조회합니다 (type 필터링 포함).
    func list(type todoType: String, serverURL: String, token: String) async throws -> [Todo] {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/todos?type=\(todoType)",
            method: .get,
            token: token
        )
    }

    /// 특정 투두를 조회합니다.
    func get(id: Int, serverURL: String, token: String) async throws -> Todo {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/todos/\(id)",
            method: .get,
            token: token
        )
    }

    /// 새 투두를 생성합니다.
    func create(
        title: String,
        type todoType: String = "personal",
        serverURL: String,
        token: String
    ) async throws -> Todo {
        let client = makeClient(serverURL: serverURL)
        let body = CreateRequest(title: title, type: todoType)
        return try await client.request(
            path: "/todos",
            method: .post,
            body: body,
            token: token
        )
    }

    /// 투두를 수정합니다.
    func update(
        id: Int,
        title: String? = nil,
        serverURL: String,
        token: String
    ) async throws -> Todo {
        let client = makeClient(serverURL: serverURL)
        let body = UpdateRequest(title: title)
        return try await client.request(
            path: "/todos/\(id)",
            method: .put,
            body: body,
            token: token
        )
    }

    /// 투두를 삭제합니다.
    func delete(id: Int, serverURL: String, token: String) async throws {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }
        let url = baseURL.appendingPathComponent("/todos/\(id)")
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = HTTPMethod.delete.rawValue
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (_, response): (Data, URLResponse)
        do {
            (_, response) = try await URLSession.shared.data(for: urlRequest)
        } catch {
            throw APIError.networkError(error)
        }

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
    }

    /// 투두 완료 상태를 토글합니다.
    func toggle(id: Int, serverURL: String, token: String) async throws -> Todo {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/todos/\(id)/toggle",
            method: .post,
            token: token
        )
    }

    // MARK: - Private

    private func makeClient(serverURL: String) -> APIClient {
        let baseURL = URL(string: serverURL) ?? URL(string: "http://localhost:8080")!
        return APIClient(baseURL: baseURL)
    }
}
