// SyncTestDoubles.swift
// PocketAideTests
//
// DLD-740: 데이터 동기화 — 테스트 공통 타입 및 테스트 더블 정의
//
// 이 파일은 SyncServiceTests, OfflineQueueTests, NetworkMonitorTests에서
// 공통으로 사용되는 테스트 대역(test doubles)을 정의합니다.
//
// =========================================================
// Production 타입 인터페이스 명세 (구현 필요)
// =========================================================
//
// [SyncChange] — ios/PocketAide/Data/Sync/SyncChange.swift
//   struct SyncChange: Codable {
//     enum Entity: String, Codable { case todo, memo }
//     enum Operation: String, Codable { case create, update, delete }
//     let entity: Entity
//     let id: String
//     let operation: Operation
//     let payload: [String: AnyCodable]   // AnyCodable 또는 [String: String]
//     let updatedAt: Date
//     enum CodingKeys: String, CodingKey { case entity, id, operation, payload
//       case updatedAt = "updated_at" }
//   }
//
// [SyncRequest] — ios/PocketAide/Data/Sync/SyncChange.swift
//   struct SyncRequest: Encodable {
//     let changes: [SyncChange]
//   }
//
// [SyncServerData] — ios/PocketAide/Data/Sync/SyncChange.swift
//   struct SyncServerData: Decodable {
//     let todos: [TodoSyncItem]
//     let memos: [MemoSyncItem]
//     let routines: [RoutineSyncItem]
//   }
//   struct TodoSyncItem: Decodable {
//     let id: Int; let title: String; let type: String; let updatedAt: Date
//     enum CodingKeys: String, CodingKey { case id, title, type
//       case updatedAt = "updated_at" }
//   }
//   struct MemoSyncItem: Decodable {
//     let id: Int; let content: String; let source: String; let updatedAt: Date
//     enum CodingKeys: String, CodingKey { case id, content, source
//       case updatedAt = "updated_at" }
//   }
//   struct RoutineSyncItem: Decodable {
//     let id: Int; let name: String; let intervalDays: Int
//     let lastDoneAt: String; let updatedAt: Date
//     enum CodingKeys: String, CodingKey { case id, name
//       case intervalDays = "interval_days"
//       case lastDoneAt = "last_done_at"
//       case updatedAt = "updated_at" }
//   }
//
// [SyncResponse] — ios/PocketAide/Data/Sync/SyncChange.swift
//   struct SyncResponse: Decodable {
//     let serverData: SyncServerData
//     enum CodingKeys: String, CodingKey { case serverData = "server_data" }
//   }
//
// [SyncService] — ios/PocketAide/Data/Sync/SyncService.swift
//   final class SyncService {
//     init(apiClient: MockSyncAPIClient, offlineQueue: MockOfflineQueueForService)
//     func sync(token: String, serverURL: String) async throws -> SyncServerData
//   }
//   NOTE: 테스트 가능하도록 APIClient와 OfflineQueue를 프로토콜로 추상화 권장
//
// [OfflineQueue] — ios/PocketAide/Data/Sync/OfflineQueue.swift
//   final class OfflineQueue {
//     init(userDefaults: UserDefaults = .standard)
//     func enqueue(_ change: SyncChange)
//     func dequeueAll() -> [SyncChange]
//     func clearAll()
//     var count: Int { get }
//   }
//
// [NetworkMonitor] — ios/PocketAide/Data/Sync/NetworkMonitor.swift
//   final class NetworkMonitor {
//     var isConnected: Bool { get }
//     var onReconnect: (() -> Void)?
//     var isConnectedPublisher: AnyPublisher<Bool, Never> { get }
//     func start()
//     func stop()
//     // 테스트 전용 메서드 (#if DEBUG 또는 internal 접근 제어)
//     func simulateConnectionChange(isConnected: Bool)
//   }
// =========================================================

import Foundation
import Combine
@testable import PocketAide

// MARK: - SyncService Protocol

/// SyncServiceProtocol은 SettingsViewModel 등에서 SyncService를 주입받을 때 사용합니다.
protocol SyncServiceProtocol {
    func sync(token: String, serverURL: String) async throws -> SyncServerData
}

// MARK: - Mock SyncService

/// MockSyncService는 SyncServiceProtocol을 구현한 테스트 대역입니다.
/// ViewModel 테스트에서 SyncService 의존성을 대체할 때 사용합니다.
final class MockSyncService: SyncServiceProtocol {

    // MARK: Properties

    var stubbedServerData: SyncServerData?
    var stubbedError: Error?
    var callCount: Int = 0
    var lastToken: String?
    var lastServerURL: String?

    // MARK: SyncServiceProtocol

    func sync(token: String, serverURL: String) async throws -> SyncServerData {
        callCount += 1
        lastToken = token
        lastServerURL = serverURL

        if let error = stubbedError {
            throw error
        }
        return stubbedServerData ?? SyncServerData(todos: [], memos: [], routines: [])
    }
}
