import XCTest

struct ApprovalDetailScreen {
    let app: XCUIApplication

    var title: XCUIElement { app.staticTexts["approval.detail.title"] }
    var countdown: XCUIElement { app.staticTexts["approval.detail.countdown"] }
    var status: XCUIElement { app.staticTexts["approval.detail.status"] }
    var mode: XCUIElement { app.staticTexts["approval.detail.mode"] }
    var processedAt: XCUIElement { app.staticTexts["approval.detail.processedAt"] }
    var notice: XCUIElement { app.staticTexts["approval.detail.notice"] }
    var approve: XCUIElement { app.buttons["approval.detail.approve"] }
    var reject: XCUIElement { app.buttons["approval.detail.reject"] }

    static func open(_ externalID: String, in app: XCUIApplication, file: StaticString = #filePath, line: UInt = #line) -> ApprovalDetailScreen {
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "The 승인 tab should be in the tab bar", file: file, line: line)
        tab.tap()
        let list = ApprovalsScreen(app: app)
        XCTAssertTrue(list.title.waitForExistence(timeout: 15), "The 승인 screen should open", file: file, line: line)
        let row = list.row(externalID)
        if !row.waitForExistence(timeout: 5) {
            list.pullToRefresh()
        }
        XCTAssertTrue(row.waitForExistence(timeout: 15), "\(externalID) should be listed", file: file, line: line)
        row.tap()
        let screen = ApprovalDetailScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "The request detail should open", file: file, line: line)
        return screen
    }

    func waitForStatus(_ label: String, timeout: TimeInterval = 15) -> Bool {
        RoutinesScreen.eventually(timeout: timeout) { status.exists && status.label == label }
    }

    func choose(_ decision: XCUIElement, thenTap label: String, file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(decision.waitForExistence(timeout: 5), "The decision button should be on screen", file: file, line: line)
        decision.tap()
        let alert = app.alerts.firstMatch
        XCTAssertTrue(alert.waitForExistence(timeout: 5), "A confirmation step should appear", file: file, line: line)
        alert.buttons[label].tap()
        XCTAssertTrue(TodoUI.waitToDisappear(alert, timeout: 5), "The confirmation should close", file: file, line: line)
    }

    func goBack() {
        app.navigationBars.buttons.element(boundBy: 0).tap()
    }
}

enum ApprovalLookup {
    static func status(_ api: BackendAPI, _ id: String) -> String? {
        api.call("GET", "/api/approvals/\(id)").json?["status"] as? String
    }

    static func callerView(_ caller: ApprovalCaller, _ id: String) -> [String: Any]? {
        BackendAPI(token: caller.key).call("GET", "/api/external/approvals/\(id)").json
    }
}
