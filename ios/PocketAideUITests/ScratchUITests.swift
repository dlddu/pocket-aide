// ScratchUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end scratch space (임시 공간) flow:
//   Tap Scratch tab → memo list visible
//   → add new text memo → appears in list
//   → tap move button → select destination (personal_todo) → todo tab shows new item
//   → confirm memo is removed from scratch space after move.
//
// DLD-729: 8-1: 임시 공간 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-730:
//   - A "tab_scratch" tab item is added to MainTabView
//   - ScratchListView is wired up with accessibilityIdentifier "scratch_list_view"
//   - Add-memo button exposes: "add_memo_button"
//   - Memo text input field exposes: "memo_text_field"
//   - Memo save button exposes: "memo_save_button"
//   - Each memo row exposes "memo_row_<content>"
//   - Move button per row exposes "memo_move_button_<content>"
//   - Move destination sheet exposes: "move_destination_sheet"
//   - Personal todo destination button exposes: "move_to_personal_todo_button"

import XCTest

final class ScratchUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // consistent with the pattern used by TodoUITests and RoutineUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Scratch Tab Navigation

    /// Tapping the Scratch tab must display the memo list view container.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_scratch" → "scratch_list_view" appears
    func test_scratchTab_displaysMemoList() throws {
        throw XCTSkip("DLD-730: ScratchScreen not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")

        // Act
        tabBar.buttons["tab_scratch"].tap()

        // Assert
        let listView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Scratch list view (scratch_list_view) should appear after tapping the Scratch tab"
        )
    }

    // MARK: - Happy Path: Add Memo

    /// Adding a new text memo must cause it to appear in the scratch memo list.
    ///
    /// Expected flow:
    ///   "tab_scratch" → tap "add_memo_button" → input sheet appears
    ///   → fill "memo_text_field" with "오늘 할 일 정리"
    ///   → tap "memo_save_button" → sheet dismisses
    ///   → scratch list contains "memo_row_오늘 할 일 정리"
    func test_scratchTab_addMemo_displaysInList() throws {
        throw XCTSkip("DLD-730: ScratchScreen not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_scratch"].tap()

        let listView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_memo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add memo button must exist")

        // Act — open the add-memo input and fill in the content
        addButton.tap()

        let textField = app.textFields["memo_text_field"]
        XCTAssertTrue(textField.waitForExistence(timeout: 5), "Memo text field must appear")
        textField.tap()
        textField.typeText("오늘 할 일 정리")

        let saveButton = app.buttons["memo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5), "Save button must appear")
        saveButton.tap()

        // Assert — the new memo must appear in the scratch list
        let memoRow = app.otherElements["memo_row_오늘 할 일 정리"]
        XCTAssertTrue(
            memoRow.waitForExistence(timeout: 5),
            "Newly added memo '오늘 할 일 정리' (memo_row_오늘 할 일 정리) should appear in the scratch list"
        )
    }

    // MARK: - Happy Path: Move Memo to Personal Todo

    /// Moving a memo to personal_todo must cause the item to appear in the
    /// Todo tab under the personal todo list.
    ///
    /// Expected flow:
    ///   "tab_scratch" → add memo "운동 계획 세우기"
    ///   → tap "memo_move_button_운동 계획 세우기"
    ///   → "move_destination_sheet" appears
    ///   → tap "move_to_personal_todo_button"
    ///   → navigate to "tab_todo"
    ///   → "todo_row_운동 계획 세우기" visible in todo list
    func test_scratchTab_moveMemo_toPersonalTodo_appearsInTodoTab() throws {
        throw XCTSkip("DLD-730: ScratchScreen not yet implemented")

        // Arrange — navigate to scratch tab and add a memo
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_scratch"].tap()

        let listView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_memo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let textField = app.textFields["memo_text_field"]
        XCTAssertTrue(textField.waitForExistence(timeout: 5))
        textField.tap()
        textField.typeText("운동 계획 세우기")

        let saveButton = app.buttons["memo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        let memoRow = app.otherElements["memo_row_운동 계획 세우기"]
        XCTAssertTrue(
            memoRow.waitForExistence(timeout: 5),
            "Memo '운동 계획 세우기' must appear before attempting to move"
        )

        // Act — tap the move button and select personal todo as destination
        let moveButton = app.buttons["memo_move_button_운동 계획 세우기"]
        XCTAssertTrue(
            moveButton.waitForExistence(timeout: 5),
            "Move button (memo_move_button_운동 계획 세우기) must exist on the memo row"
        )
        moveButton.tap()

        let destinationSheet = app.otherElements["move_destination_sheet"]
        XCTAssertTrue(
            destinationSheet.waitForExistence(timeout: 5),
            "Move destination sheet (move_destination_sheet) must appear after tapping move"
        )

        let personalTodoButton = app.buttons["move_to_personal_todo_button"]
        XCTAssertTrue(
            personalTodoButton.waitForExistence(timeout: 5),
            "Personal todo button (move_to_personal_todo_button) must be visible in the destination sheet"
        )
        personalTodoButton.tap()

        // Navigate to the Todo tab to verify the moved item
        tabBar.buttons["tab_todo"].tap()

        let todoListView = app.otherElements["todo_list_view"]
        XCTAssertTrue(
            todoListView.waitForExistence(timeout: 5),
            "Todo list view must be visible after switching to the Todo tab"
        )

        // Assert — new todo derived from memo must appear in the todo list
        let todoRow = app.otherElements["todo_row_운동 계획 세우기"]
        XCTAssertTrue(
            todoRow.waitForExistence(timeout: 5),
            "Todo '운동 계획 세우기' (todo_row_운동 계획 세우기) should appear in the Todo tab after moving from scratch"
        )
    }

    // MARK: - Happy Path: Memo Removed from Scratch After Move

    /// After a memo is moved to a todo, it must no longer appear in the
    /// scratch space list.
    ///
    /// Expected flow:
    ///   "tab_scratch" → add memo "독서 목록 정리"
    ///   → move to personal_todo via "move_to_personal_todo_button"
    ///   → return to "tab_scratch"
    ///   → "memo_row_독서 목록 정리" no longer exists in scratch list
    func test_scratchTab_moveMemo_deletedFromScratch() throws {
        throw XCTSkip("DLD-730: ScratchScreen not yet implemented")

        // Arrange — navigate to scratch tab and add a memo
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_scratch"].tap()

        let listView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_memo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let textField = app.textFields["memo_text_field"]
        XCTAssertTrue(textField.waitForExistence(timeout: 5))
        textField.tap()
        textField.typeText("독서 목록 정리")

        let saveButton = app.buttons["memo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        let memoRow = app.otherElements["memo_row_독서 목록 정리"]
        XCTAssertTrue(
            memoRow.waitForExistence(timeout: 5),
            "Memo '독서 목록 정리' must appear before attempting to move"
        )

        // Act — move the memo to personal_todo
        let moveButton = app.buttons["memo_move_button_독서 목록 정리"]
        XCTAssertTrue(
            moveButton.waitForExistence(timeout: 5),
            "Move button (memo_move_button_독서 목록 정리) must exist on the memo row"
        )
        moveButton.tap()

        let destinationSheet = app.otherElements["move_destination_sheet"]
        XCTAssertTrue(
            destinationSheet.waitForExistence(timeout: 5),
            "Move destination sheet must appear after tapping move"
        )

        let personalTodoButton = app.buttons["move_to_personal_todo_button"]
        XCTAssertTrue(
            personalTodoButton.waitForExistence(timeout: 5),
            "Personal todo button must be visible in the destination sheet"
        )
        personalTodoButton.tap()

        // Return to the scratch tab to verify the memo is gone
        tabBar.buttons["tab_scratch"].tap()
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        // Assert — moved memo must no longer exist in the scratch list
        XCTAssertFalse(
            memoRow.waitForExistence(timeout: 3),
            "Memo '독서 목록 정리' should be removed from the scratch list after being moved to personal_todo"
        )
    }
}
