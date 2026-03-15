// OfflineQueueTests.swift
// PocketAideTests
//
// Unit tests for OfflineQueue.
//
// DLD-740: 데이터 동기화 — OfflineQueue unit 테스트
//
// 테스트 대상:
//   - OfflineQueue.enqueue(_:) — 변경사항 추가
//   - OfflineQueue.dequeueAll() — 전체 변경사항 조회
//   - OfflineQueue.clearAll() — 전체 변경사항 초기화
//   - OfflineQueue.count — 큐에 저장된 항목 수
//   - 큐의 FIFO 순서 보장
//   - UserDefaults 영속화
//
// NOTE: TDD Red Phase — OfflineQueue가 아직 구현되지 않았으므로
//       이 테스트는 실패 상태입니다. 구현 후 통과해야 합니다.
//
// 구현 필요 파일: ios/PocketAide/Data/Sync/OfflineQueue.swift
//
// OfflineQueue 요구사항:
//   - init(userDefaults: UserDefaults = .standard)
//   - func enqueue(_ change: SyncChange)
//   - func dequeueAll() -> [SyncChange]
//   - func clearAll()
//   - var count: Int { get }
//   - 변경사항은 UserDefaults에 JSON으로 영속화됨

import XCTest
@testable import PocketAide

final class OfflineQueueTests: XCTestCase {

    // MARK: - Properties

