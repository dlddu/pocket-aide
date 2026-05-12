import Foundation

public struct RotationSelector: Sendable {
    public typealias Weights = [Priority: Int]

    public static let defaultWeights: Weights = [
        .high: 3,
        .normal: 2,
        .low: 1,
    ]

    private let weights: Weights

    public init(weights: Weights = RotationSelector.defaultWeights) {
        self.weights = weights
    }

    public func pick<G: RandomNumberGenerator>(
        from items: [Affirmation],
        using rng: inout G
    ) -> Affirmation? {
        guard !items.isEmpty else { return nil }

        let weighted = items.map { item -> (Affirmation, Int) in
            (item, max(0, weights[item.priority] ?? 0))
        }
        let total = weighted.reduce(0) { $0 + $1.1 }
        guard total > 0 else { return items.first }

        let roll = Int.random(in: 0..<total, using: &rng)
        var cursor = 0
        for (item, weight) in weighted {
            cursor += weight
            if roll < cursor {
                return item
            }
        }
        return weighted.last?.0
    }
}
