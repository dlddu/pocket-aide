// 검증 시나리오: test-approval-gate.md#시나리오 6
import XCTest

final class ApprovalPushUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testPushNamesTheRealKeyAndOpensTheRequestWithoutDeciding() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The runner should sign in as the app's user")
        let keyName = "infra-ci-\(token.lowercased())"
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: keyName), "A caller key should be issued")
        addTeardownBlock { _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke") }
        let device = ApprovalPushReceiver.registerDevice(api)

        let firstContext = "prod 배포 \(token) — 승인해 주세요"
        let first = try XCTUnwrap(caller.create("push-\(token)-1", context: firstContext, timeout: nil))
        let firstPush = try XCTUnwrap(
            ApprovalPushReceiver.payload(for: first, to: device),
            "Step 2: a push should be sent for the new request"
        )
        XCTAssertEqual(ApprovalPushReceiver.title(firstPush), "승인 요청 · \(keyName)", "Step 2: the title should name the real key, not Deploy Bot")
        XCTAssertEqual(ApprovalPushReceiver.body(firstPush), firstContext, "Step 2: the body should be the context")

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        app.terminate()

        let secondExternalID = "push-\(token)-2"
        let secondContext = "staging 배포 \(token) — 두 번째 요청"
        let second = try XCTUnwrap(caller.create(secondExternalID, context: secondContext, timeout: nil))
        let secondPush = try XCTUnwrap(
            ApprovalPushReceiver.payload(for: second, to: device),
            "Step 3: a push should be sent for the second request"
        )
        XCTAssertEqual(secondPush["approval_id"] as? String, second, "Step 3: the push should carry the request id")

        var link = URLComponents()
        link.scheme = "pocketaide"
        link.host = "approvals"
        link.queryItems = [URLQueryItem(name: "id", value: second)]
        app.open(try XCTUnwrap(link.url))
        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 20), "Step 3: opening the push link should launch the app")

        let detail = ApprovalDetailScreen(app: app)
        XCTAssertTrue(detail.title.waitForExistence(timeout: 20), "Step 3: the request detail should open")
        let facts = [
            ("approval.detail.context", secondContext),
            ("approval.detail.requester", "Deploy Bot"),
            ("approval.detail.key", keyName),
            ("approval.detail.externalId", secondExternalID),
        ]
        for (id, expected) in facts {
            let element = app.staticTexts[id]
            XCTAssertTrue(element.waitForExistence(timeout: 10), "Step 3: \(id) should be on the detail")
            XCTAssertTrue(element.label.contains(expected), "Step 3: \(id) should show \(expected): \(element.label)")
        }
        XCTAssertTrue(detail.waitForStatus("대기 중"), "Step 3: the detail should still read 대기 중: \(detail.status.label)")
        XCTAssertTrue(ApprovalSeed.tab(in: app).isSelected, "Step 3: the 승인 tab should be selected")
        XCTAssertEqual(ApprovalLookup.status(api, second), "PENDING", "Step 3: opening the push must not decide the request")
        XCTAssertEqual(ApprovalLookup.status(api, first), "PENDING", "The first request should stay pending too")
        app.terminate()
    }

    func testPushLinkOverAnOpenDetailSwitchesToTheNewRequest() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The runner should sign in as the app's user")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "push-swap-\(token.lowercased())"), "A caller key should be issued")
        addTeardownBlock { _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke") }

        let firstExternalID = "swap-\(token)-1"
        let secondExternalID = "swap-\(token)-2"
        let first = try XCTUnwrap(caller.create(firstExternalID, context: "첫 요청 \(token)", timeout: nil))
        let second = try XCTUnwrap(caller.create(secondExternalID, context: "둘째 요청 \(token)", timeout: nil))

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")

        let externalID = app.staticTexts["approval.detail.externalId"]
        app.open(try XCTUnwrap(Self.link(first)))
        XCTAssertTrue(externalID.waitForExistence(timeout: 20), "The first push should open its detail")
        XCTAssertTrue(externalID.label.contains(firstExternalID), "The detail should show the first request: \(externalID.label)")

        app.open(try XCTUnwrap(Self.link(second)))
        let switched = NSPredicate(format: "label CONTAINS %@", secondExternalID)
        XCTAssertEqual(
            XCTWaiter.wait(for: [expectation(for: switched, evaluatedWith: externalID)], timeout: 20), .completed,
            "A push over an open detail should switch to the new request: \(externalID.label)"
        )
        app.terminate()
    }

    private static func link(_ approvalID: String) -> URL? {
        var link = URLComponents()
        link.scheme = "pocketaide"
        link.host = "approvals"
        link.queryItems = [URLQueryItem(name: "id", value: approvalID)]
        return link.url
    }
}
