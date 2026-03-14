// WhisperAPIRecognizer.swift
// PocketAide

import Foundation

/// OpenAI Whisper API 기반 음성 인식 구현체 (stub).
///
/// 실제 API 호출은 추후 구현합니다.
final class WhisperAPIRecognizer: SpeechRecognizerProtocol {

    // MARK: - SpeechRecognizerProtocol

    var isAvailable: Bool = true
    private(set) var isRecording: Bool = false
    private(set) var transcript: String = ""
    private(set) var error: Error? = nil

    private var streamContinuation: AsyncStream<String>.Continuation?

    var transcriptUpdates: AsyncStream<String> {
        AsyncStream { [weak self] continuation in
            self?.streamContinuation = continuation
        }
    }

    // MARK: - Init

    init() {}

    // MARK: - SpeechRecognizerProtocol Methods

    func startRecording() async throws {
        guard isAvailable else {
            throw SpeechRecognizerError.notAvailable
        }
        guard !isRecording else {
            throw SpeechRecognizerError.alreadyRecording
        }
        isRecording = true
    }

    func stopRecording() {
        isRecording = false
        streamContinuation?.finish()
        streamContinuation = nil
    }
}
