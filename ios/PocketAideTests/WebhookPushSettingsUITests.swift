// 검증 시나리오: test-github-monitor.md#시나리오 9
import XCTest

final class WebhookPushSettingsUITests: XCTestCase {
    private let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
    private let settingsApp = XCUIApplication(bundleIdentifier: "com.apple.Preferences")

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func event(_ repo: String, _ number: Int, conclusion: String, token: String) -> WebhookEvent {
        WebhookEvent(
            repo: repo,
            pullRequest: number,
            commit: WebhookHistoryUI.commit(number, token: token),
            branch: "e2e/settings-\(number)",
            workflow: "Push settings",
            conclusion: conclusion
        )
    }

    private func signedInAPI() throws -> BackendAPI {
        try XCTUnwrap(BackendAPI.signIn(), "The runner should sign in as the app's user")
    }

    private func registerDevice(_ device: String, api: BackendAPI) {
        let reply = api.call("POST", "/api/device-tokens", json: ["token": device])
        XCTAssertEqual(reply.status, 201, "The device token should register: \(reply.text)")
    }

    private func restoreSettings(_ api: BackendAPI) {
        _ = api.call("PATCH", "/api/notification-settings", json: ["enabled": true, "outcomes": "both"])
    }

    private func waitForSettings(enabled: Bool, outcomes: String?, api: BackendAPI) {
        var last = ""
        for _ in 0..<20 {
            let reply = api.call("GET", "/api/notification-settings")
            last = reply.text
            if let json = reply.json, json["enabled"] as? Bool == enabled,
               outcomes == nil || json["outcomes"] as? String == outcomes {
                return
            }
            Thread.sleep(forTimeInterval: 0.5)
        }
        XCTFail("The sheet should save enabled=\(enabled) outcomes=\(outcomes ?? "-"), server has \(last)")
    }

    private func historyRepos(api: BackendAPI) -> [String] {
        let items = api.call("GET", "/api/notification-history?limit=100").json?["items"] as? [[String: Any]] ?? []
        return items.compactMap { $0["repo_full_name"] as? String }
    }

    private func waitForHistory(of repo: String, count: Int, api: BackendAPI) {
        var seen = 0
        for _ in 0..<60 {
            seen = historyRepos(api: api).filter { $0 == repo }.count
            if seen >= count { return }
            Thread.sleep(forTimeInterval: 1)
        }
        XCTFail("\(repo) should have \(count) history items whatever the settings, saw \(seen)")
    }

    private func pushTitles(to device: String, repo: String) -> [String] {
        guard let url = URL(string: "http://127.0.0.1:5558/e2e/received?device=\(device)") else { return [] }
        let done = DispatchSemaphore(value: 0)
        let box = PushReceiverBox()
        URLSession.shared.dataTask(with: url) { data, _, _ in
            box.rows = (try? JSONSerialization.jsonObject(with: data ?? Data())) as? [[String: Any]] ?? []
            done.signal()
        }.resume()
        XCTAssertTrue(done.wait(timeout: .now() + 20) == .success, "The fake APNs receiver should answer")
        return box.rows.compactMap { row in
            let aps = (row["payload"] as? [String: Any])?["aps"] as? [String: Any]
            let title = (aps?["alert"] as? [String: Any])?["title"] as? String
            return title.flatMap { $0.contains(repo) ? $0 : nil }
        }.sorted()
    }

    private func assertPushes(to device: String, repo: String, equal expected: [String]) {
        var titles: [String] = []
        for _ in 0..<15 {
            titles = pushTitles(to: device, repo: repo)
            if titles.count >= expected.count { break }
            Thread.sleep(forTimeInterval: 1)
        }
        Thread.sleep(forTimeInterval: 3)
        titles = pushTitles(to: device, repo: repo)
        XCTAssertEqual(titles, expected.sorted(), "Pushes sent for \(repo)")
    }

    private func allowNotificationsIfAsked() {
        let allow = springboard.alerts.buttons
            .matching(NSPredicate(format: "label IN %@", ["Allow", "허용"])).firstMatch
        if allow.waitForExistence(timeout: 5) {
            allow.tap()
        }
    }

    private func launchMonitor() -> XCUIApplication {
        let app = XCUIApplication()
        app.launch()
        allowNotificationsIfAsked()
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        return app
    }

    private func openSettingsSheet(in app: XCUIApplication) {
        let bell = app.buttons["prmonitor.settings.button"]
        XCTAssertTrue(bell.waitForExistence(timeout: 15), "The PR monitor header should have the bell button")
        bell.tap()
        XCTAssertTrue(app.switches["prmonitor.settings.enabled"].waitForExistence(timeout: 10), "The settings sheet should open")
    }

    private func closeSettingsSheet(in app: XCUIApplication) {
        app.buttons["완료"].firstMatch.tap()
        XCTAssertTrue(app.buttons["prmonitor.settings.button"].waitForExistence(timeout: 10), "The sheet should close")
    }

    private func setEnabled(_ enabled: Bool, in app: XCUIApplication) {
        let toggle = app.switches["prmonitor.settings.enabled"]
        if (toggle.value as? String == "1") != enabled {
            toggle.coordinate(withNormalizedOffset: CGVector(dx: 0.93, dy: 0.5)).tap()
        }
        let settled = expectation(for: NSPredicate(format: "value == %@", enabled ? "1" : "0"), evaluatedWith: toggle)
        wait(for: [settled], timeout: 10)
    }

