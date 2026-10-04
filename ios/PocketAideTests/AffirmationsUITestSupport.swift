import XCTest

struct AffirmationsScreen {
    let app: XCUIApplication

    static let screenTitle = "자주 읽어줘야 할 것"
    static let priorities = ["high", "normal", "low"]

    @discardableResult
    static func open(in app: XCUIApplication) -> AffirmationsScreen {
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = tabBar.buttons["다짐"]
        XCTAssertTrue(tab.waitForExistence(timeout: 5), "The 다짐 tab should be in the tab bar")
        tab.tap()
        let screen = AffirmationsScreen(app: app)
        XCTAssertTrue(screen.header.waitForExistence(timeout: 15), "The 다짐 screen header should appear")
        XCTAssertTrue(screen.addButton.waitForExistence(timeout: 10), "The add button should be on the 다짐 screen")
        return screen
    }

    func reenter() -> AffirmationsScreen {
        let other = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(other.waitForExistence(timeout: 5), "The PR 모니터 tab should be in the tab bar")
        other.tap()
        XCTAssertTrue(
            app.buttons["prmonitor.openprs.button"].waitForExistence(timeout: 15),
            "Leaving the 다짐 tab should land on PR 모니터"
        )
        return AffirmationsScreen.open(in: app)
    }

    var header: XCUIElement {
        app.staticTexts.matching(
            NSPredicate(format: "identifier == %@ AND label == %@", "screen.header.title", Self.screenTitle)
        ).firstMatch
    }

    var addButton: XCUIElement { app.buttons["affirmations.add.button"] }

    var rotateButton: XCUIElement {
        app.buttons.matching(
            NSPredicate(format: "identifier == %@ OR label == %@", "affirmations.hero.rotate", "다른 다짐 보기")
        ).firstMatch
    }

    func row(_ text: String) -> XCUIElement {
        app.staticTexts.matching(
            NSPredicate(format: "label == %@ AND NOT (identifier BEGINSWITH %@)", text, "affirmations.hero")
        ).firstMatch
    }

    func hero(_ text: String) -> XCUIElement {
        app.staticTexts.matching(
            NSPredicate(format: "label == %@ AND identifier BEGINSWITH %@", text, "affirmations.hero")
        ).firstMatch
    }

    func anywhere(_ text: String) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@", text)).firstMatch
    }

    @discardableResult
    func reveal(_ element: XCUIElement, attempts: Int = 6) -> Bool {
        for _ in 0..<attempts {
            if element.exists && element.isHittable { return true }
            app.swipeUp()
        }
        return element.exists && element.isHittable
    }

    func openCreateSheet() -> AffirmationSheet {
        addButton.tap()
        let sheet = AffirmationSheet(app: app)
        XCTAssertTrue(sheet.title.waitForExistence(timeout: 10), "The new-sentence sheet should open")
        XCTAssertEqual(sheet.title.label, "새 다짐")
        return sheet
    }

    func openEditSheet(_ text: String, file: StaticString = #filePath, line: UInt = #line) -> AffirmationSheet {
        let element = row(text)
        XCTAssertTrue(element.waitForExistence(timeout: 10), "'\(text)' should be listed", file: file, line: line)
        reveal(element)
        element.press(forDuration: 1.2)
        let sheet = AffirmationSheet(app: app)
        XCTAssertTrue(
            sheet.title.waitForExistence(timeout: 10),
            "Long-pressing '\(text)' should open its edit sheet",
            file: file,
            line: line
        )
        XCTAssertEqual(sheet.title.label, "우선순위 설정", file: file, line: line)
        return sheet
    }

    func add(_ text: String, priority: String? = nil) {
        let sheet = openCreateSheet()
        if let priority {
            sheet.pick(priority)
        }
        sheet.setText(text)
        sheet.save()
        XCTAssertTrue(row(text).waitForExistence(timeout: 10), "'\(text)' should be listed after saving")
    }

    func assertStored(_ text: String, priority: String, file: StaticString = #filePath, line: UInt = #line) {
        let sheet = openEditSheet(text, file: file, line: line)
        XCTAssertEqual(sheet.text, text, "The edit sheet should carry the stored sentence", file: file, line: line)
        XCTAssertEqual(
            sheet.selectedPriority(),
            priority,
            "'\(text)' should be stored with priority \(priority)",
            file: file,
            line: line
        )
        sheet.cancel()
    }

    func swipeDelete(_ text: String) {
        let element = row(text)
        XCTAssertTrue(element.waitForExistence(timeout: 10), "'\(text)' should be listed before swiping")
        reveal(element)
        element.swipeLeft()
        let delete = app.buttons.matching(
            NSPredicate(format: "label == %@ OR identifier ENDSWITH %@", "삭제", ".delete")
        ).firstMatch
        XCTAssertTrue(delete.waitForExistence(timeout: 5), "Swiping left should reveal 삭제")
        delete.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(element), "'\(text)' should leave the list after the swipe delete")
    }

    func remove(_ texts: [String]) {
        for text in texts where row(text).waitForExistence(timeout: 3) {
            swipeDelete(text)
        }
    }

    func assertRemoved(_ removed: [String], kept: String, rotations: Int = 10) {
        XCTAssertTrue(row(kept).waitForExistence(timeout: 10), "'\(kept)' should still be listed")
        for text in removed {
            XCTAssertFalse(anywhere(text).exists, "'\(text)' must not be listed or shown in the hero card")
        }
        XCTAssertTrue(rotateButton.waitForExistence(timeout: 5), "The hero card should offer the rotate button")
        for turn in 1...rotations {
            rotateButton.tap()
            for text in removed {
                XCTAssertFalse(anywhere(text).exists, "'\(text)' must not come back on hero rotation \(turn)")
            }
        }
    }
}

