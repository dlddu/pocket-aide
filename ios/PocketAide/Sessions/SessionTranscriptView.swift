import DesignSystem
import PocketAideAPI
import SwiftUI

struct SessionTranscriptView: View {
    let transcript: AgentSessionTranscript
    let area: DesignTokens.Area
    @State private var expanded: Set<String> = []

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            ForEach(Array(transcript.segments.enumerated()), id: \.offset) { index, segment in
                if index > 0 {
                    Rectangle()
                        .fill(DesignTokens.Color.rule(area))
                        .frame(height: 1)
                }
                switch segment {
                case .text(let text):
                    Text(text.trimmingCharacters(in: .newlines))
                        .font(.system(size: DesignTokens.Typography.bodySm, design: .monospaced))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.65))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .accessibilityIdentifier("sessions.detail.output.text")
                case .turn(let turn):
                    turnView(turn)
                }
            }
        }
        .accessibilityIdentifier("sessions.detail.transcript")
    }

    private func turnView(_ turn: AgentSessionTurn) -> some View {
        let key = "sessions.detail.turn.\(turn.seq)"
        return VStack(alignment: .leading, spacing: 0) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Text("#\(turn.seq)")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                if let timing = AgentSessionFormatting.turnTiming(turn) {
                    Text(timing)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.45))
                        .lineLimit(1)
                }
                Spacer(minLength: 0)
                SessionTurnStatusBadge(status: turn.status, area: area)
            }
            SessionTurnSection(
                title: "입력",
                preview: turn.inputPreview,
                full: turn.input,
                placeholder: turn.hasInput ? "빈 프롬프트" : "앞부분이 생략돼 입력을 볼 수 없어요",
                isExpanded: expanded.contains("\(key).input"),
                area: area,
                identifier: "\(key).input"
            ) { toggle("\(key).input") }
            SessionTurnSection(
                title: "출력",
                preview: turn.outputPreview,
                full: turn.output,
                placeholder: turn.status.isFinished ? "출력이 없어요" : "응답을 기다리는 중…",
                isExpanded: expanded.contains("\(key).output"),
                area: area,
                identifier: "\(key).output"
            ) { toggle("\(key).output") }
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier(key)
    }

    private func toggle(_ key: String) {
        if expanded.contains(key) {
            expanded.remove(key)
        } else {
            expanded.insert(key)
        }
    }
}

struct SessionTurnSection: View {
    let title: String
    let preview: String
    let full: String
    let placeholder: String
    let isExpanded: Bool
    let area: DesignTokens.Area
    let identifier: String
    let onToggle: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Button(action: onToggle) {
                HStack(alignment: .firstTextBaseline, spacing: DesignTokens.Spacing.sm) {
                    Image(systemName: "chevron.right")
                        .font(.system(size: DesignTokens.Typography.caption2xs, weight: .bold))
                        .rotationEffect(.degrees(isExpanded ? 90 : 0))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.45))
                    Text(title)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.accent(area))
                    if !isExpanded {
                        Text(preview.isEmpty ? placeholder : preview)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(preview.isEmpty ? 0.45 : 0.8))
                            .lineLimit(2)
                            .multilineTextAlignment(.leading)
                    }
                    Spacer(minLength: 0)
                }
                .padding(.vertical, DesignTokens.Spacing.sm)
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .accessibilityLabel("\(title) \(isExpanded ? "접기" : "펼치기")")
            .accessibilityValue(preview.isEmpty ? placeholder : preview)
            .accessibilityIdentifier("\(identifier).toggle")
            if isExpanded {
                Text(full.isEmpty ? placeholder : full)
                    .font(.system(size: DesignTokens.Typography.bodySm, design: .monospaced))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(full.isEmpty ? 0.45 : 1))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .textSelection(.enabled)
                    .padding(.leading, DesignTokens.Spacing.lg)
                    .padding(.bottom, DesignTokens.Spacing.sm)
                    .accessibilityIdentifier("\(identifier).full")
            }
        }
    }
}

struct SessionTurnStatusBadge: View {
    let status: AgentSessionTurn.Status
    let area: DesignTokens.Area

    var body: some View {
        HStack(spacing: 5) {
            if status.isFinished {
                Circle()
                    .fill(color)
                    .frame(width: 6, height: 6)
            } else {
                ProgressView().controlSize(.mini)
            }
            Text(status.label)
                .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold))
                .tracking(0.5)
                .foregroundStyle(color)
        }
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("sessions.turn.status")
    }

    private var color: Color {
        switch status {
        case .running: return DesignTokens.StatusColor.inProgress
        case .ok: return DesignTokens.StatusColor.success
        case .error, .timeout: return DesignTokens.StatusColor.failure
        case .canceled, .other: return DesignTokens.Color.ink(area).opacity(0.55)
        }
    }
}
