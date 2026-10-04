import DesignSystem
import PocketAideAPI
import SwiftUI

struct CalendarSection: View {
    let state: WidgetCalendarState
    let now: Date

    private let calendar = Calendar.current

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
            AreaLabel(area: .work, text: "다음 일정", showsDot: false)
            content
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private var content: some View {
        switch state {
        case .needsPermission:
            notice("앱에서 캘린더 접근을 허용해 주세요.")
        case .loaded(let summary):
            if let first = summary.events.first {
                VStack(alignment: .leading, spacing: 2) {
                    Text(first.title)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.work))
                        .lineLimit(1)
                    Text(UpcomingEvents.rangeLabel(first, now: now, calendar: calendar))
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.Color.ink(.work).opacity(0.75))
                    ForEach(Array(summary.events.dropFirst().enumerated()), id: \.offset) { _, event in
                        Text("\(UpcomingEvents.startLabel(event, now: now, calendar: calendar)) \(event.title)")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                            .foregroundStyle(DesignTokens.Color.ink(.work).opacity(0.75))
                            .lineLimit(1)
                    }
                    if summary.moreCount > 0 {
                        Text("+ \(summary.moreCount) 더")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                            .foregroundStyle(DesignTokens.Color.ink(.work).opacity(0.55))
                    }
                }
            } else {
                notice("다가오는 일정이 없어요.")
            }
        }
    }

    private func notice(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
            .foregroundStyle(DesignTokens.Color.ink(.work).opacity(0.6))
            .lineLimit(2)
    }
}
