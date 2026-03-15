// MemoService.swift
// PocketAide

import Foundation

/// 서버와 통신하여 메모 CRUD 및 이동 처리를 수행하는 데이터 레이어 서비스.
final class MemoService {

    // MARK: - Request Types

    struct CreateRequest: Encodable {
        let content: String
        let source: String
    }

    struct UpdateRequest: Encodable {
        let content: String
    }

    struct MoveRequest: Encodable {
        let target: String
    }

    // MARK: - CRUD

    /// 메모 목록을 조회합니다.
    func list(serverURL: String, token: String) async throws -> [Memo] {
        let client = makeClient(serverURL: serverURL)
        return try await client.request(
            path: "memos",
            method: .get,
            token: token
        )
    }

    /// 새 메모를 생성합니다.
    func create(content: String, source: String = "text", serverURL: String, token: String) async throws -> Memo {
        let client = makeClient(serverURL: serverURL)
        let body = CreateRequest(content: content, source: source)
        return try await client.request(
            path: "memos",
            method: .post,
            body: body,
            token: token
        )
    }

    /// 메모를 수정합니다.
    func update(id: Int, content: String, serverURL: String, token: String) async throws -> Memo {
        let client = makeClient(serverURL: serverURL)
        let body = UpdateRequest(content: content)
        return try await client.request(
            path: "memos/\(id)",
            method: .put,
            body: body,
            token: token
        )
    }

    /// 메모를 삭제합니다.
    func delete(id: Int, serverURL: String, token: String) async throws {
        let client = makeClient(serverURL: serverURL)
        try await client.requestVoid(
            path: "memos/\(id)",
            method: .delete,
            token: token
        )
    }

    /// 메모를 투두로 이동합니다.
    func move(id: Int, target: String, serverURL: String, token: String) async throws {
        let client = makeClient(serverURL: serverURL)
        let body = MoveRequest(target: target)
        try await client.requestVoid(
            path: "memos/\(id)/move",
            method: .post,
            body: body,
            token: token
        )
    }

    // MARK: - Private

    private func makeClient(serverURL: String) -> APIClient {
        let baseURL = URL(string: serverURL)!
        return APIClient(baseURL: baseURL)
    }
}

// MARK: - MemoServiceProtocol Conformance

extension MemoService: MemoServiceProtocol {}
