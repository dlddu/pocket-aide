import XCTest

final class HelloWorldUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testHelloWorldShowsHeader() throws {
        let app = XCUIApplication()
        app.launch()

        XCTAssertTrue(
            app.staticTexts["Hello, pocket-aide"].waitForExistence(timeout: 10),
            "Header text should appear"
        )
    }

    func testHealthBadgeReflectsBackend() throws {
        let app = XCUIApplication()
        app.launch()

        let health = app.staticTexts["HealthStatus"]
        XCTAssertTrue(health.waitForExistence(timeout: 10), "Health status label should appear")
        XCTAssertFalse(health.label.isEmpty, "Health status label should have text")
    }

    func testSignInButtonExistsBeforeAuth() throws {
        let app = XCUIApplication()
        app.launch()

        let signIn = app.buttons["SignInButton"]
        XCTAssertTrue(signIn.waitForExistence(timeout: 10))
        XCTAssertTrue(signIn.isHittable)
    }
}
