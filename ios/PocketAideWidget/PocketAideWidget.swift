import DesignSystem
import SwiftUI
import WidgetKit

struct PocketAideWidgetEntryView: View {
    let entry: PocketAideWidgetEntry

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                PlaceholderSection(area: .personal, label: "날씨")
                ruleDivider(vertical: true)
                PlaceholderSection(area: .work, label: "다음 일정")
            }
            .padding(.bottom, DesignTokens.Spacing.md)

            ruleDivider(vertical: false)

            AffirmationSection(state: entry.state)
                .padding(.vertical, DesignTokens.Spacing.md)

            ruleDivider(vertical: false)

            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                PlaceholderSection(area: .aiChat, label: "메일")
                ruleDivider(vertical: true)
                PlaceholderSection(area: .scratchpad, label: "알림")
            }
            .padding(.top, DesignTokens.Spacing.md)
        }
        .padding(DesignTokens.Spacing.lg)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        // Outer widgetURL is the fallback tap target — used for placeholder
        // sections that don't yet have their own destinations. The
        // affirmation section overrides this region with its own Link.
        .widgetURL(URL(string: "pocketaide://root"))
        .containerBackground(for: .widget) {
            DesignTokens.Color.widgetSurface()
        }
    }

    @ViewBuilder
    private func ruleDivider(vertical: Bool) -> some View {
        if vertical {
            Rectangle()
                .fill(DesignTokens.Color.widgetRule())
                .frame(width: 1)
        } else {
            Rectangle()
                .fill(DesignTokens.Color.widgetRule())
                .frame(height: 1)
        }
    }
}

@main
struct PocketAideWidget: Widget {
    let kind = "PocketAideWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: AffirmationProvider()) { entry in
            PocketAideWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("PocketAide")
        .description("하루를 한눈에 — 다짐과 일상 정보를 모아 봅니다.")
        .supportedFamilies([.systemLarge])
    }
}
