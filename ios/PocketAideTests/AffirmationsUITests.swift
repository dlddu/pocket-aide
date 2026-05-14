import XCTest

/// End-to-end coverage for the 다짐 (affirmations) tab. Relies on the shared
/// `UITestAuth.ensureSignedIn` helper to drop a token in the simulator keychain
/// once per process, then each test launches into the new TabView shell.
final class AffirmationsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        // Class execution order is alphabetical, so this class runs BEFORE
        // LoginUITests. We must populate the keychain ourselves — otherwise
        // every launch lands on LoginView.
        UITestAuth.ensureSignedIn(self)
    }

    private func launchApp(seed: String? = "1337") -> XCUIApplication {
        let app = XCUIApplication()
        if let seed {
            app.launchEnvironment["ROTATION_SEED"] = seed
        }
        app.launch()
        return app
    }

    /// Tap the affirmations tab by *position* (last of 6) rather than by label,
    /// since SwiftUI on iOS 26 may not honour the `selection:` binding's
    /// default for the last tab and we can't rely on the Korean "다짐" string
    /// being the exact button label exposed to XCUITest.
    private func selectAffirmationsTab(in app: XCUIApplication) {
        let tabsBar = app.tabBars.firstMatch
        XCTAssertTrue(tabsBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")

        // Diagnostic: attach what the tab bar actually exposes so we can read
        // it from the xcresult bundle when running in CI.
        let dump = XCTAttachment(string: tabsBar.debugDescription)
        dump.name = "tabBar.debugDescription"
        dump.lifetime = .keepAlways
        add(dump)

        let screenshot = XCTAttachment(screenshot: app.screenshot())
        screenshot.name = "before-affirmations-tap"
        screenshot.lifetime = .keepAlways
        add(screenshot)

        // Prefer the labelled button if present, otherwise fall back to the
        // last tab in the bar (affirmations is the 6th tab in RootView).
        let labelled = tabsBar.buttons["다짐"]
        if labelled.exists {
            labelled.tap()
            return
        }
        let allTabs = tabsBar.buttons.allElementsBoundByIndex
        XCTAssertFalse(allTabs.isEmpty, "Tab bar should expose at least one button")
        allTabs.last?.tap()

        let after = XCTAttachment(screenshot: app.screenshot())
        after.name = "after-affirmations-tap"
        after.lifetime = .keepAlways
        add(after)
    }

    func testAffirmationsTabIsReachable() {
        let app = launchApp()
        selectAffirmationsTab(in: app)
        XCTAssertTrue(
            app.staticTexts["screen.header.title"].waitForExistence(timeout: 15),
            "Affirmations header should appear after selecting the affirmations tab"
        )
    }

    func testAddSentenceOpensSheet() {
        let app = launchApp()
        selectAffirmationsTab(in: app)
        XCTAssertTrue(app.staticTexts["screen.header.title"].waitForExistence(timeout: 15))

        let addButton = app.buttons["affirmations.add.button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add button should be visible")
        addButton.tap()

        let sheetTitle = app.staticTexts["sheet.title"]
        XCTAssertTrue(sheetTitle.waitForExistence(timeout: 5), "Priority edit sheet should appear")

        // Cancel — closing via the cancel button is enough to verify the
        // sheet round-trips. Typing into the SwiftUI multi-line TextField
        // through XCUITest is flaky across iOS releases.
        let cancel = app.buttons["sheet.cancel.button"]
        if cancel.waitForExistence(timeout: 3) {
            cancel.tap()
        }
        XCTAssertFalse(
            app.staticTexts["sheet.title"].waitForExistence(timeout: 2),
            "Sheet should dismiss after cancel"
        )
    }

    func testRotationSeedLandsOnAffirmationsScreen() {
        let app = launchApp(seed: "424242")
        selectAffirmationsTab(in: app)
        XCTAssertTrue(
            app.staticTexts["screen.header.title"].waitForExistence(timeout: 15),
            "Same rotation seed should still render the affirmations screen header"
        )
    }
}
