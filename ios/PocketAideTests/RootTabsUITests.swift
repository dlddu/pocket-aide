import XCTest

final class RootTabsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testAffirmationsTabSelectedByDefault() throws {
        let app = XCUIApplication()
        app.launch()

        let affirmationsTab = app.tabBars.buttons["tab.affirmations"]
        XCTAssertTrue(
            affirmationsTab.waitForExistence(timeout: 10),
            "Affirmations tab should appear in tab bar"
        )
        XCTAssertTrue(affirmationsTab.isSelected, "Affirmations tab should be selected on launch")
    }

    func testAllSixTabsExist() throws {
        let app = XCUIApplication()
        app.launch()

        for tab in ["tab.chat", "tab.scratchpad", "tab.personal", "tab.work", "tab.routines", "tab.affirmations"] {
            XCTAssertTrue(
                app.tabBars.buttons[tab].waitForExistence(timeout: 5),
                "Tab \(tab) should exist"
            )
        }
    }
}
