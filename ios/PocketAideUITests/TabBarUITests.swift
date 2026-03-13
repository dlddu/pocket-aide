// TabBarUITests.swift
// PocketAideUITests
//
// XCUITest suite that verifies the TabBar skeleton shows exactly 7 tabs and
// that each tab is individually selectable.
//
// These tests are in the TDD Red phase — they will fail until the production
// UI is implemented.
//
// Expected tab identifiers (accessibility identifiers set on each TabItem):
//   "tab_home"      — Home
//   "tab_record"    — Record / Dictation
//   "tab_history"   — History
//   "tab_widget"    — Widget
//   "tab_assistant" — AI Assistant
//   "tab_settings"  — Settings
//   "tab_profile"   — Profile

import XCTest

final class TabBarUITests: XCTestCase {

    // MARK: Properties

    private var app: XCUIApplication!

    /// All expected accessibility identifiers for the seven tabs, in display
    /// order (left → right).
    private let expectedTabIdentifiers: [String] = [
        "tab_home",
        "tab_record",
        "tab_history",
        "tab_widget",
        "tab_assistant",
        "tab_settings",
        "tab_profile"
    ]

    // MARK: Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // Pass a launch argument so the app can skip authentication/onboarding
        // and land directly on the TabBar during UI tests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path

    /// The TabBar must be visible on the initial screen.
    func test_tabBar_isVisible() {
        // Arrange — app is launched

        // Act
        let tabBar = app.tabBars.firstMatch

        // Assert
        XCTAssertTrue(
            tabBar.waitForExistence(timeout: 5),
            "TabBar should be visible on launch"
        )
    }

    /// The TabBar must contain exactly 7 tabs.
    func test_tabBar_hasSevenTabs() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act
        let tabCount = tabBar.buttons.count

        // Assert
        XCTAssertEqual(
            tabCount,
            7,
            "Expected exactly 7 tabs, found \(tabCount)"
        )
    }

    /// Every expected tab identifier must exist inside the TabBar.
    func test_tabBar_containsAllExpectedTabs() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act & Assert
        for identifier in expectedTabIdentifiers {
            let tab = tabBar.buttons[identifier]
            XCTAssertTrue(
                tab.exists,
                "Tab with identifier '\(identifier)' should exist in the TabBar"
            )
        }
    }

    // MARK: - Tab Selection

    /// Tapping the Home tab must select it (it is selected by default, but the
    /// test explicitly verifies the selected state after a tap).
    func test_homeTab_isSelectedByDefault() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act
        let homeTab = tabBar.buttons["tab_home"]

        // Assert
        XCTAssertTrue(homeTab.exists, "Home tab should exist")
        XCTAssertTrue(
            homeTab.isSelected,
            "Home tab should be selected by default"
        )
    }

    /// Tapping each tab in turn must make it the selected tab.
    func test_eachTab_canBeSelected() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act & Assert
        for identifier in expectedTabIdentifiers {
            let tab = tabBar.buttons[identifier]
            XCTAssertTrue(tab.exists, "Tab '\(identifier)' must exist before tapping")

            tab.tap()

            XCTAssertTrue(
                tab.isSelected,
                "Tab '\(identifier)' should be selected after tapping"
            )
        }
    }

    /// After tapping the Record tab, the Home tab must no longer be selected.
    func test_selectingRecordTab_deselectedHomeTab() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let homeTab   = tabBar.buttons["tab_home"]
        let recordTab = tabBar.buttons["tab_record"]

        // Act
        recordTab.tap()

        // Assert
        XCTAssertTrue(recordTab.isSelected, "Record tab should be selected")
        XCTAssertFalse(homeTab.isSelected,  "Home tab should not be selected")
    }

    // MARK: - Content Area

    /// Each tab must display its corresponding content view after selection.
    /// The content view is identified by an accessibility identifier that
    /// matches the pattern `<tab_identifier>_view`.
    func test_eachTab_displaysCorrespondingContentView() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act & Assert
        for identifier in expectedTabIdentifiers {
            tabBar.buttons[identifier].tap()

            let contentView = app.otherElements["\(identifier)_view"]
            XCTAssertTrue(
                contentView.waitForExistence(timeout: 3),
                "Content view '\(identifier)_view' should appear after selecting '\(identifier)'"
            )
        }
    }

    // MARK: - Accessibility

    /// Every tab button must have a non-empty accessibility label so VoiceOver
    /// users can identify tabs.
    func test_allTabs_haveAccessibilityLabels() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act & Assert
        for identifier in expectedTabIdentifiers {
            let tab = tabBar.buttons[identifier]
            XCTAssertTrue(tab.exists, "Tab '\(identifier)' must exist")
            XCTAssertFalse(
                tab.label.isEmpty,
                "Tab '\(identifier)' must have a non-empty accessibility label"
            )
        }
    }

    // MARK: - Edge Cases

    /// Tapping the currently selected tab again must not crash or navigate
    /// away; the tab must remain selected.
    func test_tappingSelectedTabAgain_remainsSelected() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        let homeTab = tabBar.buttons["tab_home"]

        // Act — tap the already-selected tab
        homeTab.tap()
        homeTab.tap()

        // Assert
        XCTAssertTrue(homeTab.isSelected, "Home tab should remain selected after double-tap")
        XCTAssertTrue(app.tabBars.firstMatch.exists, "TabBar should still be visible")
    }

    /// The last tab (Profile) must be reachable — i.e., it must not be hidden
    /// or truncated off-screen.
    func test_lastTab_isHittable() {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act
        let profileTab = tabBar.buttons["tab_profile"]

        // Assert
        XCTAssertTrue(profileTab.exists,    "Profile tab must exist")
        XCTAssertTrue(profileTab.isHittable, "Profile tab must be hittable (not off-screen)")
    }
}
