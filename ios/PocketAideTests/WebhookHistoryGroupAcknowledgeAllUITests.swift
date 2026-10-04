// 검증 시나리오: test-github-monitor.md#시나리오 17
import XCTest

final class WebhookHistoryGroupAcknowledgeAllUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func event(_ repo: String, _ pullRequest: Int, _ workflow: String, _ conclusion: String, token: String) -> WebhookEvent {
        WebhookEvent(
            repo: repo,
            pullRequest: pullRequest,
            commit: WebhookHistoryUI.commit(pullRequest, token: token),
            branch: "feature/\(pullRequest)",
            workflow: workflow,
            conclusion: conclusion
        )
    }

    private func waitToDisappear(_ element: XCUIElement, timeout: TimeInterval = 10) -> Bool {
        let gone = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: element)
        return XCTWaiter().wait(for: [gone], timeout: timeout) == .completed
    }

    private func waitForUnread(_ expected: Int, in app: XCUIApplication) -> Int? {
        var value = WebhookHistoryUI.unreadCount(in: app)
        var polls = 0
        while value != expected && polls < 15 {
            Thread.sleep(forTimeInterval: 1)
            value = WebhookHistoryUI.unreadCount(in: app)
            polls += 1
        }
        return value
    }

    private func assertFullyAcknowledged(_ group: String, in app: XCUIApplication) {
        XCTAssertTrue(WebhookHistoryUI.reveal(group, in: app), "Group \(group) should stay listed once acknowledged")
        XCTAssertTrue(
            WebhookHistoryUI.text("모두 확인", of: group, in: app).waitForExistence(timeout: 10),
            "Group \(group) should read 모두 확인"
        )
        let labels = WebhookHistoryUI.labels(of: group, in: app)
        XCTAssertFalse(labels.contains { $0.hasPrefix("미확인") }, "Group \(group) should have no unacknowledged count: \(labels)")
        XCTAssertEqual(
            WebhookHistoryUI.acknowledgeAllButtons(of: group, in: app).count, 0,
            "A fully acknowledged group should not offer 모두 확인"
        )
        XCTAssertEqual(
            WebhookHistoryUI.itemAcknowledgeButtons(of: group, in: app).count, 0,
            "A fully acknowledged group should have no item left to acknowledge"
        )
    }

    private func arrange(groupA: String, groupC: String) {
        let app = WebhookHistoryUI.openMonitor(until: groupA, reads: "CI 3건")
        WebhookHistoryUI.tap(WebhookHistoryUI.itemAcknowledgeButtons(of: groupA, in: app).firstMatch, in: app)
        XCTAssertTrue(
            WebhookHistoryUI.text("미확인 2", of: groupA, in: app).waitForExistence(timeout: 10),
            "Acknowledging one item should leave group A with two unacknowledged"
        )
        XCTAssertTrue(WebhookHistoryUI.reveal(groupC, in: app), "Group C should be listed")
        let last = WebhookHistoryUI.itemAcknowledgeButtons(of: groupC, in: app).firstMatch
        WebhookHistoryUI.tap(last, in: app)
        XCTAssertTrue(waitToDisappear(last), "Acknowledging group C's only item should leave nothing to acknowledge in it")
        Thread.sleep(forTimeInterval: 3)
        app.terminate()
    }

    func testAcknowledgeAllClearsOnlyTheUnacknowledgedItemsOfThatGroup() {
        let token = WebhookHistoryUI.runToken()
        let repo = "dlddu/e2e-ackall-\(token)"
        let sent = [
            event(repo, 3, "Build", "success", token: token),
            event(repo, 2, "Build", "success", token: token),
            event(repo, 1, "Build", "success", token: token),
            event(repo, 1, "Lint", "success", token: token),
            event(repo, 1, "Deploy", "failure", token: token)
        ]
        WebhookHistoryUI.send(sent)
        let groupC = sent[0].groupKey
        let groupB = sent[1].groupKey
        let groupA = sent[2].groupKey
        arrange(groupA: groupA, groupC: groupC)

        var app = WebhookHistoryUI.openMonitor(until: groupA, reads: "미확인 2")
        let acknowledged = WebhookHistoryUI.labels(of: groupA, in: app).filter { $0.hasSuffix("· 확인됨") }
        XCTAssertEqual(acknowledged.count, 1, "Group A should start with one acknowledged item: \(acknowledged)")
        XCTAssertEqual(WebhookHistoryUI.acknowledgeAllButtons(of: groupA, in: app).count, 1, "Group A should offer 모두 확인")
        XCTAssertTrue(WebhookHistoryUI.reveal(groupB, in: app), "Group B should be listed")
        XCTAssertTrue(WebhookHistoryUI.text("미확인 1", of: groupB, in: app).exists, "Group B should have one unacknowledged item")
        XCTAssertEqual(WebhookHistoryUI.acknowledgeAllButtons(of: groupB, in: app).count, 1, "Group B should offer 모두 확인")
        assertFullyAcknowledged(groupC, in: app)

        app.terminate()
        app = WebhookHistoryUI.openMonitor(until: groupA, reads: "미확인 2")
        guard let before = WebhookHistoryUI.unreadCount(in: app) else {
            XCTFail("The header should show the unacknowledged badge")
            return
        }
        WebhookHistoryUI.tap(WebhookHistoryUI.acknowledgeAllButtons(of: groupA, in: app).firstMatch, in: app)
        XCTAssertEqual(waitForUnread(before - 2, in: app), before - 2, "모두 확인 should take two off the unacknowledged badge")
        XCTAssertTrue(WebhookHistoryUI.reveal(groupB, in: app), "Group B should stay listed")
        XCTAssertTrue(WebhookHistoryUI.text("미확인 1", of: groupB, in: app).exists, "Group B should keep its unacknowledged item")
        XCTAssertEqual(WebhookHistoryUI.acknowledgeAllButtons(of: groupB, in: app).count, 1, "Group B should still offer 모두 확인")
        assertFullyAcknowledged(groupA, in: app)

        app.terminate()
        app = WebhookHistoryUI.openMonitor(until: groupB, reads: "미확인 1")
        XCTAssertEqual(WebhookHistoryUI.unreadCount(in: app), before - 2, "The acknowledgements should persist across a relaunch")
        assertFullyAcknowledged(groupA, in: app)
        app.terminate()
    }
}
