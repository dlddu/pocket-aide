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

    private let memoService: MemoServiceProtocol
    private let keychainService = KeychainService()

    // MARK: - Credential Strategy

    /// 자격 증명 공급 전략.
    private enum CredentialStrategy {
        /// KeychainService에서 동적으로 로드합니다.
        case keychain
        /// 고정된 자격 증명을 사용합니다. nil이면 자격 증명 없음으로 처리합니다.
        case fixed(serverURL: String?, token: String?)
    }

    private let credentialStrategy: CredentialStrategy

    // MARK: - Init

    /// 프로덕션 초기화 — 실제 MemoService와 KeychainService를 사용합니다.
    init() {
        self.memoService = MemoService()
        self.credentialStrategy = .keychain
    }

    /// 테스트용 초기화 — MemoServiceProtocol을 주입합니다.
    /// 자격 증명은 KeychainService에서 로드합니다.
    ///
    /// 이 초기화를 사용할 때 Keychain에 유효한 자격 증명이 없으면
    /// `createMemo()` 등의 메서드는 errorMessage를 설정하고 조기 반환합니다.
    /// 자격 증명을 제어하려면 `init(memoService:serverURL:token:)`을 사용하세요.
    ///
    /// - Parameter memoService: 주입할 MemoServiceProtocol 구현체
    init(memoService: some MemoServiceProtocol) {
        self.memoService = memoService
        // 테스트에서는 Keychain이 비어 있을 수 있으므로
        // 기본 테스트 자격 증명을 제공합니다.
        self.credentialStrategy = .fixed(
            serverURL: "https://test.example.com",
            token: "test-token"
        )
    }

    /// 테스트용 초기화 — MemoServiceProtocol과 자격 증명을 모두 주입합니다.
    ///
    /// - Parameters:
    ///   - memoService: 주입할 MemoServiceProtocol 구현체
    ///   - serverURL: 서버 주소. nil이면 자격 증명 없음으로 처리합니다.
    ///   - token: 인증 토큰. nil이면 자격 증명 없음으로 처리합니다.
    init(memoService: some MemoServiceProtocol, serverURL: String?, token: String?) {
        self.memoService = memoService
        self.credentialStrategy = .fixed(serverURL: serverURL, token: token)
    }

    // MARK: - Private Helpers

    /// 현재 인증 자격 증명을 반환합니다.
    private func credentials() -> (serverURL: String, token: String)? {
        switch credentialStrategy {
        case .keychain:
            guard let url = keychainService.loadServerURL(),
                  let tok = keychainService.loadToken() else {
                return nil
            }
            return (url, tok)

        case .fixed(let serverURL, let token):
            guard let url = serverURL, let tok = token else {
                return nil
            }
            return (url, tok)
        }
    }

    // MARK: - Public Methods

    /// 메모 목록을 서버에서 불러옵니다.
    func loadMemos() async {
        guard let creds = credentials() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            if let concreteMemoService = memoService as? MemoService {
                memos = try await concreteMemoService.list(serverURL: creds.serverURL, token: creds.token)
            }
        } catch {
            errorMessage = "메모 목록을 불러오지 못했습니다."
        }
    }

    /// 새 메모를 생성합니다.
    ///
    /// - Parameters:
    ///   - content: 메모 내용
    ///   - source: 메모 출처. 기본값은 `"text"`.
    func createMemo(content: String, source: String = "text") async {
        guard let creds = credentials() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let created = try await memoService.create(
                content: content,
                source: source,
                serverURL: creds.serverURL,
                token: creds.token
            )
            memos.append(created)
        } catch {
            errorMessage = "메모를 생성하지 못했습니다."
        }
    }

    /// 메모를 삭제합니다.
    func deleteMemo(id: Int) async {
        guard let creds = credentials() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            if let concreteMemoService = memoService as? MemoService {
                try await concreteMemoService.delete(id: id, serverURL: creds.serverURL, token: creds.token)
            }
            memos.removeAll { $0.id == id }
        } catch {
            errorMessage = "메모를 삭제하지 못했습니다."
        }
    }

    /// 메모를 투두로 이동합니다.
    func moveMemo(id: Int, target: String) async {
        guard let creds = credentials() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            if let concreteMemoService = memoService as? MemoService {
                try await concreteMemoService.move(id: id, target: target, serverURL: creds.serverURL, token: creds.token)
            }
            memos.removeAll { $0.id == id }
        } catch {
            errorMessage = "메모를 이동하지 못했습니다."
        }
    }
}
