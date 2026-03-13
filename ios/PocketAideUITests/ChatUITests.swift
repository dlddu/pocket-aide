// ChatUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end AI chat (text) flow:
//   Tap Assistant tab → Chat interface visible → Type message → Send
//   → User bubble displayed → AI response bubble displayed
//   → Open model selector → Change model → Re-send → Response confirmed
//   → Switch tabs and return → Chat history persists.
//
// DLD-719: 3-1: AI 채팅 (텍스트) — e2e 테스트 작성 (skipped)
//
// NOTE: All tests in this file are skipped via `throw XCTSkip(...)` because
// the AssistantView chat UI (message input, send button, response bubbles,
// model selector) has not yet been implemented. Remove the XCTSkip throw in
// each test once the corresponding UI is ready.

import XCTest

final class ChatUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // Pass "--uitesting" so the app bypasses the auth flow and lands
        // directly on MainTabView, matching the pattern used by TabBarUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Chat Interface

    /// Tapping the Assistant tab must display the chat interface container view.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_assistant" → "assistant_chat_view" appears
    func test_assistantTab_displaysChatInterface() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")

        // Act
        tabBar.buttons["tab_assistant"].tap()

        // Assert
        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(
            chatView.waitForExistence(timeout: 5),
            "Chat interface container (assistant_chat_view) should be visible after tapping Assistant tab"
        )
    }

    /// The chat message input field must be present and editable after
    /// navigating to the Assistant tab.
    ///
    /// Expected flow:
    ///   "tab_assistant" selected → "chat_input_field" exists and is enabled
    func test_chatInput_isVisibleAndEditable() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        // Act
        let inputField = app.textFields["chat_input_field"]

        // Assert
        XCTAssertTrue(
            inputField.waitForExistence(timeout: 5),
            "Chat input field (chat_input_field) should be visible"
        )
        XCTAssertTrue(inputField.isEnabled, "Chat input field should be enabled for editing")
    }

    /// After typing a message and tapping Send, the message must appear as a
    /// user bubble in the conversation list.
    ///
    /// Expected flow:
    ///   Type "Hello AI" → tap "send_button" → "user_message_bubble" appears
    ///   with text "Hello AI"
    func test_sendMessage_displaysUserBubble() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let inputField  = app.textFields["chat_input_field"]
        let sendButton  = app.buttons["send_button"]
        XCTAssertTrue(inputField.waitForExistence(timeout: 5), "Input field must exist before sending")
        XCTAssertTrue(sendButton.exists, "Send button must exist before sending")

        // Act
        inputField.tap()
        inputField.typeText("Hello AI")
        sendButton.tap()

        // Assert
        let userBubble = app.otherElements["user_message_bubble"]
        XCTAssertTrue(
            userBubble.waitForExistence(timeout: 5),
            "User message bubble should appear after sending a message"
        )
        XCTAssertTrue(
            app.staticTexts["Hello AI"].exists,
            "The sent message text 'Hello AI' should be visible in the user bubble"
        )
    }

    /// After the user sends a message, an AI response bubble must appear in
    /// the conversation list.
    ///
    /// Expected flow:
    ///   Message sent → "ai_response_bubble" appears within timeout
    func test_sendMessage_displaysAIResponseBubble() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let inputField = app.textFields["chat_input_field"]
        let sendButton = app.buttons["send_button"]
        XCTAssertTrue(inputField.waitForExistence(timeout: 5))

        // Act
        inputField.tap()
        inputField.typeText("What is the weather today?")
        sendButton.tap()

        // Assert
        let aiBubble = app.otherElements["ai_response_bubble"]
        XCTAssertTrue(
            aiBubble.waitForExistence(timeout: 15),
            "AI response bubble (ai_response_bubble) should appear after the model responds"
        )
    }

    // MARK: - Model Selector

    /// The model selector control must be accessible from the chat interface.
    ///
    /// Expected flow:
    ///   Chat interface visible → "model_selector_button" exists and is hittable
    func test_modelSelector_isAccessible() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        // Act
        let modelSelectorButton = app.buttons["model_selector_button"]

        // Assert
        XCTAssertTrue(
            modelSelectorButton.waitForExistence(timeout: 5),
            "Model selector button (model_selector_button) should be present in the chat interface"
        )
        XCTAssertTrue(modelSelectorButton.isHittable, "Model selector button must be hittable")
    }

    /// Tapping the model selector and choosing a different model must update
    /// the displayed model name in the chat interface.
    ///
    /// Expected flow:
    ///   Tap "model_selector_button" → model picker appears →
    ///   select second model option → picker dismisses →
    ///   "selected_model_label" shows new model name
    func test_modelSelector_changesModel() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let modelSelectorButton = app.buttons["model_selector_button"]
        XCTAssertTrue(modelSelectorButton.waitForExistence(timeout: 5))

        // Capture the initial model label before changing
        let modelLabel = app.staticTexts["selected_model_label"]
        let initialModel = modelLabel.label

        // Act
        modelSelectorButton.tap()

        // The picker lists available models; select the second option
        let modelPicker = app.otherElements["model_picker"]
        XCTAssertTrue(modelPicker.waitForExistence(timeout: 5), "Model picker should appear after tapping selector")
        let secondModelOption = modelPicker.buttons.element(boundBy: 1)
        XCTAssertTrue(secondModelOption.exists, "At least two model options must be available")
        secondModelOption.tap()

        // Assert — picker should be dismissed and label should have changed
        XCTAssertFalse(modelPicker.exists, "Model picker should be dismissed after selection")
        XCTAssertTrue(
            modelLabel.waitForExistence(timeout: 3),
            "Selected model label (selected_model_label) should still be visible"
        )
        XCTAssertNotEqual(
            modelLabel.label,
            initialModel,
            "Selected model label should change after picking a different model"
        )
    }

    // MARK: - Model Change + Re-send

    /// After changing the model, sending a new message must use the newly
    /// selected model and display an AI response bubble.
    ///
    /// Expected flow:
    ///   Change model → type new message → send → AI response bubble appears
    func test_sendMessage_afterModelChange_usesNewModel() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        // Change the model first
        let modelSelectorButton = app.buttons["model_selector_button"]
        XCTAssertTrue(modelSelectorButton.waitForExistence(timeout: 5))
        modelSelectorButton.tap()

        let modelPicker = app.otherElements["model_picker"]
        XCTAssertTrue(modelPicker.waitForExistence(timeout: 5))
        let secondModelOption = modelPicker.buttons.element(boundBy: 1)
        XCTAssertTrue(secondModelOption.exists)
        let newModelName = secondModelOption.label
        secondModelOption.tap()

        // Act — send a message with the new model selected
        let inputField = app.textFields["chat_input_field"]
        let sendButton = app.buttons["send_button"]
        XCTAssertTrue(inputField.waitForExistence(timeout: 5))

        inputField.tap()
        inputField.typeText("Tell me a joke")
        sendButton.tap()

        // Assert — AI response bubble must appear indicating the model responded
        let aiBubble = app.otherElements["ai_response_bubble"]
        XCTAssertTrue(
            aiBubble.waitForExistence(timeout: 15),
            "AI response bubble should appear after sending with model '\(newModelName)'"
        )

        // The selected model label must still reflect the chosen model
        let modelLabel = app.staticTexts["selected_model_label"]
        XCTAssertEqual(
            modelLabel.label,
            newModelName,
            "Model label should remain '\(newModelName)' after sending"
        )
    }

    // MARK: - Chat History Persistence

    /// Chat messages must persist when the user switches to another tab and
    /// returns to the Assistant tab.
    ///
    /// Expected flow:
    ///   Send message → AI responds → tap "tab_home" → tap "tab_assistant"
    ///   → previous user bubble and AI bubble are still visible
    func test_chatHistory_persistsAcrossTabSwitches() throws {
        throw XCTSkip("AssistantView chat UI not yet implemented — activate after DLD-719")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let inputField = app.textFields["chat_input_field"]
        let sendButton = app.buttons["send_button"]
        XCTAssertTrue(inputField.waitForExistence(timeout: 5))

        // Send a message and wait for AI response
        inputField.tap()
        inputField.typeText("Remember this message")
        sendButton.tap()

        let aiBubble = app.otherElements["ai_response_bubble"]
        XCTAssertTrue(aiBubble.waitForExistence(timeout: 15), "AI must respond before switching tabs")

        // Act — switch away and come back
        tabBar.buttons["tab_home"].tap()

        let homeView = app.otherElements["tab_home_view"]
        XCTAssertTrue(homeView.waitForExistence(timeout: 5), "Home view must appear after switching tabs")

        tabBar.buttons["tab_assistant"].tap()

        // Assert — previous messages must still be visible
        XCTAssertTrue(
            chatView.waitForExistence(timeout: 5),
            "Chat interface should reappear after returning to Assistant tab"
        )
        XCTAssertTrue(
            app.staticTexts["Remember this message"].exists,
            "Previously sent message should still be visible after tab switch"
        )
        XCTAssertTrue(
            aiBubble.exists,
            "AI response bubble should persist after returning to the Assistant tab"
        )
    }
}
