// PocketAideLargeEntry.swift
// PocketAideWidget

import WidgetKit
import Foundation

struct PocketAideLargeEntry: TimelineEntry {
    let date: Date
    let calendarEvents: [CalendarEvent]
    let weather: WeatherInfo
    let mailSummary: String
    let sentence: String
    let notifications: [WidgetNotification]
}

struct CalendarEvent: Identifiable {
    let id: UUID
    let title: String
    let startTime: Date
}

struct WeatherInfo {
    let condition: String
    let currentTemp: Int
    let highTemp: Int
    let lowTemp: Int
}

struct WidgetNotification: Identifiable {
    let id: UUID
    let appName: String
    let body: String
}
