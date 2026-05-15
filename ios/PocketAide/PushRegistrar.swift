import Foundation
import PocketAideAPI
import UIKit
import UserNotifications

/// PushRegistrar owns the iOS-side half of the PR-monitor pipeline:
/// 1. Ask the user for notification permission.
/// 2. Register with APNs.
/// 3. POST the resulting device token to /api/device-tokens.
///
/// It also re-checks system permission on demand (for the foreground banner
/// in RootView) so the UI can prompt the user to re-enable notifications
/// without restarting the app.
@MainActor
final class PushRegistrar {
    /// Snapshot of the system push-authorization state, surfaced to the UI
    /// via AppAuthCoordinator so the user can see when notifications are off.
    enum AuthorizationState {
        case notDetermined
        case denied
        case authorized
        case unknown
    }

    private var tokenObserver: NSObjectProtocol?
    private var failureObserver: NSObjectProtocol?
    private weak var api: APIClient?

    deinit {
        if let tokenObserver { NotificationCenter.default.removeObserver(tokenObserver) }
        if let failureObserver { NotificationCenter.default.removeObserver(failureObserver) }
    }

    /// Asks for push permission (if not yet decided), registers with APNs on
    /// success, and forwards the resulting token to the backend. Returns the
    /// post-call authorization state so callers can show UI immediately.
    @discardableResult
    func register(api: APIClient) async -> AuthorizationState {
        self.api = api
        installObserversIfNeeded()

        let center = UNUserNotificationCenter.current()
        let current = await center.notificationSettings()

        switch current.authorizationStatus {
        case .denied:
            return .denied
        case .authorized, .provisional, .ephemeral:
            // Already granted: just (re-)register so we get a fresh token.
            UIApplication.shared.registerForRemoteNotifications()
            return .authorized
        case .notDetermined:
            do {
                let granted = try await center.requestAuthorization(options: [.alert, .badge, .sound])
                if granted {
                    UIApplication.shared.registerForRemoteNotifications()
                    return .authorized
                }
                return .denied
            } catch {
                return .unknown
            }
        @unknown default:
            return .unknown
        }
    }

    /// Re-reads the system authorization state without prompting. Use when
    /// the app comes back to the foreground (the user may have toggled the
    /// setting in Settings.app).
    func currentAuthorization() async -> AuthorizationState {
        let settings = await UNUserNotificationCenter.current().notificationSettings()
        switch settings.authorizationStatus {
        case .denied: return .denied
        case .authorized, .provisional, .ephemeral: return .authorized
        case .notDetermined: return .notDetermined
        @unknown default: return .unknown
        }
    }

    private func installObserversIfNeeded() {
        if tokenObserver == nil {
            tokenObserver = NotificationCenter.default.addObserver(
                forName: .pushTokenReceived,
                object: nil,
                queue: .main
            ) { [weak self] note in
                guard let self, let token = note.object as? String, let api = self.api else { return }
                Task { @MainActor in
                    do {
                        try await api.registerDeviceToken(token)
                    } catch {
                        // Surface as log only — registration retries on next
                        // launch via bootstrap(). No retry loop in the draft.
                        print("PushRegistrar: registerDeviceToken failed: \(error)")
                    }
                }
            }
        }
        if failureObserver == nil {
            failureObserver = NotificationCenter.default.addObserver(
                forName: .pushTokenRegistrationFailed,
                object: nil,
                queue: .main
            ) { note in
                if let err = note.object as? Error {
                    print("PushRegistrar: APNs registration failed: \(err)")
                }
            }
        }
    }
}
