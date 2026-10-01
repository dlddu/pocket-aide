// 검증 시나리오: 없음 (스모크/인프라)
import XCTest

// mock-exception: EXT — 실 OIDC IdP 대신 oidcmock 토큰으로 로그인한다 (docs/e2e-mocking-policy.md 허용목록)
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

    func testRelaunchWithStoredTokenLandsOnTabShell() {
        let app = XCUIApplication()
        app.launch()

        XCTAssertTrue(
            app.tabBars.firstMatch.waitForExistence(timeout: 15),
            "A stored oidcmock token should land on the signed-in TabView shell"
        )
        XCTAssertFalse(
            app.buttons["SignInButton"].exists,
            "LoginView's SignInButton should not be on screen while signed in"
        )
    }
}
