// 검증 시나리오: test-scratchpad.md#시나리오 6
import XCTest

final class ScratchpadMoveTargetsUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testEveryMemoOffersFourTargetsAndMovesOnlyWhereItsChipWasTapped() {
        let token = TodoUI.uniqueToken()
        let moves = [
            (target: "personal", memo: "세탁소 들르기 \(token)"),
            (target: "work", memo: "회의록 공유 \(token)"),
            (target: "affirmation", memo: "서두르지 않는다 \(token)"),
            (target: "routine", memo: "주말 대청소 \(token)"),
        ]
        let kept = "아직 정하지 못한 메모 \(token)"
        let all = moves.map { $0.memo } + [kept]
        let app = TodoUI.launch()

        let scratchpad = ScratchpadScreen.open(in: app)
        for move in moves.reversed() {
            scratchpad.add(move.memo)
        }
        scratchpad.add(kept)

        scratchpad.assertFourChips(for: kept)
        for move in moves {
            scratchpad.assertFourChips(for: move.memo)
            scratchpad.move(move.memo, to: move.target)
            if move.target == "affirmation" {
                let sheet = AffirmationSheet(app: app)
                XCTAssertTrue(sheet.title.waitForExistence(timeout: 10), "→ 다짐 should open the priority sheet")
                sheet.cancel()
            }
            XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.memo(move.memo)), "'\(move.memo)' should leave 임시 공간")
        }
        XCTAssertTrue(scratchpad.memo(kept).exists, "The memo whose chips were not tapped should stay in 임시 공간")

        assertDestinations(moves.map { $0.memo }, all: all, token: token, in: app)

        let back = ScratchpadScreen.open(in: app)
        XCTAssertTrue(back.memo(kept).waitForExistence(timeout: 15), "The untouched memo should still be in 임시 공간")
        for moved in moves.map { $0.memo } {
            XCTAssertFalse(back.memo(moved).exists, "'\(moved)' must not remain in 임시 공간")
        }
        back.remove([kept])
    }

    private func assertDestinations(_ moved: [String], all: [String], token: String, in app: XCUIApplication) {
        for (index, area) in [TodoUIArea.personal, .work].enumerated() {
            let mine = moved[index]
            TodoScreen.open(area, in: app).assertListing(token: token, present: [mine], absent: all.filter { $0 != mine })
        }

        let sentence = moved[2]
        let affirmations = AffirmationsScreen.open(in: app)
        XCTAssertTrue(affirmations.row(sentence).waitForExistence(timeout: 15), "The → 다짐 memo should be listed as a sentence")
        for other in all where other != sentence {
            XCTAssertFalse(affirmations.anywhere(other).exists, "'\(other)' must not be listed in 다짐")
        }
        affirmations.remove([sentence])

        let name = moved[3]
        let routines = RoutinesScreen.open(in: app)
        XCTAssertTrue(routines.card(name).reveal(), "The → 루틴 memo should be listed as a routine")
        for other in all where other != name {
            XCTAssertFalse(routines.card(other).title.exists, "'\(other)' must not be listed in 루틴")
        }
        routines.swipeDelete(routines.card(name).title)
    }
}
