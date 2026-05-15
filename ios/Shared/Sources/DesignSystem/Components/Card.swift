import SwiftUI

public extension DesignTokens {
    enum CardPadding: Sendable {
        case small
        case medium
        case large

        var value: CGFloat {
            switch self {
            case .small: return 14
            case .medium: return 16
            case .large: return 24
            }
        }

        var radius: CGFloat {
            switch self {
            case .small: return DesignTokens.Radius.card
            case .medium: return DesignTokens.Radius.card
            case .large: return 28
            }
        }
    }
}

public struct Card<Content: View>: View {
    private let area: DesignTokens.Area
    private let padding: DesignTokens.CardPadding
    private let emphasized: Bool
    private let content: Content

    public init(
        area: DesignTokens.Area,
        padding: DesignTokens.CardPadding = .medium,
        emphasized: Bool = false,
        @ViewBuilder content: () -> Content
    ) {
        self.area = area
        self.padding = padding
        self.emphasized = emphasized
        self.content = content()
    }

    public var body: some View {
        content
            .padding(padding.value)
            .background(DesignTokens.Color.card(area))
            .clipShape(RoundedRectangle(cornerRadius: padding.radius, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: padding.radius, style: .continuous)
                    .stroke(
                        emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area),
                        lineWidth: emphasized ? 2 : 1
                    )
            )
    }
}
