// WidgetSettingsView.swift
// PocketAide

import SwiftUI

struct WidgetSettingsView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Widget")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Widget")
        }
        .accessibilityIdentifier("tab_widget_view")
    }
}
