import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

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
        /// §1.11 PR 모니터 영역 — cool indigo. Used by the GitHub PR-monitor
        /// 7th-tab screen (PRD-10).
        case prMonitor
    }

    /// Cross-area semantic colors used by PR monitor cards.
    /// Defined as direct sRGB literals (not asset references) because they
    /// borrow §1.5 (forest) / §1.10 (destructive) — declaring them on the
    /// prMonitor asset would duplicate the source-of-truth in §1.11.
    public enum StatusColor {
        /// `--forest` borrowed from §1.5 routines area for CI success.
        public static let success = SwiftUI.Color(red: 0x4F / 255.0, green: 0x6E / 255.0, blue: 0x5C / 255.0)
        /// `--destructive` borrowed from §1.10 (affirmations) for CI failure.
        public static let failure = SwiftUI.Color(red: 0x9C / 255.0, green: 0x3F / 255.0, blue: 0x2D / 255.0)
        /// Deeper accent for the push-arrival pulse glow (`--accent-strong`).
        public static let arrivalGlow = SwiftUI.Color(red: 0x3D / 255.0, green: 0x2F / 255.0, blue: 0x8E / 255.0)
    }

    public enum Color {
        public static func surface(_ area: Area) -> SwiftUI.Color {
            resolved(area, "surface")
        }

        public static func ink(_ area: Area) -> SwiftUI.Color {
            resolved(area, "ink")
        }

        public static func accent(_ area: Area) -> SwiftUI.Color {
            resolved(area, "accent")
        }

        public static func rule(_ area: Area) -> SwiftUI.Color {
            resolved(area, "rule")
        }

        public static func soft(_ area: Area) -> SwiftUI.Color {
            resolved(area, "soft")
        }

        public static func card(_ area: Area) -> SwiftUI.Color {
            resolved(area, "card")
        }

        /// Try the asset catalog first, then fall back to a hardcoded literal
        /// when the asset is missing. Without this guard `Color(named:)`
        /// silently resolves to clear, which on a dark background renders as
        /// pure black — exactly the symptom we hit on PR-monitor's first
        /// build before adding fallbacks. Fallback values currently cover
        /// §1.11 prMonitor only (the area that exhibited the issue); other
        /// areas keep their existing catalog-only path.
        private static func resolved(_ area: Area, _ token: String) -> SwiftUI.Color {
            let name = asset(area, token)
            #if canImport(UIKit)
            if let ui = UIColor(named: name, in: .module, compatibleWith: nil) {
                return SwiftUI.Color(uiColor: ui)
            }
            if let fb = prMonitorFallback(area: area, token: token) {
                return fb
            }
            #endif
            return SwiftUI.Color(name, bundle: .module)
        }

        #if canImport(UIKit)
        private static let prMonitorLight: [String: (CGFloat, CGFloat, CGFloat)] = [
            "surface": (0xEE / 255.0, 0xED / 255.0, 0xF5 / 255.0),
            "ink":     (0x22 / 255.0, 0x1F / 255.0, 0x33 / 255.0),
            "accent":  (0x5B / 255.0, 0x4D / 255.0, 0xB8 / 255.0),
            "rule":    (0xD8 / 255.0, 0xD5 / 255.0, 0xE4 / 255.0),
            "soft":    (0xDD / 255.0, 0xDA / 255.0, 0xEB / 255.0),
            "card":    (0xFF / 255.0, 0xFF / 255.0, 0xFF / 255.0),
        ]

        private static let prMonitorDark: [String: (CGFloat, CGFloat, CGFloat)] = [
            "surface": (0x22 / 255.0, 0x1C / 255.0, 0x3F / 255.0),
            "ink":     (0xE0 / 255.0, 0xDC / 255.0, 0xEB / 255.0),
            "accent":  (0x84 / 255.0, 0x78 / 255.0, 0xD8 / 255.0),
            "rule":    (0x3F / 255.0, 0x36 / 255.0, 0x5B / 255.0),
            "soft":    (0x2D / 255.0, 0x27 / 255.0, 0x4E / 255.0),
            "card":    (0x2D / 255.0, 0x27 / 255.0, 0x4E / 255.0),
        ]

        private static func prMonitorFallback(area: Area, token: String) -> SwiftUI.Color? {
            guard area == .prMonitor,
                  let light = prMonitorLight[token],
                  let dark = prMonitorDark[token] else { return nil }
            let dynamic = UIColor { trait in
                let rgb = trait.userInterfaceStyle == .dark ? dark : light
                return UIColor(red: rgb.0, green: rgb.1, blue: rgb.2, alpha: 1.0)
            }
            return SwiftUI.Color(uiColor: dynamic)
        }
        #endif

        /// Destructive semantic color for the given area.
        ///
        /// Only defined for areas listed in `tokens.md` §1.10. Asking for an
        /// area without a registered destructive colorset will fall back to
        /// the asset catalog default (clear), so add the colorset before use.
        public static func destructive(_ area: Area) -> SwiftUI.Color {
            SwiftUI.Color(asset(area, "destructive"), bundle: .module)
        }

        // §1.8 — widget surface/rule live outside the area system.
        // Data-source accent inside the widget still uses `accent(area)`.
        public static func widgetSurface() -> SwiftUI.Color {
            SwiftUI.Color("widget/surface", bundle: .module)
        }

        public static func widgetRule() -> SwiftUI.Color {
            SwiftUI.Color("widget/rule", bundle: .module)
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
        case .prMonitor: return "PR · Monitor"
        }
    }
}
