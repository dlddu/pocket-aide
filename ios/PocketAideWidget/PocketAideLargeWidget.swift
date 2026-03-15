// PocketAideLargeWidget.swift
// PocketAideWidget

import WidgetKit
import SwiftUI

struct PocketAideLargeWidget: Widget {
    let kind: String = "PocketAideLargeWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: PocketAideLargeTimelineProvider()) { entry in
            LargeWidgetView(entry: entry)
        }
        .configurationDisplayName("PocketAide")
        .description("캘린더, 날씨, 메일, 문장, 알림을 한눈에")
        .supportedFamilies([.systemLarge])
    }
}
