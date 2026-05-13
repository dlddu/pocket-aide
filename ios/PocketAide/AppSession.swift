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
        if CommandLine.arguments.contains("--ui-test-skip-auth") {
            state = .signedIn
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
}
