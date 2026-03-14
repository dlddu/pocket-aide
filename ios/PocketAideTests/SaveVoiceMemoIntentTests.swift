// SaveVoiceMemoIntentTests.swift
// PocketAideTests
//
// Unit tests for the SaveVoiceMemoIntent App Intent:
//   - perform() delegates to SpeechRecognizerProtocol for voice recording
//   - transcribed text is saved via MemoService.create(source: "voice")
//   - confirmationRequired is false (no user prompt before execution)
//   - empty transcript prevents memo creation
//   - SpeechRecognizerError propagates as a user-facing IntentError
//   - haptic feedback is triggered on successful save
//
// DLD-732: 9-2: Siri Shortcut 음성 메모 — 구현 (TDD Red phase)
//
// NOTE: These tests are in the TDD Red phase — they will fail until
// code-writer implements SaveVoiceMemoIntent and MockMemoService.
//
// Targets production types:
//   - PocketAide.SaveVoiceMemoIntent      (to be created)
//   - PocketAide.MockMemoService           (to be created)
//   - PocketAide.MockSpeechRecognizer      (already exists)
//   - PocketAide.SpeechRecognizerProtocol  (already exists)
//   - PocketAide.SpeechRecognizerError     (already exists)

import XCTest
@testable import PocketAide

// MARK: - MockMemoService

/// `MemoService`의 테스트 대역.
///
/// `SaveVoiceMemoIntent`가 `MemoService.create(content:source:serverURL:token:)`를
/// 올바른 인자로 호출하는지 검증하기 위해 사용됩니다.
final class MockMemoService {

    // MARK: - Captured Arguments

    /// `create(content:source:serverURL:token:)` 호출 시 캡처한 인자.
    struct CreateCall {
        let content: String
        let source: String
    }

    private(set) var createCalls: [CreateCall] = []

    /// `create` 호출 시 throw할 에러. `nil`이면 정상 동작.
    var simulatedCreateError: Error? = nil

    /// `create` 호출 시 반환할 메모. 기본값은 더미 메모.
    var stubbedMemo: Memo = Memo(id: 1, content: "stub", source: "voice")

    // MARK: - Reset

    func reset() {
        createCalls = []
        simulatedCreateError = nil
        stubbedMemo = Memo(id: 1, content: "stub", source: "voice")
    }
}

extension MockMemoService {
    func create(content: String, source: String, serverURL: String, token: String) async throws -> Memo {
        createCalls.append(CreateCall(content: content, source: source))
        if let error = simulatedCreateError {
            throw error
        }
        return stubbedMemo
    }
}

// MARK: - SaveVoiceMemoIntentTests

/// `SaveVoiceMemoIntent.perform()` 전체 흐름을 검증하는 유닛 테스트.
///
/// 실제 Siri 런타임, 마이크, 네트워크 없이 Mock 객체만으로 모든 경로를 커버합니다.
final class SaveVoiceMemoIntentTests: XCTestCase {

    // MARK: - Properties

