import DesignSystem
import PocketAideAPI
import SwiftUI
import UIKit

struct WeatherView: View {
    @StateObject private var viewModel = WeatherViewModel()
    @Environment(\.dismiss) private var dismiss

    private let area = DesignTokens.Area.personal

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(area).ignoresSafeArea()
            VStack(spacing: 0) {
                header
                ScrollView {
                    content
                        .padding(.horizontal, DesignTokens.Spacing.xl)
                        .padding(.bottom, DesignTokens.Spacing.xxl)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
                .refreshable { await viewModel.load(useCache: false) }
            }
        }
        .task { await viewModel.load(useCache: true) }
        .onReceive(NotificationCenter.default.publisher(for: WeatherLocationAccess.didRefresh)) { _ in
            Task { await viewModel.locationDidRefresh() }
        }
        .accessibilityIdentifier("weather.screen")
    }

    private var header: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                AreaLabel(area: area, text: "날씨")
                Text(title)
                    .font(DesignTokens.Typography.font(size: 24, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                    .accessibilityIdentifier("weather.place")
            }
            Spacer()
            Button("닫기") { dismiss() }
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.accent(area))
                .accessibilityIdentifier("weather.close")
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.xl)
        .padding(.bottom, DesignTokens.Spacing.md)
    }

    private var title: String {
        if let place = viewModel.placeName, !place.isEmpty {
            return place
        }
        return "날씨"
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.state {
        case .loading:
            ProgressView()
                .frame(maxWidth: .infinity)
                .padding(.top, DesignTokens.Spacing.xxl)
        case .needsLocation:
            notice("앱에서 위치 접근을 허용해 주세요.", action: "설정 열기", identifier: "weather.settings") {
                if let url = URL(string: UIApplication.openSettingsURLString) {
                    UIApplication.shared.open(url)
                }
            }
        case .error:
            notice("잠시 후 다시 시도할게요.", action: "다시 시도", identifier: "weather.retry") {
                Task { await viewModel.load(useCache: false) }
            }
        case .loaded(let snapshot):
            forecast(snapshot)
        }
    }

    private func notice(
        _ text: String,
        action: String,
        identifier: String,
        perform: @escaping () -> Void
    ) -> some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            Text(text)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                .accessibilityIdentifier("weather.notice")
            Button(action, action: perform)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.accent(area))
                .accessibilityIdentifier(identifier)
        }
        .padding(.top, DesignTokens.Spacing.lg)
    }

    private func forecast(_ snapshot: WeatherForecastSnapshot) -> some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
            current(snapshot.forecast)
            sectionTitle("시간별 예보")
            hourly(snapshot.forecast.hours)
            sectionTitle("주간 예보")
            weekly(snapshot.forecast.days)
            Text(Weather.updatedLabel(snapshot.fetchedAt))
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .accessibilityIdentifier("weather.updated")
        }
    }

    private func current(_ forecast: WeatherForecast) -> some View {
        Card(area: area) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                HStack(alignment: .firstTextBaseline, spacing: DesignTokens.Spacing.sm) {
                    Text(Weather.temperatureLabel(forecast.summary))
                        .font(DesignTokens.Typography.font(size: 48, weight: .bold))
                        .monospacedDigit()
                        .foregroundStyle(DesignTokens.Color.ink(area))
                        .accessibilityIdentifier("weather.current.temperature")
                    Text(forecast.summary.condition)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.titleMd, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                        .accessibilityIdentifier("weather.current.condition")
                }
                Text(Weather.apparentLabel(forecast))
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                    .accessibilityIdentifier("weather.current.apparent")
                Text(Weather.detailLabel(forecast.summary))
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                    .accessibilityIdentifier("weather.current.detail")
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private func sectionTitle(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.7))
    }

    private func hourly(_ hours: [WeatherHour]) -> some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                ForEach(Array(hours.enumerated()), id: \.offset) { index, hour in
                    VStack(spacing: DesignTokens.Spacing.xs) {
                        Text(hour.time)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.6))
                        Text("\(hour.temperature)°")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                            .monospacedDigit()
                            .foregroundStyle(DesignTokens.Color.ink(area))
                        Text(hour.condition)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                        Text(Weather.chanceLabel(hour.precipitationChance))
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                            .monospacedDigit()
                            .foregroundStyle(DesignTokens.Color.accent(area))
                    }
                    .frame(minWidth: 52)
                    .accessibilityElement(children: .combine)
                    .accessibilityIdentifier("weather.hour.\(index)")
                }
            }
            .padding(DesignTokens.Spacing.md)
        }
        .background(DesignTokens.Color.card(area))
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
    }

    private func weekly(_ days: [WeatherDay]) -> some View {
        Card(area: area) {
            VStack(spacing: DesignTokens.Spacing.md) {
                ForEach(Array(days.enumerated()), id: \.offset) { index, day in
                    HStack(spacing: DesignTokens.Spacing.sm) {
                        Text(day.label)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm, weight: .semibold))
                            .foregroundStyle(DesignTokens.Color.ink(area))
                            .frame(width: 76, alignment: .leading)
                        Text(day.condition)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.75))
                            .lineLimit(1)
                        Spacer(minLength: 0)
                        Text(Weather.chanceLabel(day.precipitationChance))
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                            .monospacedDigit()
                            .foregroundStyle(DesignTokens.Color.accent(area))
                        Text("\(day.high)° / \(day.low)°")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm, weight: .semibold))
                            .monospacedDigit()
                            .foregroundStyle(DesignTokens.Color.ink(area))
                            .frame(width: 84, alignment: .trailing)
                    }
                    .accessibilityElement(children: .combine)
                    .accessibilityIdentifier("weather.day.\(index)")
                }
            }
        }
    }
}
