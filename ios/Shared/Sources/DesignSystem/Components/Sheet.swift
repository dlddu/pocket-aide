import SwiftUI

public struct Backdrop: View {
    private let area: DesignTokens.Area
    private let opacity: Double
    private let onTap: () -> Void

    public init(area: DesignTokens.Area, opacity: Double = 0.45, onTap: @escaping () -> Void = {}) {
        self.area = area
        self.opacity = opacity
        self.onTap = onTap
    }

    public var body: some View {
        DesignTokens.Color.ink(area)
            .opacity(opacity)
            .ignoresSafeArea()
            .contentShape(Rectangle())
            .onTapGesture(perform: onTap)
            .accessibilityIdentifier("sheet.backdrop")
    }
}

public struct Handle: View {
    private let area: DesignTokens.Area

    public init(area: DesignTokens.Area) {
        self.area = area
    }

    public var body: some View {
        RoundedRectangle(cornerRadius: 9999, style: .continuous)
            .fill(DesignTokens.Color.rule(area))
            .frame(width: 36, height: 4)
            .padding(.top, 10)
            .padding(.bottom, 6)
            .accessibilityIdentifier("sheet.handle")
    }
}

public struct Sheet<Content: View>: View {
    private let area: DesignTokens.Area
    private let onClose: () -> Void
    private let content: Content

    public init(
        area: DesignTokens.Area,
        onClose: @escaping () -> Void,
        @ViewBuilder content: () -> Content
    ) {
        self.area = area
        self.onClose = onClose
        self.content = content()
    }

    public var body: some View {
        ZStack(alignment: .bottom) {
            Backdrop(area: area, onTap: onClose)
            VStack(spacing: 0) {
                Handle(area: area)
                content
            }
            .frame(maxWidth: .infinity)
            .background(DesignTokens.Color.surface(area))
            .clipShape(UnevenRoundedRectangle(
                topLeadingRadius: 24,
                bottomLeadingRadius: 0,
                bottomTrailingRadius: 0,
                topTrailingRadius: 24,
                style: .continuous
            ))
            .shadow(color: Color.black.opacity(0.12), radius: 12, x: 0, y: -4)
            // NOTE: No `.accessibilityIdentifier("sheet.body")` here. On iOS 26
            // SwiftUI cascades container identifiers down to every leaf,
            // clobbering inner identifiers (sheet.title, sheet.save.button,
            // etc.) that UI tests query.
        }
        .ignoresSafeArea(.container, edges: .bottom)
    }
}
