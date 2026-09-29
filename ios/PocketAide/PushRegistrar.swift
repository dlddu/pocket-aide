import Foundation
import PocketAideAPI
import UIKit
import UserNotifications

@MainActor
final class PushRegistrar {
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
                        // Log only — bootstrap() registers again on the next launch.
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
