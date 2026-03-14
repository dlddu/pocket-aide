// AssistantView.swift
// PocketAide

import SwiftUI

// MARK: - AssistantView

struct AssistantView: View {
    var body: some View {
        ChatScreen()
    }
}

// MARK: - ChatScreen

struct ChatScreen: View {

    @StateObject private var viewModel = ChatViewModel()

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Model Selector Bar
                ModelSelectorBar(viewModel: viewModel)

                // Message List
                MessageList(messages: viewModel.messages, isLoading: viewModel.isLoading)

                // Error Banner
                if let error = viewModel.errorMessage {
                    HStack {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundColor(.orange)
                        Text(error)
                            .font(.caption)
                            .foregroundColor(.secondary)
                        Spacer()
                        Button {
                            viewModel.errorMessage = nil
                        } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundColor(.secondary)
                        }
                    }
                    .padding(.horizontal)
                    .padding(.vertical, 6)
                    .background(Color(.systemGray6))
                    .accessibilityIdentifier("error_banner")
                }

                // Voice Input Indicator (shown when recording)
                if viewModel.isVoiceRecording {
                    VoiceInputIndicator()
                }

                // Input Bar
                ChatInputBar(viewModel: viewModel)
            }
            .accessibilityIdentifier("assistant_chat_view")
            .navigationTitle("Assistant")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    NavigationLink(destination: SettingsView()) {
                        Image(systemName: "gearshape")
                    }
                    .accessibilityIdentifier("settings_button")
                }
            }
        }
        .sheet(isPresented: $viewModel.showModelPicker) {
            ModelPickerSheet(viewModel: viewModel)
        }
        .alert(
            "음성 인식 오류",
            isPresented: Binding(
                get: { viewModel.voiceErrorMessage != nil },
                set: { if !$0 { viewModel.clearVoiceError() } }
            ),
            actions: {
                Button("확인") { viewModel.clearVoiceError() }
            },
            message: {
                Text(viewModel.voiceErrorMessage ?? "")
            }
        )
    }
}

// MARK: - ModelSelectorBar

private struct ModelSelectorBar: View {

    @ObservedObject var viewModel: ChatViewModel

    var body: some View {
        HStack {
            Button(action: {
                viewModel.showModelPicker = true
            }) {
                HStack(spacing: 4) {
                    Text(viewModel.selectedModel)
                        .font(.subheadline)
                        .accessibilityIdentifier("selected_model_label")
                    Image(systemName: "chevron.down")
                        .font(.caption)
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(Color(.systemGray6))
                .cornerRadius(8)
            }
            .accessibilityIdentifier("model_selector_button")

            Spacer()
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
    }
}

// MARK: - ModelPickerSheet

private struct ModelPickerSheet: View {

    @ObservedObject var viewModel: ChatViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: 0) {
                ForEach(viewModel.availableModels, id: \.self) { model in
                    Button(action: {
                        viewModel.selectedModel = model
                        viewModel.showModelPicker = false
                    }) {
                        HStack {
                            Text(model)
                                .foregroundColor(.primary)
                            Spacer()
                            if model == viewModel.selectedModel {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.accentColor)
                            }
                        }
                        .padding()
                    }
                    Divider()
                }
                Spacer()
            }
            .accessibilityIdentifier("model_picker")
            .navigationTitle("모델 선택")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("닫기") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MARK: - MessageList

private struct MessageList: View {

    let messages: [ChatMessage]
    let isLoading: Bool

    var body: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 12) {
                    if messages.isEmpty {
                        Text("새 대화를 시작하세요")
                            .foregroundColor(.secondary)
                            .frame(maxWidth: .infinity)
                            .padding(.top, 40)
                    } else {
                        ForEach(messages) { message in
                            MessageBubble(message: message)
                                .id(message.id)
                        }
                    }

                    if isLoading {
                        TypingIndicator()
                            .id("typing-indicator")
                    }
                }
                .padding()
            }
            .onChange(of: messages.count) { _ in
                if let last = messages.last {
                    withAnimation {
                        proxy.scrollTo(last.id, anchor: .bottom)
                    }
                }
            }
        }
    }
}

// MARK: - TypingIndicator

private struct TypingIndicator: View {

    @State private var animating = false

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<3, id: \.self) { index in
                Circle()
                    .fill(Color(.systemGray3))
                    .frame(width: 8, height: 8)
                    .scaleEffect(animating ? 1.0 : 0.5)
                    .animation(
                        .easeInOut(duration: 0.6)
                            .repeatForever()
                            .delay(Double(index) * 0.2),
                        value: animating
                    )
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(Color(.systemGray5))
        .cornerRadius(16)
        .accessibilityIdentifier("typing_indicator")
        .onAppear { animating = true }
    }
}

// MARK: - VoiceInputIndicator

private struct VoiceInputIndicator: View {

    @State private var animating = false

    var body: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(Color.red)
                .frame(width: 12, height: 12)
                .scaleEffect(animating ? 1.3 : 0.8)
                .animation(
                    .easeInOut(duration: 0.6).repeatForever(autoreverses: true),
                    value: animating
                )
            Text("녹음 중...")
                .font(.caption)
                .foregroundColor(.red)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
        .background(Color.red.opacity(0.1))
        .cornerRadius(12)
        .accessibilityIdentifier("voice_input_indicator")
        .onAppear { animating = true }
    }
}

// MARK: - MessageBubble

private struct MessageBubble: View {

    let message: ChatMessage

    var isUser: Bool { message.role == .user }

    var body: some View {
        HStack {
            if isUser { Spacer() }

            Text(message.content)
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(isUser ? Color.blue : Color(.systemGray5))
                .foregroundColor(isUser ? .white : .primary)
                .cornerRadius(16)
                .accessibilityIdentifier(isUser ? "user_message_bubble" : "ai_response_bubble")

            if !isUser { Spacer() }
        }
    }
}

// MARK: - ChatInputBar

private struct ChatInputBar: View {

    @ObservedObject var viewModel: ChatViewModel

    var body: some View {
        HStack(spacing: 8) {
            TextField("메시지를 입력하세요", text: $viewModel.inputText)
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(.systemGray6))
                .cornerRadius(20)
                .accessibilityIdentifier("chat_input_field")
                .onSubmit {
                    viewModel.sendMessage()
                }

            // Mic Button
            Button(action: {
                Task {
                    try? await viewModel.toggleVoiceInput()
                }
            }) {
                Image(systemName: viewModel.isVoiceRecording ? "mic.fill" : "mic")
                    .font(.system(size: 24))
                    .foregroundColor(viewModel.isVoiceRecording ? .red : .blue)
            }
            .accessibilityIdentifier("mic_button")

            // Send Button
            Button(action: {
                viewModel.sendMessage()
            }) {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.system(size: 32))
                    .foregroundColor(viewModel.inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? .gray : .blue)
            }
            .disabled(viewModel.inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || viewModel.isLoading)
            .accessibilityIdentifier("send_button")
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
        .shadow(radius: 1)
    }
}
