import DesignSystem
import PocketAideAPI
import SwiftUI

/// PR/커밋 단위로 묶인 알림 이력 그룹 카드 (PRD-10 AC13).
/// 디자인 시스템: `Card.history-group.unacked` / `Card.history-group.acked`
/// (`docs/design-system/components.md` §5 참조).
///
/// 헤더: 그룹 키(PR 번호+제목 또는 branch@sha) · 종합 상태 · 미확인 배지.
/// 본체: 그룹 내 항목들을 시각 역순으로 노출. 항목 1개당 `PRMonitorHistoryRow`
/// 재사용. 그룹 전체가 확인 완료된 경우 본체는 접고 헤더만 표시한다.
struct PRMonitorGroupCard: View {
    let group: HistoryGroup
    let highlightedEventID: Int64?
    let onAcknowledge: (Int64) -> Void

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
                    titleLine
                    statusSummary
                }
                Spacer(minLength: DesignTokens.Spacing.sm)
                badge
            }
        }
        .padding(DesignTokens.Spacing.md)
    }

    @ViewBuilder
    private var titleLine: some View {
        if let number = group.prNumber {
            // PR 그룹: repo · #N title
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
            // 커밋 그룹 (PR 없음): repo · branch @sha
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
        // 푸시 진입 항목이 그룹 안에 있으면 그룹 카드 자체가 인디고 글로우.
        // 각 row의 ArrivalGlowOverlay는 row 안에서 그대로 살아 있어 사용자가
        // "이 그룹의 어떤 항목"에 도착했는지 한눈에 식별할 수 있다.
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
