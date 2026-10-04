import DesignSystem
import PocketAideAPI
import SwiftUI

struct PRMonitorGroupCard: View {
    let group: HistoryGroup
    let highlightedEventID: Int64?
    let onAcknowledge: (Int64) -> Void
    let onAcknowledgeGroup: (() -> Void)?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
            if !group.allAcknowledged {
                body(items: group.items)
            }
        }
        .background(cardBackground)
        .overlay(cardBorder)
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
        .overlay(groupHighlightOverlay)
    }

    @ViewBuilder
    private var header: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.xs) {
                    if isGroupHighlighted {
                        Text("방금 진입")
                            .font(DesignTokens.Typography.font(
                                size: DesignTokens.Typography.caption2xs,
                                weight: .bold
                            ))
                            .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                            .accessibilityIdentifier("prmonitor.group.\(group.id).arrival")
                    }
                    titleLine
                    statusSummary
                    if !group.allAcknowledged {
                        groupLinks
                    }
                }
                Spacer(minLength: DesignTokens.Spacing.sm)
                VStack(alignment: .trailing, spacing: 4) {
                    badge
                    ackGroupButton
                }
            }
        }
        .padding(DesignTokens.Spacing.md)
    }

    @ViewBuilder
    private var ackGroupButton: some View {
        if !group.allAcknowledged, let onAcknowledgeGroup {
            Button(action: onAcknowledgeGroup) {
                HStack(spacing: 4) {
                    Image(systemName: "checkmark")
                        .font(.system(size: 9, weight: .bold))
                    Text("모두 확인")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.caption2xs,
                            weight: .bold
                        ))
                }
                .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                .padding(.vertical, 4)
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.accent(.prMonitor), lineWidth: 1)
                )
            }
            .buttonStyle(.borderless)
            .accessibilityIdentifier("prmonitor.group.\(group.id).ack-all.button")
        }
    }

    @ViewBuilder
    private var titleLine: some View {
        if let number = group.prNumber {
            (
                Text(group.repoFullName)
                    .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                + Text(" · ")
                    .foregroundStyle(.tertiary)
                + Text("#\(number)")
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
                + Text(prTitleSuffix(group.prTitle))
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor))
            )
            .font(DesignTokens.Typography.font(
                size: DesignTokens.Typography.body,
                weight: group.allAcknowledged ? .medium : .semibold
            ))
            .strikethrough(group.allAcknowledged)
        } else {
            (
                Text(group.repoFullName)
                    .foregroundStyle(DesignTokens.Color.accent(.prMonitor))
                + Text(" · ")
                    .foregroundStyle(.tertiary)
                + Text(group.headBranch.isEmpty ? "—" : group.headBranch)
                    .foregroundStyle(DesignTokens.Color.ink(.prMonitor).opacity(0.7))
                + Text(shortSHASuffix(group.headSHA))
                    .foregroundStyle(.tertiary)
            )
            .font(DesignTokens.Typography.font(
                size: DesignTokens.Typography.body,
                weight: group.allAcknowledged ? .medium : .semibold
            ))
            .strikethrough(group.allAcknowledged)
        }
    }

    @ViewBuilder
    private var statusSummary: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            if group.successCount > 0 {
                statusDot(color: DesignTokens.StatusColor.success, label: "통과 \(group.successCount)")
            }
            if group.failureCount > 0 {
                statusDot(color: DesignTokens.StatusColor.failure, label: "실패 \(group.failureCount)")
            }
            if group.inProgressCount > 0 {
                statusDot(color: DesignTokens.StatusColor.inProgress, label: "진행 \(group.inProgressCount)")
            }
            Text("CI \(group.items.count)건")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(.secondary)
            Spacer(minLength: 0)
            Text("최근 \(relativeRecent)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(.tertiary)
        }
    }

    @ViewBuilder
    private func statusDot(color: Color, label: String) -> some View {
        HStack(spacing: 4) {
            Circle()
                .fill(color)
                .frame(width: 6, height: 6)
            Text(label)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))
                .foregroundStyle(color)
        }
    }

    @ViewBuilder
    private var badge: some View {
        if group.allAcknowledged {
            Text("모두 확인")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))
                .foregroundStyle(.secondary)
                .padding(.vertical, 4)
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .overlay(
                    RoundedRectangle(cornerRadius: 6, style: .continuous)
                        .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                )
                .accessibilityIdentifier("prmonitor.group.\(group.id).badge")
        } else {
            Text("미확인 \(group.unacknowledgedCount)")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))
                .foregroundStyle(.white)
                .padding(.vertical, 4)
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .background(DesignTokens.Color.accent(.prMonitor))
                .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
                .accessibilityIdentifier("prmonitor.group.\(group.id).badge")
        }
    }

    @ViewBuilder
    private func body(items: [NotificationHistoryItem]) -> some View {
        Rectangle()
            .fill(DesignTokens.Color.rule(.prMonitor))
            .frame(height: 1)
        VStack(spacing: 0) {
            ForEach(Array(items.enumerated()), id: \.element.id) { idx, item in
                if idx > 0 {
                    Rectangle()
                        .fill(DesignTokens.Color.rule(.prMonitor).opacity(0.6))
                        .frame(height: 1)
                }
                PRMonitorHistoryRow(
                    item: item,
                    isHighlighted: highlightedEventID == item.id,
                    onAcknowledge: { onAcknowledge(item.id) }
                )
                .padding(.horizontal, DesignTokens.Spacing.sm)
                .padding(.vertical, DesignTokens.Spacing.xs)
                .accessibilityIdentifier("prmonitor.row.\(item.id)")
            }
        }
    }

    private var cardBackground: Color {
        group.allAcknowledged
            ? DesignTokens.Color.surface(.prMonitor).opacity(0.6)
            : DesignTokens.Color.card(.prMonitor)
    }

    @ViewBuilder
    private var cardBorder: some View {
        if group.allAcknowledged {
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                .strokeBorder(
                    DesignTokens.Color.rule(.prMonitor),
                    style: StrokeStyle(lineWidth: 1, dash: [4, 3])
                )
        } else {
            ZStack(alignment: .leading) {
                RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                    .stroke(DesignTokens.Color.rule(.prMonitor), lineWidth: 1)
                Rectangle()
                    .fill(DesignTokens.Color.accent(.prMonitor))
                    .frame(width: 3)
                    .clipShape(
                        UnevenRoundedRectangle(cornerRadii: .init(
                            topLeading: DesignTokens.Radius.card,
                            bottomLeading: DesignTokens.Radius.card,
                            bottomTrailing: 0, topTrailing: 0
                        ), style: .continuous)
                    )
            }
        }
    }

    @ViewBuilder
    private var groupHighlightOverlay: some View {
        if isGroupHighlighted {
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                .stroke(DesignTokens.StatusColor.arrivalGlow.opacity(0.16), lineWidth: 4)
                .shadow(color: DesignTokens.StatusColor.arrivalGlow.opacity(0.35), radius: 12, x: 0, y: 6)
        }
    }

    private var isGroupHighlighted: Bool {
        guard let id = highlightedEventID else { return false }
        return group.items.contains(where: { $0.id == id })
    }

    private var relativeRecent: String {
        let date = Date(timeIntervalSince1970: TimeInterval(group.latestCreatedAt))
        return Self.timeFormatter.string(from: date)
    }

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "HH:mm"
        return f
    }()

    private func prTitleSuffix(_ title: String?) -> String {
        guard let title, !title.isEmpty else { return "" }
        return " \(title)"
    }

    private func shortSHASuffix(_ sha: String) -> String {
        guard !sha.isEmpty else { return "" }
        let short = String(sha.prefix(7))
        return " @\(short)"
    }
}

extension PRMonitorGroupCard {
    @ViewBuilder
    fileprivate var groupLinks: some View {
        if let prDest = group.prURL.flatMap(URL.init(string:)) {
            HStack(spacing: 4) {
                linkChip(
                    label: "PR",
                    systemImage: "arrow.triangle.pull",
                    url: prDest,
                    identifier: "prmonitor.group.\(group.id).link.pr"
                )
            }
        }
    }

    @ViewBuilder
    fileprivate func linkChip(label: String, systemImage: String, url: URL, identifier: String) -> some View {
        Link(destination: url) {
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