    private func chooseOutcomes(_ label: String, in app: XCUIApplication) {
        let segment = app.segmentedControls["prmonitor.settings.outcomes"].buttons[label]
        let choice = segment.waitForExistence(timeout: 5) ? segment : app.buttons[label].firstMatch
        XCTAssertTrue(choice.waitForExistence(timeout: 5), "The outcomes picker should offer \(label)")
        choice.tap()
        let selected = expectation(for: NSPredicate(format: "isSelected == true"), evaluatedWith: choice)
        wait(for: [selected], timeout: 10)
    }

    private func assertHistoryShows(_ events: [WebhookEvent]) {
        let app = launchMonitor()
        for event in events {
            XCTAssertTrue(
                WebhookHistoryUI.reveal(event.groupKey, in: app),
                "The history should list \(event.repo) #\(event.pullRequest ?? 0) whatever the settings"
            )
        }
        app.terminate()
    }

    private struct Stage {
        let name: String
        let outcomes: String?
        let pushesSuccess: Bool
        let pushesFailure: Bool

        var enabled: Bool { outcomes != nil }
        var label: String {
            switch outcomes {
            case "failure": return "실패만"
            case "success": return "성공만"
            default: return "둘 다"
            }
        }
    }

    private func run(_ stage: Stage, token: String, device: String, api: BackendAPI) {
        let app = launchMonitor()
        openSettingsSheet(in: app)
        setEnabled(stage.enabled, in: app)
        if stage.enabled {
            chooseOutcomes(stage.label, in: app)
        }
        waitForSettings(enabled: stage.enabled, outcomes: stage.outcomes, api: api)
        closeSettingsSheet(in: app)
        app.terminate()

        let repo = "dlddu/push-settings-\(token)-\(stage.name)"
        let passed = event(repo, 91, conclusion: "success", token: token)
        let failed = event(repo, 92, conclusion: "failure", token: token)
        WebhookHistoryUI.send([passed, failed])
        waitForHistory(of: repo, count: 2, api: api)

        var expected: [String] = []
        if stage.pushesSuccess { expected.append("CI 통과 — \(repo) #91") }
        if stage.pushesFailure { expected.append("CI 실패 — \(repo) #92") }
        assertPushes(to: device, repo: repo, equal: expected)
        assertHistoryShows([passed, failed])
    }

    func testPushArrivalFollowsNotificationSettings() throws {
        let token = WebhookHistoryUI.runToken()
        let device = (UUID().uuidString + UUID().uuidString).replacingOccurrences(of: "-", with: "").lowercased()
        let api = try signedInAPI()
        addTeardownBlock { self.restoreSettings(api) }
        restoreSettings(api)
        registerDevice(device, api: api)

        run(Stage(name: "failure", outcomes: "failure", pushesSuccess: false, pushesFailure: true),
            token: token, device: device, api: api)
        run(Stage(name: "success", outcomes: "success", pushesSuccess: true, pushesFailure: false),
            token: token, device: device, api: api)
        run(Stage(name: "off", outcomes: nil, pushesSuccess: false, pushesFailure: false),
            token: token, device: device, api: api)
    }

    private func revealCell(_ labels: [String], in list: XCUIApplication) -> XCUIElement {
        let cell = list.cells.containing(NSPredicate(format: "label IN %@", labels)).firstMatch
        let item = cell.exists ? cell : list.staticTexts.matching(NSPredicate(format: "label IN %@", labels)).firstMatch
        var drags = 0
        while !(item.exists && item.isHittable) && drags < 12 {
            list.swipeUp()
            drags += 1
        }
        XCTAssertTrue(item.exists, "Settings should list \(labels.joined(separator: " / "))")
        return item
    }

    private func setSystemNotifications(allowed: Bool) {
        settingsApp.terminate()
        settingsApp.launch()
        revealCell(["Apps", "앱"], in: settingsApp).tap()
        let search = settingsApp.searchFields.firstMatch
        if search.waitForExistence(timeout: 3) {
            search.tap()
            search.typeText("PocketAide")
        }
        revealCell(["PocketAide"], in: settingsApp).tap()
        revealCell(["Notifications", "알림"], in: settingsApp).tap()
        let allow = settingsApp.switches
            .matching(NSPredicate(format: "label IN %@", ["Allow Notifications", "알림 허용"])).firstMatch
        XCTAssertTrue(allow.waitForExistence(timeout: 10), "PocketAide's notification page should have the allow switch")
        if (allow.value as? String == "1") != allowed {
            allow.coordinate(withNormalizedOffset: CGVector(dx: 0.93, dy: 0.5)).tap()
        }
        let settled = expectation(for: NSPredicate(format: "value == %@", allowed ? "1" : "0"), evaluatedWith: allow)
        wait(for: [settled], timeout: 10)
        settingsApp.terminate()
    }

    func testDeniedSystemPermissionShowsNotice() {
        let app = launchMonitor()
        openSettingsSheet(in: app)
        XCTAssertFalse(
            app.descendants(matching: .any)["prmonitor.settings.permissionNotice"].exists,
            "With notifications allowed the sheet should not warn"
        )
        closeSettingsSheet(in: app)

        addTeardownBlock {
            self.setSystemNotifications(allowed: true)
            XCUIApplication().terminate()
        }
        setSystemNotifications(allowed: false)

        app.activate()
        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 15), "The app should come back to the front")
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        openSettingsSheet(in: app)
        let notice = app.descendants(matching: .any)["prmonitor.settings.permissionNotice"]
        XCTAssertTrue(notice.waitForExistence(timeout: 15), "The sheet should say system notifications are off")
        XCTAssertTrue(
            notice.label.contains("시스템 알림 권한이 꺼져 있어"),
            "The notice should explain the system permission: \(notice.label)"
        )
        closeSettingsSheet(in: app)
        app.terminate()
    }
}

private final class PushReceiverBox {
    var rows: [[String: Any]] = []
}
