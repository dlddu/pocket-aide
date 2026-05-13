import Foundation
import PocketAideAPI
import PocketAideAuth
import PocketAideStorage
import SwiftUI

@MainActor
final class AppSession: ObservableObject {
    enum AuthState {
        case unknown
        case signedIn
        case signedOut
    }

    @Published var state: AuthState = .unknown
    @Published var isSigningIn = false
    @Published var signInError: String?

    let api: APIClient?
    private let oidc: OIDCClient?
    private let tokenStore: TokenStoring

    init() {
        let store = KeychainTokenStore(accessGroup: nil)
        self.tokenStore = store
        if let client = try? APIClient.fromBundle(.main, tokenStore: store) {
            self.api = client
            self.oidc = OIDCClient(api: client, tokenStore: store)
        } else {
            self.api = nil
            self.oidc = nil
        }
        if CommandLine.arguments.contains("--ui-test-mock-oidc") {
            Task { await signInWithDevToken() }
        } else {
            refreshState()
        }
    }

    var isSignedIn: Bool {
        if case .signedIn = state { return true }
        return false
    }

    func refreshState() {
        do {
            if try tokenStore.load() != nil {
                state = .signedIn
            } else {
                state = .signedOut
            }
        } catch {
            state = .signedOut
        }
    }

    func signIn() async {
        guard let oidc else {
            signInError = "백엔드 설정을 확인해주세요"
            return
        }
        isSigningIn = true
        signInError = nil
        defer { isSigningIn = false }
        do {
            _ = try await oidc.signIn()
            state = .signedIn
        } catch {
            signInError = "로그인에 실패했어요"
        }
    }

    func signOut() {
        try? oidc?.signOut()
        state = .signedOut
    }

    // Hits the backend's POCKET_AIDE_DEV-only /dev/auth-token endpoint, which
    // mints an oidcmock-signed access token. Stored in the keychain so the
    // standard Authorization header path works for subsequent API calls.
    private func signInWithDevToken() async {
        guard let api else {
            state = .signedOut
            return
        }
        do {
            let bundle = try await fetchDevTokenBundle(baseURL: api.baseURL)
            try tokenStore.save(bundle)
            state = .signedIn
        } catch {
            signInError = "mock 로그인에 실패했어요"
            state = .signedOut
        }
    }

    private func fetchDevTokenBundle(baseURL: URL) async throws -> TokenBundle {
        var req = URLRequest(url: baseURL.appendingPathComponent("dev/auth-token"))
        req.httpMethod = "POST"
        let (data, response) = try await URLSession.shared.data(for: req)
        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            throw URLError(.badServerResponse)
        }
        struct DevTokenResponse: Decodable {
            let access_token: String
            let expires_in: Int
        }
        let decoded = try JSONDecoder().decode(DevTokenResponse.self, from: data)
        return TokenBundle(
            accessToken: decoded.access_token,
            refreshToken: nil,
            expiresAt: Date().addingTimeInterval(TimeInterval(decoded.expires_in))
        )
    }
}
