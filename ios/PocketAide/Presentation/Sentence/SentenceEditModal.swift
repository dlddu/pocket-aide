// SentenceEditModal.swift
// PocketAide

import SwiftUI

/// 새 문장을 추가하는 Sheet/Modal 화면.
///
/// AccessibilityIdentifier 목록:
/// - sentence_content_field    : 문장 내용 입력 필드
/// - sentence_category_picker  : 카테고리 선택 피커
/// - sentence_save_button      : 저장 버튼
struct SentenceEditModal: View {

    @ObservedObject var viewModel: SentenceViewModel
    @Environment(\.dismiss) private var dismiss

    // MARK: - Form State

    @State private var content: String = ""
    @State private var selectedCategoryId: Int = 0

    var body: some View {
        NavigationStack {
            Form {
                Section("문장 내용") {
                    TextField("문장을 입력하세요", text: $content)
                        .accessibilityIdentifier("sentence_content_field")
                }

                Section("카테고리") {
                    Picker("카테고리", selection: $selectedCategoryId) {
                        ForEach(viewModel.categories) { category in
                            Text(category.name).tag(category.id)
                        }
                    }
                    .pickerStyle(.wheel)
                    .accessibilityIdentifier("sentence_category_picker")
                }
            }
            .navigationTitle("문장 추가")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("취소") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("저장") {
                        save()
                    }
                    .disabled(content.trimmingCharacters(in: .whitespaces).isEmpty || selectedCategoryId == 0)
                    .accessibilityIdentifier("sentence_save_button")
                }
            }
            .onAppear {
                if let first = viewModel.categories.first {
                    selectedCategoryId = first.id
                }
            }
        }
    }

    // MARK: - Actions

    private func save() {
        let trimmedContent = content.trimmingCharacters(in: .whitespaces)
        guard !trimmedContent.isEmpty, selectedCategoryId != 0 else { return }

        Task {
            await viewModel.createSentence(content: trimmedContent, categoryId: selectedCategoryId)
            dismiss()
        }
    }
}
