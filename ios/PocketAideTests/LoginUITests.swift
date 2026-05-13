import XCTest

/// End-to-end coverage for the LoginView and the full OIDC sign-in flow
/// against the local oidcmock server (started by CI before this target runs).
///
/// Pre-conditions assumed by the test environment:
///   - oidcmock listening on :5556 (issuer = http://localhost:5556)
///   - backend listening on :8080, configured to verify against oidcmock
///   - BackendBaseURL in Info-Debug.plist points to http://localhost:8080
///   - URL scheme `pocketaide-dev` registered for the redirect URI
///
/// Every test asks the app to wipe the keychain on launch via the
/// `-uitest-reset-keychain` launch argument so that we always start on the
/// LoginView, independent of state left behind by previous CI runs on the
/// same simulator UDID.
final class LoginUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    private func makeApp() -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments.append("-uitest-reset-keychain")
        return app
    }

    func testLoginViewShowsBrandAndSignInButton() {
        let app = makeApp()
        app.launch()

        XCTAssertTrue(
            app.staticTexts["BrandTitle"].waitForExistence(timeout: 10),
            "Brand title should appear on the login screen"
        )
        XCTAssertTrue(
            app.staticTexts["BrandTagline"].exists,
            "Brand tagline should appear on the login screen"
        )

        let signIn = app.buttons["SignInButton"]
        XCTAssertTrue(signIn.waitForExistence(timeout: 10))
        XCTAssertTrue(signIn.isHittable, "Sign-in button should be tappable")
    }

    func testHealthBadgeReflectsBackendOnLogin() {
        let app = makeApp()
        app.launch()

        let health = app.staticTexts["HealthStatus"]
        XCTAssertTrue(health.waitForExistence(timeout: 10), "Health status label should appear")
        XCTAssertFalse(health.label.isEmpty, "Health status label should have text")
    }

    /// Walks the full OIDC dance via the oidcmock server:
    /// 1. tap Sign-in button on LoginView
    /// 2. approve the ASWebAuthenticationSession system prompt (if shown)
    /// 3. oidcmock immediately 302-redirects to the registered URL scheme
    /// 4. assert we land on the signed-in screen with the mock subject
    /// 5. sign out and confirm we return to LoginView
    func testFullSignInFlowAgainstOidcMock() {
        let app = makeApp()

        // Add an interruption monitor for the system "wants to use … to sign in"
        // alert that ASWebAuthenticationSession raises on the first attempt.
        let monitor = addUIInterruptionMonitor(withDescription: "ASWebAuth Sign In") { alert in
            for label in ["Continue", "계속", "확인", "OK"] {
                let button = alert.buttons[label]
                if button.exists {
                    button.tap()
                    return true
                }
            }
            return false
        }
        defer { removeUIInterruptionMonitor(monitor) }

        app.launch()

        let signInButton = app.buttons["SignInButton"]
        XCTAssertTrue(signInButton.waitForExistence(timeout: 10))
        signInButton.tap()

        // The ASWebAuthenticationSession consent alert is rendered by
        // springboard, not the app, so reach for it directly. Some iOS
        // releases / re-runs skip the alert entirely (consent persisted),
        // so we treat "not present" as a no-op rather than a failure.
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let continuePredicate = NSPredicate(format: "label IN %@", ["Continue", "계속", "확인", "OK"])
        let continueButton = springboard.buttons.matching(continuePredicate).firstMatch
        if continueButton.waitForExistence(timeout: 10) {
            continueButton.tap()
        }

        let signedInLabel = app.staticTexts["SignedInLabel"]
        XCTAssertTrue(
            signedInLabel.waitForExistence(timeout: 30),
            "Should land on signed-in screen after completing the OIDC flow"
        )
        XCTAssertTrue(
            signedInLabel.label.contains("mock-user-123"),
            "Subject claim should match oidcmock default subject; got: \(signedInLabel.label)"
        )

        let signOut = app.buttons["SignOutButton"]
        XCTAssertTrue(signOut.waitForExistence(timeout: 10), "Sign-out button should appear on signed-in screen")
        signOut.tap()

        XCTAssertTrue(
            app.buttons["SignInButton"].waitForExistence(timeout: 10),
            "After signing out we should be returned to the LoginView"
        )
        XCTAssertFalse(
            app.staticTexts["SignedInLabel"].exists,
            "SignedInLabel should disappear after sign-out"
        )
    }
}
