// 검증 시나리오: test-todo.md#시나리오 1
import XCTest

final class TodoAreaSeparationUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testItemAddedToOneAreaNeverAppearsInTheOther() {
        let token = TodoUI.uniqueToken()
        let personalTitle = "치과 예약 잡기 \(token)"
        let workTitle = "분기 리뷰 자료 \(token)"
        let app = TodoUI.launch()

        let personal = TodoScreen.open(.personal, in: app)
        guard let personalBefore = personal.counts() else { return XCTFail("Personal summary should parse") }
        personal.add(title: personalTitle)
        let personalAfter = TodoCounts(open: personalBefore.open + 1, done: personalBefore.done)
        XCTAssertTrue(personal.waitForCounts(personalAfter), "Personal summary should count the new personal item")

        let work = TodoScreen.open(.work, in: app)
        guard let workBefore = work.counts() else { return XCTFail("Work summary should parse") }
        work.add(title: workTitle)
        let workAfter = TodoCounts(open: workBefore.open + 1, done: workBefore.done)
        XCTAssertTrue(work.waitForCounts(workAfter), "Work summary should count the new work item")

        assertSeparated(app: app, token: token, personalTitle: personalTitle, workTitle: workTitle,
                        personalCounts: personalAfter, workCounts: workAfter, refresh: true)

        TodoUI.relaunch(app)
        assertSeparated(app: app, token: token, personalTitle: personalTitle, workTitle: workTitle,
                        personalCounts: personalAfter, workCounts: workAfter, refresh: false)
    }

    private func assertSeparated(
        app: XCUIApplication,
        token: String,
        personalTitle: String,
        workTitle: String,
        personalCounts: TodoCounts,
        workCounts: TodoCounts,
        refresh: Bool
    ) {
        for area in TodoUIArea.allCases {
            let screen = TodoScreen.open(area, in: app)
            if refresh {
                screen.pullToRefresh()
            }
            let mine = area == .personal ? personalTitle : workTitle
            let theirs = area == .personal ? workTitle : personalTitle
            screen.assertListing(token: token, present: [mine], absent: [theirs])
            let expected = area == .personal ? personalCounts : workCounts
            XCTAssertTrue(screen.waitForCounts(expected), "The \(area.rawValue) summary should count only its own items")
        }
    }
}
