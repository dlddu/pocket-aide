import Foundation
import PocketAideAPI
import PocketAideAuth
import PocketAideStorage

@MainActor
final class AppAuthCoordinator: ObservableObject {
    @Published var signedIn: Bool
    @Published var isSigningIn = false
    @Published var signInError: String?
    @Published var healthOK = false
    @Published var healthError: String?
    @Published var me: MeResponse?

    let api: APIClient?
    let oidc: OIDCClient?
    private let tokenStore: TokenStoring

    init() {
        let store = KeychainTokenStore(accessGroup: nil)
        if ProcessInfo.processInfo.arguments.contains("-uitest-reset-keychain") {
            try? store.clear()
        }
        self.tokenStore = store
        if let api = try? APIClient.fromBundle(.main, tokenStore: store) {
            self.api = api
            self.oidc = OIDCClient(api: api, tokenStore: store)
        } else {
            self.api = nil
            self.oidc = nil
            self.healthError = "BackendBaseURL missing"
        }
        self.signedIn = (try? store.load()) != nil
    }

    func bootstrap() async {
        await refreshHealth()
        if signedIn {
            await refreshMe()
        }
    }

    func refreshHealth() async {
        guard let api else { return }
        do {
            try await api.health()
            healthOK = true
            healthError = nil
        } catch {
            healthOK = false
            healthError = String(describing: error)
        }
    }

    func refreshMe() async {
        guard let api else { return }
        do {
            me = try await api.me()
        } catch {
            me = nil
        }
    }

    func signIn() async {
        guard let oidc else { return }
        isSigningIn = true
        signInError = nil
        defer { isSigningIn = false }
        do {
            _ = try await oidc.signIn()
            signedIn = true
            await refreshMe()
        } catch {
            signInError = String(describing: error)
            signedIn = (try? tokenStore.load()) != nil
        }
    }

    func signOut() {
        try? oidc?.signOut()
        me = nil
        signedIn = false
    }
}
