// SentenceService.swift
// PocketAide

import Foundation

/// 서버와 통신하여 문장 모음 CRUD를 수행하는 데이터 레이어 서비스.
final class SentenceService {

    // MARK: - Request Types

    struct CreateCategoryRequest: Encodable {
        let name: String
    }

    struct UpdateCategoryRequest: Encodable {
        let name: String
    }

    struct CreateSentenceRequest: Encodable {
        let content: String
        let categoryId: Int

        enum CodingKeys: String, CodingKey {
            case content
            case categoryId = "category_id"
        }
    }

    struct UpdateSentenceRequest: Encodable {
        let content: String
    }

    // MARK: - Category CRUD

    /// 카테고리 목록을 조회합니다.
    func listCategories(serverURL: String, token: String) async throws -> [SentenceCategory] {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "/sentences/categories",
            method: .get,
            token: token
        )
    }

    /// 새 카테고리를 생성합니다.
    func createCategory(name: String, serverURL: String, token: String) async throws -> SentenceCategory {
        let client = makeClient(serverURL: serverURL)
        let body = CreateCategoryRequest(name: name)
        return try await client.request(
            path: "/sentences/categories",
            method: .post,
            body: body,
            token: token
        )
    }

    /// 카테고리를 수정합니다.
    func updateCategory(id: Int, name: String, serverURL: String, token: String) async throws -> SentenceCategory {
        let client = makeClient(serverURL: serverURL)
        let body = UpdateCategoryRequest(name: name)
        return try await client.request(
            path: "/sentences/categories/\(id)",
            method: .put,
            body: body,
            token: token
        )
    }

    /// 카테고리를 삭제합니다.
    func deleteCategory(id: Int, serverURL: String, token: String) async throws {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }
        let url = baseURL.appendingPathComponent("/sentences/categories/\(id)")
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

    // MARK: - Sentence CRUD

    /// 문장 목록을 조회합니다 (카테고리 필터링 포함).
    func listSentences(categoryId: Int? = nil, serverURL: String, token: String) async throws -> [Sentence] {
        let client = makeClient(serverURL: serverURL)
        var path = "/sentences"
        if let categoryId = categoryId {
            path += "?category_id=\(categoryId)"
        }
        return try await client.request(
            path: path,
            method: .get,
            token: token
        )
    }

    /// 새 문장을 생성합니다.
    func createSentence(content: String, categoryId: Int, serverURL: String, token: String) async throws -> Sentence {
        let client = makeClient(serverURL: serverURL)
        let body = CreateSentenceRequest(content: content, categoryId: categoryId)
        return try await client.request(
            path: "/sentences",
            method: .post,
            body: body,
            token: token
        )
    }

    /// 문장을 수정합니다.
    func updateSentence(id: Int, content: String, serverURL: String, token: String) async throws -> Sentence {
        let client = makeClient(serverURL: serverURL)
        let body = UpdateSentenceRequest(content: content)
        return try await client.request(
            path: "/sentences/\(id)",
            method: .put,
            body: body,
            token: token
        )
    }

    /// 문장을 삭제합니다.
    func deleteSentence(id: Int, serverURL: String, token: String) async throws {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }
        let url = baseURL.appendingPathComponent("/sentences/\(id)")
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

    // MARK: - Private

    private func makeClient(serverURL: String) -> APIClient {
        let baseURL = URL(string: serverURL)!
        return APIClient(baseURL: baseURL)
    }
}
