// 검증 시나리오: test-todo.md#시나리오 3
import XCTest

final class TodoAddUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testAddWithAllFieldsAndTitleOnlyInBothAreas() {
        let token = TodoUI.uniqueToken()
        let app = TodoUI.launch()
        var added: [TodoUIArea: (full: String, bare: String)] = [:]

        for area in TodoUIArea.allCases {
            let screen = TodoScreen.open(area, in: app)
            added[area] = addBothKinds(on: screen, token: token)
        }

        for area in TodoUIArea.allCases {
            guard let titles = added[area] else { continue }
            let screen = TodoScreen.open(area, in: app)
            assertStored(on: screen, titles: titles, token: token)
        }

        TodoUI.relaunch(app)
        for area in TodoUIArea.allCases {
            guard let titles = added[area] else { continue }
            let screen = TodoScreen.open(area, in: app)
            assertStored(on: screen, titles: titles, token: token)
        }
    }

    private func addBothKinds(on screen: TodoScreen, token: String) -> (full: String, bare: String) {
        let full = "보고서 초안 \(screen.area.rawValue) \(token)"
        let bare = "우유 사기 \(screen.area.rawValue) \(token)"
        guard let before = screen.counts() else {
            XCTFail("Summary should parse")
            return (full, bare)
        }

        let sheet = screen.openCreateSheet()
        XCTAssertFalse(sheet.saveButton.isEnabled, "Save must be disabled while the title is empty")
        sheet.setTitle("   ")
        XCTAssertFalse(sheet.saveButton.isEnabled, "Save must stay disabled for a whitespace-only title")
        sheet.setTitle(full)
        sheet.setMemo("참고 자료 첨부")
        sheet.setDueDate(on: true)
        sheet.pickPriority("high")
        sheet.save()

        screen.add(title: bare)
        XCTAssertTrue(
            screen.waitForCounts(TodoCounts(open: before.open + 2, done: before.done)),
            "Both new items should count as open"
        )
        return (full, bare)
    }

    private func assertStored(on screen: TodoScreen, titles: (full: String, bare: String), token: String) {
        screen.search(token)
        let fullRow = screen.row(titles.full)
        let bareRow = screen.row(titles.bare)
        XCTAssertTrue(fullRow.waitForExistence(timeout: 10), "'\(titles.full)' should be listed")
        XCTAssertTrue(bareRow.waitForExistence(timeout: 10), "'\(titles.bare)' should be listed")

        let meta = screen.meta(titles.full)
        XCTAssertEqual(
            meta?.label,
            "\(TodoScreen.today()) · \(screen.priorityLabel("high")) · 메모: 참고 자료 첨부",
            "The full item should show due date, priority and memo"
        )
        XCTAssertFalse(
            screen.meta(titles.bare)?.exists ?? true,
            "A title-only item should have no secondary line"
        )

        let openHeader = screen.sectionHeader(done: false)
        XCTAssertTrue(openHeader.waitForExistence(timeout: 5), "The open section header should be shown")
        XCTAssertGreaterThan(fullRow.frame.minY, openHeader.frame.minY, "New items belong to the open section")
        XCTAssertLessThan(fullRow.frame.minY, bareRow.frame.minY, "Dated items sort before undated ones")
        screen.clearSearch()
        screen.dismissKeyboard()
    }
}
