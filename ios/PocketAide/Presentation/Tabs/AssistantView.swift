// AssistantView.swift
// PocketAide

import SwiftUI

struct AssistantView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Assistant")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Assistant")
        }
        .accessibilityIdentifier("tab_assistant_view")
    }
}
