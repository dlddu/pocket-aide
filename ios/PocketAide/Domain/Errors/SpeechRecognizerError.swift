// SpeechRecognizerError.swift
// PocketAide

/// 음성 인식 중 발생할 수 있는 에러 타입.
public enum SpeechRecognizerError: Error, Equatable {
    /// 이미 녹음이 진행 중일 때 다시 시작을 시도한 경우.
    case alreadyRecording
    /// 음성 인식 권한이 없는 경우.
    case permissionDenied
    /// 디바이스에서 음성 인식을 사용할 수 없는 경우.
    case notAvailable
}
