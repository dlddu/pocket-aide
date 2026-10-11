// 검증 시나리오: test-approval-gate.md#시나리오 2
import XCTest

final class ApprovalKeyLastUsedUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testCallerRequestAndPollUpdateTheKeysLastUse() throws {
        let token = TodoUI.uniqueToken()
        let name = "infra-ci-\(token)"
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: name), "A caller key should be issued")
        let created = ApprovalIDs()
        addTeardownBlock {
            created.ids.forEach { _ = api.call("POST", "/api/approvals/\($0)/reject") }
            _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        var screen = ApprovalKeysScreen.open(in: app)
        XCTAssertTrue(screen.row(name).waitForExistence(timeout: 15), "\(name) should be listed")
        XCTAssertTrue(screen.meta(name).label.hasSuffix("마지막 사용 없음"), "Before any call the key should be unused")

        let id = try XCTUnwrap(caller.create("deploy-1-\(token)", context: "prod 배포 deploy-1", timeout: 600))
        created.ids.append(id)
        let polled = try XCTUnwrap(ApprovalLookup.callerView(caller, id), "The caller should read its request back")
        XCTAssertEqual(polled["status"] as? String, "PENDING")
        XCTAssertNil(polled["processedAt"] as? Int, "A pending request should have no processedAt")
        XCTAssertNil(polled["sub"], "The caller view should carry no user identity")
        XCTAssertNil(polled["email"], "The caller view should carry no user identity")

        screen.goBack()
        XCTAssertTrue(app.staticTexts["approvals.title"].waitForExistence(timeout: 10), "Going back should land on 승인")
        screen = ApprovalKeysScreen.open(in: app)
        XCTAssertTrue(screen.row(name).waitForExistence(timeout: 15), "\(name) should be listed")
        let meta = screen.meta(name)
        XCTAssertTrue(
            RoutinesScreen.eventually(timeout: 15) { meta.label.range(of: #"마지막 사용 \d{2}:\d{2}$"#, options: .regularExpression) != nil },
            "The key list should show today's last use after the caller's calls: \(meta.label)"
        )
        let stored = try XCTUnwrap(ApprovalKeysAPI.key(api, named: name))
        XCTAssertNotNil(stored["last_used_at"] as? Int, "The server should have stamped the last use")
        app.terminate()
    }
}
