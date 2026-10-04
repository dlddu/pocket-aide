import XCTest

enum TodoUIArea: String, CaseIterable {
    case personal, work

    var tabLabel: String {
        switch self {
        case .personal: return "개인"
        case .work: return "회사"
        }
    }

    var screenTitle: String {
        switch self {
        case .personal: return "내 일"
        case .work: return "회사"
        }
    }

    var openSectionTitle: String {
        switch self {
        case .personal: return "할 일"
        case .work: return "OPEN"
        }
    }

    var doneSectionTitle: String {
        switch self {
        case .personal: return "완료"
        case .work: return "DONE"
        }
    }

    var other: TodoUIArea {
        switch self {
        case .personal: return .work
        case .work: return .personal
        }
    }
}

struct TodoCounts: Equatable {
    let open: Int
    let done: Int
}

enum TodoUI {
    static func uniqueToken() -> String {
        String(UUID().uuidString.prefix(6))
    }

    static func launch() -> XCUIApplication {
        let app = XCUIApplication()
        app.launch()
        return app
    }

    static func relaunch(_ app: XCUIApplication) {
        app.terminate()
        app.launch()
    }

    static func dismissKeyboard(in app: XCUIApplication) {
        guard app.keyboards.firstMatch.exists else { return }
        let keys = NSPredicate(format: "label IN %@", ["return", "Return", "search", "Search", "done", "Done", "완료", "검색", "확인"])
        let key = app.keyboards.buttons.matching(keys).firstMatch
        if key.exists {
            key.tap()
        } else {
            app.typeText("\n")
        }
        _ = waitToDisappear(app.keyboards.firstMatch, timeout: 5)
    }

    static func clear(_ field: XCUIElement) {
        let current = (field.value as? String) ?? ""
        let placeholder = field.placeholderValue ?? ""
        let length = current == placeholder ? 0 : current.count
        field.coordinate(withNormalizedOffset: CGVector(dx: 0.97, dy: 0.5)).tap()
        field.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: max(length * 3, 24) + 4))
    }

    static func waitToDisappear(_ element: XCUIElement, timeout: TimeInterval = 10) -> Bool {
        let expectation = XCTNSPredicateExpectation(
            predicate: NSPredicate(format: "exists == false"),
            object: element
        )
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }
}

struct TodoScreen {
    let app: XCUIApplication
    let area: TodoUIArea

    private var prefix: String { "todos.\(area.rawValue)" }

    @discardableResult
    static func open(_ area: TodoUIArea, in app: XCUIApplication) -> TodoScreen {
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let direct = tabBar.buttons[area.tabLabel]
        if direct.exists {
            direct.tap()
        } else {
            let more = tabBar.buttons["More"]
            XCTAssertTrue(more.exists, "Neither the '\(area.tabLabel)' tab nor 'More' is present")
            more.tap()
            if !tapMoreRow(area.tabLabel, in: app) {
                let back = app.navigationBars.buttons.firstMatch
                if back.exists {
                    back.tap()
                } else {
                    more.tap()
                }
                XCTAssertTrue(tapMoreRow(area.tabLabel, in: app), "'\(area.tabLabel)' row should be listed under More")
            }
        }
        let screen = TodoScreen(app: app, area: area)
        XCTAssertTrue(
            screen.header.waitForExistence(timeout: 15),
            "The \(area.rawValue) todo screen header '\(area.screenTitle)' should appear"
        )
        XCTAssertTrue(screen.addButton.waitForExistence(timeout: 10), "The add button should be on the \(area.rawValue) screen")
        return screen
    }

    private static func tapMoreRow(_ label: String, in app: XCUIApplication) -> Bool {
        let candidates: [XCUIElement] = [
            app.tables.staticTexts[label],
            app.collectionViews.staticTexts[label],
            app.tables.cells[label],
            app.collectionViews.cells[label],
        ]
        for candidate in candidates where candidate.waitForExistence(timeout: 3) {
            candidate.tap()
            return true
        }
        return false
    }

    var header: XCUIElement {
        app.staticTexts.matching(
            NSPredicate(format: "identifier == %@ AND label == %@", "screen.header.title", area.screenTitle)
        ).firstMatch
    }

    var addButton: XCUIElement { app.buttons["\(prefix).add.button"] }
    var summary: XCUIElement { app.staticTexts["\(prefix).summary"] }
    var searchField: XCUIElement { app.textFields["\(prefix).search.field"] }
    var list: XCUIElement { app.collectionViews.firstMatch }

