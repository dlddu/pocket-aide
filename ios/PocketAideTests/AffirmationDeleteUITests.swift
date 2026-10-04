// 검증 시나리오: test-affirmations.md#시나리오 3
import XCTest

final class AffirmationDeleteUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testSwipeAndSheetDeleteRemoveSentencesFromListAndRotation() {
        let token = TodoUI.uniqueToken()
        let kept = "남길 다짐 \(token)"
        let swiped = "밀어 지울 다짐 \(token)"
        let viaSheet = "시트로 지울 다짐 \(token)"
        let app = TodoUI.launch()
        var screen = AffirmationsScreen.open(in: app)
        for sentence in [kept, viaSheet, swiped] {
            screen.add(sentence)
        }

        screen.openEditSheet(viaSheet).delete()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.row(viaSheet)), "The sheet-deleted sentence should leave the list")
        screen.swipeDelete(swiped)
        XCTAssertTrue(screen.row(kept).exists, "Untouched sentences stay listed")

        let create = screen.openCreateSheet()
        XCTAssertFalse(create.deleteButton.exists, "The create-mode sheet has no delete button")
        create.cancel()

        screen = screen.reenter()
        screen.assertRemoved([swiped, viaSheet], kept: kept)

        TodoUI.relaunch(app)
        screen = AffirmationsScreen.open(in: app)
        screen.assertRemoved([swiped, viaSheet], kept: kept)

        screen.remove([kept])
    }
}
