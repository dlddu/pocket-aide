// CalendarView.swift
// PocketAide

import SwiftUI

struct CalendarView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Calendar")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Calendar")
        }
        .accessibilityIdentifier("tab_calendar_view")
    }
}
