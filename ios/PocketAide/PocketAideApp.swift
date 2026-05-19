import Foundation
import SwiftUI

@main
struct PocketAideApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var auth = AppAuthCoordinator()
    @StateObject private var deepLinkRouter = DeepLinkRouter.shared
    @Environment(\.scenePhase) private var scenePhase
    @State private var selectedTab: RootTab = .affirmations
    @State private var highlightedEventID: Int64?

    var body: some Scene {
        WindowGroup {
            RootView(selectedTab: $selectedTab, highlightedEventID: $highlightedEventID)
                .environmentObject(auth)
                .task { await auth.bootstrap() }
                .onChange(of: scenePhase) { _, newPhase in
                    if newPhase == .active {
                        Task { await auth.refreshPushAuthorization() }
                        // Drain any URL queued before the scene was active.
                        if let pending = deepLinkRouter.pendingURL {
                            NSLog("[PocketAideApp] scene active, draining pending URL %@", pending.absoluteString)
                            handleDeepLink(pending)
                            deepLinkRouter.pendingURL = nil
                        }
                    }
                }
                .onOpenURL { url in handleDeepLink(url) }
                .onChange(of: deepLinkRouter.pendingURL) { _, url in
                    guard let url else { return }
                    NSLog("[PocketAideApp] pendingURL changed -> %@", url.absoluteString)
                    handleDeepLink(url)
                    deepLinkRouter.pendingURL = nil
                }
        }
    }

    private func handleDeepLink(_ url: URL) {
        NSLog("[PocketAideApp] handleDeepLink url=%@ scheme=%@ host=%@",
              url.absoluteString, url.scheme ?? "<nil>", url.host ?? "<nil>")
        // Widget taps land here as pocketaide://<host>. OIDC callbacks use
        // the same scheme but are intercepted by ASWebAuthenticationSession,
        // so they don't reach onOpenURL.
        guard url.scheme == "pocketaide" else { return }
        switch url.host {
        case "affirmations":
            selectedTab = .affirmations
        case "pr-monitor":
            // Optional eventId query item — set so PRMonitorView can
            // highlight the matching card. PRMonitorView clears it after
            // its arrival window expires.
            if let comps = URLComponents(url: url, resolvingAgainstBaseURL: false),
               let raw = comps.queryItems?.first(where: { $0.name == "eventId" })?.value,
               let parsed = Int64(raw) {
                highlightedEventID = parsed
            } else {
                highlightedEventID = nil
            }
            selectedTab = .prMonitor
            NSLog("[PocketAideApp] selectedTab set to .prMonitor, highlightedEventID=%@",
                  highlightedEventID.map(String.init) ?? "<nil>")
        default:
            break
        }
    }
}
