// 검증 시나리오: test-affirmations.md#시나리오 5
import XCTest

final class RotationWeightedAffirmationUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHigherPrioritySentencesAreShownMoreOften() throws {
        let token = TodoUI.uniqueToken()
        let high = "자주 떠올린다 \(token)"
        let normal = "가끔 떠올린다 \(token)"
        let low = "드물게 떠올린다 \(token)"
        let sentences = [high, normal, low]
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        api.removeAllAffirmations()
        api.addAffirmation(high, priority: "high")
        api.addAffirmation(normal, priority: "normal")
        api.addAffirmation(low, priority: "low")
        addTeardownBlock { api.deleteAffirmations(sentences) }

        let app = XCUIApplication()
        // mock-exception: DET — 히어로 회전은 시스템 난수로 문장을 고르고 XCUITest 는 그 난수를 고정할 수 없다; 앱이 회전 난수에 쓸 시드 값만 주입한다 (docs/e2e-mocking-policy.md)
        app.launchEnvironment["ROTATION_SEED"] = AffirmationsScreen.rotationSeed
        app.launch()
        let screen = AffirmationsScreen.open(in: app)

        var counts = Dictionary(uniqueKeysWithValues: sentences.map { ($0, 0) })
        XCTAssertTrue(screen.rotateButton.waitForExistence(timeout: 15), "The hero card should offer the rotate button")
        _ = screen.heroSentence(among: sentences, turn: 0)
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
    }
}

private extension AffirmationsScreen {
    static let rotationSeed = "2009"
    static let rotationSamples = 30

    func heroSentence(among sentences: [String], turn: Int) -> String {
        let hero = app.staticTexts.matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label IN %@", "affirmations.hero", sentences)
        ).firstMatch
        XCTAssertTrue(hero.waitForExistence(timeout: 10), "The hero card should show one of the seeded sentences on rotation \(turn)")
        return hero.label
    }
}
