// LargeWidgetUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end iOS Large Widget flow:
//   WidgetKit Extension (PocketAideWidget) 빌드 성공 확인
//   → Timeline Provider 테스트 데이터 주입 → 5개 섹션 렌더링 확인
//     (캘린더 / 날씨 / 메일 / 문장 / 알림)
//   → 위젯 섹션 딥링크 URL(pocketaide://tab/{section}) → 앱 해당 탭 진입 확인
//
// All tests are active. WidgetKit Extension 구현 완료 (DLD-738):
//   - PocketAideWidget Extension 타겟 생성 (bundle ID: com.dlddu.PocketAide.PocketAideWidget)
//   - Widget Kind: PocketAideLargeWidget
//   - Timeline Provider: PocketAideLargeTimelineProvider
//   - Widget Entry: PocketAideLargeEntry (sections: calendar, weather, mail, sentence, notification)
//   - 딥링크 URL 스킴: pocketaide://tab/{section} 등록 (Info.plist URL Types)
//   - "widget_section_calendar" accessibilityIdentifier 노출
//   - "widget_section_weather" accessibilityIdentifier 노출
//   - "widget_section_mail" accessibilityIdentifier 노출
//   - "widget_section_sentence" accessibilityIdentifier 노출
//   - "widget_section_notification" accessibilityIdentifier 노출
//   - "--uitesting-widget" launch argument: Timeline Provider 테스트 데이터 주입 활성화
//   - 각 탭 accessibilityIdentifier: "tab_calendar", "tab_weather", "tab_mail",
//     "tab_sentence", "tab_notification"

import XCTest

