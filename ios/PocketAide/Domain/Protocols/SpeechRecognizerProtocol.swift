// SpeechRecognizerProtocol.swift
// PocketAide

/// 음성 인식기의 공개 인터페이스.
///
/// 프로덕션 구현체(`SpeechRecognizer`)와 테스트용 대역(`MockSpeechRecognizer`)
/// 모두 이 프로토콜을 준수합니다.
public protocol SpeechRecognizerProtocol: AnyObject {
    /// 현재 디바이스/권한 상태에서 음성 인식 사용 가능 여부.
    var isAvailable: Bool { get set }

    /// 현재 녹음 중인지 여부.
    var isRecording: Bool { get }

    /// 가장 최근 전사(transcription) 텍스트.
    var transcript: String { get }

    /// 마지막으로 발생한 에러.
    var error: Error? { get }

    /// 전사 텍스트가 갱신될 때마다 새 값을 방출하는 비동기 스트림.
    var transcriptUpdates: AsyncStream<String> { get }

    /// 녹음을 시작합니다.
    /// - Throws: ``SpeechRecognizerError``
    func startRecording() async throws

    /// 녹음을 중지합니다. 녹음 중이 아닐 때 호출하면 아무 동작도 하지 않습니다.
    func stopRecording()
}
