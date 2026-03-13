// KeychainService.swift
// PocketAide

import Foundation
import Security

/// Keychain을 통해 토큰과 서버 주소를 안전하게 저장/로드/삭제합니다.
final class KeychainService {

    // MARK: - Keys

    private enum Key {
        static let token     = "com.pocket-aide.auth.token"
        static let serverURL = "com.pocket-aide.auth.serverURL"
    }

    // MARK: - Token

    /// Keychain에 인증 토큰을 저장합니다.
    func save(token: String) {
        save(value: token, forKey: Key.token)
    }

    /// Keychain에서 인증 토큰을 불러옵니다.
    /// 저장된 토큰이 없으면 nil을 반환합니다.
    func loadToken() -> String? {
        return load(forKey: Key.token)
    }

    /// Keychain에서 인증 토큰을 삭제합니다.
    func deleteToken() {
        delete(forKey: Key.token)
    }

    // MARK: - Server URL

    /// Keychain에 서버 주소를 저장합니다.
    func save(serverURL: String) {
        save(value: serverURL, forKey: Key.serverURL)
    }

    /// Keychain에서 서버 주소를 불러옵니다.
    /// 저장된 값이 없으면 nil을 반환합니다.
    func loadServerURL() -> String? {
        return load(forKey: Key.serverURL)
    }

    /// Keychain에서 서버 주소를 삭제합니다.
    func deleteServerURL() {
        delete(forKey: Key.serverURL)
    }

    // MARK: - Private Helpers

    private func save(value: String, forKey key: String) {
        guard let data = value.data(using: .utf8) else { return }

        // 기존 항목 삭제 후 새로 추가
        delete(forKey: key)

        let query: [CFString: Any] = [
            kSecClass:       kSecClassGenericPassword,
            kSecAttrAccount: key,
            kSecValueData:   data
        ]
        SecItemAdd(query as CFDictionary, nil)
    }

    private func load(forKey key: String) -> String? {
        let query: [CFString: Any] = [
            kSecClass:            kSecClassGenericPassword,
            kSecAttrAccount:      key,
            kSecReturnData:       true,
            kSecMatchLimit:       kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let value = String(data: data, encoding: .utf8)
        else {
            return nil
        }

        return value
    }

    private func delete(forKey key: String) {
        let query: [CFString: Any] = [
            kSecClass:       kSecClassGenericPassword,
            kSecAttrAccount: key
        ]
        SecItemDelete(query as CFDictionary)
    }
}
