import XCTest

struct RoutinesScreen {
    let app: XCUIApplication

    static let tabLabel = "루틴"
    static let nameField = "routines.sheet.name.field"
    static let stepsSheetField = "routines.steps.sheet.field"

    @discardableResult
    static func open(in app: XCUIApplication) -> RoutinesScreen {
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let screen = RoutinesScreen(app: app)
        let direct = tabBar.buttons[tabLabel]
        if direct.exists {
            direct.tap()
        } else {
            let more = tabBar.buttons["More"]
            XCTAssertTrue(more.exists, "Neither the 루틴 tab nor 'More' is present")
            more.tap()
            if !screen.addButton.waitForExistence(timeout: 3) && !tapMoreRow(in: app) {
                let back = app.navigationBars.buttons.firstMatch
                if back.exists {
                    back.tap()
                }
                XCTAssertTrue(tapMoreRow(in: app), "The 루틴 row should be listed under More")
            }
        }
        XCTAssertTrue(screen.addButton.waitForExistence(timeout: 15), "The 새 루틴 button should be on the 루틴 screen")
        return screen
    }

    private static func tapMoreRow(in app: XCUIApplication) -> Bool {
        let candidates: [XCUIElement] = [
            app.tables.staticTexts[tabLabel],
            app.collectionViews.staticTexts[tabLabel],
            app.tables.cells[tabLabel],
            app.collectionViews.cells[tabLabel],
        ]
        for candidate in candidates where candidate.waitForExistence(timeout: 3) {
            candidate.tap()
            return true
        }
        return false
    }

    static func eventually(timeout: TimeInterval = 10, _ condition: () -> Bool) -> Bool {
        let deadline = Date().addingTimeInterval(timeout)
        repeat {
            if condition() { return true }
            Thread.sleep(forTimeInterval: 0.3)
        } while Date() < deadline
        return condition()
    }

    func reenter() -> RoutinesScreen {
        let other = app.tabBars.firstMatch.buttons["다짐"]
        XCTAssertTrue(other.waitForExistence(timeout: 5), "The 다짐 tab should be in the tab bar")
        other.tap()
        XCTAssertTrue(
            app.buttons["affirmations.add.button"].waitForExistence(timeout: 15),
            "Leaving the 루틴 tab should land on 다짐"
        )
        return RoutinesScreen.open(in: app)
    }

    var addButton: XCUIElement {
        app.buttons.matching(
            NSPredicate(format: "identifier == %@ OR label ENDSWITH %@", "routines.add.button", "새 루틴")
        ).firstMatch
    }

    var sheetTitle: XCUIElement { app.staticTexts["routines.sheet.title"] }
    var sheetSaveButton: XCUIElement { app.buttons["routines.sheet.save.button"] }
    var sheetAddStepButton: XCUIElement { app.buttons["routines.sheet.step.add.button"] }
    var stepsSheetTitle: XCUIElement { app.staticTexts["routines.steps.sheet.title"] }
    var stepsSheetSaveButton: XCUIElement { app.buttons["routines.steps.sheet.save.button"] }

    var stepsSheetDeleteButtons: [XCUIElement] {
        app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH %@", "routines.steps.sheet.delete."))
            .allElementsBoundByIndex
            .sorted { $0.frame.minY < $1.frame.minY }
    }

    func field(_ identifier: String) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: identifier).firstMatch
    }

    func type(_ text: String, into identifier: String) {
        let target = field(identifier)
        XCTAssertTrue(target.waitForExistence(timeout: 5), "\(identifier) should exist")
        target.tap()
        target.typeText(text)
        TodoUI.dismissKeyboard(in: app)
    }

    func openAddSheet() {
        addButton.tap()
        XCTAssertTrue(sheetTitle.waitForExistence(timeout: 10), "The 새 루틴 sheet should open")
    }

    func fillSteps(_ steps: [String], trailingBlank: Bool = false) {
        for (index, step) in steps.enumerated() {
            if index > 0 {
                sheetAddStepButton.tap()
            }
            type(step, into: "routines.sheet.step.\(index)")
        }
        if trailingBlank {
            sheetAddStepButton.tap()
            XCTAssertTrue(
                field("routines.sheet.step.\(steps.count)").waitForExistence(timeout: 5),
                "단계 추가 should append an empty step field"
            )
        }
    }

    func saveAddSheet() {
        XCTAssertTrue(sheetSaveButton.isEnabled, "저장 should be enabled once the routine has a name")
        sheetSaveButton.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(sheetTitle), "The 새 루틴 sheet should close after saving")
    }

    @discardableResult
    func create(name: String, steps: [String]) -> RoutineCard {
        openAddSheet()
        type(name, into: Self.nameField)
        fillSteps(steps)
        saveAddSheet()
        let created = card(name)
        XCTAssertTrue(created.reveal(), "'\(name)' should be listed after saving")
        return created
    }

    func card(_ name: String) -> RoutineCard {
        RoutineCard(app: app, name: name)
    }

    func removeAll(limit: Int = 8) {
        let summaries = app.staticTexts.matching(
            NSPredicate(format: "label CONTAINS %@ AND label ENDSWITH %@", " / ", "완료")
        )
        for _ in 0..<limit {
            let summary = summaries.firstMatch
            guard summary.waitForExistence(timeout: 2), swipeDelete(summary) else { return }
        }
    }

    @discardableResult
    func swipeDelete(_ element: XCUIElement) -> Bool {
        element.swipeLeft()
        let delete = app.buttons.matching(NSPredicate(format: "label == %@", "삭제")).firstMatch
        guard delete.waitForExistence(timeout: 5) else { return false }
        delete.tap()
        return TodoUI.waitToDisappear(element)
    }
}

