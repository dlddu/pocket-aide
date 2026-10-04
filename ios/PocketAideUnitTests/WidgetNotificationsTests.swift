import XCTest
@testable import PocketAideAPI

final class WidgetNotificationsTests: XCTestCase {
    private func item(
        _ id: Int64,
        conclusion: String = "success",
        acknowledged: Bool = false,
        createdAt: Int64 = 1_000,
        prNumber: Int? = 7,
        prTitle: String? = "fix: thing",
        branch: String = "feature/x"
    ) -> NotificationHistoryItem {
        NotificationHistoryItem(
            id: id,
            repoFullName: "dlddu/pocket-aide",
            prNumber: prNumber,
            prTitle: prTitle,
            prURL: nil,
            commitURL: nil,
            runURL: nil,
            workflowName: "CI",
            headBranch: branch,
            headSHA: "abc",
            conclusion: conclusion,
            acknowledgedAt: acknowledged ? 2_000 : nil,
            createdAt: createdAt
        )
    }

    func testLatestUnacknowledgedFirstAndRestCounted() {
        let items = [
            item(1, createdAt: 100),
            item(2, conclusion: "failure", createdAt: 300),
            item(3, createdAt: 200),
        ]
        let summary = WidgetNotifications.summarize(items, settings: NotificationSettings())
        XCTAssertEqual(summary.latest?.id, 2)
        XCTAssertEqual(summary.moreCount, 2)
    }

    func testStartEventsAndAcknowledgedAreLeftOut() {
        let items = [
            item(1, conclusion: "queued", createdAt: 500),
            item(2, conclusion: "in_progress", createdAt: 400),
            item(3, acknowledged: true, createdAt: 300),
            item(4, createdAt: 200),
        ]
        let summary = WidgetNotifications.summarize(items, settings: NotificationSettings())
        XCTAssertEqual(summary.latest?.id, 4)
        XCTAssertEqual(summary.moreCount, 0)
    }

    func testFollowsNotificationSettings() {
        let items = [
            item(1, conclusion: "success", createdAt: 300),
            item(2, conclusion: "timed_out", createdAt: 200),
            item(3, conclusion: "cancelled", createdAt: 100),
        ]
        let successOnly = WidgetNotifications.summarize(items, settings: NotificationSettings(outcomes: .success))
        XCTAssertEqual(successOnly.latest?.id, 1)
        XCTAssertEqual(successOnly.moreCount, 0)

        let failureOnly = WidgetNotifications.summarize(items, settings: NotificationSettings(outcomes: .failure))
        XCTAssertEqual(failureOnly.latest?.id, 2)
        XCTAssertEqual(failureOnly.moreCount, 0)

        let off = WidgetNotifications.summarize(items, settings: NotificationSettings(enabled: false))
        XCTAssertNil(off.latest)
        XCTAssertEqual(off.moreCount, 0)
    }

    func testEmptyHistory() {
        let summary = WidgetNotifications.summarize([], settings: NotificationSettings())
        XCTAssertNil(summary.latest)
        XCTAssertEqual(summary.moreCount, 0)
    }

    func testPushTextMatchesBackendBanner() {
        let success = item(1, conclusion: "success")
        XCTAssertEqual(PushText.title(for: success), "CI 통과 — dlddu/pocket-aide #7")
        XCTAssertEqual(PushText.body(for: success), "fix: thing")

        let failureNoPR = item(2, conclusion: "failure", prNumber: nil, prTitle: nil, branch: "main")
        XCTAssertEqual(PushText.title(for: failureNoPR), "dlddu/pocket-aide — failure")
        XCTAssertEqual(PushText.body(for: failureNoPR), "CI on main")

        let cancelledNoTitle = item(3, conclusion: "cancelled", prTitle: nil)
        XCTAssertEqual(PushText.title(for: cancelledNoTitle), "CI cancelled — dlddu/pocket-aide #7")
        XCTAssertEqual(PushText.body(for: cancelledNoTitle), "CI on feature/x")
    }
}
