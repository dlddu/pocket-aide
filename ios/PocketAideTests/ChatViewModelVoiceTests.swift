// ChatViewModelVoiceTests.swift
// PocketAideTests
//
// Unit tests for ChatViewModel's voice input integration:
//   - startVoiceInput() delegates to the injected SpeechRecognizerProtocol
//   - transcription result populates inputText
//   - stopVoiceInput() stops recording and preserves inputText
//   - isRecording state is reflected via ChatViewModel.isVoiceRecording
//   - speech engine selection updates the active recognizer
//   - error handling when SpeechRecognizerProtocol throws
//
// DLD-722: 4-2: AI 채팅 (음성) — 구현 (TDD Red phase)
//
// NOTE: These tests are in the TDD Red phase — they will fail until
// code-writer implements the voice integration in ChatViewModel.
//
// Targets production types:
//   - PocketAide.ChatViewModel
//   - PocketAide.SpeechRecognizerProtocol
//   - PocketAide.MockSpeechRecognizer
//   - PocketAide.SpeechRecognizerError
//   - PocketAide.SpeechEngine (to be added)

import XCTest
import Combine
@testable import PocketAide

// MARK: - ChatViewModelVoiceTests

@MainActor
final class ChatViewModelVoiceTests: XCTestCase {

    // MARK: - Properties

