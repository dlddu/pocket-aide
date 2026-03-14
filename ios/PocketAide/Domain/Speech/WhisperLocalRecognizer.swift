// WhisperLocalRecognizer.swift
// PocketAide

import Foundation

/// whisper.cpp 기반 로컬 음성 인식 구현체 (stub).
///
/// 실제 whisper.cpp 연동은 추후 구현합니다.
final class WhisperLocalRecognizer: SpeechRecognizerProtocol {

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
