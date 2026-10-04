// 검증 시나리오: test-todo.md#시나리오 5
import XCTest

final class TodoCompleteUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testCompleteMovesToDoneSectionAndCanBeUndone() {
        let token = TodoUI.uniqueToken()
        let app = TodoUI.launch()
        for area in TodoUIArea.allCases {
            let screen = TodoScreen.open(area, in: app)
            exerciseCompletion(on: screen, token: token, app: app)
        }
    }

    private func exerciseCompletion(on screen: TodoScreen, token: String, app: XCUIApplication) {
        let tag = "\(screen.area.rawValue) \(token)"
        let target = "택배 반품 \(tag)"
        let sibling = "영수증 정리 \(tag)"
        screen.add(title: target, memo: "편의점 접수", due: true, priority: "low")
        screen.add(title: sibling)
        let expectedMeta = "\(TodoScreen.today()) · \(screen.priorityLabel("low")) · 메모: 편의점 접수"
        guard let before = screen.counts() else { return XCTFail("Summary should parse") }

        screen.search(token)
        screen.toggle(target)?.tap()
        let completed = TodoCounts(open: before.open - 1, done: before.done + 1)
        XCTAssertTrue(screen.waitForCounts(completed), "Completing moves one item from open to done in the summary")
        assertInDoneSection(target, on: screen, expected: true)
        XCTAssertEqual(screen.meta(target)?.label, expectedMeta, "Completion keeps memo, due date and priority")
        XCTAssertFalse(isDone(sibling, on: screen), "The other item stays open")
        screen.clearSearch()
        screen.dismissKeyboard()

        TodoScreen.open(screen.area.other, in: app)
        let again = TodoScreen.open(screen.area, in: app)
        XCTAssertTrue(again.waitForCounts(completed), "Completion should persist after re-entering the tab")
        again.search(token)
        assertInDoneSection(target, on: again, expected: true)

        again.toggle(target)?.tap()
        XCTAssertTrue(again.waitForCounts(before), "Undoing restores the open and done counts")
        assertInDoneSection(target, on: again, expected: false)
        XCTAssertEqual(again.meta(target)?.label, expectedMeta, "Undo keeps memo, due date and priority")
        again.clearSearch()
        again.dismissKeyboard()
    }

    private func isDone(_ title: String, on screen: TodoScreen) -> Bool {
        let header = screen.sectionHeader(done: true)
        let row = screen.row(title)
        guard header.exists, row.waitForExistence(timeout: 5) else { return false }
        return row.frame.minY > header.frame.minY
    }

    private func assertInDoneSection(_ title: String, on screen: TodoScreen, expected: Bool) {
        if expected {
            XCTAssertTrue(screen.sectionHeader(done: true).waitForExistence(timeout: 10), "The done section should be shown")
        }
        let deadline = Date().addingTimeInterval(10)
        while isDone(title, on: screen) != expected, Date() < deadline {
            Thread.sleep(forTimeInterval: 0.5)
        }
        XCTAssertEqual(isDone(title, on: screen), expected, "'\(title)' done-section membership should be \(expected)")
    }
}
