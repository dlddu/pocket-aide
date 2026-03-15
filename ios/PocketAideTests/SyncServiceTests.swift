// SyncServiceTests.swift
// PocketAideTests
//
// Unit tests for SyncService.
//
// DLD-740: 데이터 동기화 — SyncService unit 테스트
//
// 테스트 대상:
//   - SyncService.sync(token:serverURL:) — 변경사항 업로드 + 서버 데이터 다운로드
//   - SyncService 초기화 및 의존성 주입
//   - 성공/실패/오프라인 시나리오
//
// NOTE: TDD Red Phase — SyncService가 아직 구현되지 않았으므로
//       이 테스트는 실패 상태입니다. 구현 후 통과해야 합니다.
//
// 구현 필요 파일: ios/PocketAide/Data/Sync/SyncService.swift
//
// SyncService 요구사항:
//   - init(apiClient:offlineQueue:)
//   - func sync(token: String, serverURL: String) async throws -> SyncServerData
//   - POST /sync API 호출: SyncRequest { changes: [SyncChange] }
//   - 응답: SyncResponse { server_data: SyncServerData { todos, memos, routines } }

import XCTest
@testable import PocketAide

// MARK: - Test Doubles

/// MockAPIClient는 SyncService가 의존하는 APIClient를 대체하는 테스트 대역입니다.
/// 실제 네트워크 호출 없이 원하는 응답(성공/실패)을 주입할 수 있습니다.
final class MockSyncAPIClient {

    // MARK: Properties

    /// 요청 시 반환할 SyncResponse 데이터.
    var stubbedSyncResponse: SyncResponse?

    /// 요청 시 throw할 에러.
    var stubbedError: Error?

    /// 캡처된 마지막 요청 본문.
    var capturedSyncRequest: SyncRequest?

    /// 호출 횟수.
    var callCount: Int = 0

    // MARK: Methods

    func performSync(request: SyncRequest, token: String, serverURL: String) async throws -> SyncResponse {
        callCount += 1
        capturedSyncRequest = request

        if let error = stubbedError {
            throw error
        }
        guard let response = stubbedSyncResponse else {
            // 기본 응답: 빈 server_data
            return SyncResponse(serverData: SyncServerData(todos: [], memos: [], routines: []))
        }
        return response
    }
}

/// MockOfflineQueue는 SyncService가 사용할 오프라인 큐의 테스트 대역입니다.
final class MockOfflineQueueForService {

    // MARK: Properties

    /// 큐에 저장된 변경사항 목록.
    var changes: [SyncChange] = []

    /// dequeueAll 호출 시 반환할 변경사항. nil이면 changes를 반환.
    var stubbedChanges: [SyncChange]?

    /// clearAll 호출 횟수.
    var clearAllCallCount: Int = 0

    // MARK: Methods

    func dequeueAll() -> [SyncChange] {
        return stubbedChanges ?? changes
    }

    func clearAll() {
        clearAllCallCount += 1
        changes.removeAll()
    }
}

// MARK: - SyncServiceTests

final class SyncServiceTests: XCTestCase {

    // MARK: - Properties

