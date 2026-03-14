// ChatViewModel.swift
// PocketAide

import Foundation

/// 채팅 화면의 상태와 비즈니스 로직을 관리하는 ViewModel.
@MainActor
final class ChatViewModel: ObservableObject {

    // MARK: - Available Models

    let availableModels: [String]

    // MARK: - Published State

    @Published var messages: [ChatMessage] = []
    @Published var inputText: String = ""
    @Published var selectedModel: String
    @Published var isLoading: Bool = false
    @Published var errorMessage: String? = nil
    @Published var showModelPicker: Bool = false

    // MARK: - Voice Published State

    @Published var isVoiceRecording: Bool = false
    @Published var voiceErrorMessage: String? = nil
    @Published var selectedSpeechEngine: SpeechEngine = .whisperLocal

    // MARK: - Dependencies

    private let chatService = ChatService()
    private let keychainService = KeychainService()
    private let isUITesting: Bool
    private var speechRecognizer: SpeechRecognizerProtocol?
    private var transcriptTask: Task<Void, Never>?

    // MARK: - Init

    init(
        availableModels: [String] = {
            if let modelsEnv = ProcessInfo.processInfo.environment["LLM_AVAILABLE_MODELS"],
               !modelsEnv.isEmpty {
                return modelsEnv.components(separatedBy: ",")
            }
            return ["gpt-4o", "claude-3-opus", "gemini-pro"]
        }(),
        defaultModel: String? = ProcessInfo.processInfo.environment["LLM_DEFAULT_MODEL"],
        isUITesting: Bool = CommandLine.arguments.contains("--uitesting"),
        speechRecognizer: SpeechRecognizerProtocol? = nil
    ) {
        self.availableModels = availableModels
        self.selectedModel = defaultModel ?? availableModels.first ?? "gpt-4o"
        self.isUITesting = isUITesting

        if let injected = speechRecognizer {
            self.speechRecognizer = injected
        } else if isUITesting {
            let mock = MockSpeechRecognizer()
            mock.simulatedTranscript = "음성 인식 테스트 메시지"
            self.speechRecognizer = mock
        } else {
            self.speechRecognizer = WhisperLocalRecognizer()
        }
    }

    // MARK: - Public Methods

    /// 현재 입력된 메시지를 전송합니다.
    func sendMessage() {
        let message = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !message.isEmpty, !isLoading else { return }

        inputText = ""

        // Add user message immediately
        let userMessage = ChatMessage(role: .user, content: message)
        messages.append(userMessage)

        isLoading = true
        errorMessage = nil

        Task {
            if isUITesting {
                await sendDummyMessage(userMessage: message)
            } else {
                await sendRealMessage(userMessage: message)
            }
            isLoading = false
        }
    }

    // MARK: - Voice Input Methods

    /// 음성 입력을 시작합니다.
    func startVoiceInput() async throws {
        guard let recognizer = speechRecognizer else { return }

        do {
            // 트랜스크립트 업데이트 스트림 구독
            transcriptTask?.cancel()
            transcriptTask = Task { [weak self] in
                guard let self else { return }
                for await phrase in recognizer.transcriptUpdates {
                    if Task.isCancelled { break }
                    self.inputText = phrase
                }
            }

            try await recognizer.startRecording()
            isVoiceRecording = true

            // startRecording 완료 후 transcript 반영
            let transcript = recognizer.transcript
            if !transcript.isEmpty {
                inputText = transcript
            }
        } catch {
            isVoiceRecording = false
            transcriptTask?.cancel()
            transcriptTask = nil
            voiceErrorMessage = error.localizedDescription
            throw error
        }
    }

    /// 음성 입력을 중지합니다.
    func stopVoiceInput() {
        speechRecognizer?.stopRecording()
        isVoiceRecording = false
        transcriptTask?.cancel()
        transcriptTask = nil
    }

    /// 음성 입력을 토글합니다. 현재 녹음 중이면 중지, 아니면 시작합니다.
    func toggleVoiceInput() async throws {
        if isVoiceRecording {
            stopVoiceInput()
        } else {
            try await startVoiceInput()
        }
    }

    /// 음성 에러 메시지를 초기화합니다.
    func clearVoiceError() {
        voiceErrorMessage = nil
    }

    // MARK: - Private Methods

    private func sendDummyMessage(userMessage: String) async {
        // Simulate network delay
        try? await Task.sleep(nanoseconds: 300_000_000)
        let aiMessage = ChatMessage(role: .assistant, content: "This is a dummy AI response for UI testing.")
        messages.append(aiMessage)
    }

    private func sendRealMessage(userMessage: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        // Streaming AI response
        var aiResponseContent = ""
        var aiMessageIndex: Int? = nil

        do {
            _ = try await chatService.sendMessage(
                message: userMessage,
                model: selectedModel,
                serverURL: serverURL,
                token: token
            ) { [weak self] token in
                guard let self else { return }
                Task { @MainActor in
                    if aiMessageIndex == nil {
                        let aiMessage = ChatMessage(role: .assistant, content: token)
                        self.messages.append(aiMessage)
                        aiMessageIndex = self.messages.count - 1
                        aiResponseContent = token
                    } else if let index = aiMessageIndex {
                        aiResponseContent += " " + token
                        self.messages[index] = ChatMessage(
                            id: self.messages[index].id,
                            role: .assistant,
                            content: aiResponseContent
                        )
                    }
                }
            }
        } catch {
            errorMessage = "메시지 전송에 실패했습니다."
        }
    }
}
