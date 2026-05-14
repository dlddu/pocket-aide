import XCTest

/// End-to-end coverage that exercises the LoginView and the OIDC handshake
/// against the local oidcmock server (started by CI before this target runs).
///
/// The heavy OIDC dance is performed exactly once across the entire UI test
/// process by `UITestAuth.ensureSignedIn` (shared with AffirmationsUITests).
/// Each test in this class launches on top of that keychain token, lands
/// directly on the signed-in screen, and asserts whatever it cares about.
/// `testZSignOutReturnsToLoginAndClearsToken` is prefixed `testZ` so it runs
/// last alphabetically.
///
/// Every per-test launch sets `UI_TESTS_USE_LEGACY_HOME=1` so RootView keeps
/// rendering HelloWorldView (where the legacy identifiers live), independent
/// of the new TabView shell.
///
/// Pre-conditions assumed by the test environment:
///   - oidcmock listening on :5556 (issuer = http://localhost:5556)
///   - backend listening on :8080, configured to verify against oidcmock
///   - BackendBaseURL in Info-Debug.plist points to http://localhost:8080
///   - URL scheme `pocketaide-dev` registered for the redirect URI
final class LoginUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHealthBadgeAppearsWhileSignedIn() {
        let app = XCUIApplication()
        app.launchEnvironment["UI_TESTS_USE_LEGACY_HOME"] = "1"
        app.launch()

        let health = app.staticTexts["HealthStatus"]
        XCTAssertTrue(health.waitForExistence(timeout: 10), "Health status label should appear on the signed-in screen")
        XCTAssertFalse(health.label.isEmpty)
    }

    func testSignedInScreenShowsMockSubject() {
        let app = XCUIApplication()
        app.launchEnvironment["UI_TESTS_USE_LEGACY_HOME"] = "1"
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
        app.launchEnvironment["UI_TESTS_USE_LEGACY_HOME"] = "1"
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
