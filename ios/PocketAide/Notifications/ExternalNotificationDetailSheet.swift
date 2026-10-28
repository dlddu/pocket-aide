import DesignSystem
import PocketAideAPI
import SwiftUI

struct ExternalNotificationDetailSheet: View {
    let item: ExternalNotification
    let onClose: () -> Void

    private static let area: DesignTokens.Area = .prMonitor

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                header
                Text(item.title)
                    .font(DesignTokens.Typography.font(size: 17, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .fixedSize(horizontal: false, vertical: true)
                    .padding(.top, DesignTokens.Spacing.md)
                    .accessibilityIdentifier("notifications.detail.title")
                if !item.body.isEmpty {
                    Text(item.body)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area))
                        .lineSpacing(3)
                        .fixedSize(horizontal: false, vertical: true)
                        .padding(.top, DesignTokens.Spacing.sm)
                        .accessibilityIdentifier("notifications.detail.body")
                }
                VStack(spacing: 0) {
                    metaRow("보낸 호출자", ExternalNotificationFormatting.callerLabel(item), id: "caller")
                    metaRow("요청 주체", ExternalNotificationFormatting.requesterValue(item), id: "requester")
                    metaRow("받은 시각", ExternalNotificationFormatting.fullTime(item), id: "time")
                    metaRow("도달", ExternalNotificationFormatting.deliveryLabel(item), id: "delivery", last: true)
                }
                .overlay(alignment: .top) {
                    Rectangle()
                        .fill(DesignTokens.Color.rule(Self.area))
                        .frame(height: 1)
                }
                .padding(.top, DesignTokens.Spacing.xl)
            }
            .textSelection(.enabled)
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.top, DesignTokens.Spacing.xl)
            .padding(.bottom, DesignTokens.Spacing.xl)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .background(DesignTokens.Color.card(Self.area))
        .accessibilityIdentifier("notifications.detail")
    }

    private var header: some View {
        HStack(spacing: 6) {
            Text(ExternalNotificationFormatting.callerLabel(item))
                .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .padding(.vertical, 2)
                .background(DesignTokens.Color.soft(Self.area))
                .foregroundStyle(DesignTokens.StatusColor.arrivalGlow)
                .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
            Spacer(minLength: DesignTokens.Spacing.sm)
            Button("닫기", action: onClose)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
                .accessibilityIdentifier("notifications.detail.close")
        }
    }

    private func metaRow(_ label: String, _ value: String, id: String, last: Bool = false) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            Text(value)
                .font(.system(size: 12.5, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.ink(Self.area))
                .fixedSize(horizontal: false, vertical: true)
                .accessibilityIdentifier("notifications.detail.\(id)")
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.vertical, 10)
        .overlay(alignment: .bottom) {
            if !last {
                Rectangle()
                    .fill(DesignTokens.Color.rule(Self.area))
                    .frame(height: 1)
            }
        }
    }
}
