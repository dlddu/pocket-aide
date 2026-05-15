import SwiftUI

public struct TabBarItem: View {
    private let area: DesignTokens.Area
    private let label: String
    private let systemImage: String
    private let isActive: Bool

    public init(
        area: DesignTokens.Area,
        label: String,
        systemImage: String,
        isActive: Bool
    ) {
        self.area = area
        self.label = label
        self.systemImage = systemImage
        self.isActive = isActive
    }

    public var body: some View {
        VStack(spacing: 2) {
            Image(systemName: systemImage)
                .font(.system(size: 20, weight: isActive ? .semibold : .regular))
            Text(label)
                .font(.system(
                    size: DesignTokens.Typography.caption2xs,
                    weight: isActive ? .bold : .medium
                ))
        }
        .foregroundStyle(isActive ? DesignTokens.Color.accent(area) : Color(white: 0.62))
        .frame(maxWidth: .infinity)
        .padding(.vertical, 4)
    }
}
