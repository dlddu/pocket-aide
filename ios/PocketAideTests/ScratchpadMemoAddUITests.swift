// 검증 시나리오: test-scratchpad.md#시나리오 2
import XCTest

final class ScratchpadMemoAddUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testTextMemoIsTrimmedListedFirstAndPersists() {
        let earlier = "먼저 적어 둔 메모 \(TodoUI.uniqueToken())"
        let memo = "엄마 생신 선물 \(Int.random(in: 100_000..<1_000_000))"
        let app = TodoUI.launch()
        var scratchpad = ScratchpadScreen.open(in: app)
        scratchpad.add(earlier)

        scratchpad.openAddSheet()
        XCTAssertFalse(scratchpad.sheetSaveButton.isEnabled, "저장 must be disabled while the memo is empty")
        scratchpad.sheetField.tap()
        scratchpad.sheetField.typeText(" ")
        XCTAssertFalse(scratchpad.sheetSaveButton.isEnabled, "저장 must stay disabled while the memo is only whitespace")
        scratchpad.sheetField.typeText("\(memo) ")
        XCTAssertTrue(scratchpad.sheetSaveButton.isEnabled, "저장 should be enabled once the memo has text")
        scratchpad.sheetSaveButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.sheetTitle), "The new-memo sheet should close after saving")
        scratchpad.assertListedFirst(memo, above: earlier)
        XCTAssertFalse(AffirmationSheet(app: app).title.exists, "Saving a memo must not ask where to sort it")

        scratchpad = scratchpad.reenter()
        scratchpad.assertListedFirst(memo, above: earlier)

        TodoUI.relaunch(app)
        scratchpad = ScratchpadScreen.open(in: app)
        scratchpad.assertListedFirst(memo, above: earlier)

        scratchpad.remove([memo, earlier])
    }
}
