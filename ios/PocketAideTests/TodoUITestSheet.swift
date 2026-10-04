import XCTest

struct TodoSheet {
    let app: XCUIApplication

    var title: XCUIElement { app.staticTexts["todo.sheet.title"] }
    var titleField: XCUIElement { app.textFields["todo.sheet.title.field"] }
    var memoField: XCUIElement {
        app.descendants(matching: .any).matching(identifier: "todo.sheet.memo.field").firstMatch
    }
    var dueToggle: XCUIElement { app.switches["todo.sheet.due.toggle"] }
    var saveButton: XCUIElement { app.buttons["todo.sheet.save.button"] }
    var deleteButton: XCUIElement { app.buttons["todo.sheet.delete.button"] }
    var cancelButton: XCUIElement { app.buttons["todo.sheet.cancel.button"] }

    func replace(_ field: XCUIElement, with text: String, submit: Bool = true) {
        XCTAssertTrue(field.waitForExistence(timeout: 5), "Sheet field \(field.identifier) should exist")
        field.tap()
        TodoUI.clear(field)
        if !text.isEmpty {
            field.typeText(text)
        }
        if submit {
            TodoUI.dismissKeyboard(in: app)
        }
    }

    func setTitle(_ text: String) {
        replace(titleField, with: text)
    }

    func setMemo(_ text: String) {
        replace(memoField, with: text, submit: false)
    }

    var dueIsOn: Bool {
        (dueToggle.value as? String) == "1"
    }

    func setDueDate(on: Bool) {
        XCTAssertTrue(dueToggle.waitForExistence(timeout: 5), "The due-date toggle should exist")
        if dueIsOn == on { return }
        dueToggle.coordinate(withNormalizedOffset: CGVector(dx: 0.93, dy: 0.5)).tap()
        if dueIsOn != on {
            dueToggle.switches.firstMatch.tap()
        }
        XCTAssertEqual(dueIsOn, on, "The due-date toggle should be \(on ? "on" : "off")")
    }

    func pickPriority(_ raw: String) {
        let pill = app.buttons["filter.pill.\(raw)"]
        XCTAssertTrue(pill.waitForExistence(timeout: 5), "Priority pill \(raw) should exist")
        pill.tap()
    }

    func save() {
        XCTAssertTrue(saveButton.isEnabled, "Save should be enabled")
        saveButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(title), "The sheet should close after saving")
    }

    func cancel() {
        cancelButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(title), "The sheet should close after cancel")
    }

    func delete(in testCase: XCTestCase) {
        XCTAssertTrue(deleteButton.waitForExistence(timeout: 5), "The edit sheet should offer 삭제")
        deleteButton.tap()
        if !TodoUI.waitToDisappear(title) {
            let dump = XCTAttachment(string: app.debugDescription)
            dump.name = "app-after-sheet-delete"
            dump.lifetime = .keepAlways
            testCase.add(dump)
            let shot = XCTAttachment(screenshot: app.screenshot())
            shot.name = "screen-after-sheet-delete"
            shot.lifetime = .keepAlways
            testCase.add(shot)
            XCTFail("The sheet should close after delete")
        }
    }
}
