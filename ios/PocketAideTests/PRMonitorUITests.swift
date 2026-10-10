// 검증 시나리오: 없음 (스모크/인프라)
import XCTest

/// Needs the ios-test workflow environment: it enqueues
/// `.github/fixtures/github-webhook/workflow_run.completed.json` after the first
/// sign-in and pushes that event to the simulator once the push test logs that
/// it is waiting on the home screen ("Deliver PR-monitor push").
final class PRMonitorUITests: XCTestCase {
    /// Title line the fixture produces: GitHub's workflow_run.pull_requests[]
    /// carries no PR title, so it reads "<repo> · #<number>". Matched by
    /// prefix because the group header formats the number with grouping
    /// ("#9,001") while the row does not ("#9001").
    private let fixtureTitlePrefix = "dlddu/pocket-aide-e2e · #9"

    /// The fixture's push title as the backend formats it.
    private let pushTitle = "CI 통과 — dlddu/pocket-aide-e2e #9001"

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

    /// PRD-10 AC7 · AC12 on the real receive path. Runs before the
    /// acknowledge test (alphabetical order), so the row is still unacknowledged
    /// unless a retry follows it.
    func testPushTapOpensPRMonitorWithoutAcknowledging() {
        let app = XCUIApplication()
        app.launch()
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")

        // An earlier test in the shard may already have answered the permission
        // prompt; banners need it granted.
        let allow = springboard.alerts.buttons
            .matching(NSPredicate(format: "label IN %@", ["Allow", "허용"])).firstMatch
        if allow.waitForExistence(timeout: 10) {
            allow.tap()
        }
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        XCTAssertFalse(tab.isSelected, "Launch should land on a different tab than PR 모니터")

        XCUIDevice.shared.press(.home)
        // The workflow pushes only after this line; the UUID tells a retry apart.
        NSLog("pocketaide-e2e push-ready %@", UUID().uuidString)
        let banner = springboard.descendants(matching: .any)
            .matching(NSPredicate(format: "label CONTAINS %@", pushTitle)).firstMatch
        XCTAssertTrue(banner.waitForExistence(timeout: 120), "The pushed notification should be presented")
        banner.tap()

        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 15), "Tapping the push should open the app")
        let selected = expectation(for: NSPredicate(format: "isSelected == true"), evaluatedWith: tab)
        wait(for: [selected], timeout: 10)
        XCTAssertTrue(
            fixtureTitle(in: app).waitForExistence(timeout: 15),
            "The pushed event's row should be listed"
        )
        if !app.staticTexts["확인 완료"].firstMatch.exists {
            XCTAssertTrue(
                app.buttons["확인"].firstMatch.exists,
                "A push tap must not acknowledge the row"
            )
        }
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

    private enum GitHubToken {
        static let valid = "ghp_e2e_valid"
        static let revoked = "ghp_e2e_revoked"
        static let expires = "ghp_e2e_expires"
        static let rateLimited = "ghp_e2e_ratelimited"
    }

    private func openPullRequestsSheet() -> XCUIApplication {
        let app = XCUIApplication()
        // mock-exception: EXT — 실 GitHub 은 전용 테스트 계정 PAT 가 CI 시크릿에 없어 E2E 가 부를 수 없다; 앱의 GitHubClient 를 로컬 GitHub API 스텁으로 향하게 한다 (docs/e2e-mocking-policy.md)
        app.launchEnvironment["GITHUB_API_BASE_URL"] = "http://localhost:5557"
        app.launch()
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        let open = app.buttons["prmonitor.openprs.button"]
        XCTAssertTrue(open.waitForExistence(timeout: 10), "PR 모니터 header should offer the open-PR sheet")
        open.tap()
        XCTAssertTrue(app.navigationBars["열린 PR"].waitForExistence(timeout: 10), "The open-PR sheet should be presented")
        let disconnect = app.buttons["openprs.disconnect.button"]
        if disconnect.waitForExistence(timeout: 3) {
            disconnect.tap()
        }
        XCTAssertTrue(
            app.secureTextFields["openprs.token.field"].waitForExistence(timeout: 10),
            "A disconnected sheet should show the token form"
        )
        return app
    }

    private func submitToken(_ token: String, in app: XCUIApplication) {
        let field = app.secureTextFields["openprs.token.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 10), "Token field should be shown")
        field.tap()
        field.typeText(token)
        app.buttons["openprs.connect.button"].tap()
    }

    private func element(_ identifier: String, in app: XCUIApplication) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: identifier).firstMatch
    }

    private func assertRow(_ number: Int, contains parts: [String], in app: XCUIApplication) {
        let row = element("openprs.row.dlddu/pocket-aide-e2e#\(number)", in: app)
        XCTAssertTrue(row.waitForExistence(timeout: 15), "PR #\(number) should be listed")
        for part in parts {
            XCTAssertTrue(row.label.contains(part), "PR #\(number) row should read \(part): \(row.label)")
        }
    }

    func testOpenPullRequestsConnectListsRolesCIStatusAndFilters() {
        var app = openPullRequestsSheet()

        submitToken(GitHubToken.revoked, in: app)
        let connectError = element("openprs.connect.error", in: app)
        XCTAssertTrue(connectError.waitForExistence(timeout: 15), "A rejected PAT should not connect")
        XCTAssertTrue(connectError.label.contains("토큰을 거부"), "Connect error should name the rejected token: \(connectError.label)")
        XCTAssertFalse(app.buttons["openprs.disconnect.button"].exists, "A rejected PAT must not be stored")

        app.terminate()
        app = openPullRequestsSheet()
        submitToken(GitHubToken.valid, in: app)
        let login = element("openprs.account.login", in: app)
        XCTAssertTrue(login.waitForExistence(timeout: 15), "A valid PAT should connect")
        XCTAssertEqual(login.label, "@pocket-aide-e2e")

        let all = app.buttons["filter.pill.all"]
        if all.waitForExistence(timeout: 5) && !all.isSelected {
            all.tap()
        }
        assertRow(11, contains: ["작성자", "성공"], in: app)
        assertRow(14, contains: ["작성자", "상태 없음"], in: app)
        assertRow(12, contains: ["리뷰어", "실패", "octocat"], in: app)
        assertRow(13, contains: ["리뷰어", "진행 중", "hubot"], in: app)

        let access = element("openprs.banner.access", in: app)
        XCTAssertTrue(access.waitForExistence(timeout: 5), "The SAML-hidden PR should raise the access banner")
        XCTAssertTrue(
            app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", "PR 1개는")).firstMatch.exists,
            "The access banner should count the one hidden PR"
        )

        app.buttons["filter.pill.ciFailing"].tap()
        XCTAssertTrue(element("openprs.row.dlddu/pocket-aide-e2e#12", in: app).waitForExistence(timeout: 5), "The failing PR should stay")
        XCTAssertFalse(element("openprs.row.dlddu/pocket-aide-e2e#11", in: app).exists, "Passing PRs should be filtered out")
        XCTAssertFalse(element("openprs.row.dlddu/pocket-aide-e2e#13", in: app).exists, "Pending PRs should be filtered out")
        app.buttons["filter.pill.mine"].tap()
        XCTAssertTrue(element("openprs.row.dlddu/pocket-aide-e2e#11", in: app).waitForExistence(timeout: 5), "Authored PRs should stay")
        XCTAssertFalse(element("openprs.row.dlddu/pocket-aide-e2e#12", in: app).exists, "Review requests should be filtered out")
        app.buttons["filter.pill.all"].tap()

        app.buttons["openprs.disconnect.button"].tap()
        XCTAssertTrue(app.secureTextFields["openprs.token.field"].waitForExistence(timeout: 10), "Disconnect should return to the token form")
    }

    func testOpenPullRequestsBannersForRevokedAndRateLimitedTokens() {
        let app = openPullRequestsSheet()

        submitToken(GitHubToken.expires, in: app)
        let tokenBanner = element("openprs.banner.token", in: app)
        XCTAssertTrue(tokenBanner.waitForExistence(timeout: 15), "A token GitHub rejects on refresh should raise the token banner")
        let action = app.buttons["openprs.banner.action"]
        XCTAssertEqual(action.label, "토큰 다시 연결")
        action.tap()

        submitToken(GitHubToken.rateLimited, in: app)
        XCTAssertTrue(
            element("openprs.banner.ratelimit", in: app).waitForExistence(timeout: 15),
            "An exhausted rate limit should raise the rate-limit banner"
        )
        XCTAssertEqual(app.buttons["openprs.banner.action"].label, "다시 시도")

        app.buttons["openprs.disconnect.button"].tap()
        XCTAssertTrue(app.secureTextFields["openprs.token.field"].waitForExistence(timeout: 10), "Disconnect should return to the token form")
    }
}
