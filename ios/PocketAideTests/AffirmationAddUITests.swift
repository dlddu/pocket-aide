// 검증 시나리오: test-affirmations.md#시나리오 1
import XCTest

final class AffirmationAddUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testAddedSentenceStaysListedWithItsPriority() {
        let sentence = "천천히 꾸준히 \(TodoUI.uniqueToken())"
        let app = TodoUI.launch()
        var screen = AffirmationsScreen.open(in: app)

        let blank = screen.openCreateSheet()
        XCTAssertFalse(blank.saveButton.isEnabled, "An empty sentence cannot be saved")
        blank.saveButton.tap()
        XCTAssertTrue(blank.title.exists, "Tapping 저장 with an empty sentence must keep the sheet open")
        blank.cancel()

        let sheet = screen.openCreateSheet()
        sheet.pick("high")
        sheet.setText(sentence)
        sheet.save()
        XCTAssertTrue(screen.row(sentence).waitForExistence(timeout: 10), "The saved sentence should be listed")
        XCTAssertTrue(screen.hero(sentence).waitForExistence(timeout: 5), "The hero card should show the new sentence")
        XCTAssertTrue(app.staticTexts["우선순위 높음"].exists, "The hero card should show the 높음 priority")
        screen.assertStored(sentence, priority: "high")

        screen = screen.reenter()
        screen.assertStored(sentence, priority: "high")

        TodoUI.relaunch(app)
        screen = AffirmationsScreen.open(in: app)
        screen.assertStored(sentence, priority: "high")

        screen.remove([sentence])
    }
}
