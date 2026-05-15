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

        // First attempt: launch, optionally run the OIDC dance, watch for
        // SignedInLabel.
        if attemptSignIn(longWait: 120) {
            didSignIn = true
            return
        }

        // Fallback: the OIDC callback may have written a token to the keychain
        // even though we missed the SignedInLabel render window (CI cold-start
        // can keep the simulator stalled past the assertion timeout). A fresh
        // launch reads the keychain and lands signed-in directly, so we get a
        // cheap retry without re-doing the browser dance.
        let attachment = XCTAttachment(string: "UITestAuth: retrying sign-in after first attempt missed SignedInLabel")
        attachment.lifetime = .keepAlways
        testCase.add(attachment)

        if attemptSignIn(longWait: 60) {
            didSignIn = true
            return
        }

        XCTFail("Sign-in did not complete after two attempts (OIDC dance never produced SignedInLabel)")
    }

    /// One sign-in attempt: launch the app, if already signed in return true,
    /// otherwise run the OIDC dance and wait `longWait` seconds for
    /// SignedInLabel. Returns whether the label was seen.
    private static func attemptSignIn(longWait: TimeInterval) -> Bool {
        let app = XCUIApplication()
        app.launchEnvironment["UI_TESTS_USE_LEGACY_HOME"] = "1"
        app.launch()

        if app.staticTexts["SignedInLabel"].waitForExistence(timeout: 5) {
            app.terminate()
            return true
        }

        // Run the OIDC dance only if LoginView is actually showing — if the
        // keychain seeded a token mid-launch we may never see the button.
        let signInButton = app.buttons["SignInButton"]
        if signInButton.waitForExistence(timeout: 15) {
            signInButton.tap()

            // Springboard's consent prompt is skipped on re-runs (consent
            // persisted) — treat absence as a no-op.
            let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
            let continuePredicate = NSPredicate(format: "label IN %@", ["Continue", "계속", "확인", "OK"])
            let continueButton = springboard.buttons.matching(continuePredicate).firstMatch
            if continueButton.waitForExistence(timeout: 10) {
                continueButton.tap()
            }
        }

        let seen = app.staticTexts["SignedInLabel"].waitForExistence(timeout: longWait)
        app.terminate()
        return seen
    }
}
