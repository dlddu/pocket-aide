import XCTest

final class AffirmationsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    private func launchOnAffirmationsTab() -> XCUIApplication {
        let app = XCUIApplication()
        app.launch()
        // Affirmations is the default selection via RootView.onAppear. The
        // affirmations tab button itself may live under the system "More" tab
        // (iPhone TabView folds 6+ tabs), so we don't tap it — we just wait for
        // affirmations content to render.
        return app
    }

    func testAffirmationsScreenLandsOnLaunch() throws {
        let app = launchOnAffirmationsTab()

        let addButton = app.buttons["affirmations.add"]
        XCTAssertTrue(
            addButton.waitForExistence(timeout: 10),
            "Affirmations screen add button should be visible after selecting affirmations tab"
        )
    }

    func testTapAddOpensPrioritySheet() throws {
        let app = launchOnAffirmationsTab()

        let addButton = app.buttons["affirmations.add"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 10))
        addButton.tap()

        let textField = app.textFields["priority.sheet.text"].firstMatch
        let textView = app.textViews["priority.sheet.text"].firstMatch
        XCTAssertTrue(
            textField.waitForExistence(timeout: 5) || textView.waitForExistence(timeout: 5),
            "Priority sheet text input should appear"
        )

        let saveButton = app.buttons["priority.sheet.save"]
        XCTAssertTrue(saveButton.exists, "Save button should be present")
        let cancelButton = app.buttons["priority.sheet.cancel"]
        XCTAssertTrue(cancelButton.exists, "Cancel button should be present")
    }

    func testCancelClosesSheet() throws {
        let app = launchOnAffirmationsTab()

        let addButton = app.buttons["affirmations.add"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 10))
        addButton.tap()

        let cancelButton = app.buttons["priority.sheet.cancel"]
        XCTAssertTrue(cancelButton.waitForExistence(timeout: 5))
        cancelButton.tap()

        let dismissed = NSPredicate(format: "exists == false")
        let expectation = XCTNSPredicateExpectation(predicate: dismissed, object: cancelButton)
        XCTAssertEqual(
            XCTWaiter.wait(for: [expectation], timeout: 3),
            .completed,
            "Sheet should be dismissed and cancel button should disappear"
        )
    }
}
