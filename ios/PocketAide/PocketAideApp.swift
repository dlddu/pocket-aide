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

struct RootView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator

    var body: some View {
        Group {
            if auth.signedIn {
                HelloWorldView()
            } else {
                LoginView()
            }
        }
    }
}
