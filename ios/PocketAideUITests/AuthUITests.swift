// AuthUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end authentication flow:
//   App launch → Login screen → Enter credentials → Main TabBar
//   → Logout → Return to Login screen.
//
// All tests are currently skipped because LoginView and the supporting
// authentication infrastructure have not yet been implemented.
// Remove the `throw XCTSkip(...)` line in each test to activate it once the
// corresponding production UI is in place.
//
// DLD-717: 2-1: 사용자 인증 — e2e 테스트 작성 (skipped)
//
// NOTE: Authentication tests must NOT use the "--uitesting" launch argument
// because that flag causes PocketAideApp.swift to bypass the auth flow and
// navigate directly to MainTabView. These tests need the real auth flow.

import XCTest

final class AuthUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // Do NOT pass "--uitesting" here — auth flow must not be skipped.
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path

    /// On a fresh launch (no stored session), the app must show the Login
    /// screen as the root view.
    func test_appLaunch_displaysLoginScreen() throws {
        throw XCTSkip("TODO: 인증 화면 구현 후 활성화 — DLD-717")

        // Arrange — app is launched without "--uitesting"

        // Act
        let loginView = app.otherElements["login_view"]

        // Assert
        XCTAssertTrue(
            loginView.waitForExistence(timeout: 5),
            "Login screen should be the root view on first launch"
        )
    }

    /// The Login screen must contain a text field for the server address.
    func test_loginScreen_displaysServerAddressField() throws {
        throw XCTSkip("TODO: 인증 화면 구현 후 활성화 — DLD-717")

        // Arrange — app is launched, Login screen is visible
        let loginView = app.otherElements["login_view"]
        XCTAssertTrue(loginView.waitForExistence(timeout: 5))

        // Act
        let serverField = app.textFields["server_address_field"]

        // Assert
        XCTAssertTrue(
            serverField.exists,
            "Server address input field should be present on the Login screen"
        )
    }

    /// The Login screen must contain email and password input fields.
    func test_loginScreen_displaysEmailAndPasswordFields() throws {
        throw XCTSkip("TODO: 인증 화면 구현 후 활성화 — DLD-717")

        // Arrange — app is launched, Login screen is visible
        let loginView = app.otherElements["login_view"]
        XCTAssertTrue(loginView.waitForExistence(timeout: 5))

        // Act
        let emailField    = app.textFields["email_field"]
        let passwordField = app.secureTextFields["password_field"]

        // Assert
        XCTAssertTrue(emailField.exists,    "Email input field should be present")
        XCTAssertTrue(passwordField.exists, "Password input field should be present")
    }

    /// Entering valid credentials and tapping Login must navigate the user to
    /// the main TabBar.
    func test_login_withValidCredentials_navigatesToMainTabBar() throws {
        throw XCTSkip("TODO: 인증 화면 구현 후 활성화 — DLD-717")

        // Arrange
        let loginView = app.otherElements["login_view"]
        XCTAssertTrue(loginView.waitForExistence(timeout: 5))

        let serverField   = app.textFields["server_address_field"]
        let emailField    = app.textFields["email_field"]
        let passwordField = app.secureTextFields["password_field"]
        let loginButton   = app.buttons["login_button"]

        // Act
        serverField.tap()
        serverField.typeText("http://localhost:8080")

        emailField.tap()
        emailField.typeText("user@example.com")

        passwordField.tap()
        passwordField.typeText("Secret1!")

        loginButton.tap()

        // Assert
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(
            tabBar.waitForExistence(timeout: 10),
            "Main TabBar should be visible after successful login"
        )
    }

    // MARK: - Error Cases

    /// Entering invalid credentials must display an error message on the Login
    /// screen without navigating away.
    func test_login_withInvalidCredentials_displaysErrorMessage() throws {
        throw XCTSkip("TODO: 인증 화면 구현 후 활성화 — DLD-717")

        // Arrange
        let loginView = app.otherElements["login_view"]
        XCTAssertTrue(loginView.waitForExistence(timeout: 5))

        let serverField   = app.textFields["server_address_field"]
        let emailField    = app.textFields["email_field"]
        let passwordField = app.secureTextFields["password_field"]
        let loginButton   = app.buttons["login_button"]

        // Act
        serverField.tap()
        serverField.typeText("http://localhost:8080")

        emailField.tap()
        emailField.typeText("user@example.com")

        passwordField.tap()
        passwordField.typeText("wrong-password")

        loginButton.tap()

        // Assert
        let errorMessage = app.staticTexts["login_error_message"]
        XCTAssertTrue(
            errorMessage.waitForExistence(timeout: 5),
            "An error message should appear after login with invalid credentials"
        )
        // Login screen must remain visible
        XCTAssertTrue(loginView.exists, "Login screen should still be visible after failed login")
    }

    // MARK: - Logout

    /// After a successful login, tapping the Logout button must return the
    /// user to the Login screen.
    func test_logout_returnsToLoginScreen() throws {
        throw XCTSkip("TODO: 인증 화면 구현 후 활성화 — DLD-717")

        // Arrange — perform login first
        let loginView = app.otherElements["login_view"]
        XCTAssertTrue(loginView.waitForExistence(timeout: 5))

        let serverField   = app.textFields["server_address_field"]
        let emailField    = app.textFields["email_field"]
        let passwordField = app.secureTextFields["password_field"]
        let loginButton   = app.buttons["login_button"]

        serverField.tap()
        serverField.typeText("http://localhost:8080")

        emailField.tap()
        emailField.typeText("user@example.com")

        passwordField.tap()
        passwordField.typeText("Secret1!")

        loginButton.tap()

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 10), "Should be on TabBar before logout")

        // Act — navigate to the settings/profile tab and tap Logout
        // Exact tab identifier and logout button identifier depend on the
        // final UI implementation; adjust as needed.
        let settingsTab  = tabBar.buttons["tab_settings"]
        let logoutButton = app.buttons["logout_button"]

        settingsTab.tap()
        XCTAssertTrue(logoutButton.waitForExistence(timeout: 5), "Logout button should be accessible")
        logoutButton.tap()

        // Assert
        XCTAssertTrue(
            loginView.waitForExistence(timeout: 5),
            "Login screen should reappear after logout"
        )
    }
}
