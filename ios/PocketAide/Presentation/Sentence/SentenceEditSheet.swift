// SentenceEditSheet.swift
// PocketAide

import SwiftUI

/// 기존 문장을 편집하는 Sheet 화면.
///
/// AccessibilityIdentifier 목록:
/// - sentence_edit_field    : 문장 내용 편집 필드
/// - sentence_update_button : 수정 버튼
struct SentenceEditSheet: View {

    @ObservedObject var viewModel: SentenceViewModel
    let sentence: Sentence
    @Environment(\.dismiss) private var dismiss

    // MARK: - Form State

    @State private var content: String = ""

    var body: some View {
        NavigationStack {
            Form {
                Section("문장 내용") {
                    TextField("문장을 입력하세요", text: $content)
                        .accessibilityIdentifier("sentence_edit_field")
                }
            }
            .navigationTitle("문장 수정")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("취소") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("수정") {
                        update()
                    }
                    .disabled(content.trimmingCharacters(in: .whitespaces).isEmpty)
                    .accessibilityIdentifier("sentence_update_button")
                }
            }
            .onAppear {
                content = sentence.content
            }
        }
    }

    // MARK: - Actions

    private func update() {
        let trimmedContent = content.trimmingCharacters(in: .whitespaces)
        guard !trimmedContent.isEmpty else { return }

        Task {
            await viewModel.updateSentence(id: sentence.id, content: trimmedContent)
            dismiss()
        }
    }
}
