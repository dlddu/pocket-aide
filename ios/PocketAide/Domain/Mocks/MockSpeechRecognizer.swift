// MockSpeechRecognizer.swift
// PocketAide

import Foundation

/// `SpeechRecognizerProtocol`의 테스트 대역(Test Double).
///
/// 실제 음성 인식 API 없이 녹음 동작을 시뮬레이션합니다.
/// 유닛 테스트에서 `simulatedTranscript`, `simulatedPhrases`, `simulatedError`를
/// 설정해 다양한 시나리오를 재현할 수 있습니다.
public final class MockSpeechRecognizer: SpeechRecognizerProtocol {

    // MARK: - SpeechRecognizerProtocol

    public var isAvailable: Bool = true
    public private(set) var isRecording: Bool = false
    public private(set) var transcript: String = ""
    public private(set) var error: Error? = nil

    /// `startRecording()` 호출 시 `transcript`에 즉시 반영할 텍스트.
    public var simulatedTranscript: String = ""

    /// `transcriptUpdates` 스트림으로 순차 방출할 부분 전사 문구 배열.
    public var simulatedPhrases: [String] = []

    /// `startRecording()` 호출 시 throw할 에러. `nil`이면 정상 동작.
    public var simulatedError: Error? = nil

    /// `startRecording()`이 호출된 총 횟수.
    public private(set) var startRecordingCallCount: Int = 0

    /// `stopRecording()`이 호출된 총 횟수.
    public private(set) var stopRecordingCallCount: Int = 0

    // MARK: - AsyncStream support

    /// `transcriptUpdates` 스트림의 continuation.
    /// `transcriptUpdates` 프로퍼티에 최초 접근할 때 생성됩니다.
    private var streamContinuation: AsyncStream<String>.Continuation?

    /// 스트림이 구독되기 전에 `startRecording()`이 phrases를 yield하려 했을 때
    /// 버퍼에 저장해두고, 스트림 구독 시 소비합니다.
    private var pendingPhrases: [String] = []

    /// 전사 텍스트 갱신을 수신하는 `AsyncStream<String>`.
    ///
    /// `startRecording()` 이후에 이 프로퍼티에 접근하면 이미 버퍼된
    /// phrases를 즉시 방출합니다.
    public var transcriptUpdates: AsyncStream<String> {
        let buffered = pendingPhrases
        pendingPhrases = []
        return AsyncStream { [weak self] continuation in
            self?.streamContinuation = continuation
            // 스트림 구독 전에 쌓인 phrases를 즉시 방출
            for phrase in buffered {
                continuation.yield(phrase)
            }
        }
    }

    // MARK: - Init

    public init() {}

    // MARK: - SpeechRecognizerProtocol Methods

    public func startRecording() async throws {
        startRecordingCallCount += 1

        // 1. 사용 불가 상태 확인
        guard isAvailable else {
            throw SpeechRecognizerError.notAvailable
        }

        // 2. 시뮬레이션된 에러 확인
        if let simulatedError {
            throw simulatedError
        }

        // 3. 이미 녹음 중인지 확인
        guard !isRecording else {
            throw SpeechRecognizerError.alreadyRecording
        }

        // 4. 녹음 시작
        isRecording = true
        transcript = simulatedTranscript

        // 5. simulatedPhrases 방출
        //    continuation이 이미 존재하면 직접 yield하고,
        //    아직 없으면 pendingPhrases에 버퍼링합니다.
        let phrases = simulatedPhrases
        if !phrases.isEmpty {
            if let continuation = streamContinuation {
                for phrase in phrases {
                    continuation.yield(phrase)
                }
            } else {
                pendingPhrases.append(contentsOf: phrases)
            }
        }
    }

    public func stopRecording() {
        stopRecordingCallCount += 1
        isRecording = false
        streamContinuation?.finish()
        streamContinuation = nil
    }

    // MARK: - Test Helpers

    /// 목 객체를 초기 상태로 되돌립니다.
    public func reset() {
        isRecording = false
        transcript = ""
        error = nil
        simulatedTranscript = ""
        simulatedPhrases = []
        simulatedError = nil
        startRecordingCallCount = 0
        stopRecordingCallCount = 0
        pendingPhrases = []
        streamContinuation?.finish()
        streamContinuation = nil
    }
}