    private var mockAPIClient: MockSyncAPIClient!
    private var mockQueue: MockOfflineQueueForService!

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        mockAPIClient = MockSyncAPIClient()
        mockQueue = MockOfflineQueueForService()
    }

    override func tearDown() {
        mockAPIClient = nil
        mockQueue = nil
        super.tearDown()
    }

    // MARK: - Happy Path: 동기화 성공

    /// 오프라인 큐에 변경사항이 없을 때 POST /sync를 호출하면 빈 changes 배열이 전송되어야 한다.
    ///
    /// Scenario:
    ///   오프라인 큐 = 빈 상태
    ///   → SyncRequest.changes = []
    ///   → POST /sync 호출
    ///   → 서버 데이터 반환
    func test_sync_emptyQueue_sendsEmptyChanges() async throws {
        // Arrange
        mockQueue.stubbedChanges = []
        mockAPIClient.stubbedSyncResponse = SyncResponse(
            serverData: SyncServerData(todos: [], memos: [], routines: [])
        )
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act
        let result = try await sut.sync(token: "test-token", serverURL: "https://api.example.com")

        // Assert
        XCTAssertNotNil(result, "sync should return SyncServerData")
        XCTAssertEqual(mockAPIClient.callCount, 1, "API should be called once")
        XCTAssertEqual(mockAPIClient.capturedSyncRequest?.changes.count, 0,
                       "Expected empty changes array when queue is empty")
    }

    /// 오프라인 큐에 변경사항이 있을 때 POST /sync를 호출하면 해당 변경사항이 전송되어야 한다.
    ///
    /// Scenario:
    ///   오프라인 큐 = [todo_create, memo_create]
    ///   → SyncRequest.changes = [todo_create, memo_create]
    ///   → POST /sync 호출
    ///   → 서버 데이터 반환 + 큐 초기화
    func test_sync_withQueuedChanges_sendsAllChanges() async throws {
        // Arrange
        let changes = [
            SyncChange(
                entity: .todo,
                id: "offline-todo-1",
                operation: .create,
                payload: ["title": "오프라인 투두", "type": "personal"],
                updatedAt: Date()
            ),
            SyncChange(
                entity: .memo,
                id: "offline-memo-1",
                operation: .create,
                payload: ["content": "오프라인 메모", "source": "text"],
                updatedAt: Date()
            )
        ]
        mockQueue.stubbedChanges = changes
        mockAPIClient.stubbedSyncResponse = SyncResponse(
            serverData: SyncServerData(todos: [], memos: [], routines: [])
        )
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act
        let result = try await sut.sync(token: "test-token", serverURL: "https://api.example.com")

        // Assert
        XCTAssertNotNil(result)
        XCTAssertEqual(mockAPIClient.capturedSyncRequest?.changes.count, 2,
                       "Expected 2 changes from the offline queue to be sent")
    }

    /// 동기화 성공 후 오프라인 큐가 초기화되어야 한다.
    ///
    /// Scenario:
    ///   오프라인 큐에 변경사항 존재
    ///   → 동기화 성공
    ///   → 큐 clearAll 호출됨
    func test_sync_onSuccess_clearsOfflineQueue() async throws {
        // Arrange
        mockQueue.stubbedChanges = [
            SyncChange(
                entity: .todo,
                id: "todo-1",
                operation: .create,
                payload: ["title": "테스트"],
                updatedAt: Date()
            )
        ]
        mockAPIClient.stubbedSyncResponse = SyncResponse(
            serverData: SyncServerData(todos: [], memos: [], routines: [])
        )
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act
        _ = try await sut.sync(token: "test-token", serverURL: "https://api.example.com")

        // Assert
        XCTAssertEqual(mockQueue.clearAllCallCount, 1,
                       "Offline queue should be cleared after successful sync")
    }

    /// 서버로부터 todos, memos, routines 데이터를 반환받아야 한다.
    ///
    /// Scenario:
    ///   서버가 todos: 2건, memos: 1건, routines: 1건 응답
    ///   → SyncServerData에 각 항목 수가 올바르게 포함됨
    func test_sync_returnsServerData_withAllEntityTypes() async throws {
        // Arrange
        let stubbedTodos = [
            TodoSyncItem(id: 1, title: "서버 투두 1", type: "personal", updatedAt: Date()),
            TodoSyncItem(id: 2, title: "서버 투두 2", type: "work", updatedAt: Date())
        ]
        let stubbedMemos = [
            MemoSyncItem(id: 1, content: "서버 메모 1", source: "text", updatedAt: Date())
        ]
        let stubbedRoutines = [
            RoutineSyncItem(id: 1, name: "서버 루틴 1", intervalDays: 7, lastDoneAt: "2026-01-01", updatedAt: Date())
        ]
        mockAPIClient.stubbedSyncResponse = SyncResponse(
            serverData: SyncServerData(
                todos: stubbedTodos,
                memos: stubbedMemos,
                routines: stubbedRoutines
            )
        )
        mockQueue.stubbedChanges = []
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act
        let result = try await sut.sync(token: "test-token", serverURL: "https://api.example.com")

        // Assert
        XCTAssertEqual(result.todos.count, 2, "Expected 2 todos in server data")
        XCTAssertEqual(result.memos.count, 1, "Expected 1 memo in server data")
        XCTAssertEqual(result.routines.count, 1, "Expected 1 routine in server data")
    }

    // MARK: - Error Cases

    /// 네트워크 에러 발생 시 SyncService는 에러를 throw해야 한다.
    ///
    /// Scenario:
    ///   API 호출 시 네트워크 에러 발생
    ///   → sync throws APIError.networkError
    ///   → 큐는 초기화되지 않음 (실패 시 변경사항 보존)
    func test_sync_onNetworkError_throwsError() async {
        // Arrange
        mockAPIClient.stubbedError = APIError.networkError(URLError(.notConnectedToInternet))
        mockQueue.stubbedChanges = [
            SyncChange(
                entity: .todo,
                id: "pending-1",
                operation: .create,
                payload: ["title": "미전송 투두"],
                updatedAt: Date()
            )
        ]
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act & Assert
        do {
            _ = try await sut.sync(token: "test-token", serverURL: "https://api.example.com")
            XCTFail("Expected error to be thrown on network failure")
        } catch {
            // pass — any error is acceptable
        }
    }

    /// 네트워크 에러 발생 시 오프라인 큐가 초기화되지 않아야 한다 (변경사항 보존).
    ///
    /// Scenario:
    ///   API 호출 실패
    ///   → clearAll이 호출되지 않음
    func test_sync_onNetworkError_preservesOfflineQueue() async {
        // Arrange
        mockAPIClient.stubbedError = APIError.networkError(URLError(.notConnectedToInternet))
        mockQueue.stubbedChanges = [
            SyncChange(
                entity: .todo,
                id: "pending-2",
                operation: .create,
                payload: ["title": "보존될 투두"],
                updatedAt: Date()
            )
        ]
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act
        _ = try? await sut.sync(token: "test-token", serverURL: "https://api.example.com")

        // Assert
        XCTAssertEqual(mockQueue.clearAllCallCount, 0,
                       "Offline queue must NOT be cleared when sync fails (preserve pending changes)")
    }

    /// 인증 에러(401) 발생 시 SyncService는 APIError.unauthorized를 throw해야 한다.
    ///
    /// Scenario:
    ///   서버가 401 응답
    ///   → sync throws APIError.unauthorized
    func test_sync_onUnauthorized_throwsUnauthorizedError() async {
        // Arrange
        mockAPIClient.stubbedError = APIError.unauthorized
        mockQueue.stubbedChanges = []
        let sut = SyncService(apiClient: mockAPIClient, offlineQueue: mockQueue)

        // Act & Assert
        do {
            _ = try await sut.sync(token: "expired-token", serverURL: "https://api.example.com")
            XCTFail("Expected APIError.unauthorized to be thrown")
        } catch let error as APIError {
            XCTAssertEqual(error, .unauthorized, "Expected .unauthorized error for 401 response")
        } catch {
            XCTFail("Expected APIError, got \(type(of: error)): \(error)")
        }
    }

    // MARK: - Edge Cases

    /// SyncChange의 operation 타입이 올바르게 직렬화/역직렬화되어야 한다.
    ///
    /// Scenario:
    ///   create, update, delete 각 operation 타입 검증
    func test_syncChange_operations_serializeCorrectly() throws {
        // Arrange
        let operations: [SyncChange.Operation] = [.create, .update, .delete]
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        for operation in operations {
            let change = SyncChange(
                entity: .todo,
                id: "test-\(operation)",
                operation: operation,
                payload: ["title": "테스트"],
                updatedAt: Date()
            )

            // Act
            let data = try encoder.encode(change)
            let decoded = try decoder.decode(SyncChange.self, from: data)

            // Assert
            XCTAssertEqual(decoded.operation, operation,
                           "Operation \(operation) should survive encode/decode roundtrip")
        }
    }

    /// SyncChange의 entity 타입이 올바르게 직렬화/역직렬화되어야 한다.
    ///
    /// Scenario:
    ///   todo, memo 각 entity 타입 검증
    func test_syncChange_entities_serializeCorrectly() throws {
        // Arrange
        let entities: [SyncChange.Entity] = [.todo, .memo]
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        for entity in entities {
            let change = SyncChange(
                entity: entity,
                id: "test-\(entity)",
                operation: .create,
                payload: ["title": "테스트"],
                updatedAt: Date()
            )

            // Act
            let data = try encoder.encode(change)
            let decoded = try decoder.decode(SyncChange.self, from: data)

            // Assert
            XCTAssertEqual(decoded.entity, entity,
                           "Entity \(entity) should survive encode/decode roundtrip")
        }
    }

    /// SyncChange의 updated_at이 ISO8601(RFC3339) 형식으로 직렬화되어야 한다.
    ///
    /// Scenario:
    ///   SyncChange를 JSON으로 직렬화할 때 updated_at이 ISO8601 형식
    func test_syncChange_updatedAt_encodesAsISO8601() throws {
        // Arrange
        let knownDate = Date(timeIntervalSince1970: 1_736_000_000) // 고정 timestamp
        let change = SyncChange(
            entity: .todo,
            id: "time-test",
            operation: .create,
            payload: ["title": "시간 테스트"],
            updatedAt: knownDate
        )

        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601

        // Act
        let data = try encoder.encode(change)
        let json = try JSONSerialization.jsonObject(with: data) as? [String: Any]

        // Assert
        let updatedAt = json?["updated_at"] as? String
        XCTAssertNotNil(updatedAt, "updated_at should be present as a string in JSON")
        // ISO8601 형식: "2026-01-04T12:53:20Z" 처럼 T와 Z가 포함됨
        XCTAssertTrue(updatedAt?.contains("T") == true, "updated_at should be in ISO8601 format (contains 'T')")
    }
}
