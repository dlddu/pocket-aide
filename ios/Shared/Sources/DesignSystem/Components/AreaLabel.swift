import SwiftUI

public struct AreaLabel: View {
    public enum DotShape {
        case circle
        case square
    }

    private let area: DesignTokens.Area
    private let text: String?
    private let showsDot: Bool
    private let dotShape: DotShape
    private let dotSize: CGFloat
    private let tracking: CGFloat

    public init(
        area: DesignTokens.Area,
        text: String? = nil,
        showsDot: Bool = false,
        dotShape: DotShape = .circle,
        dotSize: CGFloat = 6,
        tracking: CGFloat = 2.4
    ) {
        self.area = area
        self.text = text
        self.showsDot = showsDot
        self.dotShape = dotShape
        self.dotSize = dotSize
        self.tracking = tracking
    }

    public var body: some View {
        HStack(spacing: 6) {
            if showsDot {
                dot
                    .frame(width: dotSize, height: dotSize)
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

    @ViewBuilder private var dot: some View {
        switch dotShape {
        case .circle: Circle().fill(DesignTokens.Color.accent(area))
        case .square: Rectangle().fill(DesignTokens.Color.accent(area))
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
