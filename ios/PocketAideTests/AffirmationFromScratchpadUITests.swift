// 검증 시나리오: test-affirmations.md#시나리오 4
import XCTest

final class AffirmationFromScratchpadUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testScratchpadMemoMovesIntoAffirmationsWithNormalPriority() {
        let token = TodoUI.uniqueToken()
        let sentence = "작게 시작해도 괜찮다 \(token)"
        let keptMemo = "나중에 정리할 생각 \(token)"
        let app = TodoUI.launch()

        var scratchpad = ScratchpadScreen.open(in: app)
        scratchpad.add(keptMemo)
        scratchpad.add(sentence)

        scratchpad.move(sentence, to: "affirmation")
        let sheet = AffirmationSheet(app: app)
        XCTAssertTrue(sheet.title.waitForExistence(timeout: 10), "→ 다짐 should open the priority sheet of the new sentence")
        XCTAssertEqual(sheet.title.label, "우선순위 설정")
        XCTAssertEqual(sheet.text, sentence, "The priority sheet should carry the memo text")
        XCTAssertEqual(sheet.selectedPriority(), "normal", "A moved memo should start with priority 보통")
        sheet.cancel()
        XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.memo(sentence)), "The moved memo should leave 임시 공간")
        XCTAssertTrue(scratchpad.memo(keptMemo).exists, "The memo that was not moved should stay in 임시 공간")

        var affirmations = AffirmationsScreen.open(in: app)
        XCTAssertTrue(affirmations.row(sentence).waitForExistence(timeout: 15), "The memo text should be listed as a sentence")
        affirmations.assertStored(sentence, priority: "normal")
        XCTAssertFalse(affirmations.row(keptMemo).exists, "The memo that was not moved must not become a sentence")

        TodoUI.relaunch(app)
        scratchpad = ScratchpadScreen.open(in: app)
        XCTAssertTrue(scratchpad.memo(keptMemo).waitForExistence(timeout: 15), "The kept memo should still be in 임시 공간")
        XCTAssertFalse(scratchpad.memo(sentence).exists, "The moved memo must not come back to 임시 공간")
        affirmations = AffirmationsScreen.open(in: app)
        affirmations.assertStored(sentence, priority: "normal")

        affirmations.remove([sentence])
    }
}
