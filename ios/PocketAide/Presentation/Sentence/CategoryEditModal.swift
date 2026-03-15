// CategoryEditModal.swift
// PocketAide

import SwiftUI

/// 새 카테고리를 추가하는 Sheet/Modal 화면.
///
/// AccessibilityIdentifier 목록:
/// - category_name_field   : 카테고리 이름 입력 필드
/// - category_save_button  : 저장 버튼
struct CategoryEditModal: View {

    @ObservedObject var viewModel: SentenceViewModel
    @Environment(\.dismiss) private var dismiss

    // MARK: - Form State

    @State private var name: String = ""

    var body: some View {
        NavigationStack {
            Form {
                Section("카테고리 이름") {
                    TextField("카테고리 이름을 입력하세요", text: $name)
                        .accessibilityIdentifier("category_name_field")
                }
            }
            .navigationTitle("카테고리 추가")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("취소") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("저장") {
                        save()
                    }
                    .disabled(name.trimmingCharacters(in: .whitespaces).isEmpty)
                    .accessibilityIdentifier("category_save_button")
                }
            }
        }
    }

    // MARK: - Actions

    private func save() {
        let trimmedName = name.trimmingCharacters(in: .whitespaces)
        guard !trimmedName.isEmpty else { return }

        Task {
            await viewModel.createCategory(name: trimmedName)
            dismiss()
        }
    }
}
