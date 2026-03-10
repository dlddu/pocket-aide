// Deterministic mock implementation of SpeechRecognizerProtocol for unit tests.
// Inject pre-canned SpeechResult values; no microphone or Speech framework needed.

import Foundation

// MARK: - MockSpeechRecognizer

/// A test double for `SpeechRecognizerProtocol` that emits pre-configured
/// `SpeechResult` values without accessing the microphone or the Speech framework.
///
/// Usage:
/// ```swift
/// let mock = MockSpeechRecognizer()
/// mock.stubbedResults = [
///     SpeechResult(transcript: "hello", isFinal: false),
///     SpeechResult(transcript: "hello world", isFinal: true),
/// ]
/// mock.permissionGranted = true
/// let stream = try await mock.startRecognizing()
/// for await result in stream { /* ... */ }
/// ```
public final class MockSpeechRecognizer: SpeechRecognizerProtocol, @unchecked Sendable {

    // MARK: - Configuration

    /// Results emitted by the `AsyncStream` returned from `startRecognizing()`.
    public var stubbedResults: [SpeechResult] = []

    /// Value returned by `requestPermission()`.
    public var permissionGranted: Bool = true

    /// If set, `startRecognizing()` throws this error instead of returning a stream.
    public var startError: SpeechRecognizerError? = nil

    /// Optional delay (in nanoseconds) between each emitted result (default: 0).
    public var emissionDelayNanoseconds: UInt64 = 0

    // MARK: - Observation

    /// Number of times `startRecognizing()` has been called.
    public private(set) var startCallCount: Int = 0

    /// Number of times `stopRecognizing()` has been called.
    public private(set) var stopCallCount: Int = 0

    /// Number of times `requestPermission()` has been called.
    public private(set) var permissionCallCount: Int = 0

    // MARK: - SpeechRecognizerProtocol

    public private(set) var isRecognizing: Bool = false

    public func startRecognizing() async throws -> AsyncStream<SpeechResult> {
        startCallCount += 1

        if let error = startError {
            throw error
        }

        isRecognizing = true

        let results = stubbedResults
        let delay = emissionDelayNanoseconds

        return AsyncStream { continuation in
            Task {
                for result in results {
                    if delay > 0 {
                        try? await Task.sleep(nanoseconds: delay)
                    }
                    continuation.yield(result)
                    if result.isFinal {
                        break
                    }
                }
                continuation.finish()
            }
        }
    }

    public func stopRecognizing() {
        stopCallCount += 1
        isRecognizing = false
    }

    public func requestPermission() async -> Bool {
        permissionCallCount += 1
        return permissionGranted
    }

    // MARK: - Helpers

    /// Resets all call counts and restores default configuration.
    public func reset() {
        stubbedResults = []
        permissionGranted = true
        startError = nil
        emissionDelayNanoseconds = 0
        startCallCount = 0
        stopCallCount = 0
        permissionCallCount = 0
        isRecognizing = false
    }
}