    private var sut: SaveVoiceMemoIntent!
    private var mockRecognizer: MockSpeechRecognizer!
    private var mockMemoService: MockMemoService!

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        mockRecognizer = MockSpeechRecognizer()
        mockMemoService = MockMemoService()
        // SaveVoiceMemoIntent must accept injected dependencies for testability.
        sut = SaveVoiceMemoIntent(
            speechRecognizer: mockRecognizer,
            memoService: mockMemoService
        )
    }

    override func tearDown() {
        sut = nil
        mockRecognizer = nil
        mockMemoService = nil
        super.tearDown()
    }

    // MARK: - confirmationRequired

    /// `SaveVoiceMemoIntent`는 Siri가 "정말로 실행할까요?"를 묻지 않도록
    /// `confirmationRequired`를 `false`로 선언해야 합니다.
    ///
    /// App Intents 프레임워크에서 이 값은 타입 프로퍼티로 선언됩니다.
    func test_confirmationRequired_isFalse() {
        // Assert — compile-time guarantee: the static property must exist and be false.
        XCTAssertFalse(
            SaveVoiceMemoIntent.confirmationRequired,
            "confirmationRequired must be false so Siri executes without prompting"
        )
    }

    // MARK: - Happy Path: perform()

    /// `perform()`이 호출되면 주입된 `SpeechRecognizerProtocol.startRecording()`이
    /// 정확히 한 번 실행되어야 합니다.
    func test_perform_callsStartRecordingOnce() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "오늘 할 일 메모"

        // Act
        _ = try await sut.perform()

        // Assert
        XCTAssertEqual(
            mockRecognizer.startRecordingCallCount,
            1,
            "startRecording() must be called exactly once during perform()"
        )
    }

    /// `perform()`이 끝나면 `SpeechRecognizerProtocol.stopRecording()`이
    /// 정확히 한 번 실행되어야 합니다.
    func test_perform_callsStopRecordingOnce() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "오늘 할 일 메모"

        // Act
        _ = try await sut.perform()

        // Assert
        XCTAssertEqual(
            mockRecognizer.stopRecordingCallCount,
            1,
            "stopRecording() must be called exactly once during perform()"
        )
    }

    /// 전사된 텍스트가 `MemoService.create(content:source:serverURL:token:)`에
    /// `source: "voice"`로 전달되어야 합니다.
    func test_perform_savesMemoWithVoiceSource() async throws {
        // Arrange
        let transcript = "회의 일정 추가해줘"
        mockRecognizer.simulatedTranscript = transcript

        // Act
        _ = try await sut.perform()

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.count,
            1,
            "MemoService.create() must be called exactly once"
        )
        XCTAssertEqual(
            mockMemoService.createCalls.first?.source,
            "voice",
            "Memo must be saved with source == 'voice'"
        )
        XCTAssertEqual(
            mockMemoService.createCalls.first?.content,
            transcript,
            "Memo content must match the transcribed text"
        )
    }

    /// `perform()`이 성공하면 `IntentResult`에 간단한 완료 메시지가 포함되어야 합니다.
    /// 메시지가 비어있지 않으면 됩니다.
    func test_perform_returnsNonEmptyResultMessage() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "장보기 목록"

        // Act
        let result = try await sut.perform()

        // Assert — the intent dialog/result must carry a non-empty message string
        XCTAssertFalse(
            result.value.isEmpty,
            "perform() must return a non-empty result message on success"
        )
    }

    // MARK: - Edge Case: Empty Transcript

    /// 음성 인식 결과가 빈 문자열인 경우 `MemoService.create()`를 호출해서는 안 됩니다.
    func test_perform_emptyTranscript_doesNotSaveMemo() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = ""

        // Act
        _ = try await sut.perform()

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.count,
            0,
            "MemoService.create() must NOT be called when transcript is empty"
        )
    }

    /// 음성 인식 결과가 공백만 있는 경우에도 메모를 저장해서는 안 됩니다.
    func test_perform_whitespaceOnlyTranscript_doesNotSaveMemo() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = "   \t\n  "

        // Act
        _ = try await sut.perform()

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.count,
            0,
            "MemoService.create() must NOT be called when transcript contains only whitespace"
        )
    }

    /// 빈 전사 결과일 때 `stopRecording()`은 여전히 호출되어야 합니다.
    /// 녹음이 열린 채로 남아서는 안 됩니다.
    func test_perform_emptyTranscript_stillStopsRecording() async throws {
        // Arrange
        mockRecognizer.simulatedTranscript = ""

        // Act
        _ = try await sut.perform()

        // Assert
        XCTAssertEqual(
            mockRecognizer.stopRecordingCallCount,
            1,
            "stopRecording() must always be called, even when transcript is empty"
        )
    }

    // MARK: - Edge Case: Incremental Transcript via transcriptUpdates

    /// `transcriptUpdates` 스트림을 통한 부분 전사가 발생할 때
    /// 최종 텍스트가 메모 내용으로 저장되어야 합니다.
    func test_perform_incrementalTranscript_savesLastPhrase() async throws {
        // Arrange — simulatedPhrases 방출 후 최종 transcript도 설정
        mockRecognizer.simulatedPhrases = ["장", "장보기", "장보기 목록"]
        mockRecognizer.simulatedTranscript = "장보기 목록"

        // Act
        _ = try await sut.perform()

        // Assert — final phrase must be saved
        XCTAssertEqual(
            mockMemoService.createCalls.first?.content,
            "장보기 목록",
            "The final phrase from transcriptUpdates must be saved as memo content"
        )
    }

    // MARK: - Error Handling

    /// `SpeechRecognizerError.permissionDenied` 발생 시 `perform()`은 에러를
    /// 상위로 전파해야 합니다. `MemoService.create()`는 호출되지 않아야 합니다.
    func test_perform_permissionDenied_throwsAndDoesNotSave() async {
        // Arrange
        mockRecognizer.simulatedError = SpeechRecognizerError.permissionDenied

        // Act & Assert
        do {
            _ = try await sut.perform()
            XCTFail("perform() must throw when speech recognition permission is denied")
        } catch {
            // expected
        }

        XCTAssertEqual(
            mockMemoService.createCalls.count,
            0,
            "MemoService.create() must NOT be called when permission is denied"
        )
    }

    /// `SpeechRecognizerError.notAvailable` 발생 시 `perform()`은 에러를
    /// 상위로 전파해야 합니다.
    func test_perform_notAvailable_throwsAndDoesNotSave() async {
        // Arrange
        mockRecognizer.isAvailable = false

        // Act & Assert
        do {
            _ = try await sut.perform()
            XCTFail("perform() must throw when speech recognition is not available")
        } catch {
            // expected
        }

        XCTAssertEqual(
            mockMemoService.createCalls.count,
            0,
            "MemoService.create() must NOT be called when speech recognition is unavailable"
        )
    }

    /// `MemoService.create()` 자체가 실패해도 `perform()`은 에러를 전파해야 합니다.
    func test_perform_memoServiceFailure_throwsError() async {
        // Arrange
        mockRecognizer.simulatedTranscript = "저장 실패 테스트"
        mockMemoService.simulatedCreateError = APIError.serverError(500)

        // Act & Assert
        do {
            _ = try await sut.perform()
            XCTFail("perform() must throw when MemoService.create() fails")
        } catch {
            // expected
        }
    }

    /// 에러가 발생한 경우에도 `stopRecording()`은 반드시 호출되어야 합니다.
    /// 리소스 누수를 방지해야 합니다.
    func test_perform_onError_alwaysStopsRecording() async {
        // Arrange
        mockRecognizer.simulatedError = SpeechRecognizerError.permissionDenied

        // Act
        do { _ = try await sut.perform() } catch {}

        // Assert
        XCTAssertEqual(
            mockRecognizer.stopRecordingCallCount,
            1,
            "stopRecording() must always be called to release microphone resources, even on error"
        )
    }
}
