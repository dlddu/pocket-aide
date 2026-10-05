import DesignSystem
import PocketAideAPI
import SwiftUI

struct WeatherSection: View {
    let state: WidgetWeatherState

    var body: some View {
        Link(destination: URL(string: "pocketaide://weather")!) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                AreaLabel(area: .personal, text: label, showsDot: false)
                content
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(.plain)
    }

    private var label: String {
        if case .loaded(_, let place?) = state, !place.isEmpty {
            return place
        }
        return "날씨"
    }

    @ViewBuilder
    private var content: some View {
        switch state {
        case .loaded(let summary, _):
            VStack(alignment: .leading, spacing: 2) {
                HStack(alignment: .firstTextBaseline, spacing: DesignTokens.Spacing.xs) {
                    Text(Weather.temperatureLabel(summary))
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.h1, weight: .bold))
                        .monospacedDigit()
                        .foregroundStyle(DesignTokens.Color.ink(.personal))
                    Text(summary.condition)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.Color.ink(.personal).opacity(0.75))
                        .lineLimit(1)
                }
                Text(Weather.detailLabel(summary))
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(DesignTokens.Color.ink(.personal).opacity(0.75))
                    .lineLimit(1)
            }
        case .needsLocation:
            notice("앱에서 위치 접근을 허용해 주세요.")
        case .error:
            notice("잠시 후 다시 시도할게요.")
        }
    }

    private func notice(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(.personal).opacity(0.6))
            .lineLimit(2)
    }
}
