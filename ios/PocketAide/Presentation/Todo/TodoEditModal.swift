// TodoEditModal.swift
// PocketAide

import SwiftUI

/// 투두를 추가하는 Sheet/Modal 화면.
///
/// AccessibilityIdentifier 목록:
/// - todo_title_field  : 제목 입력 필드
/// - todo_memo_field   : 메모 입력 필드
/// - todo_date_picker  : 날짜 피커
/// - todo_save_button  : 저장 버튼
struct TodoEditModal: View {

    @ObservedObject var viewModel: TodoViewModel
    @Environment(\.dismiss) private var dismiss

    // MARK: - Form State

    @State private var title: String = ""
    @State private var memo: String = ""
    @State private var dueDate: Date = Date()
    @State private var hasDueDate: Bool = false

    var body: some View {
        NavigationStack {
            Form {
                Section("할 일 정보") {
                    TextField("제목 (예: 장보기)", text: $title)
                        .accessibilityIdentifier("todo_title_field")

                    TextField("메모 (선택)", text: $memo)
                        .accessibilityIdentifier("todo_memo_field")
                }

                Section("날짜") {
                    Toggle("마감일 설정", isOn: $hasDueDate)

                    if hasDueDate {
                        DatePicker(
                            "마감일",
                            selection: $dueDate,
                            displayedComponents: .date
                        )
                        .datePickerStyle(.graphical)
                        .accessibilityIdentifier("todo_date_picker")
                    }
                }
            }
            .navigationTitle("할 일 추가")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("취소") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("저장") {
                        save()
                    }
                    .disabled(title.trimmingCharacters(in: .whitespaces).isEmpty)
                    .accessibilityIdentifier("todo_save_button")
                }
            }
        }
    }

    // MARK: - Actions

    private func save() {
        let trimmedTitle = title.trimmingCharacters(in: .whitespaces)
        guard !trimmedTitle.isEmpty else { return }

        let trimmedMemo = memo.trimmingCharacters(in: .whitespaces)
        Task {
            await viewModel.createTodo(
                title: trimmedTitle,
                note: trimmedMemo.isEmpty ? "" : trimmedMemo
            )
            dismiss()
        }
    }
}
