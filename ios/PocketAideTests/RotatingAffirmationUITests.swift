// 검증 시나리오: test-affirmations.md#시나리오 6
import XCTest

final class RotatingAffirmationUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHeroRotatesAcrossReentriesAndOnTheRotateButton() {
        let token = TodoUI.uniqueToken()
        let sentences = ["오늘도 한 걸음 \(token)", "서두르지 않는다 \(token)", "끝까지 해 본다 \(token)"]
        let app = TodoUI.launch()
        var screen = AffirmationsScreen.open(in: app)

        screen.removeEverySentence()
        screen.assertEmptyState()
        screen = screen.reenter()
        screen.assertEmptyState()

        for sentence in sentences {
            screen.add(sentence)
        }
        XCTAssertFalse(app.staticTexts[AffirmationsScreen.emptyTitle].exists, "The empty state gives way to the hero card")

        var shown: Set<String> = [screen.settledHero(among: sentences)]
        for _ in 0..<AffirmationsScreen.rotationAttempts where shown.count < 2 {
            screen = screen.reenter()
            shown.insert(screen.settledHero(among: sentences))
        }
        XCTAssertGreaterThanOrEqual(shown.count, 2, "Re-entering the tab should surface at least two different sentences")

        let before = screen.settledHero(among: sentences)
        var after = before
        for _ in 0..<AffirmationsScreen.rotationAttempts where after == before {
            screen.rotateButton.tap()
            after = screen.shownHero(among: sentences)
        }
        XCTAssertNotEqual(after, before, "The rotate button should replace the hero sentence")

        screen.remove(sentences)
    }
}

private extension AffirmationsScreen {
    static let emptyTitle = "첫 다짐을 추가해 보세요"
    static let rotationAttempts = 12

    var anyRow: XCUIElement {
        app.descendants(matching: .any).matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND NOT (identifier ENDSWITH %@)", "affirmations.row.", ".delete")
        ).firstMatch
    }

    var heroCard: XCUIElement {
        app.descendants(matching: .any).matching(
            NSPredicate(format: "identifier BEGINSWITH %@", "affirmations.hero")
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
        XCTFail("More than \(limit) sentences were left over before the empty-state check")
    }

    func assertEmptyState() {
        XCTAssertTrue(
            app.staticTexts[Self.emptyTitle].waitForExistence(timeout: 10),
            "With no sentences the empty-state guide should replace the hero card"
        )
        XCTAssertFalse(heroCard.exists, "With no sentences there is no hero card")
        XCTAssertFalse(rotateButton.exists, "With no sentences there is no rotate button")
        XCTAssertFalse(anyRow.exists, "With no sentences the list is empty")
    }

    func shownHero(among sentences: [String]) -> String {
        XCTAssertTrue(rotateButton.waitForExistence(timeout: 15), "The hero card should be shown once sentences exist")
        let deadline = Date().addingTimeInterval(10)
        repeat {
            if let match = sentences.first(where: { hero($0).exists }) {
                return match
            }
            Thread.sleep(forTimeInterval: 0.3)
        } while Date() < deadline
        XCTFail("The hero card should show one of the seeded sentences")
        return ""
    }

    func settledHero(among sentences: [String]) -> String {
        var current = shownHero(among: sentences)
        for _ in 0..<5 {
            Thread.sleep(forTimeInterval: 1.5)
            let next = shownHero(among: sentences)
            if next == current {
                return current
            }
            current = next
        }
        return current
    }
}
