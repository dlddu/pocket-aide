// NotificationUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end notification collection (알림 모음) flow:
//   Tap Notification tab → notification list view visible
//   → app group sections (카카오톡, Slack, 문자, 메일) displayed with notification counts
//   → pull-to-refresh refreshes the list
//   → section toggle collapses and expands notification rows
//   → App Group UserDefaults mock data injected via "--uitesting" and rows verified
//   → empty state view shown when no notifications present.
//
// DLD-735: 11-1: 알림 모음 — e2e 테스트 작성
// DLD-736: 11-2: 알림 모음 — 구현 및 e2e 테스트 활성화 (activated after DLD-736)
//
// NOTE: All tests are activated after DLD-736.
//   Implemented accessibilityIdentifier contracts:
//   - "tab_notification" tab item added to MainTabView
//   - NotificationScreen wired up with accessibilityIdentifier "notification_list_view"
//   - Each app group section exposes "notification_section_<appName>"
//   - Each section header exposes a badge count label "notification_count_<appName>"
//   - Each notification row exposes "notification_row_<appName>_msg_<n>"
//   - Section toggle (collapse/expand) button exposes "notification_toggle_<appName>"
//   - Pull-to-refresh is enabled on the notification list
//   - "--uitesting" launch argument seeds App Group UserDefaults (group.com.dlddu.PocketAide) with mock notification data
//   - "--uitesting-empty-notifications" launch argument seeds App Group UserDefaults with no data, exposing "notification_empty_view"

import XCTest

