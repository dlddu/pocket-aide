// 검증 시나리오: 없음 (스모크/인프라)
import XCTest

private extension XCUIElement {
    /// `waitForExistence` + `XCTAssertFalse` can't observe a currently-visible
    /// element disappearing — it returns immediately. Use a predicate-based
    /// wait when the assertion is "this should be gone soon".
    @discardableResult
    func waitToDisappear(timeout: TimeInterval, in testCase: XCTestCase, message: String) -> Bool {
        let predicate = NSPredicate(format: "exists == false")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        let result = XCTWaiter.wait(for: [expectation], timeout: timeout)
        if result != .completed {
            XCTFail(message, file: #filePath, line: #line)
            return false
        }
        return true
    }
}

final class AffirmationsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        // Class execution order is alphabetical, so this class runs BEFORE
        // LoginUITests. We must populate the keychain ourselves — otherwise
        // every launch lands on LoginView.
        UITestAuth.ensureSignedIn(self)
    }

    private func launchApp() -> XCUIApplication {
        let app = XCUIApplication()
        app.launch()
        return app
    }

    private func selectAffirmationsTab(in app: XCUIApplication) {
        let tabsBar = app.tabBars.firstMatch
        XCTAssertTrue(tabsBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")

        let dump = XCTAttachment(string: tabsBar.debugDescription)
        dump.name = "tabBar.debugDescription"
        dump.lifetime = .keepAlways
        add(dump)

        let labelled = tabsBar.buttons["다짐"]
        if labelled.exists {
            labelled.tap()
            return
        }

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

        XCTAssertTrue(
            app.staticTexts["screen.header.title"].waitForExistence(timeout: 15),
            "Affirmations header should appear after selecting the affirmations tab"
        )
    }

    /// Try the common places where iOS surfaces the "More" overflow rows.
    /// On iOS 26 each overflow row renders as an "Other" element whose label
    /// lives on a child `StaticText`, so the table cells themselves don't
    /// match `cells["다짐"]` — we have to query by static text or by row
    /// position. Returns true if an entry was found and tapped.
    @discardableResult
    private func findAndTapAffirmationsRow(in app: XCUIApplication) -> Bool {
        let textCandidates: [XCUIElement] = [
            app.tables.staticTexts["다짐"],
            app.collectionViews.staticTexts["다짐"],
            app.staticTexts["다짐"],
        ]
        for candidate in textCandidates where candidate.waitForExistence(timeout: 3) {
            candidate.tap()
            return true
        }
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
        // selectAffirmationsTab already waits for screen.header.title.
    }

    func testAddSentenceOpensSheet() {
        let app = launchApp()
        selectAffirmationsTab(in: app)

        let addButton = app.buttons["affirmations.add.button"]
        XCTAssertTrue(addButton.waitForExistence(timeout: 5), "Add button should be visible")
        addButton.tap()

        let sheetTitle = app.staticTexts["sheet.title"]
        XCTAssertTrue(
            sheetTitle.waitForExistence(timeout: 15),
            "Priority edit sheet should appear within 15s of tapping add"
        )

        XCTAssertFalse(
            app.buttons["sheet.delete.button"].waitForExistence(timeout: 1),
            "Delete button should be hidden in create mode"
        )

        // Cancel — closing via the cancel button is enough to verify the
        // sheet round-trips. Typing into the SwiftUI multi-line TextField
        // through XCUITest is flaky across iOS releases.
        let cancel = app.buttons["sheet.cancel.button"]
        XCTAssertTrue(
            cancel.waitForExistence(timeout: 5),
            "Cancel button should be present on create-mode sheet"
        )
        cancel.tap()

        sheetTitle.waitToDisappear(
            timeout: 5,
            in: self,
            message: "Sheet should dismiss within 5s of tapping cancel"
        )
    }
}
