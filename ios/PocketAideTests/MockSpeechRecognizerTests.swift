// MockSpeechRecognizerTests.swift
// PocketAideTests
//
// Tests that verify:
//   1. `SpeechRecognizerProtocol` defines the correct interface.
//   2. `MockSpeechRecognizer` fully conforms to the protocol.
//   3. `MockSpeechRecognizer` behaves predictably in all states so it can
//      act as a reliable test double for higher-level feature tests.
//
// These tests are in the TDD Red phase — they will fail until the production
// code under test is implemented.
//
// Targets production types:
//   - PocketAide.SpeechRecognizerProtocol
//   - PocketAide.MockSpeechRecognizer
//   - PocketAide.SpeechRecognizerError

import XCTest
@testable import PocketAide

// MARK: - MockSpeechRecognizerTests

final class MockSpeechRecognizerTests: XCTestCase {

    // MARK: Properties

    private var sut: MockSpeechRecognizer!

    // MARK: Lifecycle

    override func setUp() {
        super.setUp()
        sut = MockSpeechRecognizer()
    }

    override func tearDown() {
        sut = nil
        super.tearDown()
    }

    // MARK: - Protocol Conformance

    /// `MockSpeechRecognizer` must be assignable to a variable typed as the
    /// protocol. This is a compile-time guarantee; the test will not build if
    /// conformance is absent.
    func test_mockSpeechRecognizer_conformsToProtocol() {
        // Arrange & Act
        let recognizer: any SpeechRecognizerProtocol = MockSpeechRecognizer()

        // Assert — if this line compiles, the protocol is satisfied.
        XCTAssertNotNil(recognizer)
    }

    // MARK: - Initial State

    /// Immediately after initialisation the recognizer must not be recording.
    func test_initialState_isNotRecording() {
        XCTAssertFalse(sut.isRecording)
    }

    /// The transcript must be empty before any recording has started.
    func test_initialState_transcriptIsEmpty() {
        XCTAssertEqual(sut.transcript, "")
    }

    /// No error must be present in the initial state.
    func test_initialState_hasNoError() {
        XCTAssertNil(sut.error)
    }

    // MARK: - startRecording

    /// After `startRecording()` the recognizer must report that it is active.
    func test_startRecording_setsIsRecordingToTrue() async throws {
        // Arrange — already in initial state

        // Act
        try await sut.startRecording()

        // Assert
        XCTAssertTrue(sut.isRecording)
    }

