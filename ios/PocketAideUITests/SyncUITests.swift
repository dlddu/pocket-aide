// SyncUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end data sync flow:
//   오프라인 상태에서 투두 추가/편집
//   → 온라인 복귀 → 자동 동기화 트리거
//   → 서버 데이터와 일치 확인
//   → 다른 디바이스의 변경사항 반영 확인.
//
// DLD-739: 13-1: 데이터 동기화 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-739:
//   - SyncManager 또는 동등한 오프라인-큐 + 자동 동기화 인프라가 구현됨
//   - "--uitesting-offline" 런치 인자가 네트워크를 차단하여 오프라인 모드를 시뮬레이션함
//   - "--uitesting-online" 런치 인자가 네트워크를 복원하여 동기화를 트리거함
//   - 동기화 완료 시 접근성 식별자 "sync_status_synced"가 노출됨
//   - 동기화 진행 중 접근성 식별자 "sync_status_syncing"이 노출됨
//   - 동기화 실패 시 접근성 식별자 "sync_status_error"가 노출됨
//   - 오프라인 배지/인디케이터가 "offline_indicator"로 노출됨
//   - 투두 관련 접근성 식별자는 TodoUITests.swift의 패턴을 따름

import XCTest

final class SyncUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // consistent with the pattern used by TodoUITests and ScratchUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: 오프라인 투두 추가 후 동기화

    /// 오프라인 상태에서 투두를 추가한 뒤 온라인으로 복귀하면
    /// 자동 동기화가 실행되어 서버 데이터와 일치해야 한다.
    ///
    /// Expected flow:
    ///   "--uitesting-offline" 모드로 앱 재실행
    ///   → "tab_todo" 탭 이동 → "add_todo_button" 탭
    ///   → "todo_title_field"에 "오프라인 투두" 입력 → "todo_save_button" 탭
    ///   → 투두 로컬 저장 확인 ("todo_row_오프라인 투두" 존재)
    ///   → "--uitesting-online" 모드로 전환 (앱 재시작 또는 네트워크 복원 시뮬레이션)
    ///   → "sync_status_synced" 인디케이터 대기
    ///   → "todo_row_오프라인 투두" 여전히 표시 (서버와 동기화 완료)
    func test_sync_offlineTodoAdd_syncedOnReconnect() throws {
        // Launch in offline mode to simulate no network connectivity.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-offline"]
        app.launch()

        // Navigate to the Todo tab.
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible")
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Todo list must appear")

        // Add a new todo while offline.
        let addButton = app.buttons["add_todo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add todo button must exist")
        addButton.tap()

        let titleField = app.textFields["todo_title_field"]
        XCTAssertTrue(titleField.waitForExistence(timeout: 5), "Todo title field must appear")
        titleField.tap()
        titleField.typeText("오프라인 투두")

        let saveButton = app.buttons["todo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5), "Save button must appear")
        saveButton.tap()

        // Verify the todo is saved locally.
        let todoRow = app.otherElements["todo_row_오프라인 투두"]
        XCTAssertTrue(
            todoRow.waitForExistence(timeout: 5),
            "Offline todo should be saved locally and visible in the list"
        )

        // Simulate going back online (restart with online flag).
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-online"]
        app.launch()

        // Wait for the sync completion indicator.
        let syncedIndicator = app.otherElements["sync_status_synced"]
        XCTAssertTrue(
            syncedIndicator.waitForExistence(timeout: 15),
            "Sync should complete within 15 seconds of going online (sync_status_synced)"
        )

        // Verify the offline todo still appears after sync (it was uploaded).
        tabBar.buttons["tab_todo"].tap()
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Todo list must be visible after sync"
        )
        XCTAssertTrue(
            todoRow.waitForExistence(timeout: 5),
            "Offline todo '오프라인 투두' should still appear after sync with server"
        )
    }

    // MARK: - Happy Path: 오프라인 투두 편집 후 동기화

    /// 오프라인 상태에서 기존 투두를 편집한 뒤 온라인으로 복귀하면
    /// 편집 내용이 서버에 반영되어야 한다.
    ///
    /// Expected flow:
    ///   온라인 상태에서 투두 "편집 전 제목" 생성 후 동기화 대기
    ///   → 오프라인 전환
    ///   → "todo_row_편집 전 제목" 탭 → 편집 화면에서 "편집 후 제목"으로 수정 → 저장
    ///   → 온라인 복귀 → "sync_status_synced" 대기
    ///   → "todo_row_편집 후 제목" 표시 확인
    func test_sync_offlineTodoEdit_syncedOnReconnect() throws {
        // Step 1: Add a todo while online, then wait for initial sync.
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_todo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let titleField = app.textFields["todo_title_field"]
        XCTAssertTrue(titleField.waitForExistence(timeout: 5))
        titleField.tap()
        titleField.typeText("편집 전 제목")

        let saveButton = app.buttons["todo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        let originalRow = app.otherElements["todo_row_편집 전 제목"]
        XCTAssertTrue(originalRow.waitForExistence(timeout: 5))

        let initialSyncedIndicator = app.otherElements["sync_status_synced"]
        XCTAssertTrue(
            initialSyncedIndicator.waitForExistence(timeout: 15),
            "Initial sync must complete before going offline"
        )

        // Step 2: Go offline and edit the todo.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-offline"]
        app.launch()

        tabBar.buttons["tab_todo"].tap()
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        // Open the edit sheet for the existing todo.
        originalRow.tap()

        let editField = app.textFields["todo_title_field"]
        XCTAssertTrue(editField.waitForExistence(timeout: 5), "Edit title field must appear")
        editField.clearAndEnterText("편집 후 제목")

        let updateButton = app.buttons["todo_save_button"]
        XCTAssertTrue(updateButton.waitForExistence(timeout: 5))
        updateButton.tap()

        let editedRow = app.otherElements["todo_row_편집 후 제목"]
        XCTAssertTrue(editedRow.waitForExistence(timeout: 5), "Edited todo must appear locally")

        // Step 3: Go back online and wait for sync.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-online"]
        app.launch()

        let syncedIndicator = app.otherElements["sync_status_synced"]
        XCTAssertTrue(
            syncedIndicator.waitForExistence(timeout: 15),
            "Sync must complete after going online"
        )

        // Verify the edited title is reflected after sync.
        tabBar.buttons["tab_todo"].tap()
        XCTAssertTrue(listView.waitForExistence(timeout: 5))
        XCTAssertTrue(
            editedRow.waitForExistence(timeout: 5),
            "Edited todo '편집 후 제목' should be visible after sync"
        )
    }

    // MARK: - Happy Path: 다른 디바이스 변경사항 반영

    /// 다른 디바이스(또는 서버 직접 수정)의 변경사항이 동기화 시 로컬에 반영되어야 한다.
    ///
    /// Expected flow:
    ///   서버에 "다른 디바이스 투두" 가 직접 존재하는 상태에서 앱 실행
    ///   → "--uitesting-has-remote-changes" 런치 인자로 원격 변경사항 존재 시뮬레이션
    ///   → "sync_status_synced" 대기
    ///   → "tab_todo" 이동 → "todo_row_다른 디바이스 투두" 표시 확인
    func test_sync_remoteChanges_reflectedAfterSync() throws {
        // Restart the app with a flag that simulates remote changes being available
        // on the server (e.g., injected via a test API or environment variable).
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-has-remote-changes"]
        app.launch()

        // Wait for automatic sync to complete upon launch.
        let syncedIndicator = app.otherElements["sync_status_synced"]
        XCTAssertTrue(
            syncedIndicator.waitForExistence(timeout: 15),
            "App should automatically sync on launch when network is available (sync_status_synced)"
        )

        // Navigate to the Todo tab and verify the remote todo is now visible.
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let remoteRow = app.otherElements["todo_row_다른 디바이스 투두"]
        XCTAssertTrue(
            remoteRow.waitForExistence(timeout: 5),
            "Remote todo '다른 디바이스 투두' should appear after syncing remote changes"
        )
    }

    // MARK: - Happy Path: 동기화 진행 중 인디케이터 표시

    /// 동기화가 진행 중일 때 sync_status_syncing 인디케이터가 표시되어야 한다.
    ///
    /// Expected flow:
    ///   오프라인에서 투두 1건 추가 후 온라인 복귀
    ///   → "sync_status_syncing" 인디케이터가 잠시 표시됨
    ///   → 이후 "sync_status_synced"로 전환됨
    func test_sync_showsSyncingIndicatorDuringSync() throws {
        // Prepare an offline change so that sync has something to do.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-offline"]
        app.launch()

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_todo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let titleField = app.textFields["todo_title_field"]
        XCTAssertTrue(titleField.waitForExistence(timeout: 5))
        titleField.tap()
        titleField.typeText("동기화 테스트 투두")

        let saveButton = app.buttons["todo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        // Go online — this should trigger a sync, showing the syncing indicator.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-online"]
        app.launch()

        // The syncing indicator should appear briefly during the upload.
        let syncingIndicator = app.otherElements["sync_status_syncing"]
        XCTAssertTrue(
            syncingIndicator.waitForExistence(timeout: 10),
            "Syncing indicator (sync_status_syncing) should appear while upload is in progress"
        )

        // The indicator should transition to synced once complete.
        let syncedIndicator = app.otherElements["sync_status_synced"]
        XCTAssertTrue(
            syncedIndicator.waitForExistence(timeout: 15),
            "Sync status should transition to synced (sync_status_synced) after upload completes"
        )
    }

    // MARK: - Edge Case: 오프라인 인디케이터 표시

    /// 네트워크가 없을 때 오프라인 인디케이터가 표시되어야 한다.
    ///
    /// Expected flow:
    ///   "--uitesting-offline" 모드로 앱 실행
    ///   → "offline_indicator" 접근성 요소가 화면에 표시됨
    func test_sync_offlineMode_displaysOfflineIndicator() throws {
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-offline"]
        app.launch()

        let offlineIndicator = app.otherElements["offline_indicator"]
        XCTAssertTrue(
            offlineIndicator.waitForExistence(timeout: 5),
            "Offline indicator (offline_indicator) must be visible when the device has no network"
        )
    }

    // MARK: - Edge Case: 충돌 해결 (last-write-wins)

    /// 로컬과 서버에서 같은 투두가 각각 수정된 경우, 더 최신 updated_at을 가진 쪽이 이겨야 한다.
    ///
    /// Expected flow:
    ///   "--uitesting-conflict-server-newer" 런치 인자로 서버가 더 최신인 충돌 상황 시뮬레이션
    ///   → 동기화 후 서버의 최신 값 "서버 최신 제목"이 화면에 표시됨
    ///   → 로컬의 오래된 값 "로컬 오래된 제목"은 표시되지 않음
    func test_sync_conflictResolution_serverNewerWins() throws {
        // Launch with a preconfigured conflict scenario where the server version
        // is newer than the local version of the same todo.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-conflict-server-newer"]
        app.launch()

        let syncedIndicator = app.otherElements["sync_status_synced"]
        XCTAssertTrue(
            syncedIndicator.waitForExistence(timeout: 15),
            "Sync must complete to resolve the conflict"
        )

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        // The server's newer title must win.
        let serverRow = app.otherElements["todo_row_서버 최신 제목"]
        XCTAssertTrue(
            serverRow.waitForExistence(timeout: 5),
            "Server's newer title '서버 최신 제목' must be visible after conflict resolution (last-write-wins)"
        )

        // The stale local title must NOT appear.
        let staleRow = app.otherElements["todo_row_로컬 오래된 제목"]
        XCTAssertFalse(
            staleRow.waitForExistence(timeout: 3),
            "Stale local title '로컬 오래된 제목' must NOT appear — server was newer"
        )
    }

    // MARK: - Error Case: 동기화 실패 시 에러 인디케이터 표시

    /// 서버 오류 등으로 동기화에 실패하면 sync_status_error 인디케이터가 표시되어야 한다.
    ///
    /// Expected flow:
    ///   "--uitesting-sync-error" 런치 인자로 서버 응답 오류 시뮬레이션
    ///   → 동기화 시도 후 "sync_status_error" 인디케이터가 표시됨
    func test_sync_serverError_displaysErrorIndicator() throws {
        // Launch with a simulated server error so that sync always fails.
        app.terminate()
        app.launchArguments = ["--uitesting", "--uitesting-sync-error"]
        app.launch()

        let errorIndicator = app.otherElements["sync_status_error"]
        XCTAssertTrue(
            errorIndicator.waitForExistence(timeout: 15),
            "Error indicator (sync_status_error) must appear when the sync request fails"
        )
    }
}

// MARK: - XCUIElement Helper

private extension XCUIElement {

    /// Clears any existing text in the field and types the given string.
    func clearAndEnterText(_ text: String) {
        guard let existingText = self.value as? String, !existingText.isEmpty else {
            self.tap()
            self.typeText(text)
            return
        }
        self.tap()
        let deleteString = String(repeating: XCUIKeyboardKey.delete.rawValue, count: existingText.count)
        self.typeText(deleteString)
        self.typeText(text)
    }
}
