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

        content()
            .padding(innerPadding)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.white)
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
