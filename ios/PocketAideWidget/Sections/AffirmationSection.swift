import DesignSystem
import PocketAideAPI
import SwiftUI

struct AffirmationSection: View {
    let state: WidgetAffirmationState

    var body: some View {
        // Link gives this region its own tap target on Large widgets — the
        // outer `widgetURL` covers everything else (placeholder areas) and
        // sends the user to the app root.
        Link(destination: URL(string: "pocketaide://affirmations")!) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                HStack(spacing: 6) {
                    AreaLabel(area: .affirmations, text: "오늘의 다짐", showsDot: true)
                    Spacer(minLength: 0)
                }
                Text(message)
                    .font(DesignTokens.Typography.font(size: 15, family: .serif))
                    .foregroundStyle(textColor)
                    .multilineTextAlignment(.leading)
                    .lineLimit(3)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(.plain)
    }

    private var message: String {
        switch state {
        case .loaded(let affirmation): return affirmation.text
        case .empty: return "다짐을 앱에 등록해보세요."
        case .needsLogin: return "앱에서 로그인이 필요해요."
        case .error: return "잠시 후 다시 시도할게요."
        }
    }

    private var textColor: Color {
        switch state {
        case .loaded: return DesignTokens.Color.ink(.affirmations)
        default: return DesignTokens.Color.ink(.affirmations).opacity(0.55)
        }
    }
}
