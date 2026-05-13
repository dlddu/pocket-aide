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
        bootstrap()
    }

    // Called once at launch: if a token is already on the keychain, go
    // straight to signedIn. Otherwise, ask the backend's /api/auth/config —
    // when it reports a dev_auth_token_path (POCKET_AIDE_DEV=1 backend) we
    // automatically mint an oidcmock-signed token. Real prod backends omit
    // the field, so this falls through to the LoginView.
    private func bootstrap() {
        do {
            if try tokenStore.load() != nil {
                state = .signedIn
                return
            }
        } catch {
            // fall through, treat as signed out
        }
        Task { await self.resolveInitialAuth() }
    }

    private func resolveInitialAuth() async {
        guard let api else {
            state = .signedOut
            return
        }
        do {
            let cfg = try await api.authConfig()
            if let path = cfg.devAuthTokenPath, !path.isEmpty {
                await signInWithDevToken(path: path)
                return
            }
        } catch {
            // Backend unreachable / non-dev — surface LoginView as usual.
        }
        state = .signedOut
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

    // Hits the backend's POCKET_AIDE_DEV-only token endpoint advertised via
    // /api/auth/config.dev_auth_token_path, which mints an oidcmock-signed
    // access token. Stored in the keychain so the standard Authorization
    // header path works for subsequent API calls.
    private func signInWithDevToken(path: String) async {
        guard let api else {
            state = .signedOut
            return
        }
        do {
            let bundle = try await fetchDevTokenBundle(baseURL: api.baseURL, path: path)
            try tokenStore.save(bundle)
            state = .signedIn
        } catch {
            signInError = "mock 로그인에 실패했어요"
            state = .signedOut
        }
    }

    private func fetchDevTokenBundle(baseURL: URL, path: String) async throws -> TokenBundle {
        let trimmed = path.hasPrefix("/") ? String(path.dropFirst()) : path
        var req = URLRequest(url: baseURL.appendingPathComponent(trimmed))
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
