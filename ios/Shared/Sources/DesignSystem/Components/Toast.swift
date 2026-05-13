import SwiftUI

public struct DSToast: View {
    private let area: DesignTokens.Area
    private let message: String
    private let systemImage: String?

    public init(area: DesignTokens.Area, message: String, systemImage: String? = nil) {
        self.area = area
        self.message = message
        self.systemImage = systemImage
    }

    public var body: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            if let systemImage {
                Image(systemName: systemImage)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.accent(area))
            }
            Text(message)
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.body,
                    weight: .medium
                ))
                .foregroundStyle(DesignTokens.Color.surface(area))
                .lineLimit(2)
                .multilineTextAlignment(.leading)
        }
        .padding(.horizontal, DesignTokens.Spacing.lg)
        .padding(.vertical, DesignTokens.Spacing.md)
        .background(DesignTokens.Color.ink(area))
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.cardLarge))
        .shadow(color: Color.black.opacity(0.18), radius: 10, x: 0, y: 4)
        .accessibilityIdentifier("ds.toast")
    }
}

public extension View {
    /// Shows a transient toast at the bottom of the receiver whenever `message`
    /// becomes non-nil. The toast auto-dismisses after `duration` seconds by
    /// clearing the binding.
    ///
    /// Use the area's tokens (ink background, surface text) — see
    /// components.md §9 Toast.
    func dsToast(
        area: DesignTokens.Area,
        message: Binding<String?>,
        systemImage: String? = "exclamationmark.circle.fill",
        duration: TimeInterval = 3
    ) -> some View {
        modifier(DSToastModifier(
            area: area,
            message: message,
            systemImage: systemImage,
            duration: duration
        ))
    }
}

private struct DSToastModifier: ViewModifier {
    let area: DesignTokens.Area
    @Binding var message: String?
    let systemImage: String?
    let duration: TimeInterval

    func body(content: Content) -> some View {
        content
            .overlay(alignment: .bottom) {
                if let text = message {
                    DSToast(area: area, message: text, systemImage: systemImage)
                        .padding(.horizontal, DesignTokens.Spacing.xl)
                        .padding(.bottom, DesignTokens.Spacing.xl)
                        .transition(.move(edge: .bottom).combined(with: .opacity))
                        .task(id: text) {
                            try? await Task.sleep(nanoseconds: UInt64(duration * 1_000_000_000))
                            if message == text {
                                message = nil
                            }
                        }
                        .onTapGesture { message = nil }
                }
            }
            .animation(.easeInOut(duration: 0.2), value: message)
    }
}
