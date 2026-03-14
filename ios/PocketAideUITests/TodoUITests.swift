// TodoUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end personal todo flow:
//   Tap Todo tab → list view visible → sections "진행중" / "완료"
//   → add new todo (title) → appears in list
//   → tap checkbox → todo moves to completed section
//   → swipe to delete → todo removed from list.
//
// DLD-725: 6-1: 개인 투두 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-725:
//   - A "tab_todo" tab item is added to MainTabView
//   - TodoListView is wired up with accessibilityIdentifier "todo_list_view"
//   - Section headers expose identifiers "todo_section_pending" / "todo_section_completed"
//   - Add-todo button exposes: "add_todo_button"
//   - Todo title input field exposes: "todo_title_field"
//   - Save button exposes: "todo_save_button"
//   - Each todo row exposes "todo_row_<title>" and a checkbox button "todo_checkbox_<title>"
//   - Swipe-to-delete on a todo row removes it from the list

import XCTest

final class TodoUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // consistent with the pattern used by RoutineUITests and ChatUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Todo Tab Navigation

    /// Tapping the Todo tab must display the todo list view container.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_todo" → "todo_list_view" appears
    func test_todoTab_displaysTodoList() throws {

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")

        // Act
        tabBar.buttons["tab_todo"].tap()

        // Assert
        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Todo list view (todo_list_view) should appear after tapping the Todo tab"
        )
    }

    // MARK: - Happy Path: Section Display

    /// The todo list must show a "진행중" (pending) section for todos that
    /// have not yet been completed.
    ///
    /// Expected flow:
    ///   "tab_todo" selected → "todo_list_view" visible
    ///   → "todo_section_pending" header exists
    func test_todoTab_displaysPendingSection() throws {

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Todo list must appear first")

        // Assert
        let pendingSection = app.otherElements["todo_section_pending"]
        XCTAssertTrue(
            pendingSection.waitForExistence(timeout: 5),
            "Pending section (todo_section_pending / '진행중') should be visible in the todo list"
        )
    }

    /// The todo list must show a "완료" (completed) section for todos that
    /// have been checked off.
    ///
    /// Expected flow:
    ///   "tab_todo" selected → "todo_list_view" visible
    ///   → "todo_section_completed" header exists
    func test_todoTab_displaysCompletedSection() throws {

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Todo list must appear first")

        // Assert
        let completedSection = app.otherElements["todo_section_completed"]
        XCTAssertTrue(
            completedSection.waitForExistence(timeout: 5),
            "Completed section (todo_section_completed / '완료') should be visible in the todo list"
        )
    }

    // MARK: - Happy Path: Add New Todo

    /// Adding a new todo with a title must cause it to appear in the pending
    /// section of the todo list.
    ///
    /// Expected flow:
    ///   "tab_todo" → tap "add_todo_button" → input sheet appears
    ///   → fill "todo_title_field" with "장보기"
    ///   → tap "todo_save_button" → sheet dismisses
    ///   → todo list contains a row for "장보기" in the pending section
    func test_todoTab_addNewTodo_displaysInList() throws {

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_todo"].tap()

        let listView = app.otherElements["todo_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_todo_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add todo button must exist")

        // Act — open the add-todo input and fill in the title
        addButton.tap()

        let titleField = app.textFields["todo_title_field"]
        XCTAssertTrue(titleField.waitForExistence(timeout: 5), "Todo title field must appear")
        titleField.tap()
        titleField.typeText("장보기")

        let saveButton = app.buttons["todo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5), "Save button must appear")
        saveButton.tap()

        // Assert — the new todo must appear in the pending section
        let pendingSection = app.otherElements["todo_section_pending"]
        XCTAssertTrue(pendingSection.waitForExistence(timeout: 5), "Pending section must be visible")

        let todoRow = pendingSection.otherElements["todo_row_장보기"]
        XCTAssertTrue(
            todoRow.waitForExistence(timeout: 5),
            "Newly added todo '장보기' (todo_row_장보기) should appear in the pending section"
        )
    }

    // MARK: - Happy Path: Checkbox Toggle

    /// Tapping the checkbox on a pending todo must move it to the completed
    /// section.
    ///
    /// Expected flow:
    ///   Add todo "독서"
    ///   → "todo_row_독서" appears in pending section
    ///   → tap "todo_checkbox_독서"
    ///   → "todo_row_독서" disappears from pending section
    ///   → "todo_row_독서" appears in completed section
    func test_todoTab_tapCheckbox_movesToCompletedSection() throws {

        // Arrange — navigate to todo tab and add a pending todo
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
        titleField.typeText("독서")

        let saveButton = app.buttons["todo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        // Verify the todo starts in the pending section
        let pendingSection = app.otherElements["todo_section_pending"]
        XCTAssertTrue(pendingSection.waitForExistence(timeout: 5))

        let todoRowInPending = pendingSection.otherElements["todo_row_독서"]
        XCTAssertTrue(
            todoRowInPending.waitForExistence(timeout: 5),
            "Todo '독서' should start in the pending section"
        )

        // Act — tap the checkbox to complete the todo
        let checkbox = app.buttons["todo_checkbox_독서"]
        XCTAssertTrue(
            checkbox.waitForExistence(timeout: 5),
            "Checkbox button (todo_checkbox_독서) must exist in the todo row"
        )
        checkbox.tap()

        // Assert — todo must leave the pending section
        XCTAssertFalse(
            todoRowInPending.waitForExistence(timeout: 3),
            "Todo '독서' should no longer be in the pending section after checkbox tap"
        )

        // Assert — todo must appear in the completed section
        let completedSection = app.otherElements["todo_section_completed"]
        XCTAssertTrue(completedSection.waitForExistence(timeout: 5), "Completed section must be visible")

        let todoRowInCompleted = completedSection.otherElements["todo_row_독서"]
        XCTAssertTrue(
            todoRowInCompleted.waitForExistence(timeout: 5),
            "Todo '독서' should move to the completed section ('완료') after checkbox tap"
        )
    }

    // MARK: - Swipe to Delete

    /// Swiping to delete a todo row must remove it from the list entirely.
    ///
    /// Expected flow:
    ///   Add todo "운동하기"
    ///   → swipe left on "todo_row_운동하기" → delete action appears
    ///   → tap delete action
    ///   → "todo_row_운동하기" no longer exists in the list
    func test_todoTab_swipeToDelete_removesFromList() throws {

        // Arrange — navigate to todo tab and add a todo
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
        titleField.typeText("운동하기")

        let saveButton = app.buttons["todo_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        let todoRow = app.otherElements["todo_row_운동하기"]
        XCTAssertTrue(
            todoRow.waitForExistence(timeout: 5),
            "Todo row '운동하기' must appear before swipe-to-delete"
        )

        // Act — swipe left to reveal the delete action
        todoRow.swipeLeft()

        let deleteButton = app.buttons["Delete"]
        XCTAssertTrue(
            deleteButton.waitForExistence(timeout: 5),
            "Delete button should appear after left swipe on the todo row"
        )
        deleteButton.tap()

        // Assert — todo must no longer exist in the list
        XCTAssertFalse(
            todoRow.waitForExistence(timeout: 3),
            "Todo '운동하기' should be removed from the list after swipe-to-delete"
        )
    }
}
