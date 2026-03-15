// SyncChange.swift
// PocketAide
//
// DLD-740: 데이터 동기화 — 동기화 변경사항 모델

import Foundation

// MARK: - SyncChange

/// 클라이언트에서 서버로 전송되는 단일 변경사항.
struct SyncChange: Codable, Equatable {

    // MARK: - Nested Types

    /// 동기화 대상 엔티티 타입.
    enum Entity: String, Codable, Equatable {
        case todo
        case memo
    }

    /// 동기화 작업 유형.
    enum Operation: String, Codable, Equatable {
        case create
        case update
        case delete
    }

    // MARK: - Properties

    let entity: Entity
    let id: String
    let operation: Operation
    let payload: [String: String]
    let updatedAt: Date

    // MARK: - CodingKeys

    enum CodingKeys: String, CodingKey {
        case entity
        case id
        case operation
        case payload
        case updatedAt = "updated_at"
    }
}

// MARK: - SyncRequest

/// POST /sync 요청 본문.
struct SyncRequest: Encodable {
    let changes: [SyncChange]
}

// MARK: - SyncServerData

/// POST /sync 응답의 server_data 필드.
struct SyncServerData: Decodable, Equatable {
    let todos: [TodoSyncItem]
    let memos: [MemoSyncItem]
    let routines: [RoutineSyncItem]
}

// MARK: - TodoSyncItem

/// 동기화 응답의 단일 투두 항목.
struct TodoSyncItem: Decodable, Equatable {
    let id: Int
    let title: String
    let type: String
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id, title, type
        case updatedAt = "updated_at"
    }
}

// MARK: - MemoSyncItem

/// 동기화 응답의 단일 메모 항목.
struct MemoSyncItem: Decodable, Equatable {
    let id: Int
    let content: String
    let source: String
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id, content, source
        case updatedAt = "updated_at"
    }
}

// MARK: - RoutineSyncItem

/// 동기화 응답의 단일 루틴 항목.
struct RoutineSyncItem: Decodable, Equatable {
    let id: Int
    let name: String
    let intervalDays: Int
    let lastDoneAt: String
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id, name
        case intervalDays = "interval_days"
        case lastDoneAt = "last_done_at"
        case updatedAt = "updated_at"
    }
}

// MARK: - SyncResponse

/// POST /sync 응답 전체 구조.
struct SyncResponse: Decodable {
    let serverData: SyncServerData

    enum CodingKeys: String, CodingKey {
        case serverData = "server_data"
    }

    // 테스트용 이니셜라이저
    init(serverData: SyncServerData) {
        self.serverData = serverData
    }
}
