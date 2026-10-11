import XCTest

enum AppAccount {
    static func signOut(in app: XCUIApplication) {
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        var settings = app.buttons["prmonitor.settings.button"]
        if !settings.waitForExistence(timeout: 10) {
            settings = app.buttons["알림 설정"]
        }
        XCTAssertTrue(settings.waitForExistence(timeout: 5), "The PR monitor header should offer 알림 설정")
        settings.tap()
        var row = app.buttons["prmonitor.settings.signOut"]
        if !row.waitForExistence(timeout: 10) {
            row = app.buttons.matching(NSPredicate(format: "label == %@", "로그아웃")).firstMatch
        }
        XCTAssertTrue(row.waitForExistence(timeout: 5), "The 알림 설정 sheet should offer 로그아웃")
        row.tap()
        var confirms = app.buttons.matching(identifier: "prmonitor.settings.signOut.confirm")
        if !confirms.firstMatch.waitForExistence(timeout: 10) {
            confirms = app.buttons.matching(
                NSPredicate(format: "label == %@ AND identifier != %@", "로그아웃", "prmonitor.settings.signOut")
            )
        }
        XCTAssertTrue(confirms.firstMatch.waitForExistence(timeout: 5), "로그아웃 should ask for confirmation")
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
