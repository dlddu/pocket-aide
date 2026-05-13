import SwiftUI

public struct DSCard<Content: View>: View {
    public enum Padding {
        case small
        case large
    }

    private let area: DesignTokens.Area
    private let padding: Padding
    private let emphasized: Bool
    private let content: () -> Content

    @Environment(\.colorScheme) private var colorScheme

    public init(
        area: DesignTokens.Area,
        padding: Padding = .small,
        emphasized: Bool = false,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.area = area
        self.padding = padding
        self.emphasized = emphasized
        self.content = content
    }

    public var body: some View {
        let radius = padding == .large ? DesignTokens.Radius.cardLarge : DesignTokens.Radius.card
        let innerPadding = padding == .large ? DesignTokens.Spacing.xxl : DesignTokens.Spacing.md
        // tokens.md §1.9: in dark variants the card surface is the area's --soft
        // token (e.g. affirmations dark soft = #3A2E1E). In light it stays white
        // per the original mockups.
        let cardBackground: Color = colorScheme == .dark
            ? DesignTokens.Color.soft(area)
            : .white

        content()
            .padding(innerPadding)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: radius))
            .overlay(
                RoundedRectangle(cornerRadius: radius)
                    .stroke(
                        emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area),
                        lineWidth: emphasized ? 2 : 1
                    )
            )
    }
}
