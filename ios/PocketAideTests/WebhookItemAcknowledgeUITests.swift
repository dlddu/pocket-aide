// 검증 시나리오: test-github-monitor.md#시나리오 14
import XCTest

private struct HistoryEntry {
    let id: Int64
    let acknowledgedAt: Int64?
}

final class WebhookItemAcknowledgeUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func event(_ repo: String, _ workflow: String, _ conclusion: String, token: String) -> WebhookEvent {
        WebhookEvent(
            repo: repo,
            pullRequest: 14,
            commit: WebhookHistoryUI.commit(14, token: token),
            branch: "feature/14",
            workflow: workflow,
            conclusion: conclusion
        )
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

    private func assertOneOfTwoAcknowledged(_ group: String, in app: XCUIApplication) {
        let labels = WebhookHistoryUI.labels(of: group, in: app)
        XCTAssertEqual(
            labels.filter { $0.hasSuffix("· 확인됨") }.count, 1,
            "Only the acknowledged item should read 확인됨: \(labels)"
        )
        XCTAssertEqual(
            WebhookHistoryUI.itemAcknowledgeButtons(of: group, in: app).count, 1,
            "The other item of the group should still offer 확인"
        )
        XCTAssertEqual(
            WebhookHistoryUI.acknowledgeAllButtons(of: group, in: app).count, 1,
            "The group should still offer 모두 확인 for its remaining item"
        )
    }

    private func entries(of repo: String, via api: BackendAPI) -> [HistoryEntry] {
        let reply = api.call("GET", "/api/notification-history?limit=200")
        XCTAssertEqual(reply.status, 200, "The history list should load: \(reply.text)")
        let items = reply.json?["items"] as? [[String: Any]] ?? []
        return items
            .filter { $0["repo_full_name"] as? String == repo }
            .compactMap { item in
                guard let id = (item["id"] as? NSNumber)?.int64Value else { return nil }
                return HistoryEntry(id: id, acknowledgedAt: (item["acknowledged_at"] as? NSNumber)?.int64Value)
            }
    }

    private func acknowledgeInApp(_ group: String) {
        let app = WebhookHistoryUI.openMonitor(until: group, reads: "미확인 2")
        XCTAssertTrue(WebhookHistoryUI.text("CI 2건", of: group, in: app).exists, "The group should hold both events")
        XCTAssertEqual(WebhookHistoryUI.itemAcknowledgeButtons(of: group, in: app).count, 2, "Both items should offer 확인")
        guard let before = WebhookHistoryUI.unreadCount(in: app) else {
            XCTFail("The header should show the unacknowledged badge")
            return
        }
        WebhookHistoryUI.tap(WebhookHistoryUI.itemAcknowledgeButtons(of: group, in: app).firstMatch, in: app)
        XCTAssertTrue(
            WebhookHistoryUI.text("미확인 1", of: group, in: app).waitForExistence(timeout: 10),
            "Acknowledging one item should take the group from 미확인 2 to 미확인 1"
        )
        XCTAssertEqual(waitForUnread(before - 1, in: app), before - 1, "The header badge should drop by one")
        assertOneOfTwoAcknowledged(group, in: app)
        Thread.sleep(forTimeInterval: 3)
        app.terminate()

        let relaunched = WebhookHistoryUI.openMonitor(until: group, reads: "미확인 1")
        XCTAssertEqual(WebhookHistoryUI.unreadCount(in: relaunched), before - 1, "The badge should hold across a relaunch")
        assertOneOfTwoAcknowledged(group, in: relaunched)
        relaunched.terminate()
    }

    func testItemAcknowledgeRecordsOnlyThatItemAndIsScopedAndIdempotent() throws {
        let token = WebhookHistoryUI.runToken()
        let repo = "dlddu/e2e-itemack-\(token)"
        let sent = [
            event(repo, "Build", "success", token: token),
            event(repo, "Lint", "failure", token: token)
        ]
        WebhookHistoryUI.send(sent)
        let group = sent[0].groupKey
        acknowledgeInApp(group)

        let owner = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get the signed-in user's token")
        let listed = entries(of: repo, via: owner)
        XCTAssertEqual(listed.count, 2, "The user should have both history items: \(listed)")
        let acknowledged = try XCTUnwrap(listed.first { $0.acknowledgedAt != nil }, "One item should be acknowledged")
        let pending = try XCTUnwrap(listed.first { $0.acknowledgedAt == nil }, "One item should stay unacknowledged")

        let other = try XCTUnwrap(
            BackendAPI.signIn(subject: "e2e-other-\(token)"),
            "The test runner should get a second user's token"
        )
        let foreign = other.call("POST", "/api/notification-history/\(pending.id)/ack")
        XCTAssertEqual(foreign.status, 404, "Acknowledging another user's item should be refused: \(foreign.text)")
        XCTAssertNil(
            entries(of: repo, via: owner).first { $0.id == pending.id }?.acknowledgedAt,
            "The refused acknowledgement should leave the owner's item unacknowledged"
        )

        Thread.sleep(forTimeInterval: 2)
        let again = owner.call("POST", "/api/notification-history/\(acknowledged.id)/ack")
        XCTAssertTrue((200..<300).contains(again.status), "Re-acknowledging should succeed: \(again.status) \(again.text)")
        XCTAssertEqual(
            entries(of: repo, via: owner).first { $0.id == acknowledged.id }?.acknowledgedAt,
            acknowledged.acknowledgedAt,
            "Re-acknowledging should keep the original acknowledgement time"
        )
    }
}
