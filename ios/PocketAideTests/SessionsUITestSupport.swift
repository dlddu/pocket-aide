import XCTest

struct SessionsScreen {
    let app: XCUIApplication

    static func open(in app: XCUIApplication) -> SessionsScreen {
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = tabBar.buttons["세션"]
        XCTAssertTrue(tab.waitForExistence(timeout: 5), "The 세션 tab should sit in the bottom tab bar, not under More")
        tab.tap()
        let screen = SessionsScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "The 에이전트 세션 screen should open")
        return screen
    }

    var title: XCUIElement { app.staticTexts["sessions.title"] }
    var emptyState: XCUIElement { app.staticTexts["sessions.empty.state"] }
    var createButton: XCUIElement { app.buttons["sessions.create.button"] }
    var saveButton: XCUIElement { app.buttons["sessions.create.save.button"] }
    var catalogRetry: XCUIElement { app.buttons["sessions.create.workload.retry"] }
    var listRetry: XCUIElement { app.buttons["sessions.error.retry"] }
}
