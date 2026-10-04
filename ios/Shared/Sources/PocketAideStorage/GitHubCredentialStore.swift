import Foundation
import Security

public struct GitHubCredential: Codable, Equatable, Sendable {
    public let token: String
    public let login: String

    public init(token: String, login: String) {
        self.token = token
        self.login = login
    }
}

public protocol GitHubCredentialStoring: Sendable {
    func load() throws -> GitHubCredential?
    func save(_ credential: GitHubCredential) throws
    func clear() throws
}

public final class KeychainGitHubCredentialStore: GitHubCredentialStoring, @unchecked Sendable {
    private let service: String
    private let account: String

    public init(
        service: String = "com.dlddu.PocketAide.github",
        account: String = "primary"
    ) {
        self.service = service
        self.account = account
    }

    public func load() throws -> GitHubCredential? {
        var query = baseQuery()
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        query[kSecReturnData as String] = true

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        if status == errSecItemNotFound { return nil }
        guard status == errSecSuccess, let data = result as? Data else {
            throw KeychainError.osStatus(status)
        }
        return try JSONDecoder().decode(GitHubCredential.self, from: data)
    }

    public func save(_ credential: GitHubCredential) throws {
        let data = try JSONEncoder().encode(credential)
        var query = baseQuery()

        let attrs: [String: Any] = [kSecValueData as String: data]
        let status = SecItemUpdate(query as CFDictionary, attrs as CFDictionary)
        if status == errSecSuccess { return }
        if status == errSecItemNotFound {
            query[kSecValueData as String] = data
            let addStatus = SecItemAdd(query as CFDictionary, nil)
            guard addStatus == errSecSuccess else { throw KeychainError.osStatus(addStatus) }
            return
        }
        throw KeychainError.osStatus(status)
    }

    public func clear() throws {
        let status = SecItemDelete(baseQuery() as CFDictionary)
        if status == errSecSuccess || status == errSecItemNotFound { return }
        throw KeychainError.osStatus(status)
    }

    private func baseQuery() -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly,
        ]
    }
}
