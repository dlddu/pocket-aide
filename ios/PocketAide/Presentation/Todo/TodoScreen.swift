// TodoScreen.swift
// PocketAide

import SwiftUI

/// 개인 투두 목록을 표시하는 메인 화면.
///
/// AccessibilityIdentifier 목록:
/// - todo_list_view          : root container
/// - todo_section_pending    : 진행중 섹션
/// - todo_section_completed  : 완료 섹션
/// - add_todo_button         : 우측 상단 + 버튼
/// - todo_row_{title}        : 각 투두 행
/// - todo_checkbox_{title}   : 각 행의 체크박스 버튼
struct TodoScreen: View {

    @StateObject private var viewModel = TodoViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.todos.isEmpty && !viewModel.isLoading {
                    emptyView
                } else {
                    todoListView
                }
            }
            .navigationTitle("투두")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        viewModel.showAddSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                    .accessibilityIdentifier("add_todo_button")
                }
            }
            .sheet(isPresented: $viewModel.showAddSheet) {
                TodoEditModal(viewModel: viewModel)
            }
            .task {
                await viewModel.loadTodos()
            }
            .refreshable {
                await viewModel.loadTodos()
            }
        }
        .accessibilityIdentifier("todo_list_view")
    }

    // MARK: - Subviews

    private var todoListView: some View {
        List {
            Section("진행중") {
                ForEach(viewModel.pendingTodos) { todo in
                    TodoRow(todo: todo, viewModel: viewModel)
                }
                .onDelete { indexSet in
                    let pending = viewModel.pendingTodos
                    for index in indexSet {
                        Task {
                            await viewModel.deleteTodo(id: pending[index].id)
                        }
                    }
                }
            }
            .accessibilityIdentifier("todo_section_pending")

            Section("완료") {
                ForEach(viewModel.completedTodos) { todo in
                    TodoRow(todo: todo, viewModel: viewModel)
                }
                .onDelete { indexSet in
                    let completed = viewModel.completedTodos
                    for index in indexSet {
                        Task {
                            await viewModel.deleteTodo(id: completed[index].id)
                        }
                    }
                }
            }
            .accessibilityIdentifier("todo_section_completed")
        }
    }

    private var emptyView: some View {
        VStack(spacing: 16) {
            Image(systemName: "checkmark.square")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)
            Text("할 일이 없습니다")
                .font(.headline)
            Text("+ 버튼으로 할 일을 추가해보세요")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - TodoRow

/// 투두 목록의 개별 행 뷰.
private struct TodoRow: View {

    let todo: Todo
    let viewModel: TodoViewModel

    var body: some View {
        HStack {
            Button {
                Task {
                    await viewModel.toggleTodo(id: todo.id)
                }
            } label: {
                Image(systemName: todo.isCompleted ? "checkmark.circle.fill" : "circle")
                    .foregroundStyle(todo.isCompleted ? .green : .secondary)
                    .font(.title2)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("todo_checkbox_\(todo.title)")

            Text(todo.title)
                .font(.body)
                .strikethrough(todo.isCompleted)
                .foregroundStyle(todo.isCompleted ? .secondary : .primary)

            Spacer()
        }
        .accessibilityIdentifier("todo_row_\(todo.title)")
    }
}
