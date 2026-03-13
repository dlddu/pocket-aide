// RecordView.swift
// PocketAide

import SwiftUI

struct RecordView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Record")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Record")
        }
        .accessibilityIdentifier("tab_record_view")
    }
}
