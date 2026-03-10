// Tests for SpeechRecognizerProtocol and MockSpeechRecognizer.
//
// Validates:
//   - `MockSpeechRecognizer` emits stubbed `SpeechResult` values in order.
//   - `isRecognizing` state transitions are correct.
//   - `requestPermission()` returns the configured value.
//   - `startRecognizing()` throws when `startError` is set.
//   - Call counts are accurately tracked.
//   - `reset()` restores the mock to its initial state.

import XCTest
@testable import PocketAide

final class SpeechRecognizerTests: XCTestCase {

    var mock: MockSpeechRecognizer!

    override func setUp() {
        super.setUp()
        mock = MockSpeechRecognizer()
    }

    override func tearDown() {
        mock.reset()
        mock = nil
        super.tearDown()
    }

    // MARK: - Permission

    func test_requestPermission_returnsTrue_whenGranted() async {
        // Arrange
        mock.permissionGranted = true

        // Act
        let granted = await mock.requestPermission()

        // Assert
        XCTAssertTrue(granted)
    }

    func test_requestPermission_returnsFalse_whenDenied() async {
        // Arrange
        mock.permissionGranted = false

        // Act
        let granted = await mock.requestPermission()

        // Assert
        XCTAssertFalse(granted)
    }

    func test_requestPermission_incrementsCallCount() async {
        // Act
        _ = await mock.requestPermission()
        _ = await mock.requestPermission()

        // Assert
        XCTAssertEqual(mock.permissionCallCount, 2)
    }

    // MARK: - Start / Stop state

    func test_isRecognizing_isFalse_beforeStart() {
        XCTAssertFalse(mock.isRecognizing)
    }

    func test_isRecognizing_isTrue_afterStart() async throws {
        // Arrange
        mock.stubbedResults = [SpeechResult(transcript: "hello", isFinal: true)]

        // Act
        _ = try await mock.startRecognizing()

        // Assert
        XCTAssertTrue(mock.isRecognizing)
    }

    func test_isRecognizing_isFalse_afterStop() async throws {
        // Arrange
        mock.stubbedResults = [SpeechResult(transcript: "hello", isFinal: true)]
        _ = try await mock.startRecognizing()

        // Act
        mock.stopRecognizing()

        // Assert
        XCTAssertFalse(mock.isRecognizing)
    }

    func test_startCallCount_incrementsOnEachCall() async throws {
        // Arrange
        mock.stubbedResults = [SpeechResult(transcript: "a", isFinal: true)]

        // Act
        _ = try await mock.startRecognizing()
        _ = try await mock.startRecognizing()

        // Assert
        XCTAssertEqual(mock.startCallCount, 2)
    }

    func test_stopCallCount_incrementsOnEachCall() {
        // Act
        mock.stopRecognizing()
        mock.stopRecognizing()

        // Assert
        XCTAssertEqual(mock.stopCallCount, 2)
    }

    // MARK: - Emitted results

    func test_stream_emitsSingleFinalResult() async throws {
        // Arrange
        let expected = SpeechResult(transcript: "hello world", isFinal: true)
        mock.stubbedResults = [expected]

        // Act
        var received: [SpeechResult] = []
        let stream = try await mock.startRecognizing()
        for await result in stream {
            received.append(result)
        }

        // Assert
        XCTAssertEqual(received.count, 1)
        XCTAssertEqual(received[0], expected)
    }

    func test_stream_emitsPartialThenFinalResult() async throws {
        // Arrange
        let partial = SpeechResult(transcript: "hel", isFinal: false)
        let final_  = SpeechResult(transcript: "hello", isFinal: true)
        mock.stubbedResults = [partial, final_]

        // Act
        var received: [SpeechResult] = []
        let stream = try await mock.startRecognizing()
        for await result in stream {
            received.append(result)
        }

        // Assert
        XCTAssertEqual(received.count, 2)
        XCTAssertEqual(received[0], partial)
        XCTAssertEqual(received[1], final_)
    }