final class LargeWidgetUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView.
        // "--uitesting-widget" additionally injects test data into
        // PocketAideLargeTimelineProvider so that all 5 widget sections
        // can be driven deterministically.
        app.launchArguments = ["--uitesting", "--uitesting-widget"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - WidgetKit Extension Build

    /// WidgetKit Extension 타겟(PocketAideWidget)이 성공적으로 빌드되어
    /// 앱과 함께 번들링되어야 한다.
    ///
    /// Expected flow:
    ///   앱 번들 내 PlugIns 디렉토리에 PocketAideWidget.appex 존재
    ///   → 위젯 설정 화면("widget_settings_view")이 정상 표시됨
    func test_widgetExtension_buildsSuccessfully() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating to widget settings")

        // Act — navigate to widget settings tab where the extension is configured
        let widgetSettingsTab = tabBar.buttons["tab_widget_settings"]
        XCTAssertTrue(
            widgetSettingsTab.waitForExistence(timeout: 5),
            "Widget settings tab (tab_widget_settings) must be accessible"
        )
        widgetSettingsTab.tap()

        // Assert — widget settings view must appear, confirming extension is bundled
        let widgetSettingsView = app.otherElements["widget_settings_view"]
        XCTAssertTrue(
            widgetSettingsView.waitForExistence(timeout: 5),
            "Widget settings view (widget_settings_view) should appear, confirming PocketAideWidget extension is bundled"
        )
    }

    // MARK: - Snapshot: Timeline Provider Data Injection

    /// Timeline Provider(PocketAideLargeTimelineProvider)에 테스트 데이터를
    /// 주입했을 때 Large 위젯의 5개 섹션이 모두 렌더링되어야 한다.
    ///
    /// Expected flow:
    ///   "--uitesting-widget" 실행 인수로 Timeline Provider에 테스트 데이터 주입
    ///   → PocketAideLargeEntry의 각 섹션이 위젯 프리뷰 화면에 표시됨
    ///   → "widget_section_calendar" 존재
    ///   → "widget_section_weather" 존재
    ///   → "widget_section_mail" 존재
    ///   → "widget_section_sentence" 존재
    ///   → "widget_section_notification" 존재
    func test_widgetSnapshot_timelineProviderDataInjection_rendersAllFiveSections() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating to widget preview")

        // Navigate to widget settings / preview screen
        let widgetSettingsTab = tabBar.buttons["tab_widget_settings"]
        XCTAssertTrue(
            widgetSettingsTab.waitForExistence(timeout: 5),
            "Widget settings tab (tab_widget_settings) must be accessible"
        )
        widgetSettingsTab.tap()

        let widgetSettingsView = app.otherElements["widget_settings_view"]
        XCTAssertTrue(
            widgetSettingsView.waitForExistence(timeout: 5),
            "Widget settings view must appear before checking section rendering"
        )

        // Assert — all five widget sections must be rendered with injected test data
        let calendarSection = app.otherElements["widget_section_calendar"]
        XCTAssertTrue(
            calendarSection.waitForExistence(timeout: 5),
            "Calendar section (widget_section_calendar) should be rendered in the Large Widget preview"
        )

        let weatherSection = app.otherElements["widget_section_weather"]
        XCTAssertTrue(
            weatherSection.waitForExistence(timeout: 5),
            "Weather section (widget_section_weather) should be rendered in the Large Widget preview"
        )

        let mailSection = app.otherElements["widget_section_mail"]
        XCTAssertTrue(
            mailSection.waitForExistence(timeout: 5),
            "Mail section (widget_section_mail) should be rendered in the Large Widget preview"
        )

        let sentenceSection = app.otherElements["widget_section_sentence"]
        XCTAssertTrue(
            sentenceSection.waitForExistence(timeout: 5),
            "Sentence section (widget_section_sentence) should be rendered in the Large Widget preview"
        )

        let notificationSection = app.otherElements["widget_section_notification"]
        XCTAssertTrue(
            notificationSection.waitForExistence(timeout: 5),
            "Notification section (widget_section_notification) should be rendered in the Large Widget preview"
        )
    }

    // MARK: - Deeplink: Widget Section URL → App Tab

    /// 위젯의 각 섹션 딥링크 URL(pocketaide://tab/{section})을 열었을 때
    /// 앱의 해당 탭으로 정확히 진입해야 한다.
    ///
    /// Expected flow:
    ///   pocketaide://tab/calendar 열기 → "tab_calendar" 탭 활성화
    ///   pocketaide://tab/weather  열기 → "tab_weather" 탭 활성화
    ///   pocketaide://tab/mail     열기 → "tab_mail" 탭 활성화
    ///   pocketaide://tab/sentence 열기 → "tab_sentence" 탭 활성화
    ///   pocketaide://tab/notification 열기 → "tab_notification" 탭 활성화
    func test_widgetDeeplink_sectionURL_navigatesToCorrectTab() throws {
        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before testing deeplinks")

        // Define the mapping from deeplink URL path segment to expected tab identifier
        let deeplinkSections: [(urlSection: String, expectedTabIdentifier: String)] = [
            ("calendar",     "tab_calendar"),
            ("weather",      "tab_weather"),
            ("mail",         "tab_mail"),
            ("sentence",     "tab_sentence"),
            ("notification", "tab_notification"),
        ]

        for (urlSection, expectedTabIdentifier) in deeplinkSections {
            // Act — open the widget section deeplink URL via Safari / URL scheme
            let deeplinkURL = URL(string: "pocketaide://tab/\(urlSection)")!
            app.open(deeplinkURL)

            // Assert — the corresponding tab must become the selected tab
            let targetTab = tabBar.buttons[expectedTabIdentifier]
            XCTAssertTrue(
                targetTab.waitForExistence(timeout: 5),
                "Tab (\(expectedTabIdentifier)) should exist after opening deeplink pocketaide://tab/\(urlSection)"
            )
            XCTAssertTrue(
                targetTab.isSelected,
                "Tab (\(expectedTabIdentifier)) should be selected after opening deeplink pocketaide://tab/\(urlSection)"
            )
        }
    }
}
