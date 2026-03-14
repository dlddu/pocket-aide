// ScratchScreen.swift
// PocketAide

import SwiftUI

/// 임시 공간 메모 목록을 표시하는 메인 화면.
///
/// AccessibilityIdentifier 목록:
/// - scratch_list_view         : root container
/// - add_memo_button           : 우측 상단 + 버튼
/// - memo_row_{content}        : 각 메모 행
/// - memo_move_button_{content}: 각 행의 이동 버튼
struct ScratchScreen: View {

    @StateObject private var viewModel = ScratchViewModel()
    @State private var memoToMove: Memo? = nil

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.memos.isEmpty && !viewModel.isLoading {
                    emptyView
                } else {
                    memoListView
                }
            }
            .navigationTitle("임시 공간")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        viewModel.showAddSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                    .accessibilityIdentifier("add_memo_button")
                }
            }
            .sheet(isPresented: $viewModel.showAddSheet) {
                AddMemoModal(viewModel: viewModel)
            }
            .sheet(item: $memoToMove) { memo in
                MoveDestinationSheet(memo: memo, viewModel: viewModel)
            }
            .task {
                await viewModel.loadMemos()
            }
            .refreshable {
                await viewModel.loadMemos()
            }
        }
        .accessibilityIdentifier("scratch_list_view")
    }

    // MARK: - Subviews

    private var memoListView: some View {
        List {
            ForEach(viewModel.memos) { memo in
                MemoRow(memo: memo, onMove: {
                    memoToMove = memo
                })
            }
            .onDelete { indexSet in
                let memos = viewModel.memos
                for index in indexSet {
                    Task {
                        await viewModel.deleteMemo(id: memos[index].id)
                    }
                }
            }
        }
    }

    private var emptyView: some View {
        VStack(spacing: 16) {
            Image(systemName: "note.text")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)
            Text("메모가 없습니다")
                .font(.headline)
            Text("+ 버튼으로 메모를 추가해보세요")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - MemoRow

private struct MemoRow: View {

    let memo: Memo
    let onMove: () -> Void

    /// source에 따른 아이콘.
    private var sourceIcon: String {
        memo.source == "voice" ? "mic.fill" : "pencil"
    }

    var body: some View {
        HStack {
            Image(systemName: sourceIcon)
                .foregroundStyle(.secondary)
                .font(.caption)
                .accessibilityIdentifier("memo_source_icon_\(memo.content)")

            Text(memo.content)
                .font(.body)

            Spacer()

            Button {
                onMove()
            } label: {
                Image(systemName: "arrow.right.square")
                    .foregroundStyle(.blue)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("memo_move_button_\(memo.content)")
        }
        .accessibilityIdentifier("memo_row_\(memo.content)")
    }
}

// MARK: - AddMemoModal

struct AddMemoModal: View {

    @ObservedObject var viewModel: ScratchViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var content: String = ""

    var body: some View {
        NavigationStack {
            Form {
                Section("메모 내용") {
                    TextField("메모를 입력하세요", text: $content)
                        .accessibilityIdentifier("memo_text_field")
                }
            }
            .navigationTitle("메모 추가")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("취소") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("저장") {
                        save()
                    }
                    .disabled(content.trimmingCharacters(in: .whitespaces).isEmpty)
                    .accessibilityIdentifier("memo_save_button")
                }
            }
        }
    }

    private func save() {
        let trimmedContent = content.trimmingCharacters(in: .whitespaces)
        guard !trimmedContent.isEmpty else { return }

        Task {
            await viewModel.createMemo(content: trimmedContent)
            dismiss()
        }
    }
}

// MARK: - MoveDestinationSheet

struct MoveDestinationSheet: View {

    let memo: Memo
    @ObservedObject var viewModel: ScratchViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 20) {
            Text("이동할 위치를 선택하세요")
                .font(.headline)
                .padding(.top, 20)

            Button {
                Task {
                    await viewModel.moveMemo(id: memo.id, target: "personal_todo")
                    dismiss()
                }
            } label: {
                HStack {
                    Image(systemName: "checkmark.square")
                    Text("개인 투두로")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.blue.opacity(0.1))
                .cornerRadius(10)
            }
            .accessibilityIdentifier("move_to_personal_todo_button")

            Button {
                Task {
                    await viewModel.moveMemo(id: memo.id, target: "work_todo")
                    dismiss()
                }
            } label: {
                HStack {
                    Image(systemName: "briefcase")
                    Text("회사 투두로")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.orange.opacity(0.1))
                .cornerRadius(10)
            }
            .accessibilityIdentifier("move_to_work_todo_button")

            Button(role: .destructive) {
                Task {
                    await viewModel.deleteMemo(id: memo.id)
                    dismiss()
                }
            } label: {
                HStack {
                    Image(systemName: "trash")
                    Text("삭제")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.red.opacity(0.1))
                .cornerRadius(10)
            }
            .accessibilityIdentifier("delete_memo_button")

            Spacer()
        }
        .padding()
        .accessibilityIdentifier("move_destination_sheet")
    }
}
