// MainTabView.swift
// PocketAide

import SwiftUI

/// 앱의 탭을 관리하는 메인 TabView.
///
/// 각 탭 버튼에는 XCUITest가 식별할 수 있도록
/// `accessibilityIdentifier`가 설정되어 있습니다.
struct MainTabView: View {

    @State private var selectedTab: Tab = .home

    enum Tab: String, CaseIterable {
        case home            = "tab_home"
        case record          = "tab_record"
        case history         = "tab_history"
        case widgetSettings  = "tab_widget_settings"
        case assistant       = "tab_assistant"
        case todo            = "tab_todo"
        case scratch         = "tab_scratch"
        case routine         = "tab_routine"
        case calendar        = "tab_calendar"
        case weather         = "tab_weather"
        case mail            = "tab_mail"
        case sentence        = "tab_sentence"
        case notification    = "tab_notification"
        case settings        = "tab_settings"
        case profile         = "tab_profile"

        var title: String {
            switch self {
            case .home:            return "Home"
            case .record:          return "Record"
            case .history:         return "History"
            case .widgetSettings:  return "Widget"
            case .assistant:       return "Assistant"
            case .todo:            return "Todo"
            case .scratch:         return "Scratch"
            case .routine:         return "Routine"
            case .calendar:        return "Calendar"
            case .weather:         return "Weather"
            case .mail:            return "Mail"
            case .sentence:        return "Sentence"
            case .notification:    return "알림"
            case .settings:        return "Settings"
            case .profile:         return "Profile"
            }
        }

        var systemImage: String {
            switch self {
            case .home:            return "house"
            case .record:          return "mic"
            case .history:         return "clock"
            case .widgetSettings:  return "square.grid.2x2"
            case .assistant:       return "sparkles"
            case .todo:            return "checkmark.square"
            case .scratch:         return "note.text"
            case .routine:         return "arrow.clockwise"
            case .calendar:        return "calendar"
            case .weather:         return "cloud.sun"
            case .mail:            return "envelope"
            case .sentence:        return "text.quote"
            case .notification:    return "bell"
            case .settings:        return "gearshape"
            case .profile:         return "person.circle"
            }
        }
    }

    var body: some View {
        TabView(selection: $selectedTab) {
            HomeView()
                .tabItem {
                    Label(Tab.home.title, systemImage: Tab.home.systemImage)
                }
                .tag(Tab.home)
                .accessibilityIdentifier(Tab.home.rawValue)

            RecordView()
                .tabItem {
                    Label(Tab.record.title, systemImage: Tab.record.systemImage)
                }
                .tag(Tab.record)
                .accessibilityIdentifier(Tab.record.rawValue)

            HistoryView()
                .tabItem {
                    Label(Tab.history.title, systemImage: Tab.history.systemImage)
                }
                .tag(Tab.history)
                .accessibilityIdentifier(Tab.history.rawValue)

            WidgetSettingsView()
                .tabItem {
                    Label(Tab.widgetSettings.title, systemImage: Tab.widgetSettings.systemImage)
                }
                .tag(Tab.widgetSettings)
                .accessibilityIdentifier(Tab.widgetSettings.rawValue)

            AssistantView()
                .tabItem {
                    Label(Tab.assistant.title, systemImage: Tab.assistant.systemImage)
                }
                .tag(Tab.assistant)
                .accessibilityIdentifier(Tab.assistant.rawValue)

            TodoScreen()
                .tabItem {
                    Label(Tab.todo.title, systemImage: Tab.todo.systemImage)
                }
                .tag(Tab.todo)
                .accessibilityIdentifier(Tab.todo.rawValue)

            ScratchScreen()
                .tabItem {
                    Label(Tab.scratch.title, systemImage: Tab.scratch.systemImage)
                }
                .tag(Tab.scratch)
                .accessibilityIdentifier(Tab.scratch.rawValue)

            RoutineScreen()
                .tabItem {
                    Label(Tab.routine.title, systemImage: Tab.routine.systemImage)
                }
                .tag(Tab.routine)
                .accessibilityIdentifier(Tab.routine.rawValue)

            CalendarView()
                .tabItem {
                    Label(Tab.calendar.title, systemImage: Tab.calendar.systemImage)
                }
                .tag(Tab.calendar)
                .accessibilityIdentifier(Tab.calendar.rawValue)

            WeatherView()
                .tabItem {
                    Label(Tab.weather.title, systemImage: Tab.weather.systemImage)
                }
                .tag(Tab.weather)
                .accessibilityIdentifier(Tab.weather.rawValue)

            MailView()
                .tabItem {
                    Label(Tab.mail.title, systemImage: Tab.mail.systemImage)
                }
                .tag(Tab.mail)
                .accessibilityIdentifier(Tab.mail.rawValue)

            SentenceScreen()
                .tabItem {
                    Label(Tab.sentence.title, systemImage: Tab.sentence.systemImage)
                }
                .tag(Tab.sentence)
                .accessibilityIdentifier(Tab.sentence.rawValue)

            NotificationsScreen()
                .tabItem {
                    Label(Tab.notification.title, systemImage: Tab.notification.systemImage)
                }
                .tag(Tab.notification)
                .accessibilityIdentifier(Tab.notification.rawValue)

            SettingsView()
                .tabItem {
                    Label(Tab.settings.title, systemImage: Tab.settings.systemImage)
                }
                .tag(Tab.settings)
                .accessibilityIdentifier(Tab.settings.rawValue)

            ProfileView()
                .tabItem {
                    Label(Tab.profile.title, systemImage: Tab.profile.systemImage)
                }
                .tag(Tab.profile)
                .accessibilityIdentifier(Tab.profile.rawValue)
        }
        .onOpenURL { url in
            handleDeeplink(url)
        }
    }

    private func handleDeeplink(_ url: URL) {
        guard url.scheme == "pocketaide",
              url.host == "tab",
              let section = url.pathComponents.dropFirst().first else { return }

        switch section {
        case "calendar":     selectedTab = .calendar
        case "weather":      selectedTab = .weather
        case "mail":         selectedTab = .mail
        case "sentence":     selectedTab = .sentence
        case "notification": selectedTab = .notification
        default: break
        }
    }
}