    /// 테스트 전용 UserDefaults (각 테스트마다 격리된 저장소 사용)
    private var testDefaults: UserDefaults!
    private let suiteName = "OfflineQueueTests-\(UUID().uuidString)"

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        testDefaults = UserDefaults(suiteName: suiteName)
        testDefaults.removePersistentDomain(forName: suiteName)
    }

    override func tearDown() {
        testDefaults.removePersistentDomain(forName: suiteName)
        testDefaults = nil
        super.tearDown()
    }

    // MARK: - Helpers

    private func makeSUT() -> OfflineQueue {
        return OfflineQueue(userDefaults: testDefaults)
    }

    private func makeTodoChange(
        id: String = UUID().uuidString,
        operation: SyncChange.Operation = .create,
        title: String = "테스트 투두"
    ) -> SyncChange {
        return SyncChange(
            entity: .todo,
            id: id,
            operation: operation,
            payload: ["title": title, "type": "personal"],
            updatedAt: Date()
        )
    }

    private func makeMemoChange(
        id: String = UUID().uuidString,
        operation: SyncChange.Operation = .create,
        content: String = "테스트 메모"
    ) -> SyncChange {
        return SyncChange(
            entity: .memo,
            id: id,
            operation: operation,
            payload: ["content": content, "source": "text"],
            updatedAt: Date()
        )
    }

    // MARK: - Initialisation

    /// 새로 생성된 큐는 비어 있어야 한다.
    ///
    /// Scenario:
    ///   새 OfflineQueue 생성
    ///   → count == 0
    ///   → dequeueAll() == []
    func test_init_startsEmpty() {
        // Arrange & Act
        let sut = makeSUT()

        // Assert
        XCTAssertEqual(sut.count, 0, "New queue should have count 0")
        XCTAssertTrue(sut.dequeueAll().isEmpty, "New queue should return empty array")
    }

    // MARK: - Enqueue

    /// 변경사항 1건을 enqueue하면 count가 1이 되어야 한다.
    ///
    /// Scenario:
    ///   enqueue(todo_create)
    ///   → count == 1
    func test_enqueue_singleChange_incrementsCount() {
        // Arrange
        let sut = makeSUT()
        let change = makeTodoChange()

        // Act
        sut.enqueue(change)

        // Assert
        XCTAssertEqual(sut.count, 1, "After enqueuing 1 change, count should be 1")
    }

    /// 여러 변경사항을 enqueue하면 count가 정확하게 증가해야 한다.
    ///
    /// Scenario:
    ///   enqueue 3건
    ///   → count == 3
    func test_enqueue_multipleChanges_countMatchesEnqueuedItems() {
        // Arrange
        let sut = makeSUT()

        // Act
        sut.enqueue(makeTodoChange(id: "1"))
        sut.enqueue(makeMemoChange(id: "2"))
        sut.enqueue(makeTodoChange(id: "3", operation: .update))

        // Assert
        XCTAssertEqual(sut.count, 3, "After enqueuing 3 changes, count should be 3")
    }

    /// enqueue된 변경사항이 dequeueAll()로 전부 반환되어야 한다.
    ///
    /// Scenario:
    ///   enqueue(todo_create)
    ///   dequeueAll()
    ///   → [todo_create]
    func test_enqueue_singleChange_dequeueAllReturnsIt() {
        // Arrange
        let sut = makeSUT()
        let change = makeTodoChange(id: "unique-id", title: "큐 투두")

        // Act
        sut.enqueue(change)
        let result = sut.dequeueAll()

        // Assert
        XCTAssertEqual(result.count, 1, "dequeueAll should return 1 item")
        XCTAssertEqual(result.first?.id, "unique-id", "dequeueAll should return the enqueued change")
    }

    // MARK: - FIFO Order

    /// enqueue된 순서대로 dequeueAll()에서 반환되어야 한다 (FIFO).
    ///
    /// Scenario:
    ///   enqueue A, B, C 순서로 추가
    ///   dequeueAll()
    ///   → [A, B, C] (삽입 순서 보장)
    func test_dequeueAll_preservesFIFOOrder() {
        // Arrange
        let sut = makeSUT()
        let changeA = makeTodoChange(id: "A")
        let changeB = makeMemoChange(id: "B")
        let changeC = makeTodoChange(id: "C", operation: .update)

        // Act
        sut.enqueue(changeA)
        sut.enqueue(changeB)
        sut.enqueue(changeC)
        let result = sut.dequeueAll()

        // Assert
        XCTAssertEqual(result.count, 3, "Expected 3 changes")
        XCTAssertEqual(result[0].id, "A", "First item should be A (FIFO order)")
        XCTAssertEqual(result[1].id, "B", "Second item should be B (FIFO order)")
        XCTAssertEqual(result[2].id, "C", "Third item should be C (FIFO order)")
    }

    // MARK: - ClearAll

    /// clearAll() 호출 후 count가 0이 되어야 한다.
    ///
    /// Scenario:
    ///   enqueue 2건 → clearAll()
    ///   → count == 0
    func test_clearAll_resetsCountToZero() {
        // Arrange
        let sut = makeSUT()
        sut.enqueue(makeTodoChange(id: "1"))
        sut.enqueue(makeMemoChange(id: "2"))

        // Act
        sut.clearAll()

        // Assert
        XCTAssertEqual(sut.count, 0, "After clearAll, count should be 0")
    }

    /// clearAll() 호출 후 dequeueAll()이 빈 배열을 반환해야 한다.
    ///
    /// Scenario:
    ///   enqueue 후 clearAll()
    ///   → dequeueAll() == []
    func test_clearAll_dequeueAllReturnsEmpty() {
        // Arrange
        let sut = makeSUT()
        sut.enqueue(makeTodoChange())
        sut.clearAll()

        // Act
        let result = sut.dequeueAll()

        // Assert
        XCTAssertTrue(result.isEmpty, "dequeueAll should return empty array after clearAll")
    }

    /// 빈 큐에서 clearAll()을 호출해도 에러가 발생하지 않아야 한다.
    ///
    /// Scenario:
    ///   빈 큐에서 clearAll() 호출
    ///   → 에러 없음, count 여전히 0
    func test_clearAll_onEmptyQueue_isNoOp() {
        // Arrange
        let sut = makeSUT()

        // Act & Assert (must not throw or crash)
        sut.clearAll()
        XCTAssertEqual(sut.count, 0, "clearAll on empty queue should be a no-op")
    }

    // MARK: - Persistence

    /// 큐에 enqueue된 변경사항은 새 OfflineQueue 인스턴스에서도 복원되어야 한다.
    ///
    /// Scenario:
    ///   sut1.enqueue(change) → sut1 deallocated
    ///   sut2 = 새 인스턴스 (같은 UserDefaults)
    ///   sut2.dequeueAll() → [change] 복원됨
    func test_persistence_enqueuedChanges_restoredAfterReinit() {
        // Arrange
        let changeID = "persistent-id"
        let changeTitle = "영속화 테스트 투두"

        // Act: enqueue with first instance
        let sut1 = OfflineQueue(userDefaults: testDefaults)
        sut1.enqueue(makeTodoChange(id: changeID, title: changeTitle))

        // Simulate app restart with a new instance pointing to the same UserDefaults.
        let sut2 = OfflineQueue(userDefaults: testDefaults)
        let result = sut2.dequeueAll()

        // Assert
        XCTAssertEqual(result.count, 1, "Persisted change should be restored after re-init")
        XCTAssertEqual(result.first?.id, changeID,
                       "Restored change should have the same id as the enqueued one")
        if case .todo = result.first?.entity {
            // pass
        } else {
            XCTFail("Restored change entity should be .todo")
        }
    }

    /// clearAll() 후 새 인스턴스를 생성하면 큐가 비어 있어야 한다.
    ///
    /// Scenario:
    ///   enqueue → clearAll() → 새 인스턴스
    ///   → count == 0
    func test_persistence_afterClearAll_newInstanceStartsEmpty() {
        // Arrange
        let sut1 = OfflineQueue(userDefaults: testDefaults)
        sut1.enqueue(makeTodoChange())
        sut1.clearAll()

        // Act
        let sut2 = OfflineQueue(userDefaults: testDefaults)

        // Assert
        XCTAssertEqual(sut2.count, 0, "After clearAll, new instance should start empty")
    }

    // MARK: - Entity Types

    /// todo와 memo 변경사항을 모두 enqueue하고 dequeueAll()로 반환할 수 있어야 한다.
    ///
    /// Scenario:
    ///   enqueue(todo), enqueue(memo)
    ///   dequeueAll() → [todo, memo]
    ///   각 entity 타입이 올바르게 유지됨
    func test_dequeueAll_returnsCorrectEntityTypes() {
        // Arrange
        let sut = makeSUT()
        let todoChange = makeTodoChange(id: "todo-entity")
        let memoChange = makeMemoChange(id: "memo-entity")

        // Act
        sut.enqueue(todoChange)
        sut.enqueue(memoChange)
        let result = sut.dequeueAll()

        // Assert
        XCTAssertEqual(result.count, 2)

        let todoResult = result.first { $0.id == "todo-entity" }
        let memoResult = result.first { $0.id == "memo-entity" }

        XCTAssertNotNil(todoResult, "todo change should be in the queue")
        XCTAssertNotNil(memoResult, "memo change should be in the queue")

        if let todo = todoResult {
            if case .todo = todo.entity {
                // pass
            } else {
                XCTFail("Expected todo entity, got \(todo.entity)")
            }
        }
        if let memo = memoResult {
            if case .memo = memo.entity {
                // pass
            } else {
                XCTFail("Expected memo entity, got \(memo.entity)")
            }
        }
    }

    // MARK: - Operation Types

    /// create, update, delete 각 operation 타입이 enqueue/dequeue 후 보존되어야 한다.
    ///
    /// Scenario:
    ///   create, update, delete 각 operation으로 enqueue
    ///   dequeueAll() → operation 타입 보존됨
    func test_dequeueAll_preservesOperationTypes() {
        // Arrange
        let sut = makeSUT()
        sut.enqueue(makeTodoChange(id: "op-create", operation: .create))
        sut.enqueue(makeTodoChange(id: "op-update", operation: .update))
        sut.enqueue(makeTodoChange(id: "op-delete", operation: .delete))

        // Act
        let result = sut.dequeueAll()

        // Assert
        let createChange = result.first { $0.id == "op-create" }
        let updateChange = result.first { $0.id == "op-update" }
        let deleteChange = result.first { $0.id == "op-delete" }

        XCTAssertEqual(createChange?.operation, .create, "create operation should be preserved")
        XCTAssertEqual(updateChange?.operation, .update, "update operation should be preserved")
        XCTAssertEqual(deleteChange?.operation, .delete, "delete operation should be preserved")
    }

    // MARK: - Edge Cases

    /// 같은 id의 변경사항을 여러 번 enqueue하면 모두 보존되어야 한다 (중복 허용).
    ///
    /// Scenario:
    ///   같은 id로 2번 enqueue
    ///   → count == 2 (dedup 없음)
    func test_enqueue_duplicateIds_bothPreserved() {
        // Arrange
        let sut = makeSUT()
        let sameID = "same-id"

        // Act
        sut.enqueue(makeTodoChange(id: sameID, title: "첫 번째"))
        sut.enqueue(makeTodoChange(id: sameID, operation: .update, title: "두 번째"))

        // Assert
        XCTAssertEqual(sut.count, 2,
                       "Both changes with same id should be preserved (queue does not dedup)")
    }

    /// dequeueAll()을 여러 번 호출해도 큐가 비워지지 않아야 한다 (non-destructive peek).
    ///
    /// Scenario:
    ///   enqueue(change)
    ///   dequeueAll() — 첫 번째 호출
    ///   dequeueAll() — 두 번째 호출
    ///   → 두 번 모두 [change] 반환 (큐 비워지지 않음)
    func test_dequeueAll_isNonDestructive_multipleCallsReturnSameItems() {
        // Arrange
        let sut = makeSUT()
        sut.enqueue(makeTodoChange(id: "peek-test"))

        // Act
        let firstResult = sut.dequeueAll()
        let secondResult = sut.dequeueAll()

        // Assert
        XCTAssertEqual(firstResult.count, 1, "First dequeueAll should return 1 item")
        XCTAssertEqual(secondResult.count, 1, "Second dequeueAll should also return 1 item (non-destructive)")
        XCTAssertEqual(firstResult.first?.id, secondResult.first?.id,
                       "Both calls should return the same item")
    }
}
