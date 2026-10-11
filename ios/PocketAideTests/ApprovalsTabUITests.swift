// 검증 시나리오: test-approval-gate.md#시나리오 5
import XCTest

final class ApprovalsTabUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testPendingRequestsAppearNewestFirstWithTabBadge() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "infra-ci-\(token)"), "A caller key should be issued")
        let created = ApprovalIDs()
        addTeardownBlock {
            created.ids.forEach { _ = api.call("POST", "/api/approvals/\($0)/reject") }
            _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        let app = TodoUI.launch()
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 5), "The 승인 tab should sit in the bottom tab bar, not under More")

        let externalIDs = ["first", "second", "third"].map { "\($0)-\(token)" }
        for (index, externalID) in externalIDs.enumerated() {
            let timeout: Int? = index == 1 ? nil : 1800
            created.ids.append(try XCTUnwrap(caller.create(externalID, context: "요청 \(externalID)", timeout: timeout)))
            Thread.sleep(forTimeInterval: 1.1)
        }

        tab.tap()
        let screen = ApprovalsScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "The 승인 screen should open")
        screen.assertOrder(externalIDs.reversed(), "Right after opening the 승인 tab")
        XCTAssertTrue(screen.row(externalIDs[1]).label.contains(ApprovalSeedCopy.noDeadline), "The request without a timeout should say 기한 없음")
        XCTAssertTrue(screen.row(externalIDs[0]).label.contains("남음"), "A request with a timeout should show the time left")
        screen.assertCounts(api, "With three requests pending")

        let approved = api.call("POST", "/api/approvals/\(created.ids[0])/approve")
        XCTAssertEqual(approved.status, 200, "Approving through the API should succeed: \(approved.text)")
        screen.pullToRefresh()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.row(externalIDs[0]), timeout: 15), "The approved request should leave the list")
        screen.assertCounts(api, "After one request was approved")

        XCUIDevice.shared.press(.home)
        let late = "fourth-\(token)"
        created.ids.append(try XCTUnwrap(caller.create(late, context: "요청 \(late)", timeout: 1800)))
        app.activate()
        XCTAssertTrue(screen.row(late).waitForExistence(timeout: 20), "A request made in the background should show up on return")
        screen.assertOrder([late, externalIDs[2], externalIDs[1]], "After returning to the foreground")
        screen.assertCounts(api, "After a request arrived in the background")
    }
}

final class ApprovalIDs {
    var ids: [String] = []
}

enum ApprovalSeedCopy {
    static let noDeadline = "기한 없음"
}

struct ApprovalCaller {
    let id: Int
    let key: String

    func create(_ externalID: String, context: String, timeout: Int?) -> String? {
        var body: [String: Any] = ["externalId": externalID, "context": context, "requesterName": "Deploy Bot"]
        if let timeout {
            body["timeoutSeconds"] = timeout
        }
        let reply = BackendAPI(token: key).call("POST", "/api/external/approvals", json: body)
        XCTAssertEqual(reply.status, 201, "Creating an approval request should succeed: \(reply.text)")
        return reply.json?["id"] as? String
    }
}

enum ApprovalSeed {
    static func issueKey(_ api: BackendAPI, name: String) -> ApprovalCaller? {
        let reply = api.call("POST", "/api/approval-keys", json: ["name": name])
        XCTAssertEqual(reply.status, 201, "Issuing a caller key should succeed: \(reply.text)")
        guard let key = reply.json?["key"] as? String, let id = reply.json?["id"] as? Int else { return nil }
        return ApprovalCaller(id: id, key: key)
    }

    static func pendingCount(_ api: BackendAPI) -> Int? {
        let reply = api.call("GET", "/api/approvals")
        return (reply.json?["items"] as? [Any])?.count
    }

    static func tab(in app: XCUIApplication) -> XCUIElement {
        app.tabBars.firstMatch.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "승인")).firstMatch
    }
}

struct ApprovalsScreen {
    let app: XCUIApplication

    var title: XCUIElement { app.staticTexts["approvals.title"] }
    var pendingCount: XCUIElement { app.staticTexts["approvals.pending.count"] }

    func row(_ externalID: String) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: "approvals.row.\(externalID)").firstMatch
    }

    func pullToRefresh() {
        let list = app.collectionViews.firstMatch
        XCTAssertTrue(list.waitForExistence(timeout: 5), "The 승인 list should be on screen")
        let start = list.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.15))
        start.press(forDuration: 0.1, thenDragTo: list.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.85)))
    }

    func assertOrder(_ externalIDs: [String], _ step: String, file: StaticString = #filePath, line: UInt = #line) {
        let rows = externalIDs.map(row)
        for (externalID, element) in zip(externalIDs, rows) {
            XCTAssertTrue(element.waitForExistence(timeout: 15), "\(step): \(externalID) should be listed", file: file, line: line)
        }
        let tops = rows.map(\.frame.minY)
        XCTAssertEqual(tops, tops.sorted(), "\(step): rows should be newest first \(externalIDs)", file: file, line: line)
    }

    func assertCounts(_ api: BackendAPI, _ step: String, file: StaticString = #filePath, line: UInt = #line) {
        guard let expected = ApprovalSeed.pendingCount(api) else {
            return XCTFail("\(step): the pending list API should answer", file: file, line: line)
        }
        var reading = TabBadgeReading.unreadable
        let matched = RoutinesScreen.eventually(timeout: 15) {
            reading = TabBadgeReader.read(ApprovalSeed.tab(in: app), in: app)
            return pendingCount.label == "\(expected)" && reading.count == expected
        }
        XCTAssertTrue(
            matched,
            "\(step): the header count and the 승인 tab badge should read \(expected) but read \(pendingCount.label) / \(reading)",
            file: file,
            line: line
        )
    }
}
