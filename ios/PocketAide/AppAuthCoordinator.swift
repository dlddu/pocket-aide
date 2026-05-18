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
    @Published var meError: String?
    @Published var pushAuthorizationDenied = false

    let api: APIClient?
    let oidc: OIDCClient?
    private let tokenStore: TokenStoring
    private let pushRegistrar = PushRegistrar()

    init() {
        // Share the keychain item with the widget extension. Both targets
        // declare the same `keychain-access-groups` entitlement; the value
        // here must match (with the resolved `$(AppIdentifierPrefix)`).
        // If the Info.plist key is missing in some environment we fall back
        // to nil so the app still works standalone.
        let accessGroup = Bundle.main.object(forInfoDictionaryKey: "KeychainAccessGroup") as? String
        let resolvedGroup = accessGroup.flatMap { $0.isEmpty ? nil : $0 }
        let store = KeychainTokenStore(accessGroup: resolvedGroup)
        self.tokenStore = store
        if let api = try? APIClient.fromBundle(.main, tokenStore: store) {
            let oidc = OIDCClient(api: api, tokenStore: store)
            api.setTokenRefresher(oidc)
            self.api = api
            self.oidc = oidc
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
            await registerForPush()
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
        meError = nil
        do {
            me = try await api.me()
        } catch APIError.badStatus(401, _) {
            signOut()
        } catch {
            me = nil
            meError = String(describing: error)
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
            await registerForPush()
        } catch {
            signInError = String(describing: error)
            signedIn = (try? tokenStore.load()) != nil
        }
    }

    func signOut() {
        try? oidc?.signOut()
        me = nil
        meError = nil
        signedIn = false
    }

    func refreshPushAuthorization() async {
        let state = await pushRegistrar.currentAuthorization()
        pushAuthorizationDenied = (state == .denied)
    }

    private func registerForPush() async {
        guard let api else { return }
        let state = await pushRegistrar.register(api: api)
        pushAuthorizationDenied = (state == .denied)
    }
}
