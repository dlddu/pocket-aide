import DesignSystem
import SwiftUI

struct ComingSoonView: View {
    let area: DesignTokens.Area
    let areaLabel: String

    var body: some View {
        VStack(spacing: DesignTokens.Spacing.md) {
            DSAreaLabel(area: area, text: areaLabel)
            Text("준비 중")
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.h2,
                    weight: .semibold
                ))
                .foregroundStyle(DesignTokens.Color.ink(area))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(DesignTokens.Color.surface(area))
    }
}

struct ChatTab: View {
    var body: some View { ComingSoonView(area: .aiChat, areaLabel: "AI 채팅") }
}

struct ScratchpadTab: View {
    var body: some View { ComingSoonView(area: .scratchpad, areaLabel: "임시공간") }
}

struct PersonalTab: View {
    var body: some View { ComingSoonView(area: .personal, areaLabel: "Personal") }
}

struct WorkTab: View {
    var body: some View { ComingSoonView(area: .work, areaLabel: "Work") }
}

struct RoutinesTab: View {
    var body: some View { ComingSoonView(area: .routines, areaLabel: "루틴") }
}
