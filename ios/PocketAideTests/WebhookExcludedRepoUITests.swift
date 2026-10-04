// 검증 시나리오: test-github-monitor.md#시나리오 7
import XCTest

final class WebhookExcludedRepoUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func event(_ repo: String, _ pullRequest: Int, token: String) -> WebhookEvent {
        WebhookEvent(
            repo: repo,
            pullRequest: pullRequest,
            commit: WebhookHistoryUI.commit(pullRequest, token: token),
            branch: "feature/\(pullRequest)",
            workflow: "Build",
            conclusion: "success"
        )
    }

    private func openSheet(in app: XCUIApplication) {
        let open = app.buttons["prmonitor.excluded.button"]
        XCTAssertTrue(open.waitForExistence(timeout: 10), "PR 모니터 should offer 제외 레포 관리")
        open.tap()
        XCTAssertTrue(
            app.textFields["prmonitor.excluded.input"].waitForExistence(timeout: 10),
            "The excluded repositories sheet should open"
        )
    }

    private func closeSheet(in app: XCUIApplication) {
        app.buttons["완료"].tap()
        XCTAssertTrue(
            TodoUI.waitToDisappear(app.textFields["prmonitor.excluded.input"]),
            "완료 should close the excluded repositories sheet"
        )
    }

    private func submit(_ text: String, in app: XCUIApplication) {
        let input = app.textFields["prmonitor.excluded.input"]
        input.tap()
        input.typeText(text)
        app.buttons["prmonitor.excluded.add.button"].tap()
    }

    private func rows(_ repo: String, in app: XCUIApplication) -> XCUIElementQuery {
        app.staticTexts.matching(NSPredicate(format: "label == %@", repo))
    }

    private func rejection(_ status: String, in app: XCUIApplication) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", "HTTP \(status)")).firstMatch
    }

    private func exclude(_ repo: String, token: String) {
        let app = WebhookHistoryUI.openMonitor()
        openSheet(in: app)
        submit(repo, in: app)
        XCTAssertTrue(rows(repo, in: app).firstMatch.waitForExistence(timeout: 10), "\(repo) should be listed as excluded")

        let malformed = "e2e-excl-\(token)"
        submit(malformed, in: app)
        XCTAssertTrue(rejection("400", in: app).waitForExistence(timeout: 10), "A name that is not owner/name should be refused")
        XCTAssertEqual(rows(malformed, in: app).count, 0, "The refused name should not be listed")

        submit(repo, in: app)
        XCTAssertTrue(rejection("409", in: app).waitForExistence(timeout: 10), "An already excluded repository should be refused")
        XCTAssertEqual(rows(repo, in: app).count, 1, "\(repo) should be listed once")

        closeSheet(in: app)
        app.terminate()
    }

    private func include(_ repo: String) {
        let app = WebhookHistoryUI.openMonitor()
        openSheet(in: app)
        let row = rows(repo, in: app).firstMatch
        XCTAssertTrue(row.waitForExistence(timeout: 10), "\(repo) should still be listed as excluded")
        row.swipeLeft()
        let delete = app.buttons.matching(NSPredicate(format: "label == %@", "삭제")).firstMatch
        XCTAssertTrue(delete.waitForExistence(timeout: 5), "Swiping the repository left should reveal 삭제")
        delete.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(row), "\(repo) should leave the excluded list")
        closeSheet(in: app)
        app.terminate()
    }

    func testExcludedRepositoryRecordsNothingUntilItIsIncludedAgain() {
        let token = WebhookHistoryUI.runToken()
        let excluded = "dlddu/e2e-excl-a-\(token)"
        let other = "dlddu/e2e-excl-b-\(token)"
        exclude(excluded, token: token)

        let dropped = event(excluded, 1, token: token)
        let kept = event(other, 1, token: token)
        WebhookHistoryUI.send([dropped, kept])
        var app = WebhookHistoryUI.openMonitor(until: kept.groupKey, reads: "CI 1건")
        XCTAssertEqual(
            WebhookHistoryUI.groups(of: excluded, in: app), [],
            "A workflow finishing in the excluded repository should leave no history"
        )
        XCTAssertEqual(WebhookHistoryUI.groups(of: other, in: app), [kept.groupKey], "The other repository should be recorded")
        app.terminate()

        include(excluded)
        let restored = event(excluded, 2, token: token)
        WebhookHistoryUI.send(restored)
        app = WebhookHistoryUI.openMonitor(until: restored.groupKey, reads: "CI 1건")
        XCTAssertEqual(
            WebhookHistoryUI.groups(of: excluded, in: app), [restored.groupKey],
            "Only the workflow that finished after the repository was included again should be recorded"
        )
        XCTAssertEqual(WebhookHistoryUI.groups(of: other, in: app), [kept.groupKey], "The other repository should stay recorded")
        app.terminate()
    }
}
