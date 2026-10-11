import SwiftUI

public struct ScreenHeader<LabelAccessory: View, Trailing: View>: View {
    private let area: DesignTokens.Area
    private let title: String
    private let titleFamily: DesignTokens.Typography.Family
    private let titleSize: CGFloat
    private let labelTracking: CGFloat
    private let labelShowsDot: Bool
    private let labelDotShape: AreaLabel.DotShape
    private let labelDotSize: CGFloat
    private let labelSpacing: CGFloat
    private let subtitle: String?
    private let subtitleIdentifier: String
    private let subtitleMonospaced: Bool
    private let bottomPadding: CGFloat
    private let labelAccessory: LabelAccessory
    private let trailing: Trailing

    public init(
        area: DesignTokens.Area,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans,
        titleSize: CGFloat = 24,
        labelTracking: CGFloat = 2.4,
        labelShowsDot: Bool = false,
        labelDotShape: AreaLabel.DotShape = .circle,
        labelDotSize: CGFloat = 6,
        labelSpacing: CGFloat = 2,
        subtitle: String? = nil,
        subtitleIdentifier: String = "screen.header.subtitle",
        subtitleMonospaced: Bool = false,
        bottomPadding: CGFloat = DesignTokens.Spacing.sm,
        @ViewBuilder labelAccessory: () -> LabelAccessory,
        @ViewBuilder trailing: () -> Trailing
    ) {
        self.area = area
        self.title = title
        self.titleFamily = titleFamily
        self.titleSize = titleSize
        self.labelTracking = labelTracking
        self.labelShowsDot = labelShowsDot
        self.labelDotShape = labelDotShape
        self.labelDotSize = labelDotSize
        self.labelSpacing = labelSpacing
        self.subtitle = subtitle
        self.subtitleIdentifier = subtitleIdentifier
        self.subtitleMonospaced = subtitleMonospaced
        self.bottomPadding = bottomPadding
        self.labelAccessory = labelAccessory()
        self.trailing = trailing()
    }

    public var body: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: labelSpacing) {
                HStack(spacing: 0) {
                    AreaLabel(
                        area: area,
                        showsDot: labelShowsDot,
                        dotShape: labelDotShape,
                        dotSize: labelDotSize,
                        tracking: labelTracking
                    )
                    labelAccessory
                }
                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .font(DesignTokens.Typography.font(
                            size: titleSize,
                            weight: .bold,
                            family: titleFamily
                        ))
                        .foregroundStyle(DesignTokens.Color.ink(area))
                        .tracking(-0.025 * titleSize)
                        .accessibilityIdentifier("screen.header.title")
                    if let subtitle {
                        Text(subtitle)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                            .monospaced(subtitleMonospaced)
                            .accessibilityIdentifier(subtitleIdentifier)
                    }
                }
            }
            Spacer()
            trailing
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
        .padding(.bottom, bottomPadding)
    }
}

public extension ScreenHeader where LabelAccessory == EmptyView {
    init(
        area: DesignTokens.Area,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans,
        subtitle: String? = nil,
        @ViewBuilder trailing: () -> Trailing
    ) {
        self.init(area: area, title: title, titleFamily: titleFamily, subtitle: subtitle) {
            EmptyView()
        } trailing: {
            trailing()
        }
    }
}

public extension ScreenHeader where LabelAccessory == EmptyView, Trailing == EmptyView {
    init(
        area: DesignTokens.Area,
        title: String,
        titleFamily: DesignTokens.Typography.Family = .sans,
        subtitle: String? = nil
    ) {
        self.init(area: area, title: title, titleFamily: titleFamily, subtitle: subtitle) { EmptyView() }
    }
}
