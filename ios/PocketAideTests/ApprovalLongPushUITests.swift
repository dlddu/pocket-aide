// 검증 시나리오: test-approval-gate.md#시나리오 7
import XCTest

final class ApprovalLongPushUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testLongContextPushIsCutAndBothRequestsWait() throws {
        let token = TodoUI.uniqueToken()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The runner should sign in as the app's user")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(api, name: "long-ci-\(token.lowercased())"), "A caller key should be issued")
        addTeardownBlock { _ = api.call("POST", "/api/approval-keys/\(caller.id)/revoke") }
        let device = ApprovalPushReceiver.registerDevice(api)

        let longContext = String(repeating: "배포 로그 \(token) 줄을 확인해 주세요. ", count: 400)
        XCTAssertGreaterThanOrEqual(longContext.utf8.count, 10 * 1024, "The context should be at least 10KB")
        let longExternalID = "long-\(token)"
        let started = Date()
        let long = try XCTUnwrap(caller.create(longExternalID, context: longContext, timeout: nil))
        XCTAssertLessThan(Date().timeIntervalSince(started), 5, "Step 1: creating should not wait for the push")
        let push = try XCTUnwrap(ApprovalPushReceiver.payload(for: long, to: device), "Step 1: the long request should still be pushed")
        let body = try XCTUnwrap(ApprovalPushReceiver.body(push), "Step 1: the push should have a body")
        XCTAssertTrue(body.hasSuffix("…"), "Step 1: the cut body should end with …")
        XCTAssertTrue(longContext.hasPrefix(String(body.dropLast())), "Step 1: the body should be the front of the context")
        let size = try JSONSerialization.data(withJSONObject: push).count
        XCTAssertLessThanOrEqual(size, 4096, "Step 1: the payload should fit the 4KB APNs limit, was \(size) bytes")

        let shortExternalID = "short-\(token)"
        _ = try XCTUnwrap(caller.create(shortExternalID, context: "짧은 요청 \(token)", timeout: nil))

        let app = TodoUI.launch()
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "The 승인 tab should be in the tab bar")
        tab.tap()
        let list = ApprovalsScreen(app: app)
        XCTAssertTrue(list.title.waitForExistence(timeout: 15), "The 승인 screen should open")
        if !list.row(longExternalID).waitForExistence(timeout: 5) {
            list.pullToRefresh()
        }
        XCTAssertTrue(list.row(longExternalID).waitForExistence(timeout: 15), "Step 3: the long request should be waiting")
        XCTAssertTrue(list.row(shortExternalID).waitForExistence(timeout: 15), "Step 3: the second request should be waiting")
        app.terminate()
    }
}
