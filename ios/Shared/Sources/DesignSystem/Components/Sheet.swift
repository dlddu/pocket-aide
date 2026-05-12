import SwiftUI

public struct DSBackdrop: View {
    private let area: DesignTokens.Area
    private let onTap: () -> Void

    public init(area: DesignTokens.Area, onTap: @escaping () -> Void) {
        self.area = area
        self.onTap = onTap
    }

    public var body: some View {
        DesignTokens.Color.ink(area)
            .opacity(0.45)
            .ignoresSafeArea()
            .contentShape(Rectangle())
            .onTapGesture(perform: onTap)
            .accessibilityIdentifier("ds.backdrop")
    }
}

public struct DSHandle: View {
    private let area: DesignTokens.Area

    public init(area: DesignTokens.Area) {
        self.area = area
    }

    public var body: some View {
        RoundedRectangle(cornerRadius: DesignTokens.Radius.chip)
            .fill(DesignTokens.Color.rule(area))
            .frame(width: 36, height: 4)
            .padding(.top, 10)
            .padding(.bottom, 6)
            .accessibilityIdentifier("ds.handle")
    }
}

public struct DSSheet<Content: View>: View {
    private let area: DesignTokens.Area
    @Binding private var isPresented: Bool
    private let content: () -> Content

    public init(
        area: DesignTokens.Area,
        isPresented: Binding<Bool>,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.area = area
        self._isPresented = isPresented
        self.content = content
    }

    public var body: some View {
        ZStack(alignment: .bottom) {
            if isPresented {
                DSBackdrop(area: area) { isPresented = false }
                    .transition(.opacity)

                VStack(spacing: 0) {
                    DSHandle(area: area)
                    content()
                }
                .frame(maxWidth: .infinity)
                .background(DesignTokens.Color.surface(area))
                .clipShape(
                    UnevenRoundedRectangle(
                        topLeadingRadius: DesignTokens.Radius.cardLarge,
                        topTrailingRadius: DesignTokens.Radius.cardLarge
                    )
                )
                .transition(.move(edge: .bottom))
                .accessibilityIdentifier("ds.sheet")
            }
        }
        .animation(.easeInOut(duration: 0.2), value: isPresented)
    }
}
