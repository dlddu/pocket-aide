// RoutineService.swift
// PocketAide

import Foundation

/// 서버와 통신하여 루틴 CRUD 및 완료 처리를 수행하는 데이터 레이어 서비스.
final class RoutineService {

    // MARK: - Request Types

    struct CreateRequest: Encodable {
        let name: String
        let intervalDays: Int
        let lastDoneAt: String

        enum CodingKeys: String, CodingKey {
            case name
            case intervalDays = "interval_days"
            case lastDoneAt = "last_done_at"
        }
    }

    struct UpdateRequest: Encodable {
        let name: String?
        let intervalDays: Int?
        let note: String?
        let notifyEnabled: Bool?

        enum CodingKeys: String, CodingKey {
            case name
            case intervalDays = "interval_days"
            case note
            case notifyEnabled = "notify_enabled"
        }
    }

    // MARK: - CRUD

    /// 루틴 목록을 조회합니다.
    func list(serverURL: String, token: String) async throws -> [Routine] {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/routines",
            method: .get,
            token: token
        )
    }

    /// 특정 루틴을 조회합니다.
    func get(id: Int, serverURL: String, token: String) async throws -> Routine {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/routines/\(id)",
            method: .get,
            token: token
        )
    }

    /// 새 루틴을 생성합니다.
    func create(
        name: String,
        intervalDays: Int,
        lastDoneAt: String,
        serverURL: String,
        token: String
    ) async throws -> Routine {
        let client = makeClient(serverURL: serverURL)
        let body = CreateRequest(name: name, intervalDays: intervalDays, lastDoneAt: lastDoneAt)
        return try await client.request(
            path: "/routines",
            method: .post,
            body: body,
            token: token
        )
    }

    /// 루틴을 수정합니다.
    func update(
        id: Int,
        name: String? = nil,
        intervalDays: Int? = nil,
        note: String? = nil,
        notifyEnabled: Bool? = nil,
        serverURL: String,
        token: String
    ) async throws -> Routine {
        let client = makeClient(serverURL: serverURL)
        let body = UpdateRequest(
            name: name,
            intervalDays: intervalDays,
            note: note,
            notifyEnabled: notifyEnabled
        )
        return try await client.request(
            path: "/routines/\(id)",
            method: .put,
            body: body,
            token: token
        )
    }

    /// 루틴을 삭제합니다.
    func delete(id: Int, serverURL: String, token: String) async throws {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }
        let url = baseURL.appendingPathComponent("/routines/\(id)")
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

    /// 루틴을 완료 처리합니다 (last_done_at = 오늘, next_due_date 재계산).
    func complete(id: Int, serverURL: String, token: String) async throws -> Routine {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/routines/\(id)/complete",
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