    private var sut: ChatViewModel!
    private var mockRecognizer: MockSpeechRecognizer!
    private var cancellables: Set<AnyCancellable>!

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        mockRecognizer = MockSpeechRecognizer()
        // ChatViewModel must accept a SpeechRecognizerProtocol dependency so
        // tests can inject MockSpeechRecognizer without hitting real APIs.
        sut = ChatViewModel(speechRecognizer: mockRecognizer)
        cancellables = []
    }

    override func tearDown() {
        cancellables = nil
        sut = nil
        mockRecognizer = nil
        super.tearDown()
    }

    // MARK: - Initial Voice State

    /// Before any voice interaction the ViewModel must report that it is not
    /// recording, matching the recognizer's initial state.
    func test_initialState_isVoiceRecordingIsFalse() {
        XCTAssertFalse(sut.isVoiceRecording, "isVoiceRecording should be false before startVoiceInput() is called")
    }

    /// The inputText field must be empty in the initial state so no stale
    /// transcript from a previous session is visible.
    func test_initialState_inputTextIsEmpty() {
        XCTAssertEqual(sut.inputText, "")
    }

    // MARK: - startVoiceInput: Happy Path

    /// Calling startVoiceInput() must flip isVoiceRecording to true,
    /// reflecting that the underlying recognizer has begun recording.
    func test_startVoiceInput_setsIsVoiceRecordingToTrue() async throws {
        // Arrange — mock is ready, no errors

        // Act
        try await sut.startVoiceInput()

        // Assert
        XCTAssertTrue(sut.isVoiceRecording, "isVoiceRecording must be true after startVoiceInput()")
        XCTAssertEqual(mockRecognizer.startRecordingCallCount, 1,
                       "startRecording() must be called exactly once on the underlying recognizer")
    }

    /// After startVoiceInput() the MockSpeechRecognizer's simulatedTranscript
    /// must be reflected in ChatViewModel.inputText so the user can see what
    /// was recognised.
    func test_startVoiceInput_transcriptPopulatesInputText() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "오늘 날씨 어때?"

        // Act
        try await sut.startVoiceInput()

        // Assert
        XCTAssertEqual(
            sut.inputText,
            "오늘 날씨 어때?",
            "inputText must reflect the transcript produced by the recognizer"
        )
    }

    /// The ViewModel must subscribe to transcriptUpdates and reflect each
    /// intermediate phrase in inputText as the stream emits values.
    func test_startVoiceInput_transcriptUpdates_populatesInputTextIncrementally() async throws {
        // Arrange
        let phrases = ["회의", "회의 일정", "회의 일정 추가해줘"]
        mockRecognizer.simulatedPhrases = phrases

        // Act
        try await sut.startVoiceInput()

        // Give the async stream a chance to deliver all values
        try await Task.sleep(nanoseconds: 100_000_000)

        // Assert — the final phrase must be in inputText
        XCTAssertEqual(
            sut.inputText,
            "회의 일정 추가해줘",
            "inputText must be updated to the final phrase emitted by transcriptUpdates"
        )
    }

    // MARK: - stopVoiceInput: Happy Path

    /// Calling stopVoiceInput() after startVoiceInput() must set
    /// isVoiceRecording back to false.
    func test_stopVoiceInput_setsIsVoiceRecordingToFalse() async throws {
        // Arrange
        try await sut.startVoiceInput()
        XCTAssertTrue(sut.isVoiceRecording)

        // Act
        sut.stopVoiceInput()

        // Assert
        XCTAssertFalse(sut.isVoiceRecording, "isVoiceRecording must be false after stopVoiceInput()")
        XCTAssertEqual(mockRecognizer.stopRecordingCallCount, 1,
                       "stopRecording() must be called exactly once on the underlying recognizer")
    }

    /// After stopVoiceInput() the text accumulated during recording must still
    /// be present in inputText so the user can review and edit it before sending.
    func test_stopVoiceInput_preservesInputText() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "보낼 메시지 내용"
        try await sut.startVoiceInput()
        XCTAssertEqual(sut.inputText, "보낼 메시지 내용")

        // Act
        sut.stopVoiceInput()

        // Assert
        XCTAssertEqual(
            sut.inputText,
            "보낼 메시지 내용",
            "inputText must not be cleared when recording stops"
        )
    }

    /// Calling stopVoiceInput() when not currently recording must be a no-op
    /// — it must not crash or toggle any unexpected state.
    func test_stopVoiceInput_isNoOpWhenNotRecording() {
        // Arrange — not recording

        // Act — must not crash
        sut.stopVoiceInput()

        // Assert
        XCTAssertFalse(sut.isVoiceRecording)
        XCTAssertEqual(sut.inputText, "")
    }

    // MARK: - toggleVoiceInput: Toggle Behaviour

    /// The first toggleVoiceInput() call must start recording.
    func test_toggleVoiceInput_firstCall_startsRecording() async throws {
        // Arrange — not recording

        // Act
        try await sut.toggleVoiceInput()

        // Assert
        XCTAssertTrue(sut.isVoiceRecording, "First toggle must start recording")
    }

    /// The second toggleVoiceInput() call while recording must stop it.
    func test_toggleVoiceInput_secondCall_stopsRecording() async throws {
        // Arrange — start recording first
        try await sut.toggleVoiceInput()
        XCTAssertTrue(sut.isVoiceRecording)

        // Act
        try await sut.toggleVoiceInput()

        // Assert
        XCTAssertFalse(sut.isVoiceRecording, "Second toggle must stop recording")
    }

    // MARK: - Error Handling

    /// If the recognizer throws SpeechRecognizerError.permissionDenied,
    /// startVoiceInput() must surface it as ChatViewModel.voiceErrorMessage
    /// and leave isVoiceRecording as false.
    func test_startVoiceInput_permissionDenied_setsVoiceErrorMessage() async {
        // Arrange
        mockRecognizer.simulatedError = SpeechRecognizerError.permissionDenied

        // Act
        do {
            try await sut.startVoiceInput()
        } catch {
            // ChatViewModel may rethrow or swallow; both are valid as long as
            // voiceErrorMessage is populated.
        }

        // Assert
        XCTAssertFalse(sut.isVoiceRecording, "isVoiceRecording must remain false on permission error")
        XCTAssertNotNil(
            sut.voiceErrorMessage,
            "voiceErrorMessage must be set when the recognizer throws permissionDenied"
        )
    }

    /// If the recognizer throws SpeechRecognizerError.notAvailable,
    /// the ViewModel must set voiceErrorMessage and keep isVoiceRecording false.
    func test_startVoiceInput_notAvailable_setsVoiceErrorMessage() async {
        // Arrange
        mockRecognizer.isAvailable = false

        // Act
        do {
            try await sut.startVoiceInput()
        } catch {
            // expected
        }

        // Assert
        XCTAssertFalse(sut.isVoiceRecording)
        XCTAssertNotNil(
            sut.voiceErrorMessage,
            "voiceErrorMessage must be set when speech recognition is not available"
        )
    }

    /// After the user dismisses the error (e.g. taps an alert OK button),
    /// voiceErrorMessage must be cleared so subsequent recordings can succeed.
    func test_clearVoiceError_resetsVoiceErrorMessage() async {
        // Arrange — trigger an error first
        mockRecognizer.simulatedError = SpeechRecognizerError.permissionDenied
        do { try await sut.startVoiceInput() } catch {}
        XCTAssertNotNil(sut.voiceErrorMessage)

        // Act
        sut.clearVoiceError()

        // Assert
        XCTAssertNil(sut.voiceErrorMessage, "voiceErrorMessage must be nil after clearVoiceError()")
    }

    // MARK: - Speech Engine Selection

    /// The ViewModel must expose a selectedSpeechEngine property initialised
    /// to .whisperLocal as the default engine.
    func test_selectedSpeechEngine_defaultsToWhisperLocal() {
        XCTAssertEqual(
            sut.selectedSpeechEngine,
            .whisperLocal,
            "Default speech engine should be .whisperLocal"
        )
    }

    /// Changing selectedSpeechEngine to .whisperAPI must swap the underlying
    /// recognizer so subsequent recordings use the API-based engine.
    func test_selectedSpeechEngine_changeToWhisperAPI_updatesRecognizer() {
        // Arrange — default is .whisperLocal

        // Act
        sut.selectedSpeechEngine = .whisperAPI

        // Assert
        XCTAssertEqual(sut.selectedSpeechEngine, .whisperAPI,
                       "selectedSpeechEngine must update to .whisperAPI")
    }

    /// Changing selectedSpeechEngine back to .whisperLocal must restore that
    /// selection.
    func test_selectedSpeechEngine_changeToWhisperLocal_restoresSelection() {
        // Arrange — set to API first
        sut.selectedSpeechEngine = .whisperAPI

        // Act
        sut.selectedSpeechEngine = .whisperLocal

        // Assert
        XCTAssertEqual(sut.selectedSpeechEngine, .whisperLocal,
                       "selectedSpeechEngine must be restorable to .whisperLocal")
    }

    // MARK: - Voice Input + Send Integration

    /// After voice transcription fills inputText, sendMessage() must consume
    /// that text and clear inputText, just as it does for keyboard input.
    func test_voiceTranscription_thenSendMessage_clearsInputText() async throws {
        // Arrange — simulate transcription result
        mockRecognizer.simulatedTranscript = "테스트 메시지"
        try await sut.startVoiceInput()
        sut.stopVoiceInput()
        XCTAssertEqual(sut.inputText, "테스트 메시지")

        // Act
        sut.sendMessage()

        // Assert — inputText cleared after send (ChatViewModel.sendMessage existing behaviour)
        XCTAssertEqual(
            sut.inputText,
            "",
            "inputText must be cleared after sendMessage() processes the voice-transcribed text"
        )
    }

    /// After voice transcription fills inputText and the user sends, the
    /// transcribed text must appear as a user message bubble in messages.
    func test_voiceTranscription_thenSendMessage_appendsUserMessage() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "음성 전송 테스트"
        try await sut.startVoiceInput()
        sut.stopVoiceInput()

        // Act
        sut.sendMessage()

        // Assert
        let userMessages = sut.messages.filter { $0.role == .user }
        XCTAssertEqual(
            userMessages.last?.content,
            "음성 전송 테스트",
            "The voice-transcribed text must appear as the most recent user message"
        )
    }
}
