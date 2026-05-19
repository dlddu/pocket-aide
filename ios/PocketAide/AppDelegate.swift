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

    /// Posted when the user taps a PocketAide push notification. Object is
    /// the deep-link URL synthesized from the push payload's `event_id`
    /// (PRD-10 AC7). PocketAideApp listens via .onReceive and routes the
    /// app to the PR monitor tab.
    static let pushNotificationOpened = Notification.Name("pushNotificationOpened")
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

        // Cold-start path: if the app launched in response to a push tap,
        // the system surfaces the originating notification here. Post the
        // deep link async so SwiftUI has time to install onReceive before
        // we publish — without the dispatch the broadcast arrives during
        // app construction and is dropped.
        if let response = launchOptions?[.remoteNotification] as? [AnyHashable: Any],
           let url = PRMonitorPushPayload.deepLinkURL(fromUserInfo: response) {
            DispatchQueue.main.async {
                NotificationCenter.default.post(name: .pushNotificationOpened, object: url)
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
    /// and broadcast it; PocketAideApp consumes the broadcast.
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let info = response.notification.request.content.userInfo
        if let url = PRMonitorPushPayload.deepLinkURL(fromUserInfo: info) {
            NotificationCenter.default.post(name: .pushNotificationOpened, object: url)
        }
        completionHandler()
    }
}
