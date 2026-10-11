import XCTest

struct ApprovalKeysScreen {
    let app: XCUIApplication

    var title: XCUIElement { app.staticTexts["approval.keys.title"] }
    var empty: XCUIElement { app.staticTexts["approval.keys.empty"] }
    var issueButton: XCUIElement { app.buttons["approval.keys.issue"] }
    var nameField: XCUIElement { app.textFields["approval.keys.issue.name"] }
    var submit: XCUIElement { app.buttons["approval.keys.issue.submit"] }
    var cancel: XCUIElement { app.buttons["approval.keys.issue.cancel"] }
    var issueNotice: XCUIElement { app.staticTexts["approval.keys.issue.notice"] }
    var issuedTitle: XCUIElement { app.staticTexts["approval.keys.issued.title"] }
    var issuedSecret: XCUIElement { app.staticTexts["approval.keys.issued.secret"] }
    var copy: XCUIElement { app.buttons["approval.keys.issued.copy"] }
    var done: XCUIElement { app.buttons["approval.keys.issued.done"] }

    func row(_ name: String) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: "approval.keys.row.\(name)").firstMatch
    }

    func meta(_ name: String) -> XCUIElement {
        app.staticTexts["approval.keys.meta.\(name)"]
    }

    func revoke(_ name: String) -> XCUIElement {
        app.buttons["approval.keys.revoke.\(name)"]
    }

    func revoked(_ name: String) -> XCUIElement {
        app.staticTexts["approval.keys.revoked.\(name)"]
    }

    static func open(in app: XCUIApplication, file: StaticString = #filePath, line: UInt = #line) -> ApprovalKeysScreen {
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "The 승인 tab should be in the tab bar", file: file, line: line)
        tab.tap()
        let settings = app.buttons["approvals.settings.button"]
        XCTAssertTrue(settings.waitForExistence(timeout: 15), "The 승인 header should offer 승인 게이트 설정", file: file, line: line)
        settings.tap()
        let screen = ApprovalKeysScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "승인 게이트 설정 should open", file: file, line: line)
        return screen
    }

    func issue(_ name: String, file: StaticString = #filePath, line: UInt = #line) {
        XCTAssertTrue(issueButton.waitForExistence(timeout: 10), "+ 호출자 키 발급 should be on screen", file: file, line: line)
        issueButton.tap()
        XCTAssertTrue(nameField.waitForExistence(timeout: 10), "The issue sheet should ask for a caller name", file: file, line: line)
        nameField.tap()
        nameField.typeText(name)
        submit.tap()
    }

    func secretIsNowhere(_ secret: String) -> Bool {
        app.descendants(matching: .any).matching(NSPredicate(format: "label CONTAINS %@ OR value CONTAINS %@", secret, secret)).count == 0
    }

    func goBack() {
        app.navigationBars.buttons.element(boundBy: 0).tap()
    }
}

enum ApprovalKeysAPI {
    static func keys(_ api: BackendAPI) -> [[String: Any]] {
        let reply = api.call("GET", "/api/approval-keys")
        XCTAssertEqual(reply.status, 200, "The caller key list should load: \(reply.text)")
        return reply.json?["keys"] as? [[String: Any]] ?? []
    }

    static func key(_ api: BackendAPI, named name: String) -> [String: Any]? {
        keys(api).first { $0["name"] as? String == name }
    }

    static func revokeAll(_ api: BackendAPI, named name: String) {
        guard let id = key(api, named: name)?["id"] as? Int else { return }
        _ = api.call("POST", "/api/approval-keys/\(id)/revoke")
    }
}

enum ApprovalUserSwitch {
    static func signOut(in app: XCUIApplication) {
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        let settings = app.buttons["prmonitor.settings.button"]
        XCTAssertTrue(settings.waitForExistence(timeout: 10), "The PR monitor header should offer 알림 설정")
        settings.tap()
        let row = app.buttons["prmonitor.settings.signOut"]
        XCTAssertTrue(row.waitForExistence(timeout: 10), "The 알림 설정 sheet should offer 로그아웃")
        row.tap()
        let confirms = app.buttons.matching(identifier: "prmonitor.settings.signOut.confirm")
        XCTAssertTrue(confirms.firstMatch.waitForExistence(timeout: 10), "로그아웃 should ask for confirmation")
        let confirm = confirms.allElementsBoundByIndex.first(where: { $0.isHittable }) ?? confirms.firstMatch
        confirm.tap()
        XCTAssertTrue(app.buttons["SignInButton"].waitForExistence(timeout: 20), "Signing out should land on the login screen")
    }

    static func signIn(in app: XCUIApplication) {
        let button = app.buttons["SignInButton"]
        XCTAssertTrue(button.waitForExistence(timeout: 20), "The login screen should offer sign-in")
        button.tap()
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let consent = springboard.buttons.matching(NSPredicate(format: "label IN %@", ["Continue", "계속", "확인", "OK"])).firstMatch
        if consent.waitForExistence(timeout: 10) {
            consent.tap()
        }
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 120), "Signing in should land on the tab shell")
    }

    static func switchUser(to subject: String, in app: XCUIApplication) {
        signOut(in: app)
        BackendAPI.setNextAppLoginSubject(subject)
        signIn(in: app)
    }

    static func restoreDefaultUser() {
        BackendAPI.setNextAppLoginSubject("")
        let app = XCUIApplication()
        app.launch()
        if app.tabBars.firstMatch.waitForExistence(timeout: 15) {
            signOut(in: app)
        }
        signIn(in: app)
        app.terminate()
    }
}
