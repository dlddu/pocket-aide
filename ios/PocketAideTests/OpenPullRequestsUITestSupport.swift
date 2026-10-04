import XCTest

enum OpenPullRequestsUI {
    enum Token {
        static let valid = "ghp_e2e_valid"
        static let expires = "ghp_e2e_expires"
        static let rateLimited = "ghp_e2e_ratelimited"
    }

    static let handle = "@pocket-aide-e2e"

    static func launch() -> XCUIApplication {
        let app = XCUIApplication()
        // mock-exception: EXT — 실 GitHub 은 전용 테스트 계정 PAT 가 CI 시크릿에 없어 E2E 가 부를 수 없다; 앱의 GitHubClient 를 로컬 GitHub API 스텁으로 향하게 한다 (docs/e2e-mocking-policy.md)
        app.launchEnvironment["GITHUB_API_BASE_URL"] = "http://localhost:5557"
        app.launch()
        return app
    }

    static func openSheet(in app: XCUIApplication) {
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        let open = app.buttons["prmonitor.openprs.button"]
        XCTAssertTrue(open.waitForExistence(timeout: 10), "PR 모니터 header should offer the open-PR sheet")
        open.tap()
        XCTAssertTrue(app.navigationBars["열린 PR"].waitForExistence(timeout: 10), "The open-PR sheet should be presented")
    }

    static func closeSheet(in app: XCUIApplication) {
        let done = app.navigationBars["열린 PR"].buttons["완료"]
        XCTAssertTrue(done.waitForExistence(timeout: 5), "The open-PR sheet should offer 완료")
        done.tap()
        XCTAssertTrue(app.buttons["prmonitor.openprs.button"].waitForExistence(timeout: 10), "완료 should return to PR 모니터")
    }

    static func launchDisconnected() -> XCUIApplication {
        let app = launch()
        openSheet(in: app)
        let disconnect = app.buttons["openprs.disconnect.button"]
        if disconnect.waitForExistence(timeout: 3) {
            disconnect.tap()
        }
        XCTAssertTrue(
            app.secureTextFields["openprs.token.field"].waitForExistence(timeout: 10),
            "A disconnected sheet should show the token form"
        )
        return app
    }

    static func submit(_ token: String, in app: XCUIApplication) {
        let field = app.secureTextFields["openprs.token.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 10), "Token field should be shown")
        field.tap()
        field.typeText(token)
        app.buttons["openprs.connect.button"].tap()
    }

    static func element(_ identifier: String, in app: XCUIApplication) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: identifier).firstMatch
    }

    static func row(_ number: Int, in app: XCUIApplication) -> XCUIElement {
        element("openprs.row.dlddu/pocket-aide-e2e#\(number)", in: app)
    }

    static func rows(in app: XCUIApplication) -> XCUIElementQuery {
        app.descendants(matching: .any).matching(NSPredicate(format: "identifier BEGINSWITH %@", "openprs.row."))
    }

    static func selectFilter(_ rawValue: String, in app: XCUIApplication) {
        let pill = app.buttons["filter.pill.\(rawValue)"]
        XCTAssertTrue(pill.waitForExistence(timeout: 5), "Filter \(rawValue) should be offered")
        if !pill.isSelected {
            pill.tap()
        }
        XCTAssertTrue(pill.isSelected, "Filter \(rawValue) should be selected")
    }
}
