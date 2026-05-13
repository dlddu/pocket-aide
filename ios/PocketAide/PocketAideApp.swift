import SwiftUI

@main
struct PocketAideApp: App {
    @StateObject private var session = AppSession()

    var body: some Scene {
        WindowGroup {
            Group {
                switch session.state {
                case .unknown:
                    Color.clear
                case .signedIn:
                    RootView()
                case .signedOut:
                    LoginView()
                }
            }
            .environmentObject(session)
        }
    }
}
