import XCTest

final class AffirmationsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    private func launchOnAffirmationsTab() -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments.append("--ui-test-mock-oidc")
        app.launch()
        // Affirmations is the default selection via RootView.onAppear. The
        // affirmations tab button itself may live under the system "More" tab
        // (iPhone TabView folds 6+ tabs), so we don't tap it — we just rely on
        // affirmations content being the active tab.
        return app
    }

    private func findInSheet(_ app: XCUIApplication, identifier: String, timeout: TimeInterval = 5) -> XCUIElement {
        // SwiftUI .sheet() places content in a separate accessibility container.
        // Walking descendants finds elements there, and narrowing to .button
        // ensures the returned element is hit-testable (not a wrapping VStack).
        let element = app.descendants(matching: .button)
            .matching(identifier: identifier)
            .firstMatch
        _ = element.waitForExistence(timeout: timeout)
        return element
    }

    func testAffirmationsScreenLandsOnLaunch() throws {
        let app = launchOnAffirmationsTab()

        let addButton = app.buttons["affirmations.add"]
        XCTAssertTrue(
            addButton.waitForExistence(timeout: 15),
            "Affirmations screen add button should be visible (affirmations is default tab)"
        )
    }

    func testTapAddOpensPrioritySheet() throws {
        let app = launchOnAffirmationsTab()

        let addButton = app.buttons["affirmations.add"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 15))
        addButton.tap()

        let saveButton = findInSheet(app, identifier: "priority.sheet.save", timeout: 8)
        XCTAssertTrue(saveButton.exists, "Save button should appear in the priority sheet")

        let cancelButton = findInSheet(app, identifier: "priority.sheet.cancel", timeout: 2)
        XCTAssertTrue(cancelButton.exists, "Cancel button should appear in the priority sheet")
    }

    func testCancelClosesSheet() throws {
        let app = launchOnAffirmationsTab()

        let addButton = app.buttons["affirmations.add"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 15))
        addButton.tap()

        let cancelButton = findInSheet(app, identifier: "priority.sheet.cancel", timeout: 8)
        XCTAssertTrue(cancelButton.exists, "Cancel button should be present before tap")
        cancelButton.tap()

        let dismissed = NSPredicate(format: "exists == false")
        let expectation = XCTNSPredicateExpectation(predicate: dismissed, object: cancelButton)
        XCTAssertEqual(
            XCTWaiter.wait(for: [expectation], timeout: 6),
            .completed,
            "Sheet should be dismissed and cancel button should disappear"
        )
    }
}
