// PocketAideApp.swift
// PocketAide

import SwiftUI

@main
struct PocketAideApp: App {

    /// UI 테스트 실행 중 여부. `--uitesting` 인자가 있으면 `true`.
    private let isUITesting: Bool = CommandLine.arguments.contains("--uitesting")

    /// 빈 알림 상태 UI 테스트 여부. `--uitesting-empty-notifications` 인자가 있으면 `true`.
    private let isUITestingEmptyNotifications: Bool =
        CommandLine.arguments.contains("--uitesting-empty-notifications")

    @StateObject private var authViewModel = AuthViewModel()

    init() {
        if CommandLine.arguments.contains("--uitesting") {
            injectMockData()
        }
    }

    var body: some Scene {
        WindowGroup {
            if isUITesting {
                // 인증/온보딩을 건너뛰고 TabBar로 바로 이동
                MainTabView()
                    .environmentObject(authViewModel)
            } else {
                if authViewModel.isAuthenticated {
                    MainTabView()
                        .environmentObject(authViewModel)
                } else {
                    LoginView()
                        .environmentObject(authViewModel)
                }
            }
        }
    }

    // MARK: - Mock Data Injection

    private func injectMockData() {
        injectMockNotifications()
        injectWidgetTestFlag()
    }

    private func injectWidgetTestFlag() {
        guard CommandLine.arguments.contains("--uitesting-widget") else { return }
        let defaults = UserDefaults(suiteName: NotificationRepository.appGroupIdentifier)
            ?? UserDefaults.standard
        defaults.set(true, forKey: "uitesting_widget")
    }

    private func injectMockNotifications() {
        let defaults = UserDefaults(suiteName: NotificationRepository.appGroupIdentifier)
            ?? UserDefaults.standard

        if isUITestingEmptyNotifications {
            // 빈 알림 상태 테스트: 빈 배열 주입
            let encoder = JSONEncoder()
            encoder.dateEncodingStrategy = .iso8601
            if let data = try? encoder.encode([AppNotification]()) {
                defaults.set(data, forKey: NotificationRepository.userDefaultsKey)
            }
            return
        }

        // 기본 mock 알림 데이터 주입
        let now = Date()
        let mockNotifications: [AppNotification] = [
            AppNotification(
                id: UUID(uuidString: "11111111-1111-1111-1111-111111111111")!,
                appName: "카카오톡",
                sender: "홍길동",
                body: "안녕하세요! 오늘 점심 같이 드실래요?",
                date: now.addingTimeInterval(-300)
            ),
            AppNotification(
                id: UUID(uuidString: "22222222-2222-2222-2222-222222222222")!,
                appName: "카카오톡",
                sender: "김철수",
                body: "회의 시간 변경되었습니다",
                date: now.addingTimeInterval(-1800)
            ),
            AppNotification(
                id: UUID(uuidString: "33333333-3333-3333-3333-333333333333")!,
                appName: "Slack",
                sender: "#general",
                body: "배포 완료되었습니다",
                date: now.addingTimeInterval(-600)
            ),
            AppNotification(
                id: UUID(uuidString: "44444444-4444-4444-4444-444444444444")!,
                appName: "문자",
                sender: "010-1234-5678",
                body: "인증번호: 123456",
                date: now.addingTimeInterval(-120)
            ),
            AppNotification(
                id: UUID(uuidString: "55555555-5555-5555-5555-555555555555")!,
                appName: "메일",
                sender: "team@company.com",
                body: "주간 보고서가 도착했습니다",
                date: now.addingTimeInterval(-900)
            ),
        ]

        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        if let data = try? encoder.encode(mockNotifications) {
            defaults.set(data, forKey: NotificationRepository.userDefaultsKey)
        }
    }
}
