// VoiceChatUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end AI voice chat flow:
//   Tap Assistant tab → Mic button visible → Tap mic button
//   → Voice input indicator appears → Speech recognition result fills input
//   → Send → User bubble displayed → AI response bubble displayed
//   → Tap mic again to stop recording → Text preserved
//   → Open Settings → Change speech engine → Verify selection persists.
//
// DLD-721: 4-1: AI 채팅 (음성) — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-721:
//   - ChatViewModel integrates SpeechRecognizerProtocol
//   - mic_button and voice_input_indicator are wired up in AssistantView
//   - SettingsView exposes speech_engine_selector / settings_speech_engine_picker
//   - "--uitesting" launch argument injects MockSpeechRecognizer

import XCTest

final class VoiceChatUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow (same as ChatUITests) and
        // additionally instructs the app to inject MockSpeechRecognizer so
        // that voice recognition can be driven deterministically by the UI test.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: Mic Button Presence

    /// The microphone button must be visible on the chat interface so the user
    /// can initiate voice input without navigating away.
    ///
    /// Expected flow:
    ///   TabBar visible → tap "tab_assistant" → "assistant_chat_view" appears
    ///   → "mic_button" exists and is hittable
    func test_chatInterface_displaysMicButton() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before navigating")
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5), "Chat interface must appear before checking mic button")

        // Act
        let micButton = app.buttons["mic_button"]

        // Assert
        XCTAssertTrue(
            micButton.waitForExistence(timeout: 5),
            "Microphone button (mic_button) should be visible in the chat interface"
        )
        XCTAssertTrue(micButton.isHittable, "Microphone button must be hittable")
    }

    // MARK: - Happy Path: Recording Indicator

    /// Tapping the mic button must show a visual indicator that recording is
    /// in progress, giving the user clear feedback that voice input is active.
    ///
    /// Expected flow:
    ///   Chat interface visible → tap "mic_button"
    ///   → "voice_input_indicator" appears within timeout
    func test_micButton_tap_showsVoiceInputIndicator() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let micButton = app.buttons["mic_button"]
        XCTAssertTrue(micButton.waitForExistence(timeout: 5), "Mic button must exist before tapping")

        // Act
        micButton.tap()

        // Assert
        let indicator = app.otherElements["voice_input_indicator"]
        XCTAssertTrue(
            indicator.waitForExistence(timeout: 5),
            "Voice input indicator (voice_input_indicator) should appear after tapping the mic button"
        )
    }

    // MARK: - Happy Path: Transcription fills input field

    /// After tapping the mic button and completing speech recognition, the
    /// transcribed text must appear in the chat input field so the user can
    /// review and optionally edit it before sending.
    ///
    /// Expected flow:
    ///   Tap "mic_button" → MockSpeechRecognizer emits simulatedTranscript
    ///   → "chat_input_field" contains the transcribed text
    func test_micButton_tap_voiceTranscriptionFillsInputField() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let micButton = app.buttons["mic_button"]
        XCTAssertTrue(micButton.waitForExistence(timeout: 5))

        // Act — MockSpeechRecognizer immediately sets simulatedTranscript
        // when startRecording() is called; ChatViewModel observes transcriptUpdates
        // and populates the input field accordingly.
        micButton.tap()

        // Assert — input field must contain the transcribed text
        let inputField = app.textFields["chat_input_field"]
        XCTAssertTrue(
            inputField.waitForExistence(timeout: 5),
            "Chat input field should be visible"
        )
        // Wait briefly for the transcription to propagate to the UI
        let transcriptionPredicate = NSPredicate(format: "value != ''")
        let expectation = XCTNSPredicateExpectation(predicate: transcriptionPredicate, object: inputField)
        let result = XCTWaiter().wait(for: [expectation], timeout: 5)
        XCTAssertEqual(
            result,
            .completed,
            "Chat input field should be populated with the transcription result"
        )
    }

    // MARK: - Happy Path: Voice → Send → User bubble

    /// After voice recognition fills the input field, tapping Send must display
    /// the transcribed text as a user message bubble in the conversation.
    ///
    /// Expected flow:
    ///   Tap "mic_button" → transcript fills "chat_input_field"
    ///   → tap "send_button" → "user_message_bubble" appears
    func test_voiceTranscription_send_displaysUserBubble() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let micButton = app.buttons["mic_button"]
        XCTAssertTrue(micButton.waitForExistence(timeout: 5))

        // Trigger voice recognition
        micButton.tap()

        // Wait for transcription to populate input field
        let inputField = app.textFields["chat_input_field"]
        let transcriptionPredicate = NSPredicate(format: "value != ''")
        let expectation = XCTNSPredicateExpectation(predicate: transcriptionPredicate, object: inputField)
        XCTWaiter().wait(for: [expectation], timeout: 5)

        // Act
        let sendButton = app.buttons["send_button"]
        XCTAssertTrue(sendButton.waitForExistence(timeout: 5), "Send button must exist")
        sendButton.tap()

        // Assert
        let userBubble = app.otherElements["user_message_bubble"]
        XCTAssertTrue(
            userBubble.waitForExistence(timeout: 5),
            "User message bubble should appear after sending voice-transcribed text"
        )
    }

    // MARK: - Happy Path: Voice → Send → AI response bubble

    /// After the user sends a voice-transcribed message, the AI response must
    /// appear as an AI response bubble in the conversation list.
    ///
    /// Expected flow:
    ///   Tap "mic_button" → transcript fills input → tap "send_button"
    ///   → "user_message_bubble" appears → "ai_response_bubble" appears
    func test_voiceTranscription_send_displaysAIResponseBubble() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let micButton = app.buttons["mic_button"]
        XCTAssertTrue(micButton.waitForExistence(timeout: 5))

        // Trigger voice recognition
        micButton.tap()

        // Wait for transcription
        let inputField = app.textFields["chat_input_field"]
        let transcriptionPredicate = NSPredicate(format: "value != ''")
        let expectation = XCTNSPredicateExpectation(predicate: transcriptionPredicate, object: inputField)
        XCTWaiter().wait(for: [expectation], timeout: 5)

        // Act
        let sendButton = app.buttons["send_button"]
        XCTAssertTrue(sendButton.waitForExistence(timeout: 5))
        sendButton.tap()

        // Assert
        let aiBubble = app.otherElements["ai_response_bubble"]
        XCTAssertTrue(
            aiBubble.waitForExistence(timeout: 15),
            "AI response bubble (ai_response_bubble) should appear after sending a voice message"
        )
    }

    // MARK: - Stop Recording: Text preserved in input field

    /// Tapping the mic button a second time while recording must stop voice
    /// recognition and preserve the text accumulated so far in the input field,
    /// allowing the user to edit before sending.
    ///
    /// Expected flow:
    ///   Tap "mic_button" → recording starts → "voice_input_indicator" visible
    ///   → tap "mic_button" again → "voice_input_indicator" disappears
    ///   → "chat_input_field" still contains the partial transcript
    func test_micButton_tapAgainWhileRecording_stopsAndPreservesText() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_assistant"].tap()

        let chatView = app.otherElements["assistant_chat_view"]
        XCTAssertTrue(chatView.waitForExistence(timeout: 5))

        let micButton = app.buttons["mic_button"]
        XCTAssertTrue(micButton.waitForExistence(timeout: 5))

        // Start recording
        micButton.tap()

        let indicator = app.otherElements["voice_input_indicator"]
        XCTAssertTrue(
            indicator.waitForExistence(timeout: 5),
            "Voice input indicator must appear after first tap"
        )

        // Wait for partial transcript to appear in input field
        let inputField = app.textFields["chat_input_field"]
        let transcriptionPredicate = NSPredicate(format: "value != ''")
        let expectation = XCTNSPredicateExpectation(predicate: transcriptionPredicate, object: inputField)
        XCTWaiter().wait(for: [expectation], timeout: 5)

        // Capture the partial transcript text
        let partialText = inputField.value as? String ?? ""

        // Act — tap mic again to stop recording
        micButton.tap()

        // Assert — indicator disappears but input field retains the text
        XCTAssertFalse(
            indicator.waitForExistence(timeout: 3),
            "Voice input indicator should disappear after stopping recording"
        )
        let currentText = inputField.value as? String ?? ""
        XCTAssertFalse(
            currentText.isEmpty,
            "Input field should still contain transcript text after stopping recording"
        )
        XCTAssertEqual(
            currentText,
            partialText,
            "Input field text should be preserved (not cleared) when recording stops"
        )
    }

    // MARK: - Settings: Speech Engine Change (Whisper Local → API)

    /// The user must be able to navigate to Settings and change the speech
    /// recognition engine from Whisper Local to Whisper API.
    ///
    /// Expected flow:
    ///   Navigate to Settings → "speech_engine_selector" is visible →
    ///   current engine is "Whisper Local" → tap selector →
    ///   "settings_speech_engine_picker" appears → select "Whisper API" →
    ///   picker dismisses → "speech_engine_selector" shows "Whisper API"
    func test_settings_speechEngine_changeFromWhisperLocalToAPI() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange — navigate to the Settings screen
        // Settings is assumed to be reachable via a tab or a button labeled "설정" / "Settings"
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible to navigate to Settings")
        tabBar.buttons["tab_settings"].tap()

        let speechEngineSelector = app.buttons["speech_engine_selector"]
        XCTAssertTrue(
            speechEngineSelector.waitForExistence(timeout: 5),
            "Speech engine selector (speech_engine_selector) should be visible in Settings"
        )

        // Verify initial engine is Whisper Local
        XCTAssertEqual(
            speechEngineSelector.label,
            "Whisper Local",
            "Default speech engine should be 'Whisper Local'"
        )

        // Act
        speechEngineSelector.tap()

        let enginePicker = app.otherElements["settings_speech_engine_picker"]
        XCTAssertTrue(
            enginePicker.waitForExistence(timeout: 5),
            "Speech engine picker (settings_speech_engine_picker) should appear after tapping selector"
        )

        let whisperAPIOption = enginePicker.buttons["Whisper API"]
        XCTAssertTrue(whisperAPIOption.exists, "'Whisper API' option must be available in the picker")
        whisperAPIOption.tap()

        // Assert
        XCTAssertFalse(
            enginePicker.waitForExistence(timeout: 3),
            "Speech engine picker should be dismissed after selection"
        )
        XCTAssertEqual(
            speechEngineSelector.label,
            "Whisper API",
            "Speech engine selector should display 'Whisper API' after selection"
        )
    }

    // MARK: - Settings: Speech Engine Change (API → Whisper Local)

    /// The user must be able to switch the speech recognition engine back from
    /// Whisper API to Whisper Local.
    ///
    /// Expected flow:
    ///   Navigate to Settings → change engine to "Whisper API" (precondition) →
    ///   tap "speech_engine_selector" → select "Whisper Local" →
    ///   "speech_engine_selector" shows "Whisper Local"
    func test_settings_speechEngine_changeFromAPIToWhisperLocal() throws {
        throw XCTSkip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

        // Arrange — navigate to Settings and switch to Whisper API first
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))
        tabBar.buttons["tab_settings"].tap()

        let speechEngineSelector = app.buttons["speech_engine_selector"]
        XCTAssertTrue(
            speechEngineSelector.waitForExistence(timeout: 5),
            "Speech engine selector must be visible"
        )

        // Set initial state to Whisper API if not already set
        if speechEngineSelector.label != "Whisper API" {
            speechEngineSelector.tap()
            let picker = app.otherElements["settings_speech_engine_picker"]
            XCTAssertTrue(picker.waitForExistence(timeout: 5))
            let apiOption = picker.buttons["Whisper API"]
            XCTAssertTrue(apiOption.exists, "'Whisper API' option must exist")
            apiOption.tap()
            XCTAssertFalse(picker.waitForExistence(timeout: 3), "Picker should dismiss")
        }

        XCTAssertEqual(
            speechEngineSelector.label,
            "Whisper API",
            "Precondition: speech engine must be 'Whisper API' before switching back"
        )

        // Act — switch back to Whisper Local
        speechEngineSelector.tap()

        let enginePicker = app.otherElements["settings_speech_engine_picker"]
        XCTAssertTrue(
            enginePicker.waitForExistence(timeout: 5),
            "Speech engine picker should reappear"
        )

        let whisperLocalOption = enginePicker.buttons["Whisper Local"]
        XCTAssertTrue(whisperLocalOption.exists, "'Whisper Local' option must be available")
        whisperLocalOption.tap()

        // Assert
        XCTAssertFalse(
            enginePicker.waitForExistence(timeout: 3),
            "Speech engine picker should be dismissed after selection"
        )
        XCTAssertEqual(
            speechEngineSelector.label,
            "Whisper Local",
            "Speech engine selector should display 'Whisper Local' after switching back"
        )
    }
}
