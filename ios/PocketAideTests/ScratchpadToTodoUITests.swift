// 검증 시나리오: test-todo.md#시나리오 9
import XCTest

final class ScratchpadToTodoUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testScratchpadMemosLandOnlyInTheChosenTodoArea() {
        let token = TodoUI.uniqueToken()
        let personalMemo = "장보기 목록 정리 \(token)"
        let workMemo = "주간 보고 초안 \(token)"
        let keptMemo = "나중에 읽을 글 \(token)"
        let app = TodoUI.launch()

        let scratchpad = ScratchpadScreen.open(in: app)
        for memo in [keptMemo, workMemo, personalMemo] {
            scratchpad.add(memo)
        }

        scratchpad.move(personalMemo, to: "personal")
        XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.memo(personalMemo)), "The memo sorted into 개인 should leave 임시 공간")
        scratchpad.move(workMemo, to: "work")
        XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.memo(workMemo)), "The memo sorted into 회사 should leave 임시 공간")

        let memos: [TodoUIArea: String] = [.personal: personalMemo, .work: workMemo]
        assertSorted(app: app, token: token, memos: memos, kept: keptMemo)

        TodoUI.relaunch(app)
        assertSorted(app: app, token: token, memos: memos, kept: keptMemo)
    }

    private func assertSorted(app: XCUIApplication, token: String, memos: [TodoUIArea: String], kept: String) {
        let scratchpad = ScratchpadScreen.open(in: app)
        XCTAssertTrue(scratchpad.memo(kept).waitForExistence(timeout: 15), "The memo whose chip was not tapped should stay in 임시 공간")
        for memo in memos.values {
            XCTAssertFalse(scratchpad.memo(memo).exists, "'\(memo)' must not remain in 임시 공간")
        }
        for area in TodoUIArea.allCases {
            guard let mine = memos[area], let theirs = memos[area.other] else { continue }
            TodoScreen.open(area, in: app).assertListing(token: token, present: [mine], absent: [theirs, kept])
        }
    }
}
