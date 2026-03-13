// ChatViewModel.swift
// PocketAide

import Foundation

/// 채팅 화면의 상태와 비즈니스 로직을 관리하는 ViewModel.
@MainActor
final class ChatViewModel: ObservableObject {

    // MARK: - Available Models

    let availableModels = ["mock", "gpt-4o", "claude-3-opus", "gemini-pro"]

    // MARK: - Published State

    @Published var messages: [ChatMessage] = []
    @Published var inputText: String = ""
    @Published var selectedModel: String = "mock"
    @Published var isLoading: Bool = false
    @Published var errorMessage: String? = nil
    @Published var showModelPicker: Bool = false

    // MARK: - Dependencies

    private let chatService = ChatService()
    private let keychainService = KeychainService()
    private let isUITesting: Bool

    // MARK: - Init

    init(isUITesting: Bool = CommandLine.arguments.contains("--uitesting")) {
        self.isUITesting = isUITesting
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
