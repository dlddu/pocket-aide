// LargeWidgetView.swift
// PocketAideWidget

import SwiftUI
import WidgetKit

struct LargeWidgetView: View {
    let entry: PocketAideLargeEntry

    var body: some View {
        VStack(spacing: 8) {
            // Top: Calendar + Weather (horizontal split)
            HStack(spacing: 8) {
                calendarSection
                weatherSection
            }

            // Middle: Mail
            mailSection

            // Bottom: Sentence + Notification
            HStack(spacing: 8) {
                sentenceSection
                notificationSection
            }
        }
        .padding(12)
    }

    // MARK: - Calendar Section

    private var calendarSection: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("📅 캘린더")
                .font(.caption)
                .fontWeight(.semibold)
            ForEach(entry.calendarEvents.prefix(2)) { event in
                Text(event.title)
                    .font(.caption2)
                    .lineLimit(1)
            }
            if entry.calendarEvents.isEmpty {
                Text("일정 없음")
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(8)
        .background(Color(.systemBackground).opacity(0.8))
        .cornerRadius(10)
        .accessibilityIdentifier("widget_section_calendar")
    }

    // MARK: - Weather Section

    private var weatherSection: some View {
        VStack(spacing: 4) {
            Text(entry.weather.condition)
                .font(.caption)
                .fontWeight(.semibold)
            Text("\(entry.weather.currentTemp)°")
                .font(.title2)
                .fontWeight(.bold)
            Text("↑\(entry.weather.highTemp)° ↓\(entry.weather.lowTemp)°")
                .font(.caption2)
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity)
        .padding(8)
        .background(
            LinearGradient(
                colors: [Color.blue.opacity(0.6), Color.blue.opacity(0.3)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .cornerRadius(10)
        .accessibilityIdentifier("widget_section_weather")
    }

    // MARK: - Mail Section

    private var mailSection: some View {
        HStack {
            Text("📧")
            Text(entry.mailSummary)
                .font(.caption)
                .lineLimit(2)
            Spacer()
        }
        .padding(8)
        .background(Color(.systemBackground).opacity(0.8))
        .cornerRadius(10)
        .accessibilityIdentifier("widget_section_mail")
    }

    // MARK: - Sentence Section

    private var sentenceSection: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("💬 문장")
                .font(.caption)
                .fontWeight(.semibold)
            Text(entry.sentence)
                .font(.caption2)
                .lineLimit(3)
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(8)
        .background(Color.yellow.opacity(0.2))
        .overlay(
            RoundedRectangle(cornerRadius: 10)
                .stroke(Color.orange, lineWidth: 1)
        )
        .cornerRadius(10)
        .accessibilityIdentifier("widget_section_sentence")
    }

    // MARK: - Notification Section

    private var notificationSection: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("🔔 알림")
                .font(.caption)
                .fontWeight(.semibold)
            ForEach(entry.notifications.prefix(2)) { notif in
                Text("\(notif.appName): \(notif.body)")
                    .font(.caption2)
                    .lineLimit(1)
            }
            if entry.notifications.isEmpty {
                Text("알림 없음")
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(8)
        .background(Color(.systemBackground).opacity(0.8))
        .cornerRadius(10)
        .accessibilityIdentifier("widget_section_notification")
    }
}