    func test_stream_emitsNothingWhenNoResultsStubbed() async throws {
        // Arrange
        mock.stubbedResults = []

        // Act
        var received: [SpeechResult] = []
        let stream = try await mock.startRecognizing()
        for await result in stream {
            received.append(result)
        }

        // Assert
        XCTAssertTrue(received.isEmpty)
    }

    func test_stream_stopsAfterFirstFinalResult() async throws {
        // Arrange — two results after the final one; they should not be emitted
        mock.stubbedResults = [
            SpeechResult(transcript: "yes", isFinal: true),
            SpeechResult(transcript: "should not appear", isFinal: false),
        ]

        // Act
        var received: [SpeechResult] = []
        let stream = try await mock.startRecognizing()
        for await result in stream {
            received.append(result)
        }

        // Assert
        XCTAssertEqual(received.count, 1)
        XCTAssertTrue(received[0].isFinal)
    }

    func test_stream_emitsTranscriptStrings_inOrder() async throws {
        // Arrange
        let transcripts = ["a", "ab", "abc"]
        mock.stubbedResults = transcripts.enumerated().map { idx, text in
            SpeechResult(transcript: text, isFinal: idx == transcripts.count - 1)
        }

        // Act
        var receivedTranscripts: [String] = []
        let stream = try await mock.startRecognizing()
        for await result in stream {
            receivedTranscripts.append(result.transcript)
        }

        // Assert
        XCTAssertEqual(receivedTranscripts, transcripts)
    }

    // MARK: - Error path

    func test_startRecognizing_throwsPermissionDenied() async {
        // Arrange
        mock.startError = .permissionDenied

        // Act & Assert
        do {
            _ = try await mock.startRecognizing()
            XCTFail("Expected SpeechRecognizerError.permissionDenied to be thrown")
        } catch let error as SpeechRecognizerError {
            XCTAssertEqual(error, .permissionDenied)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    func test_startRecognizing_throwsUnavailable() async {
        // Arrange
        mock.startError = .unavailable

        // Act & Assert
        do {
            _ = try await mock.startRecognizing()
            XCTFail("Expected SpeechRecognizerError.unavailable to be thrown")
        } catch let error as SpeechRecognizerError {
            XCTAssertEqual(error, .unavailable)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    func test_startRecognizing_throwsAlreadyRunning() async {
        // Arrange
        mock.startError = .alreadyRunning

        // Act & Assert
        do {
            _ = try await mock.startRecognizing()
            XCTFail("Expected SpeechRecognizerError.alreadyRunning to be thrown")
        } catch let error as SpeechRecognizerError {
            XCTAssertEqual(error, .alreadyRunning)
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    func test_startRecognizing_throwsEngineError_withMessage() async {
        // Arrange
        mock.startError = .engineError("audio hardware unavailable")

        // Act & Assert
        do {
            _ = try await mock.startRecognizing()
            XCTFail("Expected SpeechRecognizerError.engineError to be thrown")
        } catch let error as SpeechRecognizerError {
            if case .engineError(let msg) = error {
                XCTAssertEqual(msg, "audio hardware unavailable")
            } else {
                XCTFail("Expected engineError, got \(error)")
            }
        } catch {
            XCTFail("Unexpected error type: \(error)")
        }
    }

    // MARK: - Reset

    func test_reset_clearsAllState() async throws {
        // Arrange — dirty the mock
        mock.stubbedResults = [SpeechResult(transcript: "hi", isFinal: true)]
        _ = try await mock.startRecognizing()
        mock.stopRecognizing()
        _ = await mock.requestPermission()

        // Act
        mock.reset()

        // Assert
        XCTAssertFalse(mock.isRecognizing)
        XCTAssertEqual(mock.startCallCount, 0)
        XCTAssertEqual(mock.stopCallCount, 0)
        XCTAssertEqual(mock.permissionCallCount, 0)
        XCTAssertTrue(mock.stubbedResults.isEmpty)
        XCTAssertNil(mock.startError)
        XCTAssertTrue(mock.permissionGranted)
    }
}
