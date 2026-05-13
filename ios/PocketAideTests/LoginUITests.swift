import XCTest

/// End-to-end coverage that exercises the LoginView and the OIDC handshake
/// against the local oidcmock server (started by CI before this target runs).
///
/// Sign-in is performed **once** by `setUpWithError` (guarded by a static
/// flag) so that the heavy OIDC dance only happens a single time. Every
/// subsequent test launches the app on top of the keychain token left
/// behind by that one sign-in, lands directly on the signed-in screen, and
/// asserts whatever it cares about. The sign-out test is prefixed `testZ`
/// so it runs last and does not strip the token out from under siblings.
///
/// Pre-conditions assumed by the test environment:
///   - oidcmock listening on :5556 (issuer = http://localhost:5556)
///   - backend listening on :8080, configured to verify against oidcmock
///   - BackendBaseURL in Info-Debug.plist points to http://localhost:8080
///   - URL scheme `pocketaide-dev` registered for the redirect URI
final class LoginUITests: XCTestCase {
    private static var didSignIn = false

    override func setUpWithError() throws {
        continueAfterFailure = false
        if !LoginUITests.didSignIn {
            try signInOnce()
            LoginUITests.didSignIn = true
        }
    }

    /// Walks the full OIDC dance via the oidcmock server exactly once.
    /// After this returns the simulator keychain holds a token bundle and
    /// every subsequent `XCUIApplication.launch()` boots on HelloWorldView.
    private func signInOnce() throws {
        let app = XCUIApplication()

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

        // If the simulator keychain happens to already carry a token (e.g.
        // a previous test session within the same simulator boot), the app
        // skips LoginView. In that case sign-in is already a no-op.
        if app.staticTexts["SignedInLabel"].waitForExistence(timeout: 5) {
            app.terminate()
            return
        }

        let signInButton = app.buttons["SignInButton"]
        XCTAssertTrue(signInButton.waitForExistence(timeout: 30), "LoginView should expose SignInButton")
        signInButton.tap()

        // The ASWebAuthenticationSession consent alert is rendered by
        // springboard. Some iOS releases / re-runs skip it because consent
        // is persisted; treat absence as a no-op rather than a failure.
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let continuePredicate = NSPredicate(format: "label IN %@", ["Continue", "계속", "확인", "OK"])
        let continueButton = springboard.buttons.matching(continuePredicate).firstMatch
        if continueButton.waitForExistence(timeout: 10) {
            continueButton.tap()
        }

        let signedInLabel = app.staticTexts["SignedInLabel"]
        XCTAssertTrue(
            signedInLabel.waitForExistence(timeout: 30),
            "Sign-in should complete and land on HelloWorldView"
        )
        XCTAssertTrue(
            signedInLabel.label.contains("mock-user-123"),
            "Subject should match oidcmock default mock-user-123; got: \(signedInLabel.label)"
        )
        app.terminate()
    }

    func testHealthBadgeAppearsWhileSignedIn() {
        let app = XCUIApplication()
        app.launch()

        let health = app.staticTexts["HealthStatus"]
        XCTAssertTrue(health.waitForExistence(timeout: 10), "Health status label should appear on the signed-in screen")
        XCTAssertFalse(health.label.isEmpty)
    }

    func testSignedInScreenShowsMockSubject() {
        let app = XCUIApplication()
        app.launch()

        let signedInLabel = app.staticTexts["SignedInLabel"]
        XCTAssertTrue(signedInLabel.waitForExistence(timeout: 10))
        XCTAssertTrue(
            signedInLabel.label.contains("mock-user-123"),
            "Subject should match oidcmock default mock-user-123; got: \(signedInLabel.label)"
        )

        XCTAssertTrue(
            app.buttons["SignOutButton"].waitForExistence(timeout: 5),
            "SignOutButton should be present on the signed-in screen"
        )
    }

    /// Runs last (alphabetically) so it does not invalidate the token the
    /// other tests depend on. Verifies sign-out clears the keychain and
    /// returns the user to LoginView.
    func testZSignOutReturnsToLoginAndClearsToken() {
        let app = XCUIApplication()
        app.launch()

        let signOut = app.buttons["SignOutButton"]
        XCTAssertTrue(signOut.waitForExistence(timeout: 10), "SignOutButton should appear on the signed-in screen")
        signOut.tap()

        XCTAssertTrue(
            app.buttons["SignInButton"].waitForExistence(timeout: 10),
            "After signing out we should be back on LoginView with a SignInButton"
        )
        XCTAssertFalse(
            app.staticTexts["SignedInLabel"].exists,
            "SignedInLabel should disappear after sign-out"
        )
    }
}