struct AffirmationSheet {
    let app: XCUIApplication

    static let placeholder = "다짐 문장을 입력하세요"

    var title: XCUIElement { app.staticTexts["sheet.title"] }
    var textField: XCUIElement {
        app.descendants(matching: .any).matching(identifier: "sheet.text.field").firstMatch
    }
    var saveButton: XCUIElement { app.buttons["sheet.save.button"] }
    var deleteButton: XCUIElement { app.buttons["sheet.delete.button"] }
    var cancelButton: XCUIElement { app.buttons["sheet.cancel.button"] }

    func pill(_ raw: String) -> XCUIElement {
        app.buttons["filter.pill.\(raw)"]
    }

    var text: String {
        let value = (textField.value as? String) ?? ""
        return value == Self.placeholder ? "" : value
    }

    func selectedPriority() -> String? {
        AffirmationsScreen.priorities.first { pill($0).isSelected }
    }

    func pick(_ raw: String) {
        let target = pill(raw)
        XCTAssertTrue(target.waitForExistence(timeout: 5), "Priority pill \(raw) should exist")
        target.tap()
        let selected = XCTNSPredicateExpectation(predicate: NSPredicate(format: "isSelected == true"), object: target)
        XCTAssertEqual(XCTWaiter.wait(for: [selected], timeout: 5), .completed, "Priority pill \(raw) should become selected")
    }

    func setText(_ newText: String) {
        XCTAssertTrue(textField.waitForExistence(timeout: 5), "The sentence field should exist")
        textField.tap()
        var attempts = 0
        while !text.isEmpty && attempts < 3 {
            textField.coordinate(withNormalizedOffset: CGVector(dx: 0.97, dy: 0.9)).tap()
            textField.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: text.count * 3 + 8))
            attempts += 1
        }
        XCTAssertEqual(text, "", "The sentence field should be empty before typing")
        if !newText.isEmpty {
            textField.typeText(newText)
        }
    }

    func save() {
        XCTAssertTrue(saveButton.isEnabled, "저장 should be enabled once the sentence is not empty")
        saveButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(title), "The sheet should close after saving")
    }

    func cancel() {
        cancelButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(title), "The sheet should close after cancel")
    }

    func delete() {
        XCTAssertTrue(deleteButton.waitForExistence(timeout: 5), "The edit sheet should offer 삭제")
        deleteButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(title), "The sheet should close after delete")
    }
}
