import SwiftUI

@main
struct PocketAideApp: App {
    @StateObject private var auth = AppAuthCoordinator()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(auth)
                .task { await auth.bootstrap() }
        }
    }
}
