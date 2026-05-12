import Foundation
import Security

public struct TokenBundle: Codable, Equatable, Sendable {
    public let accessToken: String
    public let refreshToken: String?
    public let expiresAt: Date

    public init(accessToken: String, refreshToken: String?, expiresAt: Date) {
        self.accessToken = accessToken
        self.refreshToken = refreshToken
        self.expiresAt = expiresAt
    }
}

public protocol TokenStoring: Sendable {
    func load() throws -> TokenBundle?
    func save(_ bundle: TokenBundle) throws
    func clear() throws
}

public final class KeychainTokenStore: TokenStoring, @unchecked Sendable {
    private let service: String
    private let account: String
    private let accessGroup: String?

    public init(
        service: String = "com.dlddu.PocketAide.tokens",
        account: String = "primary",
        accessGroup: String? = nil
    ) {
        self.service = service
        self.account = account
        self.accessGroup = accessGroup
    }

    public func load() throws -> TokenBundle? {
        var query = baseQuery()
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        query[kSecReturnData as String] = true

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        if status == errSecItemNotFound { return nil }
        guard status == errSecSuccess, let data = result as? Data else {
            throw KeychainError.osStatus(status)
        }
        return try JSONDecoder().decode(TokenBundle.self, from: data)
    }

    public func save(_ bundle: TokenBundle) throws {
        let data = try JSONEncoder().encode(bundle)
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
        var q: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly,
        ]
        if let accessGroup {
            q[kSecAttrAccessGroup as String] = accessGroup
        }
        return q
    }
}

public enum KeychainError: Error, CustomStringConvertible {
    case osStatus(OSStatus)

    public var description: String {
        switch self {
        case .osStatus(let s): return "keychain OSStatus=\(s)"
        }
    }
}
