import SwiftUI

public struct AreaLabel: View {
    private let area: DesignTokens.Area
    private let text: String?
    private let showsDot: Bool

    public init(area: DesignTokens.Area, text: String? = nil, showsDot: Bool = false) {
        self.area = area
        self.text = text
        self.showsDot = showsDot
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
                .tracking(2.4)
                .textCase(.uppercase)
                .foregroundStyle(DesignTokens.Color.accent(area))
        }
    }

    private var defaultText: String {
        switch area {
        case .personal: return "Personal"
        case .work: return "Work"
        case .aiChat: return "채팅"
        case .scratchpad: return "임시공간"
        case .routines: return "루틴"
        case .affirmations: return "다짐"
        case .voice: return "Voice"
        case .system: return "System"
        }
    }
}
