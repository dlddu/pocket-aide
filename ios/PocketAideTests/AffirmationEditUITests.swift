// 검증 시나리오: test-affirmations.md#시나리오 2
import XCTest

final class AffirmationEditUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testEditChangesOnlyTheTargetAndCancelChangesNothing() {
        let token = TodoUI.uniqueToken()
        let original = "물 자주 마시기 \(token)"
        let edited = "물 한 잔 먼저 \(token)"
        let untouched = "일찍 잠들기 \(token)"
        let discarded = "버려질 문장 \(token)"
        let app = TodoUI.launch()
        var screen = AffirmationsScreen.open(in: app)
        screen.add(untouched)
        screen.add(original)
        screen.assertStored(original, priority: "normal")

        let edit = screen.openEditSheet(original)
        edit.pick("low")
        edit.setText(edited)
        edit.save()
        XCTAssertTrue(screen.row(edited).waitForExistence(timeout: 10), "The edited sentence should be listed")
        XCTAssertFalse(screen.anywhere(original).exists, "The old sentence should be gone")

        let cancelled = screen.openEditSheet(untouched)
        cancelled.pick("high")
        cancelled.setText(discarded)
        cancelled.cancel()
        XCTAssertTrue(screen.row(untouched).waitForExistence(timeout: 5), "A cancelled edit must keep the sentence")
        XCTAssertFalse(screen.anywhere(discarded).exists, "A cancelled edit must not save the typed sentence")

        screen = screen.reenter()
        assertEdited(on: screen, edited: edited, untouched: untouched, gone: [original, discarded])

        TodoUI.relaunch(app)
        screen = AffirmationsScreen.open(in: app)
        assertEdited(on: screen, edited: edited, untouched: untouched, gone: [original, discarded])

        screen.remove([edited, untouched])
    }

    private func assertEdited(on screen: AffirmationsScreen, edited: String, untouched: String, gone: [String]) {
        screen.assertStored(edited, priority: "low")
        screen.assertStored(untouched, priority: "normal")
        for text in gone {
            XCTAssertFalse(screen.anywhere(text).exists, "'\(text)' must not be listed")
        }
    }
}
