import DesignSystem
import SwiftUI

struct RootView: View {
    enum Tab: String, CaseIterable {
        case chat
        case scratchpad
        case personal
        case work
        case routines
        case affirmations

        var area: DesignTokens.Area {
            switch self {
            case .chat: return .aiChat
            case .scratchpad: return .scratchpad
            case .personal: return .personal
            case .work: return .work
            case .routines: return .routines
            case .affirmations: return .affirmations
            }
        }

        var label: String {
            switch self {
            case .chat: return "채팅"
            case .scratchpad: return "임시공간"
            case .personal: return "개인"
            case .work: return "회사"
            case .routines: return "루틴"
            case .affirmations: return "다짐"
            }
        }

        var systemImage: String {
            switch self {
            case .chat: return "bubble.left.fill"
            case .scratchpad: return "doc.text"
            case .personal: return "person.fill"
            case .work: return "briefcase.fill"
            case .routines: return "arrow.triangle.2.circlepath"
            case .affirmations: return "heart.fill"
            }
        }

        var accessibilityIdentifier: String {
            "tab.\(rawValue)"
        }
    }

    @State private var selected: Tab = .affirmations

    var body: some View {
        TabView(selection: $selected) {
            ChatTab()
                .tabItem { Label(Tab.chat.label, systemImage: Tab.chat.systemImage) }
                .tag(Tab.chat)
                .accessibilityIdentifier(Tab.chat.accessibilityIdentifier)

            ScratchpadTab()
                .tabItem { Label(Tab.scratchpad.label, systemImage: Tab.scratchpad.systemImage) }
                .tag(Tab.scratchpad)
                .accessibilityIdentifier(Tab.scratchpad.accessibilityIdentifier)

            PersonalTab()
                .tabItem { Label(Tab.personal.label, systemImage: Tab.personal.systemImage) }
                .tag(Tab.personal)
                .accessibilityIdentifier(Tab.personal.accessibilityIdentifier)

            WorkTab()
                .tabItem { Label(Tab.work.label, systemImage: Tab.work.systemImage) }
                .tag(Tab.work)
                .accessibilityIdentifier(Tab.work.accessibilityIdentifier)

            RoutinesTab()
                .tabItem { Label(Tab.routines.label, systemImage: Tab.routines.systemImage) }
                .tag(Tab.routines)
                .accessibilityIdentifier(Tab.routines.accessibilityIdentifier)

            AffirmationsView()
                .tabItem { Label(Tab.affirmations.label, systemImage: Tab.affirmations.systemImage) }
                .tag(Tab.affirmations)
                .accessibilityIdentifier(Tab.affirmations.accessibilityIdentifier)
        }
        .tint(DesignTokens.Color.accent(selected.area))
    }
}

#Preview {
    RootView()
}
