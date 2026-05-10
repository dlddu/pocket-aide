import DesignSystem
import SwiftUI
import WidgetKit

struct PocketAideEntry: TimelineEntry {
    let date: Date
    let greeting: String
}

struct PocketAideProvider: TimelineProvider {
    func placeholder(in _: Context) -> PocketAideEntry {
        PocketAideEntry(date: Date(), greeting: "Hello, pocket-aide")
    }

    func getSnapshot(in _: Context, completion: @escaping (PocketAideEntry) -> Void) {
        completion(PocketAideEntry(date: Date(), greeting: "Hello, pocket-aide"))
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<PocketAideEntry>) -> Void) {
        let entry = PocketAideEntry(date: Date(), greeting: "Hello, pocket-aide")
        completion(Timeline(entries: [entry], policy: .after(Date().addingTimeInterval(60 * 60))))
    }
}

struct PocketAideWidgetEntryView: View {
    let entry: PocketAideProvider.Entry

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            Text("POCKET AIDE")
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold))
                .tracking(2)
                .foregroundStyle(DesignTokens.Color.accent(.aiChat))
            Text(entry.greeting)
                .font(.system(size: DesignTokens.Typography.titleMd, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(.aiChat))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
        .padding(DesignTokens.Spacing.lg)
        .containerBackground(for: .widget) {
            DesignTokens.Color.surface(.aiChat)
        }
    }
}

@main
struct PocketAideWidget: Widget {
    let kind = "PocketAideWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: PocketAideProvider()) { entry in
            PocketAideWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("PocketAide")
        .description("Hello from pocket-aide.")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}