final class NotificationUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // and additionally seeds App Group UserDefaults (group.com.dlddu.PocketAide)
        // with mock notification data for 카카오톡, Slack, 문자, and 메일.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Notification Tab Navigation

    /// Tapping the Notification tab must display the notification list view container.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_notification" → "notification_list_view" appears
    func test_notificationTab_displaysNotificationList() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")

        // Act
        tabBar.buttons["tab_notification"].tap()

        // Assert
        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Notification list view (notification_list_view) should appear after tapping the Notification tab"
        )
    }

    // MARK: - Happy Path: App Group Sections Display

    /// All four app group sections (카카오톡, Slack, 문자, 메일) must be visible
    /// in the notification list after tapping the Notification tab.
    ///
    /// Expected flow:
    ///   "tab_notification" selected → "notification_list_view" visible
    ///   → "notification_section_카카오톡" appears
    ///   → "notification_section_Slack" appears
    ///   → "notification_section_문자" appears
    ///   → "notification_section_메일" appears
    func test_notificationTab_displaysAllAppGroupSections() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list must appear first")

        // Assert — all four app group sections must be present
        let kakaoSection = app.otherElements["notification_section_카카오톡"]
        XCTAssertTrue(
            kakaoSection.waitForExistence(timeout: 5),
            "카카오톡 section (notification_section_카카오톡) should be visible in the notification list"
        )

        let slackSection = app.otherElements["notification_section_Slack"]
        XCTAssertTrue(
            slackSection.waitForExistence(timeout: 5),
            "Slack section (notification_section_Slack) should be visible in the notification list"
        )

        let smsSection = app.otherElements["notification_section_문자"]
        XCTAssertTrue(
            smsSection.waitForExistence(timeout: 5),
            "문자 section (notification_section_문자) should be visible in the notification list"
        )

        let mailSection = app.otherElements["notification_section_메일"]
        XCTAssertTrue(
            mailSection.waitForExistence(timeout: 5),
            "메일 section (notification_section_메일) should be visible in the notification list"
        )
    }

    // MARK: - Happy Path: Notification Count per Group

    /// Each app group section header must display a notification count badge
    /// reflecting the number of notifications seeded by "--uitesting".
    ///
    /// Expected flow:
    ///   "tab_notification" selected → "notification_list_view" visible
    ///   → "notification_count_카카오톡" label exists and is non-empty
    ///   → "notification_count_Slack" label exists and is non-empty
    ///   → "notification_count_문자" label exists and is non-empty
    ///   → "notification_count_메일" label exists and is non-empty
    func test_notificationTab_displaysNotificationCountPerSection() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list must appear first")

        // Assert — each section header must expose a non-empty count badge label
        let kakaoCount = app.staticTexts["notification_count_카카오톡"]
        XCTAssertTrue(
            kakaoCount.waitForExistence(timeout: 5),
            "카카오톡 count label (notification_count_카카오톡) must be present in the section header"
        )
        XCTAssertFalse(
            kakaoCount.label.isEmpty,
            "카카오톡 count label must not be empty when notifications are seeded"
        )

        let slackCount = app.staticTexts["notification_count_Slack"]
        XCTAssertTrue(
            slackCount.waitForExistence(timeout: 5),
            "Slack count label (notification_count_Slack) must be present in the section header"
        )
        XCTAssertFalse(
            slackCount.label.isEmpty,
            "Slack count label must not be empty when notifications are seeded"
        )

        let smsCount = app.staticTexts["notification_count_문자"]
        XCTAssertTrue(
            smsCount.waitForExistence(timeout: 5),
            "문자 count label (notification_count_문자) must be present in the section header"
        )
        XCTAssertFalse(
            smsCount.label.isEmpty,
            "문자 count label must not be empty when notifications are seeded"
        )

        let mailCount = app.staticTexts["notification_count_메일"]
        XCTAssertTrue(
            mailCount.waitForExistence(timeout: 5),
            "메일 count label (notification_count_메일) must be present in the section header"
        )
        XCTAssertFalse(
            mailCount.label.isEmpty,
            "메일 count label must not be empty when notifications are seeded"
        )
    }

    // MARK: - Happy Path: Pull-to-Refresh

    /// Pulling down on the notification list must trigger a refresh and keep
    /// the notification list visible afterward.
    ///
    /// Expected flow:
    ///   "tab_notification" selected → "notification_list_view" visible
    ///   → swipe down from the top of the list
    ///   → refresh completes → "notification_list_view" remains visible
    func test_notificationTab_pullToRefresh_listRemainsVisible() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list must appear before pull-to-refresh")

        // Act — simulate pull-to-refresh by swiping down from the top of the list
        listView.swipeDown()

        // Assert — the list must still be visible after the refresh gesture completes
        XCTAssertTrue(
            listView.waitForExistence(timeout: 5),
            "Notification list (notification_list_view) must remain visible after pull-to-refresh"
        )
    }

    // MARK: - Happy Path: Section Collapse

    /// Tapping the toggle button on a section must hide the notification rows
    /// for that app group.
    ///
    /// Expected flow:
    ///   "tab_notification" selected → "notification_list_view" visible
    ///   → "notification_row_카카오톡_msg_1" visible (expanded by default)
    ///   → tap "notification_toggle_카카오톡"
    ///   → "notification_row_카카오톡_msg_1" no longer exists
    func test_notificationTab_sectionToggle_collapsesRows() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list must appear first")

        // Confirm that the first 카카오톡 row is visible before collapsing
        let firstRow = app.otherElements["notification_row_카카오톡_msg_1"]
        XCTAssertTrue(
            firstRow.waitForExistence(timeout: 5),
            "카카오톡 first row (notification_row_카카오톡_msg_1) must be visible in the expanded section"
        )

        // Act — tap the section toggle button to collapse the 카카오톡 section
        let toggleButton = app.buttons["notification_toggle_카카오톡"]
        XCTAssertTrue(
            toggleButton.waitForExistence(timeout: 5),
            "Section toggle button (notification_toggle_카카오톡) must exist in the section header"
        )
        toggleButton.tap()

        // Assert — the first row must be hidden after collapsing
        XCTAssertFalse(
            firstRow.waitForExistence(timeout: 3),
            "카카오톡 first row (notification_row_카카오톡_msg_1) should be hidden after the section is collapsed"
        )
    }

    // MARK: - Happy Path: Section Expand

    /// Tapping the toggle button twice on a section must first collapse and
    /// then re-expand the notification rows for that app group.
    ///
    /// Expected flow:
    ///   "tab_notification" selected → "notification_list_view" visible
    ///   → tap "notification_toggle_카카오톡" → rows hidden
    ///   → tap "notification_toggle_카카오톡" again → "notification_row_카카오톡_msg_1" reappears
    func test_notificationTab_sectionToggle_expandsRowsAfterCollapse() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list must appear first")

        let toggleButton = app.buttons["notification_toggle_카카오톡"]
        XCTAssertTrue(
            toggleButton.waitForExistence(timeout: 5),
            "Section toggle button (notification_toggle_카카오톡) must exist in the section header"
        )

        let firstRow = app.otherElements["notification_row_카카오톡_msg_1"]
        XCTAssertTrue(
            firstRow.waitForExistence(timeout: 5),
            "카카오톡 first row must be visible before collapsing"
        )

        // Act — collapse the section
        toggleButton.tap()
        XCTAssertFalse(
            firstRow.waitForExistence(timeout: 3),
            "카카오톡 first row should be hidden after first toggle tap"
        )

        // Act — expand the section again
        toggleButton.tap()

        // Assert — the first row must reappear after expanding
        XCTAssertTrue(
            firstRow.waitForExistence(timeout: 5),
            "카카오톡 first row (notification_row_카카오톡_msg_1) should reappear after the section is expanded again"
        )
    }

    // MARK: - Happy Path: App Group UserDefaults Data Injection

    /// Mock notification data injected via "--uitesting" into App Group UserDefaults
    /// (group.com.dlddu.PocketAide) must result in at least one row visible
    /// for each of the four app groups.
    ///
    /// Expected flow:
    ///   App launched with "--uitesting" → seeds App Group UserDefaults
    ///   → "tab_notification" selected → "notification_list_view" visible
    ///   → "notification_row_카카오톡_msg_1" exists
    ///   → "notification_row_Slack_msg_1" exists
    ///   → "notification_row_문자_msg_1" exists
    ///   → "notification_row_메일_msg_1" exists
    func test_notificationTab_userDefaultsDataInjection_allAppGroupsHaveRows() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list must appear first")

        // Assert — each app group must expose at least its first seeded notification row
        let kakaoRow = app.otherElements["notification_row_카카오톡_msg_1"]
        XCTAssertTrue(
            kakaoRow.waitForExistence(timeout: 5),
            "카카오톡 first notification row (notification_row_카카오톡_msg_1) must exist after UserDefaults data injection"
        )

        let slackRow = app.otherElements["notification_row_Slack_msg_1"]
        XCTAssertTrue(
            slackRow.waitForExistence(timeout: 5),
            "Slack first notification row (notification_row_Slack_msg_1) must exist after UserDefaults data injection"
        )

        let smsRow = app.otherElements["notification_row_문자_msg_1"]
        XCTAssertTrue(
            smsRow.waitForExistence(timeout: 5),
            "문자 first notification row (notification_row_문자_msg_1) must exist after UserDefaults data injection"
        )

        let mailRow = app.otherElements["notification_row_메일_msg_1"]
        XCTAssertTrue(
            mailRow.waitForExistence(timeout: 5),
            "메일 first notification row (notification_row_메일_msg_1) must exist after UserDefaults data injection"
        )
    }

    // MARK: - Edge Case: Empty State

    /// When no notification data is present in App Group UserDefaults,
    /// the empty state view must be displayed instead of section rows.
    ///
    /// Expected flow:
    ///   App launched with "--uitesting-empty-notifications"
    ///   → seeds App Group UserDefaults with no data
    ///   → "tab_notification" selected → "notification_list_view" visible
    ///   → "notification_empty_view" appears
    ///   → no "notification_section_카카오톡" row is visible
    func test_notificationTab_emptyNotifications_displaysEmptyView() throws {
        // Arrange — relaunch with the empty-notifications flag so no data is seeded
        // into App Group UserDefaults (group.com.dlddu.PocketAide).
        app.launchArguments = ["--uitesting", "--uitesting-empty-notifications"]
        app.launch()

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        // Act
        tabBar.buttons["tab_notification"].tap()

        let listView = app.otherElements["notification_list_view"]
        XCTAssertTrue(listView.waitForExistence(timeout: 5), "Notification list view must appear even when empty")

        // Assert — empty state view must be present
        let emptyView = app.otherElements["notification_empty_view"]
        XCTAssertTrue(
            emptyView.waitForExistence(timeout: 5),
            "Empty state view (notification_empty_view) should appear when no notifications are present"
        )

        // Assert — no app group section should be visible
        XCTAssertFalse(
            app.otherElements["notification_section_카카오톡"].waitForExistence(timeout: 3),
            "카카오톡 section should not be visible when there are no notifications"
        )
    }
}