    func row(_ title: String) -> XCUIElement {
        app.staticTexts.matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label == %@", "\(prefix).row.", title)
        ).firstMatch
    }

    func rowID(_ title: String) -> String? {
        let element = row(title)
        guard element.waitForExistence(timeout: 10) else { return nil }
        let marker = "\(prefix).row."
        return String(element.identifier.dropFirst(marker.count))
    }

    func meta(_ title: String) -> XCUIElement? {
        guard let id = rowID(title) else { return nil }
        return app.staticTexts["\(prefix).row.\(id).meta"]
    }

    func toggle(_ title: String) -> XCUIElement? {
        guard let id = rowID(title) else { return nil }
        return app.buttons["\(prefix).row.\(id).toggle"]
    }

    func sectionHeader(done: Bool) -> XCUIElement {
        let title = done ? area.doneSectionTitle : area.openSectionTitle
        return app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", "\(title) · ")).firstMatch
    }

    func counts() -> TodoCounts? {
        guard summary.waitForExistence(timeout: 10) else { return nil }
        let numbers = summary.label
            .split(whereSeparator: { !$0.isNumber })
            .compactMap { Int($0) }
        guard numbers.count == 2 else { return nil }
        return TodoCounts(open: numbers[0], done: numbers[1])
    }

    func waitForCounts(_ expected: TodoCounts, timeout: TimeInterval = 10) -> Bool {
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline {
            if counts() == expected { return true }
            Thread.sleep(forTimeInterval: 0.5)
        }
        return counts() == expected
    }

    @discardableResult
    func reveal(_ element: XCUIElement, attempts: Int = 8) -> Bool {
        for _ in 0..<attempts {
            if element.exists && element.isHittable { return true }
            list.swipeUp()
        }
        return element.exists && element.isHittable
    }

    func pullToRefresh() {
        let start = list.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.15))
        let end = list.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.85))
        start.press(forDuration: 0.1, thenDragTo: end)
    }

    func search(_ text: String) {
        XCTAssertTrue(searchField.waitForExistence(timeout: 5), "The search field should be on the \(area.rawValue) screen")
        searchField.tap()
        clearText(searchField)
        if !text.isEmpty {
            searchField.typeText(text)
        }
    }

    func clearSearch() {
        searchField.tap()
        clearText(searchField)
    }

    func dismissKeyboard() {
        TodoUI.dismissKeyboard(in: app)
    }

    func clearText(_ field: XCUIElement) {
        TodoUI.clear(field)
    }

    func assertListing(
        token: String,
        present: [String],
        absent: [String],
        file: StaticString = #filePath,
        line: UInt = #line
    ) {
        search(token)
        for title in present {
            XCTAssertTrue(
                row(title).waitForExistence(timeout: 10),
                "'\(title)' should be listed in the \(area.rawValue) area",
                file: file,
                line: line
            )
        }
        for title in absent {
            XCTAssertFalse(
                row(title).waitForExistence(timeout: present.isEmpty ? 3 : 1),
                "'\(title)' must not be listed in the \(area.rawValue) area",
                file: file,
                line: line
            )
        }
        clearSearch()
        dismissKeyboard()
    }

    func buttonLabels(in container: XCUIElement) -> [String] {
        container.buttons.allElementsBoundByIndex
            .filter { !$0.identifier.hasSuffix(".toggle") }
            .map(\.label)
            .sorted()
    }

    func openCreateSheet() -> TodoSheet {
        addButton.tap()
        let sheet = TodoSheet(app: app)
        XCTAssertTrue(sheet.title.waitForExistence(timeout: 10), "The new-todo sheet should open")
        XCTAssertEqual(sheet.title.label, "새 할 일")
        return sheet
    }

    func openEditSheet(_ title: String) -> TodoSheet {
        let element = row(title)
        XCTAssertTrue(element.waitForExistence(timeout: 10), "'\(title)' should be listed before editing")
        reveal(element)
        element.tap()
        let sheet = TodoSheet(app: app)
        XCTAssertTrue(sheet.title.waitForExistence(timeout: 10), "The edit sheet should open for '\(title)'")
        XCTAssertEqual(sheet.title.label, "할 일 편집")
        return sheet
    }

    func add(title: String, memo: String? = nil, due: Bool = false, priority: String? = nil) {
        let sheet = openCreateSheet()
        sheet.setTitle(title)
        if let memo {
            sheet.setMemo(memo)
        }
        if due {
            sheet.setDueDate(on: true)
        }
        if let priority {
            sheet.pickPriority(priority)
        }
        sheet.save()
        XCTAssertTrue(row(title).waitForExistence(timeout: 10), "'\(title)' should be listed after saving")
    }

    static func today() -> String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = .current
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.string(from: Date())
    }

    func priorityLabel(_ raw: String) -> String {
        if area == .work { return raw.uppercased() }
        switch raw {
        case "high": return "우선"
        case "normal": return "보통"
        default: return "낮음"
        }
    }
}
