// 검증 시나리오: test-todo.md#시나리오 6
import XCTest

final class TodoDeleteUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testSwipeAndSheetDeleteRemoveItemsForGood() {
        let token = TodoUI.uniqueToken()
        let app = TodoUI.launch()
        var expectations: [TodoUIArea: (kept: [String], deleted: [String])] = [:]
        for area in TodoUIArea.allCases {
            let screen = TodoScreen.open(area, in: app)
            expectations[area] = exerciseDelete(on: screen, token: token)
        }

        for area in TodoUIArea.allCases {
            guard let expected = expectations[area] else { continue }
            TodoScreen.open(area, in: app).assertListing(token: token, present: expected.kept, absent: expected.deleted)
        }

        TodoUI.relaunch(app)
        for area in TodoUIArea.allCases {
            guard let expected = expectations[area] else { continue }
            TodoScreen.open(area, in: app).assertListing(token: token, present: expected.kept, absent: expected.deleted)
        }
    }

    private func exerciseDelete(on screen: TodoScreen, token: String) -> (kept: [String], deleted: [String]) {
        let tag = "\(screen.area.rawValue) \(token)"
        let swiped = "오래된 메모 \(tag)"
        let viaSheet = "끝난 약속 \(tag)"
        let kept = "남길 항목 \(tag)"
        for title in [kept, viaSheet, swiped] {
            screen.add(title: title)
        }
        guard let before = screen.counts() else {
            XCTFail("Summary should parse")
            return ([kept], [swiped, viaSheet])
        }

        screen.search(token)
        let row = screen.row(swiped)
        XCTAssertTrue(row.waitForExistence(timeout: 10), "'\(swiped)' should be listed before swiping")
        screen.dismissKeyboard()
        row.swipeLeft()
        let swipeDelete = screen.list.buttons["삭제"]
        XCTAssertTrue(swipeDelete.waitForExistence(timeout: 5), "Swiping left should reveal 삭제")
        swipeDelete.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(row), "The swiped item should leave the list")
        XCTAssertTrue(
            screen.waitForCounts(TodoCounts(open: before.open - 1, done: before.done)),
            "The summary should drop the swiped item"
        )

        screen.openEditSheet(viaSheet).delete()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.row(viaSheet)), "The sheet-deleted item should leave the list")
        XCTAssertTrue(
            screen.waitForCounts(TodoCounts(open: before.open - 2, done: before.done)),
            "The summary should drop the sheet-deleted item"
        )
        XCTAssertTrue(screen.row(kept).exists, "Untouched items stay listed")
        screen.clearSearch()
        screen.dismissKeyboard()

        let create = screen.openCreateSheet()
        XCTAssertFalse(create.deleteButton.exists, "The create-mode sheet has no delete button")
        create.cancel()
        return ([kept], [swiped, viaSheet])
    }
}
