// 검증 시나리오: test-affirmations.md#시나리오 5
import XCTest

final class RotationWeightedAffirmationUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHigherPrioritySentencesAreShownMoreOften() {
        let token = TodoUI.uniqueToken()
        let high = "자주 떠올린다 \(token)"
        let normal = "가끔 떠올린다 \(token)"
        let low = "드물게 떠올린다 \(token)"
        let sentences = [high, normal, low]
        let app = TodoUI.launch()
        let screen = AffirmationsScreen.open(in: app)

        screen.removeEverySentence()
        screen.add(high, priority: "high")
        screen.add(normal, priority: "normal")
        screen.add(low, priority: "low")

        var counts = Dictionary(uniqueKeysWithValues: sentences.map { ($0, 0) })
        XCTAssertTrue(screen.rotateButton.waitForExistence(timeout: 15), "The hero card should offer the rotate button")
        for turn in 1...AffirmationsScreen.rotationSamples {
            screen.rotateButton.tap()
            let shown = screen.heroSentence(among: sentences, turn: turn)
            counts[shown, default: 0] += 1
        }
        let highCount = counts[high] ?? 0
        let normalCount = counts[normal] ?? 0
        let lowCount = counts[low] ?? 0
        let summary = "high \(highCount) · normal \(normalCount) · low \(lowCount) of \(AffirmationsScreen.rotationSamples)"
        XCTContext.runActivity(named: "Hero exposure counts: \(summary)") { _ in }

        XCTAssertEqual(highCount + normalCount + lowCount, AffirmationsScreen.rotationSamples, "Every rotation should show one of the seeded sentences (\(summary))")
        XCTAssertGreaterThan(highCount, normalCount, "높음 should be shown more often than 보통 (\(summary))")
        XCTAssertGreaterThan(normalCount, lowCount, "보통 should be shown more often than 낮음 (\(summary))")
        XCTAssertGreaterThan(lowCount, 0, "낮음 should still be shown at least once (\(summary))")

        screen.remove(sentences)
    }
}

private extension AffirmationsScreen {
    static let rotationSamples = 300

    var anyRow: XCUIElement {
        app.descendants(matching: .any).matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND NOT (identifier ENDSWITH %@)", "affirmations.row.", ".delete")
        ).firstMatch
    }

    func removeEverySentence(limit: Int = 60) {
        for _ in 0..<limit {
            let target = anyRow
            guard target.waitForExistence(timeout: 5) else { return }
            let identifier = target.identifier
            reveal(target)
            target.swipeLeft()
            let delete = app.buttons.matching(
                NSPredicate(format: "label == %@ OR identifier ENDSWITH %@", "삭제", ".delete")
            ).firstMatch
            XCTAssertTrue(delete.waitForExistence(timeout: 5), "Swiping a sentence left should reveal 삭제")
            delete.tap()
            let removed = app.descendants(matching: .any).matching(identifier: identifier).firstMatch
            XCTAssertTrue(TodoUI.waitToDisappear(removed), "The swiped sentence should leave the list")
        }
        XCTFail("More than \(limit) sentences were left over before seeding the three priorities")
    }

    func heroSentence(among sentences: [String], turn: Int) -> String {
        let hero = app.staticTexts.matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label IN %@", "affirmations.hero", sentences)
        ).firstMatch
        XCTAssertTrue(hero.waitForExistence(timeout: 10), "The hero card should show one of the seeded sentences on rotation \(turn)")
        return hero.label
    }
}
