// SettingsView.swift
// PocketAide

import SwiftUI

// MARK: - SyncStatus

/// 동기화 상태를 나타내는 열거형.
enum SyncStatus {
    case synced
    case syncing
    case error

    var displayText: String {
        switch self {
        case .synced:  return "Synced"
        case .syncing: return "Syncing..."
        case .error:   return "Sync Error"
        }
    }

    var accessibilityIdentifier: String {
        switch self {
        case .synced:  return "sync_status_synced"
        case .syncing: return "sync_status_syncing"
        case .error:   return "sync_status_error"
        }
    }

    var iconName: String {
        switch self {
        case .synced:  return "checkmark.circle"
        case .syncing: return "arrow.triangle.2.circlepath"
        case .error:   return "exclamationmark.circle"
        }
    }

    var color: Color {
        switch self {
        case .synced:  return .green
        case .syncing: return .blue
        case .error:   return .red
        }
    }
}

// MARK: - SettingsView

struct SettingsView: View {

    // MARK: - Accessibility Identifier Constants

    /// Siri Shortcut 설정 안내 섹션의 accessibilityIdentifier.
    static let shortcutSetupSectionIdentifier = "shortcut_setup_section"

    /// Siri Shortcut 추가 버튼의 accessibilityIdentifier.
    static let shortcutAddToSiriButtonIdentifier = "shortcut_add_to_siri_button"

    // MARK: - Properties

    @EnvironmentObject var authViewModel: AuthViewModel

    @AppStorage("selectedSpeechEngine") private var selectedEngine: String = SpeechEngine.whisperLocal.rawValue
    @State private var showEnginePicker: Bool = false

    /// 현재 동기화 상태. 외부에서 주입 가능 (UI 테스트 지원).
    @State var syncStatus: SyncStatus = .synced

    /// 마지막 동기화 시각.
    @State var lastSyncDate: Date? = nil

    /// 네트워크 연결 상태.
    @State var isNetworkConnected: Bool = true

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

                // Speech Engine Section
                VStack(alignment: .leading, spacing: 8) {
                    Text("Speech Engine")
                        .font(.headline)
                    Button(action: {
                        showEnginePicker = true
                    }) {
                        Text(selectedEngine)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 8)
                            .background(Color(.systemGray6))
                            .cornerRadius(8)
                            .foregroundColor(.primary)
                    }
                    .accessibilityIdentifier("speech_engine_selector")
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal)

                // Siri Shortcut Setup Section
                VStack(alignment: .leading, spacing: 8) {
                    Text("Siri Shortcut")
                        .font(.headline)
                    Text("음성 메모를 Siri Shortcut으로 빠르게 저장하세요.")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                    Button(action: {
                        // Siri Shortcut 추가 액션
                    }) {
                        HStack {
                            Image(systemName: "mic.fill")
                            Text("Add to Siri")
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 8)
                        .background(Color(.systemGray6))
                        .cornerRadius(8)
                        .foregroundColor(.primary)
                    }
                    .accessibilityIdentifier(SettingsView.shortcutAddToSiriButtonIdentifier)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal)
                .accessibilityIdentifier(SettingsView.shortcutSetupSectionIdentifier)

                // Sync Status Section
                VStack(alignment: .leading, spacing: 8) {
                    Text("Sync")
                        .font(.headline)

                    // 동기화 상태 표시
                    HStack(spacing: 8) {
                        Image(systemName: syncStatus.iconName)
                            .foregroundColor(syncStatus.color)
                        Text(syncStatus.displayText)
                            .font(.body)
                            .foregroundColor(syncStatus.color)
                    }
                    .accessibilityIdentifier(syncStatus.accessibilityIdentifier)

                    // 마지막 동기화 시각 표시
                    if let lastSync = lastSyncDate {
                        Text("Last synced: \(lastSync.formatted(date: .omitted, time: .shortened))")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }

                    // 연결 상태 표시
                    HStack(spacing: 4) {
                        Circle()
                            .fill(isNetworkConnected ? Color.green : Color.gray)
                            .frame(width: 8, height: 8)
                        Text(isNetworkConnected ? "Connected" : "Offline")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
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
        .sheet(isPresented: $showEnginePicker) {
            SpeechEnginePickerSheet(selectedEngine: $selectedEngine, isPresented: $showEnginePicker)
        }
    }
}

// MARK: - SpeechEnginePickerSheet

private struct SpeechEnginePickerSheet: View {

    @Binding var selectedEngine: String
    @Binding var isPresented: Bool

    var body: some View {
        VStack(spacing: 0) {
            ForEach(SpeechEngine.allCases, id: \.self) { engine in
                Button(action: {
                    selectedEngine = engine.rawValue
                    isPresented = false
                }) {
                    HStack {
                        Text(engine.rawValue)
                            .foregroundColor(.primary)
                        Spacer()
                        if selectedEngine == engine.rawValue {
                            Image(systemName: "checkmark")
                                .foregroundColor(.accentColor)
                        }
                    }
                    .padding()
                }
                .accessibilityLabel(engine.rawValue)
                Divider()
            }
            Spacer()
        }
        .accessibilityIdentifier("settings_speech_engine_picker")
        .presentationDetents([.medium])
    }
}
