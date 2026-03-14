// SpeechEngine.swift
// PocketAide

import Foundation

/// 음성 인식 엔진 선택.
///
/// `@AppStorage`와 호환되도록 `String` raw value를 사용합니다.
public enum SpeechEngine: String, CaseIterable, Equatable {
    case whisperLocal = "Whisper Local"
    case whisperAPI = "Whisper API"
}
