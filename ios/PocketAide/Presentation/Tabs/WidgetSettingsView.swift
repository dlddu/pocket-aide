// WidgetSettingsView.swift
// PocketAide

import SwiftUI

struct WidgetSettingsView: View {

    private let isUITestingWidget: Bool = CommandLine.arguments.contains("--uitesting-widget")

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 12) {
                    Text("Widget Preview")
                        .font(.headline)

                    if isUITestingWidget {
                        widgetPreview
                    } else {
                        Text("홈 화면에 위젯을 추가하세요")
                            .foregroundColor(.secondary)
                    }
                }
                .padding()
            }
            .navigationTitle("Widget")
        }
        .accessibilityIdentifier("widget_settings_view")
    }

    private var widgetPreview: some View {
        VStack(spacing: 8) {
            HStack(spacing: 8) {
                // Calendar section
                VStack(alignment: .leading, spacing: 4) {
                    Text("📅 캘린더")
                        .font(.caption)
                        .fontWeight(.semibold)
                    Text("팀 미팅")
                        .font(.caption2)
                    Text("점심 약속")
                        .font(.caption2)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(8)
                .background(Color(.systemGray6))
                .cornerRadius(10)
                .accessibilityIdentifier("widget_section_calendar")

                // Weather section
                VStack(spacing: 4) {
                    Text("맑음")
                        .font(.caption)
                        .fontWeight(.semibold)
                    Text("22°")
                        .font(.title2)
                        .fontWeight(.bold)
                    Text("↑26° ↓14°")
                        .font(.caption2)
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

            // Mail section
            HStack {
                Text("📧")
                Text("주간 보고서가 도착했습니다")
                    .font(.caption)
                Spacer()
            }
            .padding(8)
            .background(Color(.systemGray6))
            .cornerRadius(10)
            .accessibilityIdentifier("widget_section_mail")

            HStack(spacing: 8) {
                // Sentence section
                VStack(alignment: .leading, spacing: 4) {
                    Text("💬 문장")
                        .font(.caption)
                        .fontWeight(.semibold)
                    Text("오늘 하루도 최선을 다하자")
                        .font(.caption2)
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

                // Notification section
                VStack(alignment: .leading, spacing: 4) {
                    Text("🔔 알림")
                        .font(.caption)
                        .fontWeight(.semibold)
                    Text("카카오톡: 안녕하세요!")
                        .font(.caption2)
                    Text("Slack: 배포 완료")
                        .font(.caption2)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(8)
                .background(Color(.systemGray6))
                .cornerRadius(10)
                .accessibilityIdentifier("widget_section_notification")
            }
        }
        .padding()
        .background(Color(.systemGray5))
        .cornerRadius(16)
    }
}
