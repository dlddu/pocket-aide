// ScratchViewModel.swift
// PocketAide

import Foundation
import SwiftUI

/// 임시 공간 화면의 상태와 비즈니스 로직을 관리하는 ViewModel.
@MainActor
final class ScratchViewModel: ObservableObject {

    // MARK: - Published State

    @Published var memos: [Memo] = []
    @Published var isLoading = false
    @Published var errorMessage: String? = nil
    @Published var showAddSheet = false

    // MARK: - Dependencies

    private let memoService = MemoService()
    private let keychainService = KeychainService()

    // MARK: - Init

    init() {}

    // MARK: - Public Methods

    /// 메모 목록을 서버에서 불러옵니다.
    func loadMemos() async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            memos = try await memoService.list(serverURL: serverURL, token: token)
        } catch {
            errorMessage = "메모 목록을 불러오지 못했습니다."
        }
    }

    /// 새 텍스트 메모를 생성합니다.
    func createMemo(content: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let created = try await memoService.create(
                content: content,
                source: "text",
                serverURL: serverURL,
                token: token
            )
            memos.append(created)
        } catch {
            errorMessage = "메모를 생성하지 못했습니다."
        }
    }

    /// 메모를 삭제합니다.
    func deleteMemo(id: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await memoService.delete(id: id, serverURL: serverURL, token: token)
            memos.removeAll { $0.id == id }
        } catch {
            errorMessage = "메모를 삭제하지 못했습니다."
        }
    }

    /// 메모를 투두로 이동합니다.
    func moveMemo(id: Int, target: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await memoService.move(id: id, target: target, serverURL: serverURL, token: token)
            memos.removeAll { $0.id == id }
        } catch {
            errorMessage = "메모를 이동하지 못했습니다."
        }
    }
}
