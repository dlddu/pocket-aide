// Protocol-based abstraction for speech recognition.
// The live implementation will wrap Apple's Speech framework while the mock
// implementation allows unit tests to inject predetermined transcripts.

import Foundation

// MARK: - SpeechRecognizerError

/// Errors that a SpeechRecognizer implementation may throw or emit.
public enum SpeechRecognizerError: Error, Equatable {
    /// The user has not granted microphone / speech recognition permission.
    case permissionDenied

    /// The recognizer could not be started (e.g. locale not supported).
    case unavailable

    /// Recognition was already in progress when `startRecognizing()` was called.
    case alreadyRunning

    /// An underlying engine error occurred.
    case engineError(String)
}

// MARK: - SpeechRecognizerProtocol

/// Defines the contract for speech-to-text functionality.
///
/// Conforming types publish incremental `SpeechResult` values via an
/// `AsyncStream` so that the caller can render partial transcripts in real time.
public protocol SpeechRecognizerProtocol: AnyObject, Sendable {

    /// Whether the recognizer is currently capturing audio.
    var isRecognizing: Bool { get }

    /// Begins capturing and recognizing speech.
    ///
    /// - Returns: An `AsyncStream` of `SpeechResult` values. The stream ends
    ///   when `stopRecognizing()` is called or a terminal error occurs.
    /// - Throws: `SpeechRecognizerError` if the session cannot be started.
    func startRecognizing() async throws -> AsyncStream<SpeechResult>

    /// Stops the current recognition session and closes the stream.
    func stopRecognizing()

    /// Requests the appropriate system permissions.
    ///
    /// - Returns: `true` if the user granted both microphone and speech recognition access.
    func requestPermission() async -> Bool
}
