// AuthViewModel.swift
// PocketAide

import Foundation

/// 인증 상태와 로그인/로그아웃 로직을 관리하는 ViewModel.
@MainActor
final class AuthViewModel: ObservableObject {

    // MARK: - Published State

    @Published var isAuthenticated: Bool = false
    @Published var isLoading: Bool = false
    @Published var errorMessage: String? = nil

    // MARK: - Input Fields

    @Published var serverAddress: String = ""
    @Published var email: String = ""
    @Published var password: String = ""

    // MARK: - Dependencies

    private let authService = AuthService()
    private let keychainService = KeychainService()

    // MARK: - Init

    init() {
        checkStoredCredentials()
    }

    // MARK: - Public Methods

    /// Keychain에 저장된 토큰과 서버 주소를 확인해 자동 로그인합니다.
    func checkStoredCredentials() {
        if let token = keychainService.loadToken(), !token.isEmpty,
           let savedServerURL = keychainService.loadServerURL(), !savedServerURL.isEmpty {
            serverAddress = savedServerURL
            isAuthenticated = true
        }
    }

    /// 입력된 자격 증명으로 로그인을 시도합니다.
    func login() {
        guard !serverAddress.isEmpty, !email.isEmpty, !password.isEmpty else {
            errorMessage = "서버 주소, 이메일, 비밀번호를 모두 입력해주세요."
            return
        }

        isLoading = true
        errorMessage = nil

        Task {
            do {
                let token = try await authService.login(
                    email: email,
                    password: password,
                    serverURL: serverAddress
                )
                keychainService.save(token: token)
                keychainService.save(serverURL: serverAddress)
                isAuthenticated = true
            } catch {
                errorMessage = "로그인에 실패했습니다. 자격 증명을 확인해주세요."
            }
            isLoading = false
        }
    }

    /// 로그아웃합니다. Keychain에서 토큰과 서버 주소를 삭제합니다.
    func logout() {
        keychainService.deleteToken()
        keychainService.deleteServerURL()
        isAuthenticated = false
        email = ""
        password = ""
        errorMessage = nil
    }
}