struct RoutineCard {
    let app: XCUIApplication
    let name: String

    var title: XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@", name)).firstMatch
    }

    @discardableResult
    func reveal(attempts: Int = 6) -> Bool {
        guard title.waitForExistence(timeout: 10) else {
            for _ in 0..<attempts where !title.exists {
                app.swipeUp()
            }
            return title.exists
        }
        return true
    }

    private func nearestBelowTitle(_ query: XCUIElementQuery) -> XCUIElement? {
        let top = title.frame.minY - 1
        return query.allElementsBoundByIndex
            .filter { $0.frame.minY >= top }
            .min { $0.frame.minY < $1.frame.minY }
    }

    var summary: String? {
        nearestBelowTitle(app.staticTexts.matching(
            NSPredicate(format: "label CONTAINS %@ AND label ENDSWITH %@", " / ", "완료")
        ))?.label
    }

    var progress: String? {
        let top = title.frame.minY
        return app.staticTexts.matching(NSPredicate(format: "label IN %@ OR label ENDSWITH %@", ["대기", "완료"], "%"))
            .allElementsBoundByIndex
            .min { abs($0.frame.minY - top) < abs($1.frame.minY - top) }?.label
    }

    var stepsButton: XCUIElement? {
        nearestBelowTitle(app.buttons.matching(
            NSPredicate(format: "label ENDSWITH %@ OR identifier ENDSWITH %@", "단계", ".steps.button")
        ))
    }

    func step(_ stepTitle: String) -> XCUIElement {
        app.buttons.matching(NSPredicate(format: "label ENDSWITH %@", stepTitle)).firstMatch
    }

    func isChecked(_ stepTitle: String) -> Bool {
        (step(stepTitle).value as? String) == "checked"
    }

    func order(of stepTitles: [String]) -> [String] {
        stepTitles
            .filter { step($0).exists }
            .sorted { step($0).frame.minY < step($1).frame.minY }
    }

    func waitFor(summary expected: String, progress expectedProgress: String) -> Bool {
        RoutinesScreen.eventually { summary == expected && progress == expectedProgress }
    }

    func assertState(
        summary expected: String,
        progress expectedProgress: String,
        steps: [String],
        file: StaticString = #filePath,
        line: UInt = #line
    ) {
        XCTAssertTrue(reveal(), "'\(name)' should be listed on the 루틴 screen", file: file, line: line)
        XCTAssertTrue(
            waitFor(summary: expected, progress: expectedProgress),
            "'\(name)' should read '\(expected)' / '\(expectedProgress)' but reads '\(summary ?? "-")' / '\(progress ?? "-")'",
            file: file,
            line: line
        )
        XCTAssertEqual(order(of: steps), steps, "The steps of '\(name)' should be listed in this order", file: file, line: line)
    }

    func toggle(_ stepTitle: String, file: StaticString = #filePath, line: UInt = #line) {
        let target = step(stepTitle)
        XCTAssertTrue(target.waitForExistence(timeout: 10), "Step '\(stepTitle)' should be on the card", file: file, line: line)
        target.tap()
    }

    func openStepsSheet(in screen: RoutinesScreen, file: StaticString = #filePath, line: UInt = #line) {
        guard reveal(), let button = stepsButton else {
            return XCTFail("'\(name)' should offer the 단계 button", file: file, line: line)
        }
        button.tap()
        XCTAssertTrue(screen.stepsSheetTitle.waitForExistence(timeout: 10), "The steps sheet should open", file: file, line: line)
        XCTAssertEqual(screen.stepsSheetTitle.label, "\(name) · 단계", file: file, line: line)
    }
}
