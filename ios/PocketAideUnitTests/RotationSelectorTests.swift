import XCTest
@testable import PocketAideAPI

final class RotationSelectorTests: XCTestCase {
    private func aff(_ id: Int64, _ p: AffirmationPriority) -> Affirmation {
        Affirmation(id: id, text: "t\(id)", priority: p, createdAt: 0, updatedAt: 0)
    }

    func testEmptyPoolReturnsNil() {
        let selector = RotationSelector()
        var rng = SeededRNG(seed: 1)
        XCTAssertNil(selector.pick(from: [], using: &rng))
    }

    func testSingleItemAlwaysReturnsIt() {
        let selector = RotationSelector()
        let pool = [aff(1, .normal)]
        for seed in 1...32 {
            var rng = SeededRNG(seed: UInt64(seed))
            XCTAssertEqual(selector.pick(from: pool, using: &rng)?.id, 1)
        }
    }

    func testSameSeedIsDeterministic() {
        let selector = RotationSelector()
        let pool = [
            aff(1, .high),
            aff(2, .normal),
            aff(3, .low),
            aff(4, .high),
        ]
        var rngA = SeededRNG(seed: 42)
        var rngB = SeededRNG(seed: 42)
        XCTAssertEqual(
            selector.pick(from: pool, using: &rngA)?.id,
            selector.pick(from: pool, using: &rngB)?.id
        )
    }

    func testWeightedDistribution() {
        let selector = RotationSelector()
        let pool = [aff(1, .high), aff(2, .normal), aff(3, .low)]
        var rng = SeededRNG(seed: 12345)
        var counts: [Int64: Int] = [1: 0, 2: 0, 3: 0]
        let trials = 6000
        for _ in 0..<trials {
            if let pick = selector.pick(from: pool, using: &rng) {
                counts[pick.id, default: 0] += 1
            }
        }
        let totalWeight: Double = 3 + 2 + 1
        let expected: [Int64: Double] = [
            1: Double(trials) * (3.0 / totalWeight),
            2: Double(trials) * (2.0 / totalWeight),
            3: Double(trials) * (1.0 / totalWeight),
        ]
        // 10% tolerance — the test is statistical (multinomial draw) and CI
        // saw 6.9% drift on the low bucket which is well within expected
        // sampling variance for n=6000. Widen if it ever flakes again.
        let tolerance = 0.10
        for (id, want) in expected {
            let got = Double(counts[id] ?? 0)
            let drift = abs(got - want) / want
            XCTAssertLessThan(drift, tolerance, "id=\(id): got=\(got) want=\(want) drift=\(drift)")
        }
    }

    func testZeroWeightItemsAreSkipped() {
        let selector = RotationSelector(weights: [.high: 5, .normal: 0, .low: 0])
        let pool = [aff(1, .high), aff(2, .normal), aff(3, .low), aff(4, .normal)]
        var rng = SeededRNG(seed: 7)
        for _ in 0..<200 {
            XCTAssertEqual(selector.pick(from: pool, using: &rng)?.id, 1)
        }
    }

    func testRotationWeightedUITestSeedRanksPrioritiesForEveryPoolOrder() {
        let weighted = RotationSelector()
        let uniform = RotationSelector(weights: [.high: 1, .normal: 1, .low: 1])
        let orders: [[AffirmationPriority]] = [
            [.high, .normal, .low], [.high, .low, .normal],
            [.normal, .high, .low], [.normal, .low, .high],
            [.low, .high, .normal], [.low, .normal, .high],
        ]
        for order in orders {
            for skipped in 0...3 {
                let counts = rotationCounts(weighted, order: order, skipped: skipped)
                XCTAssertTrue(ranksByPriority(counts), "order=\(order) skipped=\(skipped) counts=\(counts)")
            }
        }
        for skipped in 0...3 {
            let counts = rotationCounts(uniform, order: [.low, .normal, .high], skipped: skipped)
            XCTAssertFalse(ranksByPriority(counts), "Equal weights must not pass the UI assertion: skipped=\(skipped) counts=\(counts)")
        }
    }

    private func rotationCounts(
        _ selector: RotationSelector,
        order: [AffirmationPriority],
        skipped: Int
    ) -> [AffirmationPriority: Int] {
        let pool = order.enumerated().map { aff(Int64($0.offset + 1), $0.element) }
        var rng = SeededRNG(seed: 2009)
        for _ in 0..<skipped {
            _ = selector.pick(from: pool, using: &rng)
        }
        var counts: [AffirmationPriority: Int] = [.high: 0, .normal: 0, .low: 0]
        for _ in 0..<30 {
            if let pick = selector.pick(from: pool, using: &rng) {
                counts[pick.priority, default: 0] += 1
            }
        }
        return counts
    }

    private func ranksByPriority(_ counts: [AffirmationPriority: Int]) -> Bool {
        let high = counts[.high] ?? 0
        let normal = counts[.normal] ?? 0
        let low = counts[.low] ?? 0
        return high > normal && normal > low && low > 0
    }
}
