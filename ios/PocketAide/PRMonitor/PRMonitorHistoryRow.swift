import DesignSystem
import PocketAideAPI
import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

/// One row of the PR-monitor history list. Visually distinguishes unacked
/// (`Card.history-item.unacked`: surface + left indigo stripe + ack button)
/// vs. acked (`Card.history-item.acked`: dashed + dim + strikethrough). Push
/// arrival adds a 5-second indigo glow overlay (PRD-10 AC7).
///
/// Pulled out of `PRMonitorView.swift` to keep that file under SwiftLint's
/// 400-line file_length limit.
struct PRMonitorHistoryRow: View {
    let item: NotificationHistoryItem
    let isHighlighted: Bool
    let onAcknowledge: () -> Void

    var body: some View {
        let acked = item.acknowledgedAt != nil
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            header
            titleLine
            metaLine
            actionsRow
        }
        .padding(DesignTokens.Spacing.md)
        .background(rowBackground)
        .overlay(rowBorder)
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(highlightOverlay)
        .opacity(acked ? 0.7 : 1.0)
    }

    @ViewBuilder
    private var header: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            statusBadge
            statusLabel
            Spacer()
            Text(timestampString)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(.secondary)
        }
    }

    @ViewBuilder
    private var statusBadge: some View {
        Circle()
            .fill(statusColor)
            .frame(width: 18, height: 18)
            .overlay(
                Image(systemName: badgeIcon)
                    .font(.system(size: 9, weight: .bold))
                    .foregroundStyle(.white)
            )
    }

    @ViewBuilder
    private var statusLabel: some View {
        if item.acknowledgedAt != nil {
            Text("\(verdictLabel) · 확인됨")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                .foregroundStyle(.secondary)
                .accessibilityIdentifier("prmonitor.row.\(item.id).status")
        } else {
            Text(verdictLabel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                .foregroundStyle(statusColor)
                .accessibilityIdentifier("prmonitor.row.\(item.id).status")
        }
    }

    @ViewBuilder
    private var titleLine: some View {
        Text(titleString)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: item.acknowledgedAt == nil ? .semibold : .medium))
            .foregroundStyle(item.acknowledgedAt == nil
                ? DesignTokens.Color.ink(.prMonitor)
                : DesignTokens.Color.ink(.prMonitor).opacity(0.55))
            .strikethrough(item.acknowledgedAt != nil)
            .accessibilityIdentifier("prmonitor.row.\(item.id).title")
    }

    @ViewBuilder
    private var metaLine: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Text(item.workflowName.isEmpty ? "workflow" : item.workflowName)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(.secondary)
            if !item.headBranch.isEmpty {
                Text("·")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(.tertiary)
                Text(item.headBranch)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(.secondary)
            }
            if let ackedAt = item.acknowledgedAt {
                Spacer()
                Text("확인 \(relativeAck(ackedAt))")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                    .foregroundStyle(.secondary)
            }
        }
    }

    @ViewBuilder
    private var actionsRow: some View {
        if item.acknowledgedAt == nil {
            HStack(spacing: DesignTokens.Spacing.sm) {
                if let url = item.prURL {
                    linkChip(
                        label: "PR",
                        systemImage: "arrow.triangle.pull",
                        url: url,
                        identifier: "prmonitor.row.\(item.id).link.pr"
                    )
                }
                if let url = item.commitURL {
                    linkChip(
                        label: "커밋",
                        systemImage: "circle.fill",
                        url: url,
                        identifier: "prmonitor.row.\(item.id).link.commit"
                    )
                }
                if let url = item.runURL {
                    linkChip(
                        label: "런",
                        systemImage: "play.fill",
                        url: url,
                        identifier: "prmonitor.row.\(item.id).link.run"
                    )
                }
                Spacer()
                Button {
                    onAcknowledge()
                } label: {
                    Label("확인", systemImage: "checkmark")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                        .padding(.vertical, 6)
                        .padding(.horizontal, DesignTokens.Spacing.md)
                        .background(DesignTokens.Color.accent(.prMonitor))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.borderless)
                .accessibilityIdentifier("prmonitor.row.\(item.id).ack.button")
            }
        }
    }

    @ViewBuilder
    private func linkChip(label: String, systemImage: String, url: String, identifier: String) -> some View {
        // Link instead of Button so SwiftUI's List doesn't merge the link tap
        // into the row's primary action — i.e. tapping the chip must open
        // GitHub WITHOUT also triggering the explicit 확인 button next to it
        // (AC12: external link taps never acknowledge). Each chip carries an
        // action-specific symbol so the destination is recognisable at a
        // glance: PR(pull arrow), commit(dot), run(play).
        if let dest = URL(string: url) {
            Link(destination: dest) {
                HStack(spacing: 4) {
                    Image(systemName: systemImage)
                        .font(.system(size: 9, weight: .semibold))
                    Text(label)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                }
                .padding(.vertical, 4)
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                )
                .foregroundStyle(.secondary)
            }
            .buttonStyle(.borderless)
            .accessibilityIdentifier(identifier)
        }
    }

    private var rowBackground: some View {
        item.acknowledgedAt == nil
            ? DesignTokens.Color.card(.prMonitor)
            : DesignTokens.Color.surface(.prMonitor).opacity(0.6)
    }

    @ViewBuilder
    private var rowBorder: some View {
        if item.acknowledgedAt == nil {
            ZStack(alignment: .leading) {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                Rectangle()
                    .fill(DesignTokens.Color.accent(.prMonitor))
                    .frame(width: 3)
                    .clipShape(
                        UnevenRoundedRectangle(cornerRadii: .init(
                            topLeading: 12, bottomLeading: 12,
                            bottomTrailing: 0, topTrailing: 0
                        ), style: .continuous)
                    )
            }
        } else {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(
                    DesignTokens.Color.rule(.prMonitor),
                    style: StrokeStyle(lineWidth: 1, dash: [4, 3])
                )
        }
    }

    @ViewBuilder
    private var highlightOverlay: some View {
        if isHighlighted {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .stroke(DesignTokens.StatusColor.arrivalGlow.opacity(0.55), lineWidth: 3)
                .blur(radius: 1.5)
        }
    }

    private var statusColor: Color {
        switch item.conclusion.lowercased() {
        case "success": return DesignTokens.StatusColor.success
        case "failure", "timed_out", "cancelled", "action_required": return DesignTokens.StatusColor.failure
        default: return DesignTokens.StatusColor.failure
        }
    }

    private var badgeIcon: String {
        switch item.conclusion.lowercased() {
        case "success": return "checkmark"
        default: return "xmark"
        }
    }

    private var verdictLabel: String {
        switch item.conclusion.lowercased() {
        case "success": return "CI 통과"
        case "failure": return "CI 실패"
        case "cancelled": return "CI 취소"
        case "timed_out": return "CI 타임아웃"
        default: return "CI \(item.conclusion)"
        }
    }

    private var titleString: String {
        if let number = item.prNumber, let title = item.prTitle, !title.isEmpty {
            return "\(item.repoFullName) · #\(number) \(title)"
        }
        if let number = item.prNumber {
            return "\(item.repoFullName) · #\(number)"
        }
        // No PR linked — fallback (AC6 PR-less case).
        if item.headBranch.isEmpty {
            return item.repoFullName
        }
        return "\(item.repoFullName) · \(item.headBranch)"
    }

    private var timestampString: String {
        Self.timeFormatter.string(from: Date(timeIntervalSince1970: TimeInterval(item.createdAt)))
    }

    private func relativeAck(_ epoch: Int64) -> String {
        let date = Date(timeIntervalSince1970: TimeInterval(epoch))
        return Self.relativeFormatter.localizedString(for: date, relativeTo: Date())
    }

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "HH:mm"
        return f
    }()

    private static let relativeFormatter = RelativeDateTimeFormatter()
}
