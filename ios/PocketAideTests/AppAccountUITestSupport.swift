import XCTest

enum AppAccount {
    static func signOut(in app: XCUIApplication) {
        XCTAssertTrue(trySignOut(in: app, settle: 20), "Signing out should land on the login screen")
    }

    @discardableResult
    static func trySignOut(in app: XCUIApplication, settle: TimeInterval) -> Bool {
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        guard tab.waitForExistence(timeout: 15) else { return false }
        tab.tap()
        var settings = app.buttons["prmonitor.settings.button"]
        if !settings.waitForExistence(timeout: 10) {
            settings = app.buttons["알림 설정"]
        }
        guard settings.waitForExistence(timeout: settle) else { return false }
        settings.tap()
        var row = app.buttons["prmonitor.settings.signOut"]
        if !row.waitForExistence(timeout: 10) {
            row = app.buttons.matching(NSPredicate(format: "label == %@", "로그아웃")).firstMatch
        }
        guard row.waitForExistence(timeout: settle) else { return false }
        row.tap()
        var confirms = app.buttons.matching(identifier: "prmonitor.settings.signOut.confirm")
        if !confirms.firstMatch.waitForExistence(timeout: 10) {
            confirms = app.buttons.matching(
                NSPredicate(format: "label == %@ AND identifier != %@", "로그아웃", "prmonitor.settings.signOut")
            )
        }
        guard confirms.firstMatch.waitForExistence(timeout: settle) else { return false }
        let confirm = confirms.allElementsBoundByIndex.first(where: { $0.isHittable }) ?? confirms.firstMatch
        confirm.tap()
        return app.buttons["SignInButton"].waitForExistence(timeout: settle)
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
        for _ in 0..<3 {
            app.launch()
            let signedIn = app.tabBars.firstMatch.waitForExistence(timeout: 15)
            if !signedIn || trySignOut(in: app, settle: 60) {
                signIn(in: app)
                app.terminate()
                return
            }
            app.terminate()
        }
        XCTFail("The app should sign out to restore the default user within three launches")
    }
}
