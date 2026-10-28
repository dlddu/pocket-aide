import Foundation
import PocketAideAPI
import PocketAideAuth
import PocketAideStorage
import UIKit
import UserNotifications

extension Notification.Name {
    /// Posted when APNs hands us a device token. Object is the hex-encoded
    /// token (String).
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
        UNUserNotificationCenter.current().setNotificationCategories([Self.approvalCategory])

        // Cold-start path: the system surfaces the originating notification
        // here when the app launched in response to a push tap. Hand the URL
        // to DeepLinkRouter — it's an @Published store, so SwiftUI picks it
        // up when the scene mounts even if the assignment happens before the
        // scene has installed its observer.
        if let response = launchOptions?[.remoteNotification] as? [AnyHashable: Any] {
            NSLog("[AppDelegate] cold-start launchOptions push payload: %@", String(describing: response))
            if let url = PushDeepLink.url(fromUserInfo: response) {
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

    /// PRD-10 AC7: a notification tap routes the app to the matching PR-monitor
    /// item but MUST NOT acknowledge it (AC12 explicit-button rule).
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let info = response.notification.request.content.userInfo
        NSLog("[AppDelegate] didReceive response userInfo=%@", String(describing: info))
        if let decision = ApprovalNotificationAction.decision(forActionIdentifier: response.actionIdentifier),
           let approvalID = ApprovalPushPayload.approvalID(from: info) {
            decideFromNotification(approvalID, decision, replacing: response.notification.request, completion: completionHandler)
            return
        }
        if let url = PushDeepLink.url(fromUserInfo: info) {
            Task { @MainActor in
                DeepLinkRouter.shared.receive(url)
            }
        }
        completionHandler()
    }

    private static var approvalCategory: UNNotificationCategory {
        UNNotificationCategory(
            identifier: ApprovalNotificationAction.categoryIdentifier,
            actions: [
                UNNotificationAction(
                    identifier: ApprovalNotificationAction.approveIdentifier,
                    title: ApprovalNotificationAction.title(.approve),
                    options: [.authenticationRequired]
                ),
                UNNotificationAction(
                    identifier: ApprovalNotificationAction.rejectIdentifier,
                    title: ApprovalNotificationAction.title(.reject),
                    options: [.destructive]
                ),
            ],
            intentIdentifiers: []
        )
    }

    private func decideFromNotification(
        _ approvalID: String,
        _ decision: ApprovalDecision,
        replacing request: UNNotificationRequest,
        completion: @escaping () -> Void
    ) {
        Task { @MainActor in
            let result: Result<ApprovalDecisionOutcome, Error>
            do {
                guard let api = Self.makeActionAPI() else { throw APIError.missingBaseURL }
                result = .success(try await api.decideApproval(id: approvalID, decision))
            } catch {
                result = .failure(error)
            }
            let content = UNMutableNotificationContent()
            content.title = request.content.title
            content.body = ApprovalNotificationAction.resultNotice(result)
            content.userInfo = request.content.userInfo
            try? await UNUserNotificationCenter.current().add(
                UNNotificationRequest(identifier: request.identifier, content: content, trigger: nil)
            )
            completion()
        }
    }

    @MainActor
    private static func makeActionAPI() -> APIClient? {
        let accessGroup = (Bundle.main.object(forInfoDictionaryKey: "KeychainAccessGroup") as? String)
            .flatMap { $0.isEmpty ? nil : $0 }
        let store = KeychainTokenStore(accessGroup: accessGroup)
        guard let api = try? APIClient.fromBundle(.main, tokenStore: store) else { return nil }
        api.setTokenRefresher(OIDCClient(api: api, tokenStore: store))
        return api
    }
}
