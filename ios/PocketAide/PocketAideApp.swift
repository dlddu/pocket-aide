import SwiftUI

@main
struct PocketAideApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var auth = AppAuthCoordinator()
    @Environment(\.scenePhase) private var scenePhase

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(auth)
                .task { await auth.bootstrap() }
                .onChange(of: scenePhase) { _, newPhase in
                    if newPhase == .active {
                        Task { await auth.refreshPushAuthorization() }
                    }
                }
        }
    }
}
