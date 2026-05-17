import Foundation
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
}
