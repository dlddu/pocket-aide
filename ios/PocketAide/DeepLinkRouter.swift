import Foundation

/// Holds a deep-link URL that was received outside of SwiftUI's
/// `.onOpenURL` (e.g. from a UNUserNotificationCenter delegate callback or
/// from launchOptions during cold start) until the scene can consume it.
///
/// Replaces the earlier NotificationCenter.publisher → .onReceive flow,
/// which had subtle timing windows where the publisher fired before SwiftUI
/// installed the subscriber (foreground tap → background→foreground
/// transition) and lost the deep link.
@MainActor
final class DeepLinkRouter: ObservableObject {
    static let shared = DeepLinkRouter()
    @Published var pendingURL: URL?

    func receive(_ url: URL) {
        NSLog("[DeepLinkRouter] receive %@", url.absoluteString)
        pendingURL = url
    }
}
