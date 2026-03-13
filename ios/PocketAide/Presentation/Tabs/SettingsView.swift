// SettingsView.swift
// PocketAide

import SwiftUI

struct SettingsView: View {

    @EnvironmentObject var authViewModel: AuthViewModel

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Server")
                        .font(.headline)
                    Text(authViewModel.serverAddress.isEmpty ? "Not configured" : authViewModel.serverAddress)
                        .font(.body)
                        .foregroundColor(.secondary)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal)

                VStack(alignment: .leading, spacing: 8) {
                    Text("Status")
                        .font(.headline)
                    Text(authViewModel.isAuthenticated ? "Logged in" : "Not logged in")
                        .font(.body)
                        .foregroundColor(authViewModel.isAuthenticated ? .green : .secondary)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal)

                Button(action: {
                    authViewModel.logout()
                }) {
                    Text("Logout")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .tint(.red)
                .padding(.horizontal)
                .accessibilityIdentifier("logout_button")

                Spacer()
            }
            .padding(.top)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Settings")
        }
        .accessibilityIdentifier("tab_settings_view")
    }
}
