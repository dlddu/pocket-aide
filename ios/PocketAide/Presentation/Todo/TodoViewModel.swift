// TodoViewModel.swift
// PocketAide

import Foundation

/// 투두 화면의 상태와 비즈니스 로직을 관리하는 ViewModel.
@MainActor
final class TodoViewModel: ObservableObject {

    // MARK: - Published State

    @Published var todos: [Todo] = []
    @Published var isLoading = false
    @Published var errorMessage: String? = nil
    @Published var showAddSheet = false

    // MARK: - Computed Sections

    /// 완료되지 않은 투두 (진행중)
    var pendingTodos: [Todo] { todos.filter { !$0.isCompleted } }

    /// 완료된 투두
    var completedTodos: [Todo] { todos.filter { $0.isCompleted } }

    // MARK: - Dependencies

    private let todoService = TodoService()
    private let keychainService = KeychainService()

    // MARK: - Init

    init() {}

    // MARK: - Public Methods

    /// 투두 목록을 서버에서 불러옵니다.
    func loadTodos() async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            todos = try await todoService.list(type: "personal", serverURL: serverURL, token: token)
        } catch {
            errorMessage = "투두 목록을 불러오지 못했습니다."
        }
    }

    /// 새 투두를 생성합니다.
    func createTodo(title: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let created = try await todoService.create(
                title: title,
                type: "personal",
                serverURL: serverURL,
                token: token
            )
            todos.append(created)
        } catch {
            errorMessage = "투두를 생성하지 못했습니다."
        }
    }

    /// 투두를 수정합니다.
    func updateTodo(id: Int, title: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let updated = try await todoService.update(
                id: id,
                title: title,
                serverURL: serverURL,
                token: token
            )
            if let idx = todos.firstIndex(where: { $0.id == id }) {
                todos[idx] = updated
            }
        } catch {
            errorMessage = "투두를 수정하지 못했습니다."
        }
    }

    /// 투두를 삭제합니다.
    func deleteTodo(id: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await todoService.delete(id: id, serverURL: serverURL, token: token)
            todos.removeAll { $0.id == id }
        } catch {
            errorMessage = "투두를 삭제하지 못했습니다."
        }
    }

    /// 투두 완료 상태를 토글합니다.
    func toggleTodo(id: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let toggled = try await todoService.toggle(id: id, serverURL: serverURL, token: token)
            if let idx = todos.firstIndex(where: { $0.id == id }) {
                todos[idx] = toggled
            }
        } catch {
            errorMessage = "투두 상태 변경에 실패했습니다."
        }
    }
}
