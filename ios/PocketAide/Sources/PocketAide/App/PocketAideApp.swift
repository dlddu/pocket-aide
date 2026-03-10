// Application entry point — compiled only on Apple platforms that ship SwiftUI.
// On Linux (CI) this file is excluded from compilation by the canImport guard.

#if canImport(SwiftUI)
import SwiftUI

@main
public struct PocketAideApp: App {
    public init() {}

    public var body: some Scene {
        WindowGroup {
            MainTabView()
        }
    }
}
#endif
