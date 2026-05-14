import SwiftUI

public enum DesignTokens {
    public enum Area: String, CaseIterable, Sendable {
        case personal
        case work
        case aiChat
        case scratchpad
        case routines
        case affirmations
        case voice
        case system
    }

    public enum Color {
        public static func surface(_ area: Area) -> SwiftUI.Color {
            SwiftUI.Color(asset(area, "surface"), bundle: .module)
        }

        public static func ink(_ area: Area) -> SwiftUI.Color {
            SwiftUI.Color(asset(area, "ink"), bundle: .module)
        }

        public static func accent(_ area: Area) -> SwiftUI.Color {
            SwiftUI.Color(asset(area, "accent"), bundle: .module)
        }

        public static func rule(_ area: Area) -> SwiftUI.Color {
            SwiftUI.Color(asset(area, "rule"), bundle: .module)
        }

        public static func soft(_ area: Area) -> SwiftUI.Color {
            SwiftUI.Color(asset(area, "soft"), bundle: .module)
        }
    }

    public enum Spacing {
        public static let xs: CGFloat = 4
        public static let sm: CGFloat = 8
        public static let md: CGFloat = 12
        public static let lg: CGFloat = 16
        public static let xl: CGFloat = 20
        public static let xxl: CGFloat = 24
    }

    public enum Radius {
        public static let chip: CGFloat = 999
        public static let key: CGFloat = 6
        public static let card: CGFloat = 16
        public static let cardLarge: CGFloat = 24
        public static let bubble: CGFloat = 20
        public static let device: CGFloat = 56
    }

    public enum Typography {
        public static let caption2xs: CGFloat = 10
        public static let captionXs: CGFloat = 11
        public static let captionSm: CGFloat = 12
        public static let bodySm: CGFloat = 13
        public static let body: CGFloat = 14
        public static let bodyLg: CGFloat = 15
        public static let titleMd: CGFloat = 16
        public static let h2: CGFloat = 22
        public static let h1: CGFloat = 27

        public enum Family: Sendable {
            case sans
            case serif
        }

        public static func font(
            size: CGFloat,
            weight: Font.Weight = .regular,
            family: Family = .sans
        ) -> Font {
            switch family {
            case .sans:
                return .system(size: size, weight: weight)
            case .serif:
                return .system(size: size, weight: weight, design: .serif)
            }
        }
    }

    private static func asset(_ area: Area, _ token: String) -> String {
        "\(area.rawValue)/\(token)"
    }
}

public extension DesignTokens.Area {
    var label: String {
        switch self {
        case .personal: return "Personal"
        case .work: return "Work"
        case .aiChat: return "AI Chat"
        case .scratchpad: return "Scratchpad"
        case .routines: return "Routines"
        case .affirmations: return "Affirmations"
        case .voice: return "Voice"
        case .system: return "System"
        }
    }
}
