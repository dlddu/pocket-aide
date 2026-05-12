import SwiftUI

public struct DSAreaLabel: View {
    private let area: DesignTokens.Area
    private let text: String

    public init(area: DesignTokens.Area, text: String) {
        self.area = area
        self.text = text
    }

    public var body: some View {
        Text(text)
            .font(DesignTokens.Typography.font(
                size: DesignTokens.Typography.captionXs,
                weight: .bold
            ))
            .tracking(2.4)
            .textCase(.uppercase)
            .foregroundStyle(DesignTokens.Color.accent(area))
    }
}
