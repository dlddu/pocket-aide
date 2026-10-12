import DesignSystem
import PocketAideAPI
import SwiftUI

struct ApprovalAutoResponseHeader: View {
    let mode: ApprovalAutoResponse
    let area: DesignTokens.Area

    var body: some View {
        HStack(spacing: 4) {
            Text("자동 응답")
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
            Text(mode.label)
                .fontWeight(.bold)
                .foregroundStyle(DesignTokens.Color.accent(area))
        }
        .font(.system(size: 11.5, design: .monospaced))
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("approvals.autoResponse.header")
    }
}

struct ApprovalAutoResponseBanner: View {
    let title: String
    let notice: String
    let area: DesignTokens.Area
    let turnOff: () -> Void

    var body: some View {
        HStack(alignment: .center, spacing: 10) {
            Image(systemName: "exclamationmark.circle")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.accent(area))
            Text("\(Text(title).bold()) \(notice)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityIdentifier("approvals.autoResponse.banner.text")
            Button("끄기", action: turnOff)
                .buttonStyle(.borderless)
                .font(.system(size: 11, weight: .bold, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.accent(area))
                .accessibilityIdentifier("approvals.autoResponse.off")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.accent(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("approvals.autoResponse.banner")
    }
}

struct ApprovalEmptyState: View {
    let isLoading: Bool
    let showsMessage: Bool
    let mode: ApprovalAutoResponse
    let area: DesignTokens.Area

    var body: some View {
        Group {
            if isLoading {
                ProgressView().frame(maxWidth: .infinity)
            } else if showsMessage {
                VStack(spacing: 6) {
                    Text("대기 중인 승인 요청이 없습니다")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                        .accessibilityIdentifier("approvals.empty.state")
                    Text("자동 응답 모드 \(Text(mode.label).bold()) — \(mode.outcomeNotice)")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                        .multilineTextAlignment(.center)
                        .accessibilityIdentifier("approvals.empty.mode")
                }
                .frame(maxWidth: .infinity)
                .padding(.top, DesignTokens.Spacing.xxl)
            }
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }
}

struct ApprovalAutoResponseOption: View {
    let mode: ApprovalAutoResponse
    let selected: Bool
    let area: DesignTokens.Area
    let choose: () -> Void

    var body: some View {
        Button(action: choose) {
            HStack {
                Text(mode.optionLabel)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                Spacer()
                Image(systemName: selected ? "largecircle.fill.circle" : "circle")
                    .foregroundStyle(selected ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area))
            }
            .padding(.horizontal, 14)
            .padding(.vertical, DesignTokens.Spacing.md)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(selected ? .isSelected : [])
        .accessibilityIdentifier("approval.settings.mode.\(mode.rawValue)")
    }
}
