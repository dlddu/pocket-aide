// 검증 시나리오: test-github-monitor.md#시나리오 12
import XCTest

final class WebhookUserScopedHistoryUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func event(_ repo: String, pullRequest: Int?, seed: Int, workflow: String, token: String) -> WebhookEvent {
        WebhookEvent(
            repo: repo,
            pullRequest: pullRequest,
            commit: WebhookHistoryUI.commit(seed, token: token),
            branch: pullRequest.map { "feature/\($0)" } ?? "main",
            workflow: workflow,
            conclusion: "success"
        )
    }

    private func history(via api: BackendAPI) -> [[String: Any]] {
        var all: [[String: Any]] = []
        var before: Int64?
        for _ in 0..<20 {
            let path = "/api/notification-history?limit=100" + (before.map { "&before=\($0)" } ?? "")
            let reply = api.call("GET", path)
            XCTAssertEqual(reply.status, 200, "The history list should load: \(reply.text)")
            let page = reply.json?["items"] as? [[String: Any]] ?? []
            all += page
            guard page.count == 100, let last = (page.last?["id"] as? NSNumber)?.int64Value else { break }
            before = last
        }
        return all
    }

    private func items(of repo: String, via api: BackendAPI) -> [[String: Any]] {
        history(via: api).filter { $0["repo_full_name"] as? String == repo }
    }

    private func waitForItems(of repo: String, count: Int, via api: BackendAPI) {
        var found = items(of: repo, via: api).count
        var polls = 0
        while found < count && polls < 60 {
            Thread.sleep(forTimeInterval: 1)
            found = items(of: repo, via: api).count
            polls += 1
        }
        XCTAssertEqual(found, count, "The consumer should store \(count) history items for \(repo)")
    }

    private func acknowledgeEverything(via api: BackendAPI) {
        for item in history(via: api) where item["acknowledged_at"] == nil {
            guard let id = (item["id"] as? NSNumber)?.int64Value else { continue }
            let reply = api.call("POST", "/api/notification-history/\(id)/ack")
            XCTAssertTrue((200..<300).contains(reply.status), "Acknowledging \(id) should succeed: \(reply.text)")
        }
        let left = history(via: api).filter { $0["acknowledged_at"] == nil }.count
        XCTAssertEqual(left, 0, "Every earlier history item should be acknowledged before the scenario starts")
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

    private func top(of group: String, in app: XCUIApplication) -> CGFloat {
        WebhookHistoryUI.texts(of: group, in: app).allElementsBoundByIndex.map(\.frame.minY).min() ?? .infinity
    }

    private func link(_ label: String, of group: String, in app: XCUIApplication) -> XCUIElement {
        app.descendants(matching: .any).matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label ENDSWITH %@", WebhookHistoryUI.groupPrefix + group, label)
        ).firstMatch
    }

    private func assertLinks(of group: String, withPullRequest: Bool, in app: XCUIApplication) {
        for label in ["커밋", "런"] {
            let chip = link(label, of: group, in: app)
            XCTAssertTrue(chip.waitForExistence(timeout: 10), "Group \(group) should offer the \(label) link")
            XCTAssertTrue(chip.isHittable, "The \(label) link of \(group) should be reachable")
        }
        XCTAssertEqual(
            link("PR", of: group, in: app).exists, withPullRequest,
            "Group \(group) should offer a PR link only when the event has a pull request"
        )
    }

    private func assertUnreadFirst(unread: [String], read: String, in app: XCUIApplication) {
        XCTAssertTrue(
            app.staticTexts["미확인 · \(unread.count)개 그룹"].waitForExistence(timeout: 10),
            "The unacknowledged section should hold exactly the new groups"
        )
        let tops = unread.map { top(of: $0, in: app) }
        XCTAssertEqual(tops, tops.sorted(), "Unacknowledged groups should be newest first: \(zip(unread, tops).map { "\($0)=\($1)" })")
        for group in unread {
            XCTAssertTrue(WebhookHistoryUI.text("미확인 1", of: group, in: app).exists, "Group \(group) should be emphasised as 미확인 1")
            XCTAssertEqual(WebhookHistoryUI.itemAcknowledgeButtons(of: group, in: app).count, 1, "Group \(group) should list its item with 확인")
        }
        XCTAssertTrue(WebhookHistoryUI.reveal(read, in: app), "The acknowledged group should be listed")
        let header = app.staticTexts["확인 완료"]
        XCTAssertTrue(header.exists, "The acknowledged section header should be shown")
        XCTAssertLessThan(header.frame.minY, top(of: read, in: app), "The acknowledged group should sit under 확인 완료")
        if let last = unread.last, WebhookHistoryUI.texts(of: last, in: app).firstMatch.exists {
            XCTAssertLessThan(top(of: last, in: app), header.frame.minY, "Unacknowledged groups should sit above 확인 완료")
        }
        let labels = WebhookHistoryUI.labels(of: read, in: app)
        XCTAssertTrue(labels.contains("모두 확인"), "The acknowledged group should read 모두 확인: \(labels)")
        XCTAssertFalse(labels.contains("Build"), "The acknowledged group should be collapsed to its header: \(labels)")
        XCTAssertEqual(WebhookHistoryUI.acknowledgeAllButtons(of: read, in: app).count, 0, "The acknowledged group should offer no 모두 확인 button")
    }

    private static func signOut(in app: XCUIApplication) {
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        var settings = app.buttons["prmonitor.settings.button"]
        if !settings.waitForExistence(timeout: 10) {
            settings = app.buttons["알림 설정"]
        }
        XCTAssertTrue(settings.waitForExistence(timeout: 5), "The PR monitor header should offer 알림 설정")
        settings.tap()
        var row = app.buttons["prmonitor.settings.signOut"]
        if !row.waitForExistence(timeout: 10) {
            row = app.buttons.matching(NSPredicate(format: "label == %@", "로그아웃")).firstMatch
        }
        XCTAssertTrue(row.waitForExistence(timeout: 5), "The 알림 설정 sheet should offer 로그아웃")
        row.tap()
        var confirm = app.buttons["prmonitor.settings.signOut.confirm"]
        if !confirm.waitForExistence(timeout: 10) {
            confirm = app.buttons.matching(
                NSPredicate(format: "label == %@ AND identifier != %@", "로그아웃", "prmonitor.settings.signOut")
            ).firstMatch
        }
        XCTAssertTrue(confirm.waitForExistence(timeout: 5), "로그아웃 should ask for confirmation")
        confirm.tap()
        XCTAssertTrue(app.buttons["SignInButton"].waitForExistence(timeout: 20), "Signing out should land on the login screen")
    }

    private static func signIn(in app: XCUIApplication) {
        let button = app.buttons["SignInButton"]
        XCTAssertTrue(button.waitForExistence(timeout: 20), "The login screen should offer sign-in")
        button.tap()
        let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let consent = springboard.buttons.matching(NSPredicate(format: "label IN %@", ["Continue", "계속", "확인", "OK"])).firstMatch
        if consent.waitForExistence(timeout: 10) {
            consent.tap()
        }
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 120), "Signing in should land on the tab shell")
    }

    private static func restoreDefaultUser() {
        BackendAPI.setNextAppLoginSubject("")
        let app = XCUIApplication()
        app.launch()
        if app.tabBars.firstMatch.waitForExistence(timeout: 15) {
            signOut(in: app)
        }
        signIn(in: app)
        app.terminate()
    }

    func testCompletedEventsListUnacknowledgedFirstAndStayWithTheirUser() throws {
        let token = WebhookHistoryUI.runToken()
        let repo = "dlddu/e2e-userscope-\(token)"
        let owner = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get the signed-in user's token")

        let earlier = event(repo, pullRequest: 1, seed: 1, workflow: "Build", token: token)
        WebhookHistoryUI.send(earlier)
        waitForItems(of: repo, count: 1, via: owner)
        acknowledgeEverything(via: owner)

        let pulled = event(repo, pullRequest: 2, seed: 2, workflow: "Lint", token: token)
        let pushed = event(repo, pullRequest: nil, seed: 3, workflow: "Deploy", token: token)
        WebhookHistoryUI.send([pulled, pushed])
        waitForItems(of: repo, count: 3, via: owner)

        let app = WebhookHistoryUI.openMonitor(until: pushed.groupKey, reads: "미확인 1")
        XCTAssertEqual(
            WebhookHistoryUI.groups(of: repo, in: app).intersection([pushed.groupKey, pulled.groupKey]),
            [pushed.groupKey, pulled.groupKey],
            "Both new events should be in the history"
        )
        assertLinks(of: pushed.groupKey, withPullRequest: false, in: app)
        assertLinks(of: pulled.groupKey, withPullRequest: true, in: app)
        XCTAssertEqual(waitForUnread(2, in: app), 2, "The header badge should equal the two unacknowledged items")
        assertUnreadFirst(unread: [pushed.groupKey, pulled.groupKey], read: earlier.groupKey, in: app)

        addTeardownBlock { Self.restoreDefaultUser() }
        let other = "e2e-userscope-other-\(token)"
        Self.signOut(in: app)
        BackendAPI.setNextAppLoginSubject(other)
        Self.signIn(in: app)

        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        let empty = app.descendants(matching: .any).matching(identifier: "prmonitor.empty.state").firstMatch
        XCTAssertTrue(empty.waitForExistence(timeout: 20), "The other user should start with an empty history")
        XCTAssertTrue(WebhookHistoryUI.groups(of: repo, in: app).isEmpty, "The first user's history should not show for the other user")
        XCTAssertFalse(
            app.descendants(matching: .any).matching(identifier: "prmonitor.unread.badge").firstMatch.exists,
            "The other user should have no unacknowledged badge"
        )

        let second = try XCTUnwrap(BackendAPI.signIn(subject: other), "The test runner should get the other user's token")
        XCTAssertTrue(items(of: repo, via: second).isEmpty, "The other user should own no history rows of \(repo)")
        XCTAssertEqual(items(of: repo, via: owner).count, 3, "The first user's rows should stay with the first user")
        app.terminate()
    }
}
