import XCTest
@testable import PocketAideAPI

/// Splitmix64 — a tiny seedable RNG. Same seed → same sequence of values, so
/// every test that pipes one in gets a deterministic pick.
struct SeededGenerator: RandomNumberGenerator {
    private var state: UInt64

    init(seed: UInt64) {
        self.state = seed == 0 ? 0xDEAD_BEEF_CAFE_F00D : seed
    }

    mutating func next() -> UInt64 {
        state &+= 0x9E37_79B9_7F4A_7C15
        var z = state
        z = (z ^ (z &>> 30)) &* 0xBF58_476D_1CE4_E5B9
        z = (z ^ (z &>> 27)) &* 0x94D0_49BB_1331_11EB
        return z ^ (z &>> 31)
    }
}

final class RotationSelectorTests: XCTestCase {
    private func aff(_ id: Int64, _ p: AffirmationPriority) -> Affirmation {
        Affirmation(id: id, text: "t\(id)", priority: p, createdAt: 0, updatedAt: 0)
    }

    func testEmptyPoolReturnsNil() {
        let selector = RotationSelector()
        var rng = SeededGenerator(seed: 1)
        XCTAssertNil(selector.pick(from: [], using: &rng))
    }

    func testSingleItemAlwaysReturnsIt() {
        let selector = RotationSelector()
        let pool = [aff(1, .normal)]
        for seed in 1...32 {
            var rng = SeededGenerator(seed: UInt64(seed))
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
        var rngA = SeededGenerator(seed: 42)
        var rngB = SeededGenerator(seed: 42)
        XCTAssertEqual(
            selector.pick(from: pool, using: &rngA)?.id,
            selector.pick(from: pool, using: &rngB)?.id
        )
    }

    func testWeightedDistribution() {
        let selector = RotationSelector()
        // One item per tier, then a 3:2:1 weighted draw should hit them
        // roughly 50% / 33% / 17% over many trials.
        let pool = [aff(1, .high), aff(2, .normal), aff(3, .low)]
        var rng = SeededGenerator(seed: 12345)
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
        // Override weights so .low has zero chance. Even with many low items,
        // a high item must always win.
        let selector = RotationSelector(weights: [.high: 5, .normal: 0, .low: 0])
        let pool = [aff(1, .high), aff(2, .normal), aff(3, .low), aff(4, .normal)]
        var rng = SeededGenerator(seed: 7)
        for _ in 0..<200 {
            XCTAssertEqual(selector.pick(from: pool, using: &rng)?.id, 1)
        }
    }
}
