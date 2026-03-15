// PocketAideLargeTimelineProvider.swift
// PocketAideWidget

import WidgetKit
import Foundation

struct PocketAideLargeTimelineProvider: TimelineProvider {

    // Read a flag from App Group to determine if we should use test data
    private var isUITesting: Bool {
        let defaults = UserDefaults(suiteName: "group.com.dlddu.PocketAide")
        return defaults?.bool(forKey: "uitesting_widget") ?? false
    }

    func placeholder(in context: Context) -> PocketAideLargeEntry {
        Self.sampleEntry()
    }

    func getSnapshot(in context: Context, completion: @escaping (PocketAideLargeEntry) -> Void) {
        completion(Self.sampleEntry())
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<PocketAideLargeEntry>) -> Void) {
        let entry: PocketAideLargeEntry
        if isUITesting {
            entry = Self.sampleEntry()
        } else {
            entry = loadRealData()
        }
        let nextUpdate = Calendar.current.date(byAdding: .minute, value: 15, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
        completion(timeline)
    }

    private func loadRealData() -> PocketAideLargeEntry {
        let defaults = UserDefaults(suiteName: "group.com.dlddu.PocketAide")

        // Read notifications from App Group
        var notifications: [WidgetNotification] = []
        if let data = defaults?.data(forKey: "notifications") {
            let decoder = JSONDecoder()
            decoder.dateDecodingStrategy = .iso8601
            if let appNotifications = try? decoder.decode([AppNotificationDTO].self, from: data) {
                notifications = appNotifications.prefix(3).map {
                    WidgetNotification(id: $0.id, appName: $0.appName, body: $0.body)
                }
            }
        }

        // Read sentence from App Group
        let sentence = defaults?.string(forKey: "widget_sentence") ?? "오늘의 문장이 없습니다"

        // Read mail summary from App Group
        let mail = defaults?.string(forKey: "widget_mail") ?? "새 메일이 없습니다"

        return PocketAideLargeEntry(
            date: Date(),
            calendarEvents: [],
            weather: WeatherInfo(condition: "맑음", currentTemp: 20, highTemp: 25, lowTemp: 15),
            mailSummary: mail,
            sentence: sentence,
            notifications: notifications
        )
    }

    static func sampleEntry() -> PocketAideLargeEntry {
        PocketAideLargeEntry(
            date: Date(),
            calendarEvents: [
                CalendarEvent(id: UUID(), title: "팀 미팅", startTime: Date()),
                CalendarEvent(id: UUID(), title: "점심 약속", startTime: Date().addingTimeInterval(3600))
            ],
            weather: WeatherInfo(condition: "맑음", currentTemp: 22, highTemp: 26, lowTemp: 14),
            mailSummary: "주간 보고서가 도착했습니다",
            sentence: "오늘 하루도 최선을 다하자",
            notifications: [
                WidgetNotification(id: UUID(), appName: "카카오톡", body: "안녕하세요!"),
                WidgetNotification(id: UUID(), appName: "Slack", body: "배포 완료")
            ]
        )
    }
}

// DTO for decoding AppNotification from App Group UserDefaults
private struct AppNotificationDTO: Codable {
    let id: UUID
    let appName: String
    let sender: String
    let body: String
    let date: Date

    enum CodingKeys: String, CodingKey {
        case id
        case appName = "app_name"
        case sender
        case body
        case date
    }
}
