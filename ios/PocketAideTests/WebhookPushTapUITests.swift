// 검증 시나리오: test-github-monitor.md#시나리오 8
import XCTest

final class WebhookPushTapUITests: XCTestCase {
    private let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func event(_ repo: String, _ number: Int, token: String) -> WebhookEvent {
        WebhookEvent(
            repo: repo,
            pullRequest: number,
            commit: WebhookHistoryUI.commit(number, token: token),
            branch: "e2e/push-\(number)",
            workflow: "Push tap",
            conclusion: "success"
        )
    }

    private func pushTitle(of event: WebhookEvent) -> String {
        "CI 통과 — \(event.repo) #\(event.pullRequest ?? 0)"
    }

    private func monitorTab(in app: XCUIApplication) -> XCUIElement {
        app.tabBars.firstMatch.buttons["PR 모니터"]
    }

    private func arrival(in app: XCUIApplication) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@", "방금 진입")).firstMatch
    }

    private func allowNotificationsIfAsked() {
        let allow = springboard.alerts.buttons
            .matching(NSPredicate(format: "label IN %@", ["Allow", "허용"])).firstMatch
        if allow.waitForExistence(timeout: 5) {
            allow.tap()
        }
    }

    private func leaveMonitorTab(in app: XCUIApplication) {
        let other = app.tabBars.firstMatch.buttons["다짐"]
        XCTAssertTrue(other.waitForExistence(timeout: 10), "다짐 should be a direct tab")
        other.tap()
        XCTAssertFalse(monitorTab(in: app).isSelected, "PR 모니터 should be left before the next push")
    }

    private func tapBanner(for event: WebhookEvent) {
        let banner = springboard.descendants(matching: .any)
            .matching(NSPredicate(format: "label CONTAINS %@", pushTitle(of: event))).firstMatch
        XCTAssertTrue(banner.waitForExistence(timeout: 90), "The push for \(event.repo) should be presented")
        banner.tap()
    }

    private func assertOpenedMonitor(_ app: XCUIApplication) {
        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 15), "Tapping the push should bring the app forward")
        let selected = expectation(for: NSPredicate(format: "isSelected == true"), evaluatedWith: monitorTab(in: app))
        wait(for: [selected], timeout: 10)
    }

    private func assertHighlighted(_ event: WebhookEvent, in app: XCUIApplication) {
        XCTAssertTrue(
            WebhookHistoryUI.text("방금 진입", of: event.groupKey, in: app).waitForExistence(timeout: 5),
            "The pushed event's group card should be highlighted"
        )
    }

    private func assertStillUnread(_ event: WebhookEvent, unread expected: Int, in app: XCUIApplication) {
        XCTAssertEqual(
            WebhookHistoryUI.itemAcknowledgeButtons(of: event.groupKey, in: app).count, 1,
            "A push tap must leave the item unacknowledged"
        )
        var value = WebhookHistoryUI.unreadCount(in: app)
        var polls = 0
        while value != expected && polls < 15 {
            Thread.sleep(forTimeInterval: 1)
            value = WebhookHistoryUI.unreadCount(in: app)
            polls += 1
        }
        XCTAssertEqual(value, expected, "A push tap must not lower the unacknowledged badge")
    }

    private func waitForHighlightToClear(in app: XCUIApplication) {
        let cleared = expectation(for: NSPredicate(format: "exists == false"), evaluatedWith: arrival(in: app))
        wait(for: [cleared], timeout: 15)
    }

    func testPushTapHighlightsItemWithoutAcknowledging() {
        let token = WebhookHistoryUI.runToken()
        let closed = event("dlddu/push-e2e-\(token)", 81, token: token)
        let foreground = event("dlddu/push-e2e-\(token)", 82, token: token)
        let plain = event("dlddu/push-e2e-noid-\(token)", 83, token: token)

        let app = XCUIApplication()
        app.launch()
        allowNotificationsIfAsked()
        XCTAssertTrue(monitorTab(in: app).waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        monitorTab(in: app).tap()
        let before = WebhookHistoryUI.unreadCount(in: app) ?? 0

        app.terminate()
        WebhookHistoryUI.send(closed)
        tapBanner(for: closed)
        assertOpenedMonitor(app)
        assertHighlighted(closed, in: app)
        assertStillUnread(closed, unread: before + 1, in: app)
        waitForHighlightToClear(in: app)

        leaveMonitorTab(in: app)
        WebhookHistoryUI.send(foreground)
        tapBanner(for: foreground)
        assertOpenedMonitor(app)
        assertHighlighted(foreground, in: app)
        assertStillUnread(foreground, unread: before + 2, in: app)
        waitForHighlightToClear(in: app)

        leaveMonitorTab(in: app)
        WebhookHistoryUI.send(plain)
        tapBanner(for: plain)
        assertOpenedMonitor(app)
        Thread.sleep(forTimeInterval: 2)
        XCTAssertFalse(arrival(in: app).exists, "A push without an event id should open the tab without a highlight")
        app.terminate()
    }
}
