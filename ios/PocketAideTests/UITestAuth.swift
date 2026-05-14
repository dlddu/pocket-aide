import XCTest

/// Shared OIDC sign-in helper used by every UI test class.
///
/// Sign-in is heavy (`ASWebAuthenticationSession` round trip against the
/// oidcmock server) and the resulting token lives in the simulator keychain
/// which persists across tests. We perform the dance exactly once per UI test
/// process via a static guard, then every test launches on top of that token.
///
/// IMPORTANT: this helper launches the app under
/// `UI_TESTS_USE_LEGACY_HOME=1` for sign-in only, because the verification it
/// runs (`SignedInLabel.waitForExistence`) targets `HelloWorldView`. Per-test
/// launches are free to drop that env var and exercise the new TabView shell.
enum UITestAuth {
    private static var didSignIn = false

    static func ensureSignedIn(_ testCase: XCTestCase) {
        guard !didSignIn else { return }
        let app = XCUIApplication()
        app.launchEnvironment["UI_TESTS_USE_LEGACY_HOME"] = "1"

        let monitor = testCase.addUIInterruptionMonitor(withDescription: "ASWebAuth Sign In") { alert in
            for label in ["Continue", "계속", "확인", "OK"] {
                let button = alert.buttons[label]
                if button.exists {
                    button.tap()
                    return true
                }
            }
            return false
        }
        defer { testCase.removeUIInterruptionMonitor(monitor) }

        app.launch()

        // If a prior session already left a token in the keychain we're done.
        if app.staticTexts["SignedInLabel"].waitForExistence(timeout: 5) {
            didSignIn = true
            app.terminate()
            return
        }

        let signInButton = app.buttons["SignInButton"]
        XCTAssertTrue(signInButton.waitForExistence(timeout: 30), "LoginView should expose SignInButton")
        signInButton.tap()

        // Springboard's consent prompt is skipped on re-runs (consent
        // persisted) — treat absence as a no-op.
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let continuePredicate = NSPredicate(format: "label IN %@", ["Continue", "계속", "확인", "OK"])
        let continueButton = springboard.buttons.matching(continuePredicate).firstMatch
        if continueButton.waitForExistence(timeout: 10) {
            continueButton.tap()
        }

        let signedInLabel = app.staticTexts["SignedInLabel"]
        XCTAssertTrue(
            signedInLabel.waitForExistence(timeout: 120),
            "Sign-in should complete and land on HelloWorldView"
        )
        XCTAssertTrue(
            signedInLabel.label.contains("mock-user-123"),
            "Subject should match oidcmock default mock-user-123; got: \(signedInLabel.label)"
        )
        didSignIn = true
        app.terminate()
    }
}
