// RoutineEditScreen.swift
// PocketAide

import SwiftUI

/// 루틴을 추가하거나 수정하는 Sheet/Modal 화면.
///
/// AccessibilityIdentifier 목록:
/// - routine_name_field      : 이름 입력 필드
/// - routine_interval_field  : 주기 입력 필드
/// - routine_last_done_field : 마지막 수행일 입력 필드
/// - routine_save_button     : 저장 버튼
struct RoutineEditScreen: View {

    @ObservedObject var viewModel: RoutineViewModel
    @Environment(\.dismiss) private var dismiss

    // MARK: - Editing Target (nil = create mode)

    var editingRoutine: Routine? = nil

    // MARK: - Form State

    @State private var name: String = ""
    @State private var intervalDays: String = "1"
    @State private var lastDoneAt: String = ""

    private var isEditing: Bool { editingRoutine != nil }

    var body: some View {
        NavigationStack {
            Form {
                Section("루틴 정보") {
                    TextField("이름 (예: 샤워)", text: $name)
                        .accessibilityIdentifier("routine_name_field")

                    TextField("주기 (일)", text: $intervalDays)
                        .keyboardType(.numberPad)
                        .accessibilityIdentifier("routine_interval_field")

                    TextField("마지막 수행일 (YYYY-MM-DD)", text: $lastDoneAt)
                        .keyboardType(.numbersAndPunctuation)
                        .accessibilityIdentifier("routine_last_done_field")
                }
            }
            .navigationTitle(isEditing ? "루틴 수정" : "루틴 추가")
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
                    .accessibilityIdentifier("routine_save_button")
                }
            }
            .onAppear {
                if let routine = editingRoutine {
                    name = routine.name
                    intervalDays = "\(routine.intervalDays)"
                    lastDoneAt = routine.lastDoneAt
                } else {
                    lastDoneAt = todayString()
                }
            }
        }
    }

    // MARK: - Actions

    private func save() {
        let trimmedName = name.trimmingCharacters(in: .whitespaces)
        guard !trimmedName.isEmpty else { return }

        let days = Int(intervalDays) ?? 1
        let dateStr = lastDoneAt.isEmpty ? todayString() : lastDoneAt

        Task {
            if let routine = editingRoutine {
                await viewModel.updateRoutine(id: routine.id, name: trimmedName, intervalDays: days)
            } else {
                await viewModel.createRoutine(name: trimmedName, intervalDays: days, lastDoneAt: dateStr)
            }
            dismiss()
        }
    }

    private func todayString() -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = TimeZone(identifier: "UTC")
        return formatter.string(from: Date())
    }
}
