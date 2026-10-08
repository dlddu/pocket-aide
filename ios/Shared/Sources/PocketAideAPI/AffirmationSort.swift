import Foundation

public enum AffirmationSortOrder: String, CaseIterable, Sendable, Hashable {
    case priority
    case newest

    public func sorted(_ items: [Affirmation]) -> [Affirmation] {
        switch self {
        case .priority:
            return items.sorted { lhs, rhs in
                let left = Self.rank(lhs.priority)
                let right = Self.rank(rhs.priority)
                return left != right ? left < right : Self.newer(lhs, rhs)
            }
        case .newest:
            return items.sorted(by: Self.newer)
        }
    }

    private static func rank(_ priority: AffirmationPriority) -> Int {
        switch priority {
        case .high: return 0
        case .normal: return 1
        case .low: return 2
        }
    }

    private static func newer(_ lhs: Affirmation, _ rhs: Affirmation) -> Bool {
        (lhs.createdAt, lhs.id) > (rhs.createdAt, rhs.id)
    }
}
