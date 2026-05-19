import Foundation
import SwiftUI

@main
struct PocketAideApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var auth = AppAuthCoordinator()
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
                    }
                }
                .onOpenURL { url in handleDeepLink(url) }
                .onReceive(NotificationCenter.default.publisher(for: .pushNotificationOpened)) { note in
                    if let url = note.object as? URL {
                        handleDeepLink(url)
                    }
                }
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
        case "pr-monitor":
            selectedTab = .prMonitor
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
        default:
            break
        }
    }
}
