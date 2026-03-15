// SentenceUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end sentence collection flow:
//   Tap Sentence tab → sentence list view visible → categories with sentences displayed
//   → add new category → add new sentence → edit sentence → long-press to delete sentence.
//
// DLD-733: 10-1: 문장 모음 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-734:
//   - A "tab_sentence" tab item is added to MainTabView
//   - SentenceListView is wired up with accessibilityIdentifier "sentence_list_view"
//   - Category sections expose identifier "sentence_category_section_<name>"
//   - Add-category sheet exposes: "add_category_button", "category_name_field",
//     "category_save_button"
//   - Add-sentence sheet exposes: "add_sentence_button", "sentence_content_field",
//     "sentence_category_picker", "sentence_save_button"
//   - Each sentence row exposes "sentence_row_<content>"
//   - Edit button per row exposes "sentence_edit_button_<content>"
//   - Edit sheet exposes: "sentence_edit_field", "sentence_update_button"
//   - Long-press on a sentence row reveals a delete context menu action

import XCTest

final class SentenceUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // consistent with the pattern used by RoutineUITests and TodoUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Sentence Tab Navigation

    /// Tapping the Sentence tab must display the sentence list view container.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_sentence" → "sentence_list_view" appears
    func test_sentenceTab_displaysSentenceList() throws {
        throw XCTSkip("DLD-734: SentenceScreen not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")

        // Act
        tabBar.buttons["tab_sentence"].tap()

        // Assert
        let listView = app.otherElements["sentence_list_view"]
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Sentence list view (sentence_list_view) should appear after tapping the Sentence tab"
        )
    }

    // MARK: - Happy Path: Category Section Display

    /// The sentence list must display sentences grouped by category section.
    ///
    /// Expected flow:
    ///   "tab_sentence" selected → "sentence_list_view" visible
    ///   → at least one "sentence_category_section_<name>" section exists
    func test_sentenceTab_displaysCategoriesWithSentences() throws {
        throw XCTSkip("DLD-734: SentenceScreen not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_sentence"].tap()

        let listView = app.otherElements["sentence_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Sentence list must appear first")

        // Add a category so at least one section exists
        let addCategoryButton = app.buttons["add_category_button"]
        XCTAssertTrue(addCategoryButton.waitForExistence(timeout: 5), "Add category button must exist")
        addCategoryButton.tap()

        let categoryNameField = app.textFields["category_name_field"]
        XCTAssertTrue(categoryNameField.waitForExistence(timeout: 5), "Category name field must appear")
        categoryNameField.tap()
        categoryNameField.typeText("인사말")

        let categorySaveButton = app.buttons["category_save_button"]
        XCTAssertTrue(categorySaveButton.waitForExistence(timeout: 5), "Category save button must appear")
        categorySaveButton.tap()

        // Assert — the newly added category section must be visible
        let categorySection = app.otherElements["sentence_category_section_인사말"]
        XCTAssertTrue(
            categorySection.waitForExistence(timeout: 5),
            "Category section (sentence_category_section_인사말) should appear in the sentence list"
        )
    }

    // MARK: - Happy Path: Add Category

    /// Adding a new category must cause it to appear as a section in the
    /// sentence list.
    ///
    /// Expected flow:
    ///   "tab_sentence" → tap "add_category_button" → sheet appears
    ///   → fill "category_name_field" with "감사 표현"
    ///   → tap "category_save_button" → sheet dismisses
    ///   → sentence list contains "sentence_category_section_감사 표현"
    func test_sentenceTab_addCategory_displaysInList() throws {
        throw XCTSkip("DLD-734: SentenceScreen not yet implemented")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_sentence"].tap()

        let listView = app.otherElements["sentence_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        let addCategoryButton = app.buttons["add_category_button"]
        XCTAssertTrue(addCategoryButton.waitForExistence(timeout: 5), "Add category button must exist")

        // Act — open the add-category sheet and fill in the form
        addCategoryButton.tap()

        let categoryNameField = app.textFields["category_name_field"]
        XCTAssertTrue(categoryNameField.waitForExistence(timeout: 5), "Category name field must appear")
        categoryNameField.tap()
        categoryNameField.typeText("감사 표현")

        let categorySaveButton = app.buttons["category_save_button"]
        XCTAssertTrue(categorySaveButton.waitForExistence(timeout: 5), "Category save button must appear")
        categorySaveButton.tap()

        // Assert — the new category section must appear in the list
        let categorySection = app.otherElements["sentence_category_section_감사 표현"]
        XCTAssertTrue(
            categorySection.waitForExistence(timeout: 5),
            "Newly added category '감사 표현' (sentence_category_section_감사 표현) should appear in the sentence list"
        )
    }

    // MARK: - Happy Path: Add Sentence

    /// Adding a new sentence with content and a selected category must cause
    /// it to appear in the corresponding category section of the sentence list.
    ///
    /// Expected flow:
    ///   "tab_sentence" → add category "인사말" → tap "add_sentence_button"
    ///   → sheet appears → fill "sentence_content_field" with "안녕하세요"
    ///   → select "인사말" in "sentence_category_picker"
    ///   → tap "sentence_save_button" → sheet dismisses
    ///   → "sentence_row_안녕하세요" appears under the "인사말" section
    func test_sentenceTab_addSentence_displaysInList() throws {
        throw XCTSkip("DLD-734: SentenceScreen not yet implemented")

        // Arrange — navigate to sentence tab and add a category first
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_sentence"].tap()

        let listView = app.otherElements["sentence_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        // Add category "인사말"
        let addCategoryButton = app.buttons["add_category_button"]
        XCTAssertTrue(addCategoryButton.waitForExistence(timeout: 5))
        addCategoryButton.tap()

        let categoryNameField = app.textFields["category_name_field"]
        XCTAssertTrue(categoryNameField.waitForExistence(timeout: 5))
        categoryNameField.tap()
        categoryNameField.typeText("인사말")

        let categorySaveButton = app.buttons["category_save_button"]
        XCTAssertTrue(categorySaveButton.waitForExistence(timeout: 5))
        categorySaveButton.tap()

        let categorySection = app.otherElements["sentence_category_section_인사말"]
        XCTAssertTrue(categorySection.waitForExistence(timeout: 5), "Category section must appear before adding sentence")

        // Act — open the add-sentence sheet and fill in the form
        let addSentenceButton = app.buttons["add_sentence_button"]
        XCTAssertTrue(addSentenceButton.waitForExistence(timeout: 5), "Add sentence button must exist")
        addSentenceButton.tap()

        let contentField = app.textFields["sentence_content_field"]
        XCTAssertTrue(contentField.waitForExistence(timeout: 5), "Sentence content field must appear")
        contentField.tap()
        contentField.typeText("안녕하세요")

        let categoryPicker = app.pickers["sentence_category_picker"]
        XCTAssertTrue(categoryPicker.waitForExistence(timeout: 5), "Category picker must appear")
        categoryPicker.pickerWheels.firstMatch.adjust(toPickerWheelValue: "인사말")

        let sentenceSaveButton = app.buttons["sentence_save_button"]
        XCTAssertTrue(sentenceSaveButton.waitForExistence(timeout: 5), "Sentence save button must appear")
        sentenceSaveButton.tap()

        // Assert — the new sentence must appear in the "인사말" category section
        let sentenceRow = categorySection.otherElements["sentence_row_안녕하세요"]
        XCTAssertTrue(
            sentenceRow.waitForExistence(timeout: 5),
            "Newly added sentence '안녕하세요' (sentence_row_안녕하세요) should appear under the '인사말' category section"
        )
    }

    // MARK: - Happy Path: Edit Sentence

    /// Editing a sentence via the edit button must update the displayed
    /// content in the sentence list.
    ///
    /// Expected flow:
    ///   Add category "인사말" → add sentence "안녕하세요"
    ///   → tap "sentence_edit_button_안녕하세요" → edit sheet appears
    ///   → clear "sentence_edit_field" and type "안녕히 가세요"
    ///   → tap "sentence_update_button" → sheet dismisses
    ///   → "sentence_row_안녕하세요" is gone
    ///   → "sentence_row_안녕히 가세요" appears in the list
    func test_sentenceTab_editSentence_updatesInList() throws {
        throw XCTSkip("DLD-734: SentenceScreen not yet implemented")

        // Arrange — navigate to sentence tab, add category and sentence
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_sentence"].tap()

        let listView = app.otherElements["sentence_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        // Add category
        let addCategoryButton = app.buttons["add_category_button"]
        XCTAssertTrue(addCategoryButton.waitForExistence(timeout: 5))
        addCategoryButton.tap()

        let categoryNameField = app.textFields["category_name_field"]
        XCTAssertTrue(categoryNameField.waitForExistence(timeout: 5))
        categoryNameField.tap()
        categoryNameField.typeText("인사말")

        let categorySaveButton = app.buttons["category_save_button"]
        XCTAssertTrue(categorySaveButton.waitForExistence(timeout: 5))
        categorySaveButton.tap()

        XCTAssertTrue(app.otherElements["sentence_category_section_인사말"].waitForExistence(timeout: 5))

        // Add sentence
        let addSentenceButton = app.buttons["add_sentence_button"]
        XCTAssertTrue(addSentenceButton.waitForExistence(timeout: 5))
        addSentenceButton.tap()

        let contentField = app.textFields["sentence_content_field"]
        XCTAssertTrue(contentField.waitForExistence(timeout: 5))
        contentField.tap()
        contentField.typeText("안녕하세요")

        let categoryPicker = app.pickers["sentence_category_picker"]
        XCTAssertTrue(categoryPicker.waitForExistence(timeout: 5))
        categoryPicker.pickerWheels.firstMatch.adjust(toPickerWheelValue: "인사말")

        let sentenceSaveButton = app.buttons["sentence_save_button"]
        XCTAssertTrue(sentenceSaveButton.waitForExistence(timeout: 5))
        sentenceSaveButton.tap()

        let sentenceRow = app.otherElements["sentence_row_안녕하세요"]
        XCTAssertTrue(sentenceRow.waitForExistence(timeout: 5), "Sentence '안녕하세요' must appear before editing")

        // Act — tap the edit button for "안녕하세요"
        let editButton = app.buttons["sentence_edit_button_안녕하세요"]
        XCTAssertTrue(
            editButton.waitForExistence(timeout: 5),
            "Edit button (sentence_edit_button_안녕하세요) must exist in the sentence row"
        )
        editButton.tap()

        let editField = app.textFields["sentence_edit_field"]
        XCTAssertTrue(editField.waitForExistence(timeout: 5), "Edit field must appear")
        editField.tap()
        // Clear existing text and type the new content
        editField.press(forDuration: 1.0)
        app.menuItems["Select All"].tap()
        editField.typeText("안녕히 가세요")

        let updateButton = app.buttons["sentence_update_button"]
        XCTAssertTrue(updateButton.waitForExistence(timeout: 5), "Update button must appear")
        updateButton.tap()

        // Assert — the old sentence row must be gone
        XCTAssertFalse(
            sentenceRow.waitForExistence(timeout: 3),
            "Old sentence '안녕하세요' should no longer appear after edit"
        )

        // Assert — the updated sentence row must appear
        let updatedRow = app.otherElements["sentence_row_안녕히 가세요"]
        XCTAssertTrue(
            updatedRow.waitForExistence(timeout: 5),
            "Updated sentence '안녕히 가세요' (sentence_row_안녕히 가세요) should appear after editing"
        )
    }

    // MARK: - Edge Case: Long-Press Delete Sentence

    /// Long-pressing a sentence row must reveal a delete context menu action,
    /// and after confirming deletion the sentence must be removed from the list.
    ///
    /// Expected flow:
    ///   Add category "인사말" → add sentence "반갑습니다"
    ///   → long-press "sentence_row_반갑습니다" → context menu appears with "삭제" action
    ///   → tap "삭제" → confirmation alert appears → confirm deletion
    ///   → "sentence_row_반갑습니다" is removed from the list
    func test_sentenceTab_longPressDeleteSentence_removedFromList() throws {
        throw XCTSkip("DLD-734: SentenceScreen not yet implemented")

        // Arrange — navigate to sentence tab, add category and sentence
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_sentence"].tap()

        let listView = app.otherElements["sentence_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5))

        // Add category
        let addCategoryButton = app.buttons["add_category_button"]
        XCTAssertTrue(addCategoryButton.waitForExistence(timeout: 5))
        addCategoryButton.tap()

        let categoryNameField = app.textFields["category_name_field"]
        XCTAssertTrue(categoryNameField.waitForExistence(timeout: 5))
        categoryNameField.tap()
        categoryNameField.typeText("인사말")

        let categorySaveButton = app.buttons["category_save_button"]
        XCTAssertTrue(categorySaveButton.waitForExistence(timeout: 5))
        categorySaveButton.tap()

        XCTAssertTrue(app.otherElements["sentence_category_section_인사말"].waitForExistence(timeout: 5))

        // Add sentence
        let addSentenceButton = app.buttons["add_sentence_button"]
        XCTAssertTrue(addSentenceButton.waitForExistence(timeout: 5))
        addSentenceButton.tap()

        let contentField = app.textFields["sentence_content_field"]
        XCTAssertTrue(contentField.waitForExistence(timeout: 5))
        contentField.tap()
        contentField.typeText("반갑습니다")

        let categoryPicker = app.pickers["sentence_category_picker"]
        XCTAssertTrue(categoryPicker.waitForExistence(timeout: 5))
        categoryPicker.pickerWheels.firstMatch.adjust(toPickerWheelValue: "인사말")

        let sentenceSaveButton = app.buttons["sentence_save_button"]
        XCTAssertTrue(sentenceSaveButton.waitForExistence(timeout: 5))
        sentenceSaveButton.tap()

        let sentenceRow = app.otherElements["sentence_row_반갑습니다"]
        XCTAssertTrue(
            sentenceRow.waitForExistence(timeout: 5),
            "Sentence '반갑습니다' must appear before long-press delete"
        )

        // Act — long-press to reveal the delete context menu
        sentenceRow.press(forDuration: 1.0)

        let deleteAction = app.buttons["삭제"]
        XCTAssertTrue(
            deleteAction.waitForExistence(timeout: 5),
            "Delete context menu action ('삭제') should appear after long-pressing the sentence row"
        )
        deleteAction.tap()

        // If a confirmation alert appears, confirm the deletion
        let confirmButton = app.alerts.buttons["삭제"]
        if confirmButton.waitForExistence(timeout: 3) {
            confirmButton.tap()
        }

        // Assert — the sentence row must be removed from the list
        XCTAssertFalse(
            sentenceRow.waitForExistence(timeout: 3),
            "Deleted sentence '반갑습니다' should be removed from the sentence list after long-press delete"
        )
    }
}
