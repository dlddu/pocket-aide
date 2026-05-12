import SwiftUI

public struct DSTabBarItem: View {
    private let area: DesignTokens.Area
    private let label: String
    private let systemImage: String
    private let isActive: Bool

    public init(area: DesignTokens.Area, label: String, systemImage: String, isActive: Bool) {
        self.area = area
        self.label = label
        self.systemImage = systemImage
        self.isActive = isActive
    }

    public var body: some View {
        VStack(spacing: 2) {
            Image(systemName: systemImage)
                .font(.system(size: 18, weight: isActive ? .semibold : .regular))
            Text(label)
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.caption2xs,
                    weight: isActive ? .bold : .medium
                ))
        }
        .foregroundStyle(
            isActive
                ? DesignTokens.Color.accent(area)
                : DesignTokens.Color.ink(area).opacity(0.45)
        )
    }
}
