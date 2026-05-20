import Foundation
import PocketAideAPI
import UIKit
import UserNotifications

extension Notification.Name {
    /// Posted when APNs hands us a device token. Object is the hex-encoded
    /// token (String). Listeners are PushRegistrar (to forward to the backend)
    /// and anyone who wants to react to fresh tokens.
    static let pushTokenReceived = Notification.Name("pushTokenReceived")

    /// Posted when APNs registration fails. Object is the underlying Error.
    static let pushTokenRegistrationFailed = Notification.Name("pushTokenRegistrationFailed")
}

/// AppDelegate exists solely to receive the APNs callbacks SwiftUI's App
/// scene cannot handle directly. Everything else is wired through the
/// SwiftUI lifecycle.
final class AppDelegate: NSObject, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        UNUserNotificationCenter.current().delegate = self

        // Cold-start path: the system surfaces the originating notification
        // here when the app launched in response to a push tap. Hand the URL
        // to DeepLinkRouter — it's an @Published store, so SwiftUI picks it
        // up when the scene mounts even if the assignment happens before the
        // scene has installed its observer.
        if let response = launchOptions?[.remoteNotification] as? [AnyHashable: Any] {
            NSLog("[AppDelegate] cold-start launchOptions push payload: %@", String(describing: response))
            if let url = PRMonitorPushPayload.deepLinkURL(fromUserInfo: response) {
                Task { @MainActor in
                    DeepLinkRouter.shared.receive(url)
                }
            }
        }
        return true
    }

    func application(
        _ application: UIApplication,
        didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
    ) {
        let hex = deviceToken.map { String(format: "%02x", $0) }.joined()
        NotificationCenter.default.post(name: .pushTokenReceived, object: hex)
    }

    func application(
        _ application: UIApplication,
        didFailToRegisterForRemoteNotificationsWithError error: Error
    ) {
        NotificationCenter.default.post(name: .pushTokenRegistrationFailed, object: error)
    }

    /// Show banner + sound when a push arrives while the app is foreground.
    /// Without this, foreground pushes are silently swallowed.
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        completionHandler([.banner, .sound, .badge])
    }

    /// Fires when the user taps a notification (foreground or background
    /// launch). PRD-10 AC7: the tap must route the app to the matching
    /// PR-monitor item but MUST NOT acknowledge it (AC12 explicit-button
    /// rule). We synthesize a deep-link URL from the payload's `event_id`
    /// and hand it to DeepLinkRouter; the scene observes that store.
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let info = response.notification.request.content.userInfo
        NSLog("[AppDelegate] didReceive response userInfo=%@", String(describing: info))
        if let url = PRMonitorPushPayload.deepLinkURL(fromUserInfo: info) {
            Task { @MainActor in
                DeepLinkRouter.shared.receive(url)
            }
        } else {
            NSLog("[AppDelegate] didReceive response: no event_id in payload, skipping deep link")
        }
        completionHandler()
    }
}
