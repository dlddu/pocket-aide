import Foundation
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
    private var onToken: (@MainActor (String) -> Void)?

    deinit {
        if let tokenObserver { NotificationCenter.default.removeObserver(tokenObserver) }
        if let failureObserver { NotificationCenter.default.removeObserver(failureObserver) }
    }

    /// 알림 권한을 확인·요청하고 APNs 등록을 건다. 토큰이 오면 `onToken` 이 받아 서버 등록을 맡는다.
    @discardableResult
    func register(onToken: @escaping @MainActor (String) -> Void) async -> AuthorizationState {
        self.onToken = onToken
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
                guard let self, let token = note.object as? String else { return }
                Task { @MainActor in self.onToken?(token) }
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
