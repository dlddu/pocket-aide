import DesignSystem
import PocketAideAPI
import SwiftUI
import UIKit

struct PRMonitorNotificationSettingsSheet: View {
    @ObservedObject var viewModel: PRMonitorViewModel
    let pushAuthorizationDenied: Bool
    @Binding var isPresented: Bool
    let onSignOut: () -> Void

    @State private var confirmingSignOut = false

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                if pushAuthorizationDenied {
                    permissionNotice
                }
                enabledRow
                outcomesRow
                if let message = viewModel.settingsError {
                    Text(message)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.StatusColor.failure)
                        .accessibilityIdentifier("prmonitor.settings.error")
                }
                Text("설정과 무관하게 모든 CI 완료 이벤트는 이력에 남습니다.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.55))
                Spacer()
                signOutRow
            }
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.top, DesignTokens.Spacing.lg)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(DesignTokens.Color.surface(.prMonitor))
            .navigationTitle("알림 설정")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("완료") { isPresented = false }
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                }
            }
        }
    }

    private var signOutRow: some View {
        Button("로그아웃", role: .destructive) {
            confirmingSignOut = true
        }
        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
        .frame(maxWidth: .infinity)
        .padding(.bottom, DesignTokens.Spacing.lg)
        .accessibilityIdentifier("prmonitor.settings.signOut")
        .confirmationDialog(
            "로그아웃할까요?",
            isPresented: $confirmingSignOut,
            titleVisibility: .visible
        ) {
            Button("로그아웃", role: .destructive) {
                isPresented = false
                onSignOut()
            }
            .accessibilityIdentifier("prmonitor.settings.signOut.confirm")
            Button("취소", role: .cancel) {}
        } message: {
            Text("이 기기에 저장된 GitHub 토큰도 함께 지워집니다. 다른 계정으로 다시 로그인할 수 있습니다.")
        }
    }

    private var permissionNotice: some View {
        Button {
            if let url = URL(string: UIApplication.openSettingsURLString) {
                UIApplication.shared.open(url)
            }
        } label: {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Image(systemName: "bell.slash.fill")
                Text("시스템 알림 권한이 꺼져 있어 아래 설정과 관계없이 푸시가 도착하지 않습니다. 설정에서 켜기")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .multilineTextAlignment(.leading)
                Spacer(minLength: 0)
            }
            .padding(DesignTokens.Spacing.md)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(DesignTokens.Color.accent(.prMonitor).opacity(0.12))
            .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
            .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("prmonitor.settings.permissionNotice")
    }

    private var enabledRow: some View {
        Toggle(isOn: Binding(
            get: { viewModel.notificationSettings.enabled },
            set: { newValue in Task { await viewModel.setNotificationsEnabled(newValue) } }
        )) {
            Text("CI 완료 푸시 받기")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
        }
        .tint(DesignTokens.Color.accent(.prMonitor))
        .disabled(viewModel.isLoadingSettings)
        .accessibilityIdentifier("prmonitor.settings.enabled")
    }

    private var outcomesRow: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            Text("받을 결과")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.7))
            Picker("받을 결과", selection: Binding(
                get: { viewModel.notificationSettings.outcomes },
                set: { newValue in Task { await viewModel.setNotificationOutcomes(newValue) } }
            )) {
                ForEach(NotificationOutcomes.allCases) { outcome in
                    Text(outcome.label).tag(outcome)
                }
            }
            .pickerStyle(.segmented)
            .accessibilityIdentifier("prmonitor.settings.outcomes")
        }
        .disabled(viewModel.isLoadingSettings || !viewModel.notificationSettings.enabled)
        .opacity(viewModel.notificationSettings.enabled ? 1 : 0.45)
    }
}
