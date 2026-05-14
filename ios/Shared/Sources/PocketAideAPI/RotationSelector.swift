import Foundation

/// Pure function that picks a single affirmation from a pool, weighted by
/// priority (high : normal : low = 3 : 2 : 1 by default).
///
/// Pure so its output is deterministic given the same seeded
/// `RandomNumberGenerator`. The view model injects a generator at runtime so
/// production gets `SystemRandomNumberGenerator` and tests get a seeded one.
public struct RotationSelector: Sendable {
    public let weights: [AffirmationPriority: Int]

    public init(weights: [AffirmationPriority: Int] = [.high: 3, .normal: 2, .low: 1]) {
        self.weights = weights
    }

    public func pick<G: RandomNumberGenerator>(
        from pool: [Affirmation],
        using generator: inout G
    ) -> Affirmation? {
        guard !pool.isEmpty else { return nil }
        let total = pool.reduce(0) { $0 + weight(for: $1.priority) }
        guard total > 0 else { return pool.first }
        let target = Int.random(in: 0..<total, using: &generator)
        var running = 0
        for item in pool {
            running += weight(for: item.priority)
            if target < running {
                return item
            }
        }
        return pool.last
    }

    private func weight(for priority: AffirmationPriority) -> Int {
        max(weights[priority] ?? 0, 0)
    }
}
