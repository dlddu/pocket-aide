import DesignSystem
import PocketAideAPI
import SwiftUI

struct PriorityDots: View {
    enum Layout {
        case vertical
        case horizontal
    }

    let priority: Priority
    var layout: Layout = .vertical

    private var filledCount: Int {
        switch priority {
        case .high: return 3
        case .normal: return 2
        case .low: return 1
        }
    }

    var body: some View {
        let dots = (0..<3).map { index in
            Circle()
                .fill(
                    index < filledCount
                        ? DesignTokens.Color.accent(.affirmations)
                        : DesignTokens.Color.accent(.affirmations).opacity(0.3)
                )
                .frame(width: 6, height: 6)
        }

        Group {
            switch layout {
            case .vertical:
                VStack(spacing: 2) {
                    ForEach(0..<3) { i in dots[i] }
                }
            case .horizontal:
                HStack(spacing: 2) {
                    ForEach(0..<3) { i in dots[i] }
                }
            }
        }
        .accessibilityIdentifier("priority.dots.\(priority.rawValue)")
    }
}
