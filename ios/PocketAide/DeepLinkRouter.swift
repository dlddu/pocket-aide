import Foundation

/// Holds a deep-link URL that was received outside of SwiftUI's
/// `.onOpenURL` (e.g. from a UNUserNotificationCenter delegate callback or
/// from launchOptions during cold start) until the scene can consume it.
///
/// A NotificationCenter publisher → `.onReceive` hand-off loses links that fire
/// before SwiftUI installs the subscriber (e.g. a tap during the
/// background→foreground transition), so the URL is stored instead.
@MainActor
final class DeepLinkRouter: ObservableObject {
    static let shared = DeepLinkRouter()
    @Published var pendingURL: URL?

    func receive(_ url: URL) {
        NSLog("[DeepLinkRouter] receive %@", url.absoluteString)
        pendingURL = url
    }
}
