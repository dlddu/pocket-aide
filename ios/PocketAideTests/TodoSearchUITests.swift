// 검증 시나리오: test-todo.md#시나리오 2
import XCTest

final class TodoSearchUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testSearchOnlyLooksInsideTheCurrentArea() {
        let token = TodoUI.uniqueToken()
        let personalTitle = "세금 신고 \(token)"
        let workTitle = "세금계산서 발행 \(token)"
        let workMemoTitle = "분기 정산 \(token)"
        let workMemo = "세금 관련"
        let app = TodoUI.launch()

        let personal = TodoScreen.open(.personal, in: app)
        personal.add(title: personalTitle)
        let work = TodoScreen.open(.work, in: app)
        work.add(title: workTitle)
        work.add(title: workMemoTitle, memo: workMemo)

        let personalAgain = TodoScreen.open(.personal, in: app)
        personalAgain.search("세금")
        XCTAssertTrue(personalAgain.row(personalTitle).waitForExistence(timeout: 10), "Personal search should find the personal item")
        XCTAssertFalse(personalAgain.row(workTitle).exists, "Personal search must not surface work items")
        XCTAssertFalse(personalAgain.row(workMemoTitle).exists, "Personal search must not surface work memos")
        personalAgain.clearSearch()
        personalAgain.dismissKeyboard()

        let workAgain = TodoScreen.open(.work, in: app)
        workAgain.search("세금")
        XCTAssertTrue(workAgain.row(workTitle).waitForExistence(timeout: 10), "Work search should match titles")
        XCTAssertTrue(workAgain.row(workMemoTitle).waitForExistence(timeout: 5), "Work search should match memos")
        XCTAssertFalse(workAgain.row(personalTitle).exists, "Work search must not surface personal items")

        workAgain.search("없는말")
        XCTAssertTrue(TodoUI.waitToDisappear(workAgain.row(workTitle), timeout: 5), "A query without matches should hide every item")
        let anyRow = app.staticTexts.matching(NSPredicate(format: "identifier BEGINSWITH %@", "todos.work.row."))
        XCTAssertEqual(anyRow.count, 0, "A query without matches should leave no rows")

        workAgain.clearSearch()
        workAgain.dismissKeyboard()
        XCTAssertTrue(workAgain.row(workTitle).waitForExistence(timeout: 10), "Clearing the query should restore the full list")
        XCTAssertTrue(workAgain.reveal(workAgain.row(workMemoTitle)), "Clearing the query should restore the memo-matched item")
    }
}