    /// Calling `startRecording()` while already recording must throw
    /// `SpeechRecognizerError.alreadyRecording`.
    func test_startRecording_throwsWhenAlreadyRecording() async throws {
        // Arrange
        try await sut.startRecording()

        // Act & Assert
        do {
            try await sut.startRecording()
            XCTFail("Expected SpeechRecognizerError.alreadyRecording to be thrown")
        } catch let error as SpeechRecognizerError {
            XCTAssertEqual(error, .alreadyRecording)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    /// When `simulatedError` is set, `startRecording()` must propagate it and
    /// leave `isRecording` as `false`.
    func test_startRecording_throwsSimulatedError() async {
        // Arrange
        sut.simulatedError = SpeechRecognizerError.permissionDenied

        // Act & Assert
        do {
            try await sut.startRecording()
            XCTFail("Expected simulated error to be thrown")
        } catch let error as SpeechRecognizerError {
            XCTAssertEqual(error, .permissionDenied)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }

        XCTAssertFalse(sut.isRecording)
    }

    // MARK: - stopRecording

    /// After a start/stop cycle the recognizer must report it is no longer
    /// recording.
    func test_stopRecording_setsIsRecordingToFalse() async throws {
        // Arrange
        try await sut.startRecording()
        XCTAssertTrue(sut.isRecording)

        // Act
        sut.stopRecording()

        // Assert
        XCTAssertFalse(sut.isRecording)
    }

    /// Calling `stopRecording()` when not recording must be a no-op (must not
    /// crash or throw).
    func test_stopRecording_isNoOpWhenNotRecording() {
        // Arrange — already stopped

        // Act — must not crash
        sut.stopRecording()

        // Assert
        XCTAssertFalse(sut.isRecording)
    }

    // MARK: - Transcript Simulation

    /// Injecting a canned transcript via `simulatedTranscript` must make that
    /// text appear in `transcript` once recording starts.
    func test_simulatedTranscript_appearsInTranscriptAfterStart() async throws {
        // Arrange
        sut.simulatedTranscript = "hello world"

        // Act
        try await sut.startRecording()

        // Assert
        XCTAssertEqual(sut.transcript, "hello world")
    }

    /// After `stopRecording()` the transcript must retain its last value so
    /// callers can read it at any point.
    func test_transcript_retainsValueAfterStop() async throws {
        // Arrange
        sut.simulatedTranscript = "retained text"
        try await sut.startRecording()

        // Act
        sut.stopRecording()

        // Assert
        XCTAssertEqual(sut.transcript, "retained text")
    }

    /// Calling `reset()` must clear the transcript and return the mock to its
    /// initial state.
    func test_reset_clearsTranscriptAndState() async throws {
        // Arrange
        sut.simulatedTranscript = "some text"
        try await sut.startRecording()
        sut.stopRecording()

        // Act
        sut.reset()

        // Assert
        XCTAssertEqual(sut.transcript, "")
        XCTAssertFalse(sut.isRecording)
        XCTAssertNil(sut.error)
    }

    // MARK: - Call Count Tracking

    /// The mock must track how many times `startRecording()` was called so
    /// that tests can assert interaction counts.
    func test_startRecordingCallCount_incrementsOnEachCall() async throws {
        // Arrange
        XCTAssertEqual(sut.startRecordingCallCount, 0)

        // Act
        try await sut.startRecording()
        sut.stopRecording()
        try await sut.startRecording()

        // Assert
        XCTAssertEqual(sut.startRecordingCallCount, 2)
    }

    /// The mock must track how many times `stopRecording()` was called.
    func test_stopRecordingCallCount_incrementsOnEachCall() async throws {
        // Arrange
        try await sut.startRecording()

        // Act
        sut.stopRecording()
        sut.stopRecording() // no-op second call

        // Assert
        XCTAssertEqual(sut.stopRecordingCallCount, 2)
    }

    // MARK: - Availability

    /// The protocol must expose an `isAvailable` property. The mock should
    /// allow it to be toggled to simulate devices where speech recognition is
    /// unavailable.
    func test_isAvailable_defaultsToTrue() {
        XCTAssertTrue(sut.isAvailable)
    }

    /// When `isAvailable` is `false`, `startRecording()` must throw
    /// `SpeechRecognizerError.notAvailable`.
    func test_startRecording_throwsNotAvailableWhenUnavailable() async {
        // Arrange
        sut.isAvailable = false

        // Act & Assert
        do {
            try await sut.startRecording()
            XCTFail("Expected SpeechRecognizerError.notAvailable to be thrown")
        } catch let error as SpeechRecognizerError {
            XCTAssertEqual(error, .notAvailable)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    // MARK: - AsyncStream / Transcript Updates

    /// `transcriptUpdates` must be an `AsyncStream<String>` that emits each
    /// partial transcript as it changes during a recording session.
    func test_transcriptUpdates_emitsSimulatedPhrases() async throws {
        // Arrange
        let phrases = ["Hi", "Hi there", "Hi there world"]
        sut.simulatedPhrases = phrases
        var received: [String] = []

        // Act
        try await sut.startRecording()
        for await partial in sut.transcriptUpdates {
            received.append(partial)
            if received.count == phrases.count { break }
        }

        // Assert
        XCTAssertEqual(received, phrases)
    }
}
