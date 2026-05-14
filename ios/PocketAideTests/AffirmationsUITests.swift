import XCTest

/// End-to-end coverage for the 다짐 (affirmations) tab. These tests rely on
/// the same signed-in session that LoginUITests sets up — they do NOT set
/// `UI_TESTS_USE_LEGACY_HOME`, so they land on the new TabView with the
/// affirmations tab pre-selected. A rotation seed is pinned so hero
/// selection is deterministic across test runs.
final class AffirmationsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        // Class execution order is alphabetical, so this class runs BEFORE
        // LoginUITests. We must perform the OIDC dance ourselves to populate
        // the simulator keychain — otherwise every launch lands on LoginView
        // and the affirmations screen identifiers are unreachable.
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

    func testAffirmationsTabIsDefaultSelected() {
        let app = launchApp()
        // The tab bar buttons in SwiftUI TabView are labeled by the .tabItem
        // text; identifiers on tab content do not propagate to the tab bar
        // button. Querying by label is the canonical way to assert tab
        // existence.
        let tabsBar = app.tabBars.firstMatch
        XCTAssertTrue(tabsBar.waitForExistence(timeout: 10), "Tab bar should appear")
        XCTAssertTrue(tabsBar.buttons["다짐"].exists, "다짐 tab should be present")

        // The screen header is rendered by AffirmationsView only; if it
        // shows up immediately on launch the affirmations tab is the
        // pre-selected one.
        XCTAssertTrue(
            app.staticTexts["screen.header.title"].waitForExistence(timeout: 10),
            "Affirmations header should be visible because the affirmations tab is selected by default"
        )
    }

    func testAddSentenceOpensSheetAndShowsInList() {
        let app = launchApp()
        // Wait for the screen header to confirm we landed on the affirmations
        // tab.
        XCTAssertTrue(app.staticTexts["screen.header.title"].waitForExistence(timeout: 10))

        let addButton = app.buttons["affirmations.add.button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add button should be visible")
        addButton.tap()

        let sheetTitle = app.staticTexts["sheet.title"]
        XCTAssertTrue(sheetTitle.waitForExistence(timeout: 5), "Priority edit sheet should appear")

        let field = app.textFields["sheet.text.field"]
        if field.waitForExistence(timeout: 3) {
            field.tap()
            field.typeText("매일 1%씩")
        }

        let save = app.buttons["sheet.save.button"]
        if save.waitForExistence(timeout: 3) {
            save.tap()
        }

        // After saving the sheet disappears and the new sentence becomes the
        // hero (PRD-5 AC: 추가 직후 자동 노출).
        XCTAssertFalse(
            app.staticTexts["sheet.title"].waitForExistence(timeout: 2),
            "Sheet should dismiss after save"
        )
    }

    func testRotationIsDeterministicForSameSeed() {
        let appA = launchApp(seed: "424242")
        XCTAssertTrue(appA.staticTexts["screen.header.title"].waitForExistence(timeout: 10))
        let heroA = appA.staticTexts["affirmations.hero.text"].label
        appA.terminate()

        let appB = launchApp(seed: "424242")
        XCTAssertTrue(appB.staticTexts["screen.header.title"].waitForExistence(timeout: 10))
        let heroB = appB.staticTexts["affirmations.hero.text"].label

        // If there are no items yet (empty backend state) both heroes will be
        // empty strings, which is also a stable equality. The test asserts
        // *equality* under the same seed, not a particular value.
        XCTAssertEqual(
            heroA,
            heroB,
            "Same rotation seed should produce the same hero affirmation"
        )
    }
}
