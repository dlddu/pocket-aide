import XCTest

final class RootTabsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testAffirmationsTabSelectedByDefault() throws {
        let app = XCUIApplication()
        app.launch()

        let affirmationsTab = app.tabBars.buttons["다짐"]
        XCTAssertTrue(
            affirmationsTab.waitForExistence(timeout: 15),
            "Affirmations tab should appear in tab bar"
        )
        XCTAssertTrue(affirmationsTab.isSelected, "Affirmations tab should be selected on launch")
    }

    func testAllSixTabsExist() throws {
        let app = XCUIApplication()
        app.launch()

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear")

        for label in ["채팅", "임시공간", "개인", "회사", "루틴", "다짐"] {
            XCTAssertTrue(
                tabBar.buttons[label].waitForExistence(timeout: 5),
                "Tab \(label) should exist"
            )
        }
    }
}
