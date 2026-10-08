import XCTest

struct ScratchpadScreen {
    let app: XCUIApplication

    static let screenTitle = "임시 공간"
    static let tabLabel = "임시공간"
    static let chipLabels = ["personal": "→ 개인", "work": "→ 회사", "affirmation": "→ 다짐", "routine": "→ 루틴"]

    @discardableResult
    static func open(in app: XCUIApplication) -> ScratchpadScreen {
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = tabBar.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", tabLabel)).firstMatch
        XCTAssertTrue(tab.waitForExistence(timeout: 5), "The 임시공간 tab should be in the tab bar")
        tab.tap()
        let screen = ScratchpadScreen(app: app)
        XCTAssertTrue(screen.header.waitForExistence(timeout: 15), "The 임시 공간 screen header should appear")
        XCTAssertTrue(screen.addButton.waitForExistence(timeout: 10), "The add button should be on the 임시 공간 screen")
        return screen
    }

    var header: XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@", Self.screenTitle)).firstMatch
    }

    var addButton: XCUIElement {
        app.buttons.matching(
            NSPredicate(format: "identifier == %@ OR label ENDSWITH %@", "scratchpad.add.button", "새 메모")
        ).firstMatch
    }

    var sheetTitle: XCUIElement { app.staticTexts["scratchpad.sheet.title"] }
    var sheetField: XCUIElement {
        app.descendants(matching: .any).matching(identifier: "scratchpad.sheet.text.field").firstMatch
    }
    var sheetSaveButton: XCUIElement { app.buttons["scratchpad.sheet.save.button"] }

    func memo(_ text: String) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@", text)).firstMatch
    }

    func add(_ text: String) {
        addButton.tap()
        XCTAssertTrue(sheetTitle.waitForExistence(timeout: 10), "The new-memo sheet should open")
        XCTAssertTrue(sheetField.waitForExistence(timeout: 5), "The memo field should exist")
        sheetField.tap()
        sheetField.typeText(text)
        XCTAssertTrue(sheetSaveButton.isEnabled, "저장 should be enabled once the memo is not empty")
        sheetSaveButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(sheetTitle), "The new-memo sheet should close after saving")
        XCTAssertTrue(memo(text).waitForExistence(timeout: 10), "'\(text)' should be listed in 임시 공간 after saving")
    }

    func chip(_ target: String, for text: String) -> XCUIElement? {
        let element = memo(text)
        guard element.waitForExistence(timeout: 10), let label = Self.chipLabels[target] else { return nil }
        let floor = element.frame.maxY - 1
        return app.buttons.matching(NSPredicate(format: "label == %@", label)).allElementsBoundByIndex
            .filter { $0.frame.minY >= floor }
            .min { $0.frame.minY < $1.frame.minY }
    }

    func move(_ text: String, to target: String, file: StaticString = #filePath, line: UInt = #line) {
        guard let chip = chip(target, for: text) else {
            return XCTFail("'\(text)' should be listed with a \(target) chip under it", file: file, line: line)
        }
        chip.tap()
    }
}

extension ScratchpadScreen {
    func openAddSheet() {
        addButton.tap()
        XCTAssertTrue(sheetTitle.waitForExistence(timeout: 10), "The new-memo sheet should open")
        XCTAssertTrue(sheetField.waitForExistence(timeout: 5), "The memo field should exist")
    }

    func reenter() -> ScratchpadScreen {
        AffirmationsScreen.open(in: app)
        return ScratchpadScreen.open(in: app)
    }

    var todayHeader: XCUIElement {
        app.descendants(matching: .any).matching(NSPredicate(format: "label BEGINSWITH %@", "오늘")).firstMatch
    }

    func chips(_ target: String) -> [XCUIElement] {
        guard let label = Self.chipLabels[target] else { return [] }
        return app.buttons.matching(NSPredicate(format: "label == %@", label)).allElementsBoundByIndex
    }

    func assertListedFirst(_ text: String, above other: String, file: StaticString = #filePath, line: UInt = #line) {
        let element = memo(text)
        XCTAssertTrue(element.waitForExistence(timeout: 15), "'\(text)' should be listed in 임시 공간", file: file, line: line)
        XCTAssertTrue(memo(other).waitForExistence(timeout: 5), "'\(other)' should still be listed", file: file, line: line)
        XCTAssertTrue(todayHeader.exists, "The 오늘 section should head the list", file: file, line: line)
        let top = element.frame.minY
        XCTAssertLessThan(todayHeader.frame.minY, top, "'\(text)' should sit under the 오늘 header", file: file, line: line)
        XCTAssertLessThan(top, memo(other).frame.minY, "The newest memo should sit above '\(other)'", file: file, line: line)
        XCTAssertTrue(
            chips("personal").allSatisfy { $0.frame.minY > top },
            "No other memo card should sit above '\(text)'",
            file: file,
            line: line
        )
    }

    func assertFourChips(for text: String, file: StaticString = #filePath, line: UInt = #line) {
        let element = memo(text)
        XCTAssertTrue(element.waitForExistence(timeout: 10), "'\(text)' should be listed in 임시 공간", file: file, line: line)
        let aligned = RoutinesScreen.eventually {
            let found = Self.chipLabels.keys.compactMap { chip($0, for: text) }
            guard found.count == Self.chipLabels.count, let first = found.first else { return false }
            let gap = first.frame.minY - element.frame.maxY
            return gap < 40 && found.allSatisfy { abs($0.frame.minY - first.frame.minY) < 2 }
        }
        XCTAssertTrue(
            aligned,
            "'\(text)' should carry the four chips → 개인 · → 회사 · → 다짐 · → 루틴 on its own card",
            file: file,
            line: line
        )
    }

    func swipeDelete(_ text: String, file: StaticString = #filePath, line: UInt = #line) {
        let element = memo(text)
        XCTAssertTrue(element.waitForExistence(timeout: 10), "'\(text)' should be listed before swiping", file: file, line: line)
        let delete = app.buttons.matching(NSPredicate(format: "label == %@", "삭제")).firstMatch
        element.swipeLeft()
        if !delete.waitForExistence(timeout: 3) {
            let start = element.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
            start.press(forDuration: 0.1, thenDragTo: start.withOffset(CGVector(dx: -240, dy: 0)))
        }
        XCTAssertTrue(delete.waitForExistence(timeout: 5), "Swiping a memo left should reveal 삭제", file: file, line: line)
        delete.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(element), "'\(text)' should leave 임시 공간 after 삭제", file: file, line: line)
    }

    func remove(_ texts: [String]) {
        for text in texts where memo(text).waitForExistence(timeout: 3) {
            swipeDelete(text)
        }
    }
}
