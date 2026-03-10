// Tests for AppGroupStorage.
//
// `AppGroupStorage` is initialised with `suiteName: nil` so it falls back to
// `UserDefaults.standard`. This makes the tests portable to Linux CI where
// no App Group entitlement can be registered.
//
// Each test uses a unique key prefix via the `testSuiteID` to avoid
// cross-contamination between parallel test runs.

import XCTest
@testable import PocketAide

final class AppGroupStorageTests: XCTestCase {

    /// Use a transient, in-process UserDefaults domain to keep tests isolated.
    private var storage: AppGroupStorage!
    private let isolatedSuite = "com.test.AppGroupStorageTests.\(UUID().uuidString)"

    override func setUp() {
        super.setUp()
        // Pass the isolated suite name so each test class run gets a clean domain.
        storage = AppGroupStorage(suiteName: isolatedSuite)
    }

    override func tearDown() {
        // Remove all keys written during the test.
        let defaults = UserDefaults(suiteName: isolatedSuite)
        AppGroupKey.allCases.forEach { key in
            defaults?.removeObject(forKey: key.rawValue)
        }
        storage = nil
        super.tearDown()
    }

    // MARK: - String

    func test_setString_andRetrieveIt() {
        // Arrange
        let value = "Hello from PocketAide"

        // Act
        storage.set(value, forKey: .lastChatMessage)
        let retrieved = storage.string(forKey: .lastChatMessage)

        // Assert
        XCTAssertEqual(retrieved, value)
    }

    func test_getString_returnsNil_whenKeyAbsent() {
        // Act
        let result = storage.string(forKey: .lastChatMessage)

        // Assert
        XCTAssertNil(result)
    }

    func test_setString_overwritesPreviousValue() {
        // Arrange
        storage.set("first", forKey: .lastChatMessage)

        // Act
        storage.set("second", forKey: .lastChatMessage)
        let result = storage.string(forKey: .lastChatMessage)

        // Assert
        XCTAssertEqual(result, "second")
    }

    func test_setString_preservesEmptyString() {
        // Act
        storage.set("", forKey: .lastChatMessage)
        let result = storage.string(forKey: .lastChatMessage)

        // Assert
        XCTAssertEqual(result, "")
    }

    // MARK: - Integer

    func test_setInteger_andRetrieveIt() {
        // Arrange
        let value = 42

        // Act
        storage.set(value, forKey: .pendingNoteCount)
        let retrieved = storage.integer(forKey: .pendingNoteCount)

        // Assert
        XCTAssertEqual(retrieved, value)
    }

    func test_getInteger_returnsNil_whenKeyAbsent() {
        // Act
        let result = storage.integer(forKey: .pendingNoteCount)

        // Assert
        XCTAssertNil(result)
    }

    func test_setInteger_zero_isDistinguishableFromAbsent() {
        // Act
        storage.set(0, forKey: .pendingNoteCount)
        let result = storage.integer(forKey: .pendingNoteCount)

        // Assert — 0 was explicitly stored, so result must be non-nil
        XCTAssertNotNil(result)
        XCTAssertEqual(result, 0)
    }

    func test_setInteger_negativeValue() {
        // Act
        storage.set(-5, forKey: .pendingNoteCount)
        let result = storage.integer(forKey: .pendingNoteCount)

        // Assert
        XCTAssertEqual(result, -5)
    }

    // MARK: - Date

    func test_setDate_andRetrieveIt() {
        // Arrange
        let now = Date()

        // Act
        storage.set(now, forKey: .lastSyncDate)
        let retrieved = storage.date(forKey: .lastSyncDate)

        // Assert
        XCTAssertNotNil(retrieved)
        // Allow 1 ms tolerance for floating-point round-trip.
        XCTAssertEqual(retrieved!.timeIntervalSince1970, now.timeIntervalSince1970, accuracy: 0.001)
    }

    func test_getDate_returnsNil_whenKeyAbsent() {
        // Act
        let result = storage.date(forKey: .lastSyncDate)

        // Assert
        XCTAssertNil(result)
    }

    func test_setDate_overwritesPreviousValue() {
        // Arrange
        let first  = Date(timeIntervalSince1970: 1_000_000)
        let second = Date(timeIntervalSince1970: 2_000_000)
        storage.set(first, forKey: .lastSyncDate)

        // Act
        storage.set(second, forKey: .lastSyncDate)
        let result = storage.date(forKey: .lastSyncDate)

        // Assert
        XCTAssertEqual(result?.timeIntervalSince1970, second.timeIntervalSince1970, accuracy: 0.001)
    }

    // MARK: - Remove

    func test_remove_deletesStringValue() {
        // Arrange
        storage.set("to be removed", forKey: .lastChatMessage)
        XCTAssertNotNil(storage.string(forKey: .lastChatMessage))

        // Act
        storage.remove(forKey: .lastChatMessage)

        // Assert
        XCTAssertNil(storage.string(forKey: .lastChatMessage))
    }

    func test_remove_deletesIntegerValue() {
        // Arrange
        storage.set(99, forKey: .pendingNoteCount)
        XCTAssertNotNil(storage.integer(forKey: .pendingNoteCount))

        // Act
        storage.remove(forKey: .pendingNoteCount)

        // Assert
        XCTAssertNil(storage.integer(forKey: .pendingNoteCount))
    }

    func test_remove_onAbsentKey_doesNotThrow() {
        // Act & Assert — must not crash
        storage.remove(forKey: .lastChatMessage)
        XCTAssertNil(storage.string(forKey: .lastChatMessage))
    }

    // MARK: - Key isolation

    func test_differentKeys_doNotInterfere() {
        // Act
        storage.set("chat value", forKey: .lastChatMessage)
        storage.set(7, forKey: .pendingNoteCount)

        // Assert
        XCTAssertEqual(storage.string(forKey: .lastChatMessage), "chat value")
        XCTAssertEqual(storage.integer(forKey: .pendingNoteCount), 7)
    }

    // MARK: - App Group identifier

    func test_appGroupIdentifier_matchesExpectedValue() {
        XCTAssertEqual(AppGroupStorage.appGroupIdentifier, "group.com.dlddu.pocket-aide")
    }
}

// MARK: - AppGroupKey CaseIterable (for teardown convenience)

extension AppGroupKey: CaseIterable {
    public static var allCases: [AppGroupKey] {
        [.lastChatMessage, .pendingNoteCount, .lastSyncDate]
    }
}
