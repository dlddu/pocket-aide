// 검증 시나리오: test-github-monitor.md#시나리오 15
import XCTest

final class WebhookHistoryExternalLinkUITests: XCTestCase {
    private let safari = XCUIApplication(bundleIdentifier: "com.apple.mobilesafari")

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    override func tearDown() {
        safari.terminate()
        super.tearDown()
    }

    private func link(_ label: String, of group: String, in app: XCUIApplication) -> XCUIElement {
        app.descendants(matching: .any).matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label ENDSWITH %@", WebhookHistoryUI.groupPrefix + group, label)
        ).firstMatch
    }

    private func openExternally(_ label: String, of group: String, in app: XCUIApplication) {
        let chip = link(label, of: group, in: app)
        XCTAssertTrue(chip.waitForExistence(timeout: 10), "The unacknowledged item should offer the \(label) link")
        WebhookHistoryUI.tap(chip, in: app)
        XCTAssertTrue(safari.wait(for: .runningForeground, timeout: 60), "Tapping \(label) should open the GitHub page in Safari")
        Thread.sleep(forTimeInterval: 3)
        app.activate()
        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 30), "The app should come back to the front")
    }

    private func assertStillUnacknowledged(_ group: String, unread: Int, in app: XCUIApplication) {
        XCTAssertTrue(
            WebhookHistoryUI.text("미확인 1", of: group, in: app).waitForExistence(timeout: 10),
            "The group should keep its one unacknowledged item"
        )
        let labels = WebhookHistoryUI.labels(of: group, in: app)
        XCTAssertFalse(labels.contains { $0.hasSuffix("· 확인됨") }, "The item should not read as acknowledged: \(labels)")
        XCTAssertEqual(
            WebhookHistoryUI.itemAcknowledgeButtons(of: group, in: app).count, 1,
            "The item should still offer 확인"
        )
        XCTAssertEqual(
            WebhookHistoryUI.acknowledgeAllButtons(of: group, in: app).count, 1,
            "The group should still offer 모두 확인"
        )
        XCTAssertEqual(WebhookHistoryUI.unreadCount(in: app), unread, "The unacknowledged badge should not move")
    }

    func testOpeningExternalLinksLeavesTheItemUnacknowledged() {
        let token = WebhookHistoryUI.runToken()
        let event = WebhookEvent(
            repo: "dlddu/e2e-link-\(token)",
            pullRequest: 1,
            commit: WebhookHistoryUI.commit(1, token: token),
            branch: "feature/1",
            workflow: "Build",
            conclusion: "success"
        )
        WebhookHistoryUI.send(event)
        let group = event.groupKey

        var app = WebhookHistoryUI.openMonitor(until: group, reads: "미확인 1")
        guard let before = WebhookHistoryUI.unreadCount(in: app) else {
            XCTFail("The header should show the unacknowledged badge")
            return
        }
        for label in ["PR", "런"] {
            openExternally(label, of: group, in: app)
            assertStillUnacknowledged(group, unread: before, in: app)
        }

        app.terminate()
        app = WebhookHistoryUI.openMonitor(until: group, reads: "미확인 1")
        assertStillUnacknowledged(group, unread: before, in: app)
        app.terminate()
    }
}
