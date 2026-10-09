import SwiftUI

public struct AreaLabel: View {
    private let area: DesignTokens.Area
    private let text: String?
    private let showsDot: Bool
    private let tracking: CGFloat

    public init(area: DesignTokens.Area, text: String? = nil, showsDot: Bool = false, tracking: CGFloat = 2.4) {
        self.area = area
        self.text = text
        self.showsDot = showsDot
        self.tracking = tracking
    }

    public var body: some View {
        HStack(spacing: 6) {
            if showsDot {
                Circle()
                    .fill(DesignTokens.Color.accent(area))
                    .frame(width: 6, height: 6)
            }
            Text(text ?? defaultText)
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.captionXs,
                    weight: .bold,
                    family: .sans
                ))
                .tracking(tracking)
                .textCase(.uppercase)
                .foregroundStyle(DesignTokens.Color.accent(area))
        }
    }

    private var defaultText: String {
        switch area {
        case .personal: return "Personal"
        case .work: return "Work"
        case .aiChat: return "채팅"
        case .scratchpad: return "임시 공간"
        case .routines: return "루틴"
        case .affirmations: return "다짐"
        case .voice: return "Voice"
        case .system: return "System"
        case .prMonitor: return "PR · Monitor"
        }
    }
}
