// 검증 시나리오: test-todo.md#시나리오 8
import XCTest

final class TodoNoCrossAreaMoveUITests: XCTestCase {
    private let sheetControls: Set<String> = [
        "todo.sheet.title.field",
        "todo.sheet.memo.field",
        "todo.sheet.due.toggle",
        "filter.pill.unset",
        "filter.pill.high",
        "filter.pill.normal",
        "filter.pill.low",
        "todo.sheet.save.button",
        "todo.sheet.delete.button",
        "todo.sheet.cancel.button",
    ]

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testNoControlMovesAnItemToTheOtherArea() {
        let token = TodoUI.uniqueToken()
        let app = TodoUI.launch()
        for area in TodoUIArea.allCases {
            let screen = TodoScreen.open(area, in: app)
            let title = "영역 고정 \(area.rawValue) \(token)"
            screen.add(title: title)
            guard let counts = screen.counts() else { return XCTFail("Summary should parse") }

            assertEditSheetHasNoAreaControl(on: screen, title: title)
            assertSwipeOffersOnlyDelete(on: screen, title: title, token: token)
            assertLongPressAndDragDoNothing(on: screen, title: title, token: token, app: app)

            let again = TodoScreen.open(area, in: app)
            XCTAssertTrue(again.waitForCounts(counts), "The \(area.rawValue) summary should be unchanged")
            again.assertListing(token: token, present: [title], absent: [])
            TodoScreen.open(area.other, in: app).assertListing(token: token, present: [], absent: [title])
        }
    }

    private func assertEditSheetHasNoAreaControl(on screen: TodoScreen, title: String) {
        let sheet = screen.openEditSheet(title)
        for identifier in sheetControls {
            XCTAssertTrue(
                screen.app.descendants(matching: .any)[identifier].waitForExistence(timeout: 5),
                "Edit sheet should offer \(identifier)"
            )
        }
        let sheetIdentifiers = Set(
            screen.app.descendants(matching: .any).allElementsBoundByIndex
                .map(\.identifier)
                .filter { $0.hasPrefix("todo.sheet.") || $0.hasPrefix("filter.pill.") }
        )
        let extra = sheetIdentifiers.subtracting(sheetControls).subtracting(["todo.sheet.title"])
        XCTAssertTrue(extra.isEmpty, "Edit sheet must not offer other inputs, found \(extra.sorted())")
        let switchLabels = Set(screen.app.switches.allElementsBoundByIndex.map(\.label))
        XCTAssertEqual(switchLabels, ["마감일"], "The due-date toggle is the only switch")
        XCTAssertEqual(screen.app.segmentedControls.count, 0, "No segmented area picker")
        XCTAssertEqual(screen.app.pickers.count, 0, "No area picker")
        sheet.cancel()
    }

    private func assertSwipeOffersOnlyDelete(on screen: TodoScreen, title: String, token: String) {
        screen.search(token)
        screen.dismissKeyboard()
        let row = screen.row(title)
        XCTAssertTrue(row.waitForExistence(timeout: 10), "'\(title)' should be listed")
        let resting = screen.buttonLabels(in: screen.list)
        row.swipeLeft()
        XCTAssertTrue(screen.list.buttons["삭제"].waitForExistence(timeout: 5), "Swiping left should reveal 삭제")
        let revealed = screen.buttonLabels(in: screen.list)
        var added = revealed
        for label in resting {
            if let index = added.firstIndex(of: label) {
                added.remove(at: index)
            }
        }
        XCTAssertEqual(added, ["삭제"], "The swipe should reveal exactly one action, 삭제")
        row.swipeRight()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.list.buttons["삭제"], timeout: 5), "Swiping back should hide the action")
        screen.clearSearch()
        screen.dismissKeyboard()
    }

    private func assertLongPressAndDragDoNothing(on screen: TodoScreen, title: String, token: String, app: XCUIApplication) {
        screen.search(token)
        screen.dismissKeyboard()
        let row = screen.row(title)
        XCTAssertTrue(row.waitForExistence(timeout: 10), "'\(title)' should be listed")
        let before = screen.buttonLabels(in: app)
        row.press(forDuration: 1.5)
        XCTAssertFalse(TodoSheet(app: app).title.waitForExistence(timeout: 2), "Long press must not open a sheet")
        XCTAssertEqual(screen.buttonLabels(in: app), before, "Long press must not reveal a menu")

        let target = app.tabBars.firstMatch.coordinate(withNormalizedOffset: CGVector(dx: 0.9, dy: 0.5))
        row.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5)).press(forDuration: 1.0, thenDragTo: target)
        let back = TodoScreen.open(screen.area, in: app)
        back.assertListing(token: token, present: [title], absent: [])
    }
}
