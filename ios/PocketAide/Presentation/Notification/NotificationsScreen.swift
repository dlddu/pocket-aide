// NotificationsScreen.swift
// PocketAide

import SwiftUI

/// 알림 모음 목록을 표시하는 메인 화면 (읽기 전용).
///
/// AccessibilityIdentifier 목록:
/// - notification_list_view                     : 루트 List/NavigationStack
/// - notification_section_{appName}             : 각 앱 섹션
/// - notification_count_{appName}               : 건수 레이블
/// - notification_row_{appName}_msg_{n}         : 각 알림 행 (1-based)
/// - notification_toggle_{appName}              : 섹션 접기/펼치기 버튼
/// - notification_empty_view                    : 빈 상태 뷰
struct NotificationsScreen: View {

    @StateObject private var viewModel = NotificationsViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.groupedNotifications.isEmpty {
                    emptyView
                } else {
                    notificationListView
                }
            }
            .navigationTitle("알림")
            .navigationBarTitleDisplayMode(.large)
            .task {
                viewModel.load()
            }
            .refreshable {
                viewModel.load()
            }
        }
        .accessibilityIdentifier("notification_list_view")
    }

    // MARK: - Subviews

    private var notificationListView: some View {
        List {
            ForEach(viewModel.sortedAppNames, id: \.self) { appName in
                let notifications = viewModel.groupedNotifications[appName] ?? []
                NotificationSectionView(
                    appName: appName,
                    notifications: notifications,
                    isExpanded: viewModel.binding(for: appName)
                )
            }
        }
    }

    private var emptyView: some View {
        VStack(spacing: 16) {
            Image(systemName: "bell.slash")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)
            Text("알림이 없습니다")
                .font(.headline)
            Text("새로운 알림이 도착하면 여기에 표시됩니다")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .accessibilityIdentifier("notification_empty_view")
    }
}

// MARK: - NotificationSectionView

private struct NotificationSectionView: View {

    let appName: String
    let notifications: [AppNotification]
    @Binding var isExpanded: Bool

    private var emoji: String {
        NotificationRepository.emoji(for: appName)
    }

    var body: some View {
        Section {
            DisclosureGroup(
                isExpanded: $isExpanded,
                content: {
                    ForEach(Array(notifications.enumerated()), id: \.element.id) { index, notification in
                        NotificationRow(notification: notification)
                            .accessibilityIdentifier("notification_row_\(appName)_msg_\(index + 1)")
                    }
                },
                label: {
                    HStack {
                        Text("\(emoji) \(appName)")
                            .font(.headline)
                        Spacer()
                        Text("\(notifications.count)")
                            .font(.caption.bold())
                            .padding(.horizontal, 8)
                            .padding(.vertical, 2)
                            .background(Color.accentColor.opacity(0.15))
                            .foregroundStyle(Color.accentColor)
                            .clipShape(Capsule())
                            .accessibilityIdentifier("notification_count_\(appName)")
                    }
                }
            )
            .accessibilityIdentifier("notification_toggle_\(appName)")
        }
        .accessibilityIdentifier("notification_section_\(appName)")
    }
}

// MARK: - NotificationRow

private struct NotificationRow: View {

    let notification: AppNotification

    var body: some View {
        Button {
            if let scheme = NotificationRepository.urlScheme(for: notification.appName),
               let url = URL(string: scheme) {
                UIApplication.shared.open(url)
            }
        } label: {
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text(notification.sender)
                        .font(.subheadline.bold())
                        .foregroundStyle(.primary)
                    Spacer()
                    Text(notification.date.relativeFormatted)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                if !notification.body.isEmpty {
                    Text(notification.body)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Date Extension

private extension Date {

    var relativeFormatted: String {
        let formatter = RelativeDateTimeFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.unitsStyle = .short
        return formatter.localizedString(for: self, relativeTo: Date())
    }
}

// MARK: - NotificationsViewModel

final class NotificationsViewModel: ObservableObject {

    @Published var groupedNotifications: [String: [AppNotification]] = [:]
    @Published private var expandedApps: Set<String> = []

    private let repository: NotificationRepository

    init(repository: NotificationRepository = NotificationRepository()) {
        self.repository = repository
    }

    var sortedAppNames: [String] {
        groupedNotifications.keys.sorted()
    }

    func load() {
        let grouped = repository.fetchGroupedByApp()
        groupedNotifications = grouped
        // 새로 로드 시 아직 상태가 없는 앱은 펼쳐진 상태로 초기화
        for appName in grouped.keys where !expandedApps.contains(appName) {
            expandedApps.insert(appName)
        }
    }

    func binding(for appName: String) -> Binding<Bool> {
        Binding(
            get: { self.expandedApps.contains(appName) },
            set: { newValue in
                if newValue {
                    self.expandedApps.insert(appName)
                } else {
                    self.expandedApps.remove(appName)
                }
            }
        )
    }
}
