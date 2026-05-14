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

    /// Navigate to the affirmations screen.
    ///
    /// iOS 26 SwiftUI `TabView` on iPhone portrait shows at most 4 tabs in the
    /// bottom bar plus a "More" entry — the remaining tabs (루틴 and 다짐 in
    /// our 6-tab shell) live behind "More". Default `selection: .affirmations`
    /// causes the system to mark "More" selected on launch and surface the
    /// overflow list from which the user picks 다짐. Real users do the same
    /// two-step navigation, so we mirror that here.
    private func selectAffirmationsTab(in app: XCUIApplication) {
        let tabsBar = app.tabBars.firstMatch
        XCTAssertTrue(tabsBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")

        // Diagnostic: attach what the tab bar exposes so any future failure
        // surfaces tab-bar state in the xcresult bundle.
        let dump = XCTAttachment(string: tabsBar.debugDescription)
        dump.name = "tabBar.debugDescription"
        dump.lifetime = .keepAlways
        add(dump)

        // Path 1: affirmations is directly in the bar.
        let labelled = tabsBar.buttons["다짐"]
        if labelled.exists {
            labelled.tap()
            return
        }

        // Path 2: affirmations is in the "More" overflow. The More tab is
        // already selected by default (selection binding == .affirmations),
        // so the overflow list should already be rendered. If it isn't,
        // tap "More" to bring it up.
        if !findAndTapAffirmationsRow(in: app) {
            let more = tabsBar.buttons["More"]
            XCTAssertTrue(more.exists, "Neither '다짐' tab nor 'More' tab is present")
            more.tap()
            _ = findAndTapAffirmationsRow(in: app)
        }

        let after = XCTAttachment(screenshot: app.screenshot())
        after.name = "after-affirmations-tap"
        after.lifetime = .keepAlways
        add(after)
    }

    /// Try the common places where iOS surfaces the "More" overflow rows.
    /// On iOS 26 each overflow row renders as an "Other" element whose label
    /// lives on a child `StaticText`, so the table cells themselves don't
    /// match `cells["다짐"]` — we have to query by static text or by row
    /// position. Returns true if an entry was found and tapped.
    @discardableResult
    private func findAndTapAffirmationsRow(in app: XCUIApplication) -> Bool {
        // Direct StaticText lookups handle the iOS 26 More table layout where
        // labels sit on a child element rather than the row itself.
        let textCandidates: [XCUIElement] = [
            app.tables.staticTexts["다짐"],
            app.collectionViews.staticTexts["다짐"],
            app.staticTexts["다짐"],
        ]
        for candidate in textCandidates where candidate.waitForExistence(timeout: 3) {
            candidate.tap()
            return true
        }
        // Fallback: legacy queries in case future iOS versions promote the
        // overflow rows to proper cells/buttons.
        let elementCandidates: [XCUIElement] = [
            app.tables.cells["다짐"],
            app.collectionViews.cells["다짐"],
            app.buttons["다짐"],
        ]
        for candidate in elementCandidates where candidate.exists {
            candidate.tap()
            return true
        }
        return false
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

        // Create-mode sheet must not surface a destructive action — delete only
        // makes sense for existing items, and PriorityEditSheet wires
        // `sheet.delete.button` conditionally on `.edit` mode.
        XCTAssertFalse(
            app.buttons["sheet.delete.button"].waitForExistence(timeout: 1),
            "Delete button should be hidden in create mode"
        )

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
