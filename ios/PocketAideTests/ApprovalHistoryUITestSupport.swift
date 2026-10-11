import XCTest

struct ApprovalHistoryScreen {
    static let rowPrefix = "approvals.history.row."

    let app: XCUIApplication

    var title: XCUIElement { app.staticTexts["approvals.history.title"] }
    var loadingMore: XCUIElement { app.descendants(matching: .any).matching(identifier: "approvals.history.loadingMore").firstMatch }

    func row(_ externalID: String) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: Self.rowPrefix + externalID).firstMatch
    }

    static func open(in app: XCUIApplication, file: StaticString = #filePath, line: UInt = #line) -> ApprovalHistoryScreen {
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "The 승인 tab should be in the tab bar", file: file, line: line)
        tab.tap()
        let button = app.buttons["approvals.history.button"]
        if !button.waitForExistence(timeout: 15) {
            tab.tap()
        }
        XCTAssertTrue(button.waitForExistence(timeout: 30), "The 승인 header should offer 처리 이력", file: file, line: line)
        button.tap()
        let screen = ApprovalHistoryScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "처리 이력 should open", file: file, line: line)
        return screen
    }

    func visibleRows() -> [(id: String, minY: CGFloat)] {
        let elements = app.descendants(matching: .any)
            .matching(NSPredicate(format: "identifier BEGINSWITH %@", Self.rowPrefix))
            .allElementsBoundByIndex
        return elements
            .filter { $0.exists && $0.frame.height > 0 }
            .map { (id: String($0.identifier.dropFirst(Self.rowPrefix.count)), minY: $0.frame.minY) }
            .sorted { $0.minY < $1.minY }
    }

    func scrollThrough(expecting count: Int, maxSwipes: Int = 40) -> (order: [String], duplicates: [String]) {
        var order: [String] = []
        var duplicates: [String] = []
        let list = app.collectionViews.firstMatch
        for _ in 0...maxSwipes {
            var frames: [String: CGFloat] = [:]
            for row in visibleRows() {
                if let seen = frames[row.id], abs(seen - row.minY) > 1 {
                    duplicates.append(row.id)
                }
                frames[row.id] = row.minY
                if !order.contains(row.id) {
                    order.append(row.id)
                }
            }
            if order.count >= count && !loadingMore.exists {
                break
            }
            list.swipeUp()
        }
        return (order, duplicates)
    }

    func goBack() {
        app.navigationBars.buttons.element(boundBy: 0).tap()
    }
}

enum ApprovalHistoryAPI {
    static func externalIDs(_ api: BackendAPI) -> [String] {
        var ids: [String] = []
        var cursor: String?
        for _ in 0..<50 {
            var path = "/api/approvals/history"
            if let cursor, let encoded = cursor.addingPercentEncoding(withAllowedCharacters: .alphanumerics) {
                path += "?cursor=" + encoded
            }
            let reply = api.call("GET", path)
            XCTAssertEqual(reply.status, 200, "The history should load: \(reply.text)")
            let items = reply.json?["items"] as? [[String: Any]] ?? []
            ids += items.compactMap { $0["external_id"] as? String }
            cursor = reply.json?["next_cursor"] as? String
            if cursor == nil { break }
        }
        return ids
    }

    static func item(_ api: BackendAPI, externalID: String) -> [String: Any]? {
        let reply = api.call("GET", "/api/approvals/history")
        let items = reply.json?["items"] as? [[String: Any]] ?? []
        return items.first { $0["external_id"] as? String == externalID }
    }
}
