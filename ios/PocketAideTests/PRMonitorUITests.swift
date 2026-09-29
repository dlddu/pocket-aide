// 검증 시나리오: 없음 (스모크/인프라)
// 등재: docs/product/doc-tracker/ 최신 월 파일 「## e2e 매핑」 → 「비-시나리오(스모크·인프라) 등재」.
import XCTest

/// End-to-end coverage for the PR-monitor pipeline on the real path:
/// GitHub workflow_run envelope → SQS → backend consumer → notification
/// history → history API → PR 모니터 screen → 「확인」 acknowledge.
///
/// Pre-conditions assumed by the test environment (ios-test workflow):
///   - backend consumer long-polls the local SQS queue (start-test-sqs)
///   - the fixture `.github/fixtures/github-webhook/workflow_run.completed.json`
///     is enqueued once the first user row exists (i.e. after sign-in)
final class PRMonitorUITests: XCTestCase {
    /// Title line the fixture produces: GitHub's workflow_run.pull_requests[]
    /// carries no PR title, so it reads "<repo> · #<number>". Matched by
    /// prefix because the group header formats the number with grouping
    /// ("#9,001") while the row does not ("#9001").
    private let fixtureTitlePrefix = "dlddu/pocket-aide-e2e · #9"

    private func fixtureTitle(in app: XCUIApplication) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", fixtureTitlePrefix)).firstMatch
    }

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func openPRMonitor() -> XCUIApplication {
        let app = XCUIApplication()
        app.launch()
        let tabsBar = app.tabBars.firstMatch
        XCTAssertTrue(tabsBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = tabsBar.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 5), "PR 모니터 should be a direct tab")
        tab.tap()
        return app
    }

    // Elements are matched by label: the CI database is fresh, so the fixture
    // is the only history event and its labels are unique on screen.
    func testWebhookEventReachesHistoryAndCanBeAcknowledged() {
        var app = openPRMonitor()

        // The event is enqueued after the first sign-in and consumed within
        // seconds; relaunch to re-fetch until the row shows up.
        var attempts = 0
        while !fixtureTitle(in: app).waitForExistence(timeout: 10) && attempts < 6 {
            attempts += 1
            app.terminate()
            app = openPRMonitor()
        }
        XCTAssertTrue(
            fixtureTitle(in: app).exists,
            "History row for the webhook fixture should appear on the PR 모니터 screen"
        )

        // An acknowledged group moves under the 「확인 완료」 section header
        // and collapses to its card header, so the section header is the
        // observable acknowledged state. -test-iterations 2 may re-run this
        // test after the row was already acknowledged.
        let acknowledgedSection = app.staticTexts["확인 완료"].firstMatch
        if !acknowledgedSection.exists {
            XCTAssertTrue(
                app.staticTexts["CI 통과"].firstMatch.exists,
                "Fixture conclusion=success should render as CI 통과 on the unacknowledged row"
            )
            let ack = app.buttons["확인"].firstMatch
            XCTAssertTrue(ack.waitForExistence(timeout: 5), "Unacknowledged row should offer the 확인 button")
            ack.tap()
        }
        XCTAssertTrue(
            acknowledgedSection.waitForExistence(timeout: 10),
            "Acknowledged group should move under 확인 완료"
        )

        // Acknowledgement is server-side: a fresh launch re-fetches history.
        app.terminate()
        app = openPRMonitor()
        XCTAssertTrue(
            fixtureTitle(in: app).waitForExistence(timeout: 15),
            "Row should still be listed after relaunch"
        )
        XCTAssertTrue(
            app.staticTexts["확인 완료"].firstMatch.exists,
            "Acknowledgement should persist via the history API"
        )
        XCTAssertFalse(
            app.staticTexts["CI 통과"].firstMatch.exists,
            "No unacknowledged row should remain after the persisted acknowledgement"
        )
    }
}
