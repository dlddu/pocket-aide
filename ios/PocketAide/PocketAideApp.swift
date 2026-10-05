import Foundation
import SwiftUI
import UIKit

@main
struct PocketAideApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var auth = AppAuthCoordinator()
    @StateObject private var deepLinkRouter = DeepLinkRouter.shared
    @Environment(\.scenePhase) private var scenePhase
    @State private var selectedTab: RootTab = .affirmations
    @State private var highlightedEventID: Int64?
    @State private var showsWeather = false

    var body: some Scene {
        WindowGroup {
            RootView(selectedTab: $selectedTab, highlightedEventID: $highlightedEventID)
                .environmentObject(auth)
                .sheet(isPresented: $showsWeather) { WeatherView() }
                .task { await auth.bootstrap() }
                .onChange(of: scenePhase) { oldPhase, newPhase in
                    NSLog("[PocketAideApp] scenePhase %@ -> %@",
                          String(describing: oldPhase), String(describing: newPhase))
                    if newPhase == .active {
                        Task { await auth.refreshPushAuthorization() }
                        if auth.signedIn {
                            WeatherLocationAccess.start()
                        }
                        drainPendingURL(label: "scene-active")
                    }
                }
                .onOpenURL { url in handleDeepLink(url) }
                .onChange(of: deepLinkRouter.pendingURL) { _, url in
                    guard url != nil else { return }
                    drainPendingURL(label: "router-onChange")
                }
        }
    }

    /// Pull the pending deep link out of the router and apply it on the next
    /// main-runloop turn. Async-dispatch is the safety belt that fixes the
    /// "background → push tap → wrong tab" symptom: when SwiftUI processes
    /// the scene activation, the TabView re-installs its selection binding
    /// AFTER the @StateObject's value change has fired. A synchronous
    /// selectedTab = .prMonitor here loses the race; deferring one turn
    /// guarantees the TabView is ready to observe the update.
    private func drainPendingURL(label: String) {
        guard let pending = deepLinkRouter.pendingURL else {
            NSLog("[PocketAideApp] drainPendingURL(%@) called with no pending URL", label)
            return
        }
        NSLog("[PocketAideApp] drainPendingURL(%@) %@", label, pending.absoluteString)
        deepLinkRouter.pendingURL = nil
        DispatchQueue.main.async {
            handleDeepLink(pending)
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
            showsWeather = false
            selectedTab = .affirmations
        case "weather":
            showsWeather = true
        case "calendar":
            showsWeather = false
            openCalendar()
        case "pr-monitor":
            showsWeather = false
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

    private func openCalendar() {
        let seconds = Int(Date().timeIntervalSinceReferenceDate)
        guard let url = URL(string: "calshow:\(seconds)") else { return }
        UIApplication.shared.open(url) { opened in
            if !opened {
                selectedTab = .affirmations
            }
        }
    }
}
