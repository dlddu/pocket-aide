// HistoryView.swift
// PocketAide

import SwiftUI

struct HistoryView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("History")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("History")
        }
        .accessibilityIdentifier("tab_history_view")
    }
}
