// RoutineUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end routine management flow:
//   Tap Routine tab → list view visible → sections "곧 해야 할 것" / "여유 있음"
//   → add new routine (name, period, last done date) → appears in list
//   → left swipe to complete → D-day updated → routine moves to relaxed section.
//
// DLD-723: 5-1: 루틴 관리 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-723:
//   - A "tab_routine" tab item is added to MainTabView
//   - RoutineListView is wired up with accessibilityIdentifier "routine_list_view"
//   - Section headers expose identifiers "routine_section_urgent" / "routine_section_relaxed"
//   - Add-routine sheet exposes: "add_routine_button", "routine_name_field",
//     "routine_interval_field", "routine_last_done_field", "routine_save_button"
//   - Each routine row exposes "routine_row_<name>" and "routine_dday_label"
//   - Left swipe on a routine row reveals a "complete_button" action

import XCTest

final class RoutineUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // consistent with the pattern used by ChatUITests and VoiceChatUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Routine Tab Navigation

    /// Tapping the Routine tab must display the routine list view container.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_routine" → "routine_list_view" appears
    func test_routineTab_displaysRoutineList() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")

        // Act
        tabBar.buttons["tab_routine"].tap()

        // Assert
        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Routine list view (routine_list_view) should appear after tapping the Routine tab"
        )
    }

    // MARK: - Happy Path: Section Display

    /// The routine list must show a "곧 해야 할 것" (urgent) section for
    /// routines whose D-day is approaching or overdue.
    ///
    /// Expected flow:
    ///   "tab_routine" selected → "routine_list_view" visible
    ///   → "routine_section_urgent" header exists
    func test_routineTab_displaysUrgentSection() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Routine list must appear first")

        // Assert
        let urgentSection = app.otherElements["routine_section_urgent"]
        XCTAssertTrue(
            urgentSection.waitForExistence(timeout: 5),
            "Urgent section (routine_section_urgent / '곧 해야 할 것') should be visible in the routine list"
        )
    }

    /// The routine list must show a "여유 있음" (relaxed) section for routines
    /// that have been completed recently and whose next due date is still far.
    ///
    /// Expected flow:
    ///   "tab_routine" selected → "routine_list_view" visible
    ///   → "routine_section_relaxed" header exists
    func test_routineTab_displaysRelaxedSection() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Routine list must appear first")

        // Assert
        let relaxedSection = app.otherElements["routine_section_relaxed"]
        XCTAssertTrue(
            relaxedSection.waitForExistence(timeout: 5),
            "Relaxed section (routine_section_relaxed / '여유 있음') should be visible in the routine list"
        )
    }

    // MARK: - Happy Path: Add New Routine

    /// Adding a new routine with a name, interval (주기), and last-done date
    /// must cause it to appear in the routine list.
    ///
    /// Expected flow:
    ///   "tab_routine" → tap "add_routine_button" → sheet appears
    ///   → fill "routine_name_field" with "샤워"
    ///   → fill "routine_interval_field" with "1"
    ///   → fill "routine_last_done_field" with "2026-03-13"
    ///   → tap "routine_save_button" → sheet dismisses
    ///   → routine list contains a row for "샤워"
    func test_routineTab_addNewRoutine_displaysInList() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_routine_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add routine button must exist")

        // Act — open the add-routine sheet and fill in the form
        addButton.tap()

        let nameField = app.textFields["routine_name_field"]
        XCTAssertTrue(nameField.waitForExistence(timeout: 5), "Routine name field must appear")
        nameField.tap()
        nameField.typeText("샤워")

        let intervalField = app.textFields["routine_interval_field"]
        XCTAssertTrue(intervalField.waitForExistence(timeout: 5), "Routine interval field must appear")
        intervalField.tap()
        intervalField.typeText("1")

        let lastDoneField = app.textFields["routine_last_done_field"]
        XCTAssertTrue(lastDoneField.waitForExistence(timeout: 5), "Last done date field must appear")
        lastDoneField.tap()
        lastDoneField.typeText("2026-03-13")

        let saveButton = app.buttons["routine_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5), "Save button must appear")
        saveButton.tap()

        // Assert — the new routine must appear in the list
        let routineRow = app.otherElements["routine_row_샤워"]
        XCTAssertTrue(
            routineRow.waitForExistence(timeout: 5),
            "Newly added routine '샤워' (routine_row_샤워) should appear in the routine list"
        )
    }

    /// After adding a new routine, the routine row must display the routine's
    /// name and its interval period.
    ///
    /// Expected flow:
    ///   Add routine "스트레칭" with interval 3
    ///   → routine row shows static text "스트레칭" and period indicator "3일"
    func test_routineTab_addRoutine_nameAndPeriodVisible() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_routine_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let nameField = app.textFields["routine_name_field"]
        XCTAssertTrue(nameField.waitForExistence(timeout: 5))
        nameField.tap()
        nameField.typeText("스트레칭")

        let intervalField = app.textFields["routine_interval_field"]
        XCTAssertTrue(intervalField.waitForExistence(timeout: 5))
        intervalField.tap()
        intervalField.typeText("3")

        let lastDoneField = app.textFields["routine_last_done_field"]
        XCTAssertTrue(lastDoneField.waitForExistence(timeout: 5))
        lastDoneField.tap()
        lastDoneField.typeText("2026-03-11")

        let saveButton = app.buttons["routine_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        // Assert — routine name and period must be visible in the list
        let routineRow = app.otherElements["routine_row_스트레칭"]
        XCTAssertTrue(routineRow.waitForExistence(timeout: 5), "Routine row for '스트레칭' must appear")

        XCTAssertTrue(
            app.staticTexts["스트레칭"].exists,
            "Routine name '스트레칭' should be visible in the list row"
        )
        // Period is displayed as "3일" (3 days)
        XCTAssertTrue(
            app.staticTexts["3일"].exists,
            "Routine interval '3일' should be visible in the list row"
        )
    }

    /// After adding a new routine, the routine row must display a D-day label
    /// so the user can see how many days remain until the next due date.
    ///
    /// Expected flow:
    ///   Add routine "세탁" with interval=7, last_done="2026-03-07"
    ///   → today is 2026-03-14, so next_due_date = 2026-03-14, d_day = D-0
    ///   → routine row shows "routine_dday_label" with text "D-0" or "D+0"
    func test_routineTab_addRoutine_dDayDisplayed() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_routine_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let nameField = app.textFields["routine_name_field"]
        XCTAssertTrue(nameField.waitForExistence(timeout: 5))
        nameField.tap()
        nameField.typeText("세탁")

        let intervalField = app.textFields["routine_interval_field"]
        XCTAssertTrue(intervalField.waitForExistence(timeout: 5))
        intervalField.tap()
        intervalField.typeText("7")

        let lastDoneField = app.textFields["routine_last_done_field"]
        XCTAssertTrue(lastDoneField.waitForExistence(timeout: 5))
        lastDoneField.tap()
        lastDoneField.typeText("2026-03-07")

        let saveButton = app.buttons["routine_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        // Assert — the D-day label must appear in the routine row
        let routineRow = app.otherElements["routine_row_세탁"]
        XCTAssertTrue(routineRow.waitForExistence(timeout: 5), "Routine row '세탁' must appear")

        let dDayLabel = routineRow.staticTexts["routine_dday_label"]
        XCTAssertTrue(
            dDayLabel.waitForExistence(timeout: 5),
            "D-day label (routine_dday_label) should be visible inside the routine row"
        )
        XCTAssertFalse(
            dDayLabel.label.isEmpty,
            "D-day label should contain a non-empty value (e.g. 'D-0')"
        )
    }

    // MARK: - Happy Path: Left Swipe to Complete

    /// Left-swiping a routine row must reveal a complete action and, after
    /// tapping it, the routine's D-day label must update to reflect that the
    /// routine was just completed (d_day resets to interval_days).
    ///
    /// Expected flow:
    ///   Routine "양치" exists with overdue D-day
    ///   → swipe left on "routine_row_양치" → "complete_button" appears
    ///   → tap "complete_button"
    ///   → D-day label updates (e.g. becomes "D-1" for interval=1)
    func test_routineTab_swipeLeftToComplete_updatesDDay() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange — navigate to routine tab and add a routine with a past due date
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_routine_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let nameField = app.textFields["routine_name_field"]
        XCTAssertTrue(nameField.waitForExistence(timeout: 5))
        nameField.tap()
        nameField.typeText("양치")

        let intervalField = app.textFields["routine_interval_field"]
        XCTAssertTrue(intervalField.waitForExistence(timeout: 5))
        intervalField.tap()
        intervalField.typeText("1")

        // Use a date far in the past so the routine is overdue and in the urgent section
        let lastDoneField = app.textFields["routine_last_done_field"]
        XCTAssertTrue(lastDoneField.waitForExistence(timeout: 5))
        lastDoneField.tap()
        lastDoneField.typeText("2026-03-01")

        let saveButton = app.buttons["routine_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        let routineRow = app.otherElements["routine_row_양치"]
        XCTAssertTrue(routineRow.waitForExistence(timeout: 5), "Routine row '양치' must appear")

        // Capture the D-day label text before completing
        let dDayLabel = routineRow.staticTexts["routine_dday_label"]
        XCTAssertTrue(dDayLabel.waitForExistence(timeout: 5), "D-day label must exist before swipe")
        let dDayBefore = dDayLabel.label

        // Act — left swipe to reveal the complete action
        routineRow.swipeLeft()

        let completeButton = app.buttons["complete_button"]
        XCTAssertTrue(
            completeButton.waitForExistence(timeout: 5),
            "Complete button (complete_button) should appear after left swipe"
        )
        completeButton.tap()

        // Assert — D-day label must update
        XCTAssertTrue(dDayLabel.waitForExistence(timeout: 5))
        let dDayAfter = dDayLabel.label
        XCTAssertNotEqual(
            dDayAfter,
            dDayBefore,
            "D-day label should update after completing the routine (before: \(dDayBefore), after: \(dDayAfter))"
        )
    }

    /// After completing a routine via left swipe, the routine must move from
    /// the "곧 해야 할 것" section to the "여유 있음" section.
    ///
    /// Expected flow:
    ///   Overdue routine "세수" is in "routine_section_urgent"
    ///   → swipe left → tap "complete_button"
    ///   → "세수" row disappears from urgent section
    ///   → "세수" row appears in "routine_section_relaxed"
    func test_routineTab_swipeLeftToComplete_movesRoutineToRelaxedSection() throws {
        throw XCTSkip("DLD-723: Routine tab not yet implemented")

        // Arrange — navigate to routine tab and add a routine with a past due date
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_routine"].tap()

        let listView = app.otherElements["routine_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addButton = app.buttons["add_routine_button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5))
        addButton.tap()

        let nameField = app.textFields["routine_name_field"]
        XCTAssertTrue(nameField.waitForExistence(timeout: 5))
        nameField.tap()
        nameField.typeText("세수")

        let intervalField = app.textFields["routine_interval_field"]
        XCTAssertTrue(intervalField.waitForExistence(timeout: 5))
        intervalField.tap()
        intervalField.typeText("14")

        // Past due date so the routine starts in the urgent section
        let lastDoneField = app.textFields["routine_last_done_field"]
        XCTAssertTrue(lastDoneField.waitForExistence(timeout: 5))
        lastDoneField.tap()
        lastDoneField.typeText("2026-02-01")

        let saveButton = app.buttons["routine_save_button"]
        XCTAssertTrue(saveButton.waitForExistence(timeout: 5))
        saveButton.tap()

        // Verify the routine starts in the urgent section
        let urgentSection = app.otherElements["routine_section_urgent"]
        XCTAssertTrue(urgentSection.waitForExistence(timeout: 5), "Urgent section must be visible")

        let routineRowInUrgent = urgentSection.otherElements["routine_row_세수"]
        XCTAssertTrue(
            routineRowInUrgent.waitForExistence(timeout: 5),
            "Overdue routine '세수' should start in the urgent section"
        )

        // Act — swipe left and complete
        routineRowInUrgent.swipeLeft()

        let completeButton = app.buttons["complete_button"]
        XCTAssertTrue(
            completeButton.waitForExistence(timeout: 5),
            "Complete button (complete_button) should appear after left swipe"
        )
        completeButton.tap()

        // Assert — routine must leave the urgent section
        XCTAssertFalse(
            routineRowInUrgent.waitForExistence(timeout: 5),
            "Completed routine '세수' should no longer be in the urgent section"
        )

        // Assert — routine must appear in the relaxed section
        let relaxedSection = app.otherElements["routine_section_relaxed"]
        XCTAssertTrue(relaxedSection.waitForExistence(timeout: 5), "Relaxed section must be visible")

        let routineRowInRelaxed = relaxedSection.otherElements["routine_row_세수"]
        XCTAssertTrue(
            routineRowInRelaxed.waitForExistence(timeout: 5),
            "Completed routine '세수' should move to the relaxed section ('여유 있음')"
        )
    }
}
