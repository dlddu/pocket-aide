import DesignSystem
import SwiftUI

struct PlaceholderTab: View {
    let area: DesignTokens.Area
    let title: String

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(area).ignoresSafeArea()
            VStack(spacing: DesignTokens.Spacing.md) {
                AreaLabel(area: area)
                Text(title)
                    .font(DesignTokens.Typography.font(
                        size: DesignTokens.Typography.h2,
                        weight: .bold
                    ))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                Text("준비 중")
                    .font(DesignTokens.Typography.font(
                        size: DesignTokens.Typography.body,
                        weight: .medium
                    ))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
            }
            .accessibilityIdentifier("placeholder.\(area.rawValue)")
        }
    }
}

struct ChatTab: View {
    var body: some View { PlaceholderTab(area: .aiChat, title: "AI 채팅") }
}

struct ScratchpadTab: View {
    var body: some View { PlaceholderTab(area: .scratchpad, title: "임시공간") }
}

struct PersonalTab: View {
    var body: some View { PlaceholderTab(area: .personal, title: "개인") }
}

struct WorkTab: View {
    var body: some View { PlaceholderTab(area: .work, title: "회사") }
}

struct RoutinesTab: View {
    var body: some View { PlaceholderTab(area: .routines, title: "루틴") }
}
