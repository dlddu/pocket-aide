// 검증 시나리오: test-todo.md#시나리오 4
import XCTest

final class TodoEditUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testEditClearsOptionalFieldsAndCancelKeepsTheItem() {
        let token = TodoUI.uniqueToken()
        let app = TodoUI.launch()
        for area in TodoUIArea.allCases {
            let screen = TodoScreen.open(area, in: app)
            exerciseEdit(on: screen, token: token, app: app)
        }
    }

    private func exerciseEdit(on screen: TodoScreen, token: String, app: XCUIApplication) {
        let tag = "\(screen.area.rawValue) \(token)"
        let original = "점검 일정 \(tag)"
        let renamed = "점검 일정 변경 \(tag)"
        let untouched = "회의록 정리 \(tag)"
        let discarded = "취소될 제목 \(tag)"
        screen.add(title: original, memo: "지난 메모", due: true, priority: "normal")
        screen.add(title: untouched)
        XCTAssertEqual(
            screen.meta(original)?.label,
            "\(TodoScreen.today()) · \(screen.priorityLabel("normal")) · 메모: 지난 메모",
            "The seeded item should start with due date, normal priority and memo"
        )

        let edit = screen.openEditSheet(original)
        edit.setTitle(renamed)
        edit.setDueDate(on: false)
        edit.pickPriority("unset")
        edit.setMemo("")
        edit.save()
        XCTAssertTrue(screen.row(renamed).waitForExistence(timeout: 10), "The renamed title should be listed")
        XCTAssertFalse(screen.row(original).exists, "The old title should be gone")
        XCTAssertFalse(screen.meta(renamed)?.exists ?? true, "Clearing due date, priority and memo removes the secondary line")

        let cancelled = screen.openEditSheet(untouched)
        cancelled.setTitle(discarded)
        cancelled.cancel()
        XCTAssertTrue(screen.row(untouched).waitForExistence(timeout: 5), "A cancelled edit must keep the original title")
        XCTAssertFalse(screen.row(discarded).exists, "A cancelled edit must not save the typed title")

        TodoScreen.open(screen.area.other, in: app)
        let again = TodoScreen.open(screen.area, in: app)
        again.assertListing(token: token, present: [renamed, untouched], absent: [original, discarded])
        XCTAssertFalse(again.meta(renamed)?.exists ?? true, "The edit should persist after re-entering the tab")
    }
}
