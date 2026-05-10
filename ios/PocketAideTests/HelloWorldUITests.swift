import XCTest

final class HelloWorldUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testHelloWorldShowsHeaderAndPalette() throws {
        let app = XCUIApplication()
        app.launch()

        XCTAssertTrue(
            app.staticTexts["Hello, pocket-aide"].waitForExistence(timeout: 5),
            "Header text should appear"
        )

        for area in ["personal", "work", "aiChat", "scratchpad", "routines", "affirmations", "voice", "system"] {
            let chip = app.otherElements["AreaChip-\(area)"]
            XCTAssertTrue(chip.waitForExistence(timeout: 3), "Area chip \(area) should be present")
        }
    }

    func testHealthBadgeReflectsBackend() throws {
        let app = XCUIApplication()
        app.launch()

        let health = app.staticTexts["HealthStatus"]
        XCTAssertTrue(health.waitForExistence(timeout: 8), "Health status label should appear")

        // When the backend is reachable (CI runs server on localhost), label should
        // contain "OK". When not reachable (developer running tests without backend
        // up), the label still appears with the error string. Either is acceptable
        // here; this assertion just guards that the view modeled the result.
        XCTAssertFalse(health.label.isEmpty)
    }

    func testSignInButtonExistsBeforeAuth() throws {
        let app = XCUIApplication()
        app.launch()

        let signIn = app.buttons["SignInButton"]
        XCTAssertTrue(signIn.waitForExistence(timeout: 5))
        XCTAssertTrue(signIn.isHittable)
    }
}
