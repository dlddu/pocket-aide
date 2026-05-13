import XCTest

final class RootTabsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testAffirmationsContentRendersByDefault() throws {
        // Affirmations is set as the default selection in RootView.onAppear.
        // iPhone's system TabView folds tabs 5+ into a "More" tab, so we cannot
        // assert on the affirmations tab button itself. Instead, we verify that
        // the affirmations view content is rendered on launch.
        let app = XCUIApplication()
        app.launch()

        XCTAssertTrue(
            app.buttons["affirmations.add"].waitForExistence(timeout: 15),
            "Affirmations screen should render on launch (it is the default selection)"
        )
    }

    func testTabBarHasMultipleTabs() throws {
        let app = XCUIApplication()
        app.launch()

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear")
        // On iPhone with 6 tabs, system TabView shows 4 + "More". Just assert
        // that at least the first four placeholder tabs are visible.
        for label in ["채팅", "임시공간", "개인", "회사"] {
            XCTAssertTrue(
                tabBar.buttons[label].waitForExistence(timeout: 5),
                "Tab \(label) should exist in the visible tab bar"
            )
        }
    }
}
