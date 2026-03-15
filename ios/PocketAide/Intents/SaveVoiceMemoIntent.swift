// SaveVoiceMemoIntent.swift
// PocketAide

import Foundation

/// Siri Shortcut에서 음성 메모를 녹음하고 저장하는 Intent.
///
/// 이 구조체는 의존성 주입을 통해 테스트 가능하도록 설계되었습니다.
/// `SpeechRecognizerProtocol`과 `MemoServiceProtocol`을 주입받아
/// 실제 마이크/네트워크 없이 유닛 테스트할 수 있습니다.
struct SaveVoiceMemoIntent {

    // MARK: - Static Properties

    /// Siri가 실행 전 "정말로 실행할까요?"를 묻지 않도록 `false`로 설정합니다.
    static let confirmationRequired: Bool = false

    // MARK: - Dependencies

    private let speechRecognizer: SpeechRecognizerProtocol
    private let memoService: MemoServiceProtocol

    // MARK: - Credential Strategy

    /// 자격 증명 공급 전략.
    private enum CredentialStrategy {
        /// KeychainService에서 동적으로 로드합니다.
        case keychain(KeychainService)
        /// 고정된 자격 증명을 사용합니다 (테스트용).
        case fixed(serverURL: String, token: String)
    }

    private let credentialStrategy: CredentialStrategy

    // MARK: - Init

    /// 프로덕션 초기화 — KeychainService를 통해 자격 증명을 동적으로 로드합니다.
    init(speechRecognizer: SpeechRecognizerProtocol,
         memoService: MemoServiceProtocol,
         keychainService: KeychainService) {
        self.speechRecognizer = speechRecognizer
        self.memoService = memoService
        self.credentialStrategy = .keychain(keychainService)
    }

    /// 테스트용 초기화 — 고정 자격 증명을 사용합니다.
    ///
    /// 테스트에서는 Keychain이 비어 있을 수 있으므로
    /// 기본 테스트 자격 증명을 제공합니다.
    init(speechRecognizer: SpeechRecognizerProtocol,
         memoService: MemoServiceProtocol) {
        self.speechRecognizer = speechRecognizer
        self.memoService = memoService
        self.credentialStrategy = .fixed(
            serverURL: "https://test.example.com",
            token: "test-token"
        )
    }

    // MARK: - Private Helpers

    /// 현재 인증 자격 증명을 반환합니다.
    private func credentials() -> (serverURL: String, token: String)? {
        switch credentialStrategy {
        case .keychain(let keychainService):
            guard let url = keychainService.loadServerURL(),
                  let tok = keychainService.loadToken() else {
                return nil
            }
            return (url, tok)

        case .fixed(let serverURL, let token):
            return (serverURL, token)
        }
    }

    // MARK: - Perform

    /// Intent를 실행합니다.
    ///
    /// 실행 흐름:
    /// 1. `stopRecording()`을 defer로 등록하여 항상 호출되도록 보장
    /// 2. `startRecording()` 호출 — 권한 없거나 사용 불가 시 throw
    /// 3. `transcript` 읽기
    /// 4. 빈 텍스트/공백이면 빈 결과 반환 (저장 안 함)
    /// 5. `memoService.create(source: "voice")` 호출
    /// 6. 성공 메시지가 담긴 `IntentResult` 반환
    ///
    /// - Returns: 완료 메시지가 담긴 `IntentResult`
    /// - Throws: `SpeechRecognizerError` 또는 `APIError`
    func perform() async throws -> IntentResult {
        defer {
            speechRecognizer.stopRecording()
        }

        try await speechRecognizer.startRecording()

        let transcript = speechRecognizer.transcript
        let trimmed = transcript.trimmingCharacters(in: .whitespacesAndNewlines)

        guard !trimmed.isEmpty else {
            return IntentResult(value: "")
        }

        guard let creds = credentials() else {
            return IntentResult(value: "서버 주소 또는 인증 토큰이 없습니다.")
        }

        _ = try await memoService.create(
            content: trimmed,
            source: "voice",
            serverURL: creds.serverURL,
            token: creds.token
        )

        return IntentResult(value: "음성 메모가 저장되었습니다.")
    }
}
