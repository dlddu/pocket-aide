import SwiftUI

@main
struct PocketAideApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var auth = AppAuthCoordinator()
    @Environment(\.scenePhase) private var scenePhase
    @State private var selectedTab: RootTab = .affirmations

    var body: some Scene {
        WindowGroup {
            RootView(selectedTab: $selectedTab)
                .environmentObject(auth)
                .task { await auth.bootstrap() }
                .onChange(of: scenePhase) { _, newPhase in
                    if newPhase == .active {
                        Task { await auth.refreshPushAuthorization() }
                    }
                }
                .onOpenURL { url in handleDeepLink(url) }
        }
    }

    private func handleDeepLink(_ url: URL) {
        // Widget taps land here as pocketaide://<host>. OIDC callbacks use
        // the same scheme but are intercepted by ASWebAuthenticationSession,
        // so they don't reach onOpenURL.
        guard url.scheme == "pocketaide" else { return }
        switch url.host {
        case "affirmations":
            selectedTab = .affirmations
        default:
            break
        }
    }
}
