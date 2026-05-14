import SwiftUI

public struct ScreenHeader<Trailing: View>: View {
    private let area: DesignTokens.Area
    private let title: String
    private let titleFamily: DesignTokens.Typography.Family
    private let trailing: Trailing

    public init(
        area: DesignTokens.Area,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans,
        @ViewBuilder trailing: () -> Trailing
    ) {
        self.area = area
        self.title = title
        self.titleFamily = titleFamily
        self.trailing = trailing()
    }

    public var body: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 4) {
                AreaLabel(area: area)
                Text(title)
                    .font(DesignTokens.Typography.font(
                        size: 24,
                        weight: .bold,
                        family: titleFamily
                    ))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                    .tracking(-0.01 * 24)
                    .accessibilityIdentifier("screen.header.title")
            }
            Spacer()
            trailing
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
        .padding(.bottom, DesignTokens.Spacing.sm)
    }
}

public extension ScreenHeader where Trailing == EmptyView {
    init(
        area: DesignTokens.Area,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans
    ) {
        self.init(area: area, title: title, titleFamily: titleFamily) { EmptyView() }
    }
}
