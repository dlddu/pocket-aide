// MailView.swift
// PocketAide

import SwiftUI

struct MailView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Mail")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Mail")
        }
        .accessibilityIdentifier("tab_mail_view")
    }
}
