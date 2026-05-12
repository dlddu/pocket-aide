import XCTest
@testable import PocketAideAPI

final class RotationSelectorTests: XCTestCase {
    private func makeAffirmation(id: Int64, priority: Priority) -> Affirmation {
        Affirmation(id: id, text: "a-\(id)", priority: priority, createdAt: 0, updatedAt: 0)
    }

    func testEmptyArrayReturnsNil() {
        var rng = SeededGenerator(seed: 1)
        XCTAssertNil(RotationSelector().pick(from: [], using: &rng))
    }

    func testSingleItemAlwaysReturned() {
        let item = makeAffirmation(id: 1, priority: .low)
        var rng = SeededGenerator(seed: 1)
        for _ in 0..<32 {
            XCTAssertEqual(RotationSelector().pick(from: [item], using: &rng)?.id, item.id)
        }
    }

    func testSameSeedProducesSameSequence() {
        let items: [Affirmation] = [
            makeAffirmation(id: 1, priority: .high),
            makeAffirmation(id: 2, priority: .normal),
            makeAffirmation(id: 3, priority: .low),
        ]
        let selector = RotationSelector()

        var rngA = SeededGenerator(seed: 42)
        var rngB = SeededGenerator(seed: 42)

        for _ in 0..<16 {
            let pickA = selector.pick(from: items, using: &rngA)
            let pickB = selector.pick(from: items, using: &rngB)
            XCTAssertEqual(pickA?.id, pickB?.id)
        }
    }

    func testWeightDistribution() {
        // With weights 3:2:1, over many trials each priority should approximate
        // 50% / 33% / 17%. We allow ±5 percentage points.
        let items: [Affirmation] = [
            makeAffirmation(id: 1, priority: .high),
            makeAffirmation(id: 2, priority: .normal),
            makeAffirmation(id: 3, priority: .low),
        ]
        let selector = RotationSelector()
        var rng = SeededGenerator(seed: 1234)
        let trials = 6000

        var counts: [Priority: Int] = [.high: 0, .normal: 0, .low: 0]
        for _ in 0..<trials {
            guard let picked = selector.pick(from: items, using: &rng) else {
                XCTFail("unexpected nil pick")
                return
            }
            counts[picked.priority, default: 0] += 1
        }

        let highRatio = Double(counts[.high]!) / Double(trials)
        let normalRatio = Double(counts[.normal]!) / Double(trials)
        let lowRatio = Double(counts[.low]!) / Double(trials)

        XCTAssertEqual(highRatio, 3.0 / 6.0, accuracy: 0.05)
        XCTAssertEqual(normalRatio, 2.0 / 6.0, accuracy: 0.05)
        XCTAssertEqual(lowRatio, 1.0 / 6.0, accuracy: 0.05)
    }
}

struct SeededGenerator: RandomNumberGenerator {
    private var state: UInt64

    init(seed: UInt64) {
        self.state = seed == 0 ? 0x9E3779B97F4A7C15 : seed
    }

    mutating func next() -> UInt64 {
        // SplitMix64 — deterministic, well-distributed.
        state &+= 0x9E3779B97F4A7C15
        var z = state
        z = (z ^ (z &>> 30)) &* 0xBF58476D1CE4E5B9
        z = (z ^ (z &>> 27)) &* 0x94D049BB133111EB
        return z ^ (z &>> 31)
    }
}
