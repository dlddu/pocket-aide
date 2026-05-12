import SwiftUI

public struct DSScreenHeader<Trailing: View>: View {
    private let area: DesignTokens.Area
    private let areaLabel: String
    private let title: String
    private let titleFamily: DesignTokens.Typography.Family
    private let trailing: () -> Trailing

    public init(
        area: DesignTokens.Area,
        areaLabel: String,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans,
        @ViewBuilder trailing: @escaping () -> Trailing
    ) {
        self.area = area
        self.areaLabel = areaLabel
        self.title = title
        self.titleFamily = titleFamily
        self.trailing = trailing
    }

    public var body: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 2) {
                DSAreaLabel(area: area, text: areaLabel)
                Text(title)
                    .font(DesignTokens.Typography.font(
                        size: 24,
                        weight: .bold,
                        family: titleFamily
                    ))
                    .foregroundStyle(DesignTokens.Color.ink(area))
            }
            Spacer(minLength: DesignTokens.Spacing.md)
            trailing()
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
        .padding(.bottom, DesignTokens.Spacing.sm)
    }
}

public extension DSScreenHeader where Trailing == EmptyView {
    init(
        area: DesignTokens.Area,
        areaLabel: String,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans
    ) {
        self.init(
            area: area,
            areaLabel: areaLabel,
            title: title,
            titleFamily: titleFamily,
            trailing: { EmptyView() }
        )
    }
}
