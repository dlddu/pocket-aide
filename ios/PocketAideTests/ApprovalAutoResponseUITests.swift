// 검증 시나리오: test-approval-gate.md#시나리오 13
import XCTest

final class ApprovalAutoResponseUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testAutoModeAppliesOnlyToLaterRequestsWithoutPushAndShowsABanner() throws {
        let token = TodoUI.uniqueToken()
        let subject = "e2e-approval-auto-\(token.lowercased())"
        let user = try XCTUnwrap(BackendAPI.signIn(subject: subject), "The test runner should get the auto-response user's token")
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(user, name: "infra-ci-\(token.lowercased())"), "A caller key should be issued")
        addTeardownBlock {
            _ = user.call("PATCH", "/api/approval-settings", json: ["auto_response": "NONE"])
            _ = user.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }
        let device = ApprovalPushReceiver.registerDevice(user)

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        addTeardownBlock { AppAccount.restoreDefaultUser() }
        AppAccount.signOut(in: app)
        BackendAPI.setNextAppLoginSubject(subject)
        AppAccount.signIn(in: app)

        let screen = ApprovalsScreen(app: app)
        let tab = ApprovalSeed.tab(in: app)
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "The 승인 tab should be in the tab bar")
        tab.tap()
        let emptyMode = app.staticTexts["approvals.empty.mode"]
        XCTAssertTrue(emptyMode.waitForExistence(timeout: 15), "An empty 승인 tab should show the current auto-response mode")
        XCTAssertTrue(emptyMode.label.contains("자동 응답 모드 없음"), "The empty state should read 없음: \(emptyMode.label)")
        XCTAssertTrue(header(in: app).label.contains("없음"), "The header should read 자동 응답 없음: \(header(in: app).label)")

        let pending = "p-\(token)"
        let pendingID = try XCTUnwrap(caller.create(pending, context: "기존 대기 요청 \(pending)", timeout: 600))
        XCTAssertNotNil(ApprovalPushReceiver.payload(for: pendingID, to: device), "Precondition: with no auto mode P should be pushed")

        choose("AUTO_APPROVE", in: app, user: user, step: "Step 1")
        let banner = bannerText(in: app)
        XCTAssertTrue(banner.waitForExistence(timeout: 15), "Step 2: the auto-response banner should be on the 승인 tab")
        XCTAssertTrue(banner.label.contains("자동 승인 모드가 켜져 있습니다."), "Step 2: the banner should name 자동 승인: \(banner.label)")
        XCTAssertTrue(screen.row(pending).waitForExistence(timeout: 15), "Step 2: P should still be listed")
        XCTAssertTrue(header(in: app).label.contains("자동 승인"), "Step 2: the header should read 자동 승인: \(header(in: app).label)")

        let approved = try create("q-\(token)", caller: caller, expecting: "APPROVED", step: "Step 3")
        XCTAssertNil(ApprovalPushReceiver.payload(for: approved, to: device, timeout: 5), "Step 3: no push should be sent for Q")
        XCTAssertEqual(ApprovalLookup.status(user, pendingID), "PENDING", "Step 2: P must not be decided by the new mode")

        let history = ApprovalHistoryScreen.open(in: app)
        let approvedRow = history.row("q-\(token)")
        XCTAssertTrue(approvedRow.waitForExistence(timeout: 15), "Step 4: Q should be in 처리 이력")
        XCTAssertTrue(approvedRow.label.contains("자동 승인"), "Step 4: Q should read 자동 승인: \(approvedRow.label)")
        history.goBack()

        let off = app.buttons["approvals.autoResponse.off"]
        XCTAssertTrue(off.waitForExistence(timeout: 15), "Step 5: the banner should offer 끄기")
        off.tap()
        XCTAssertTrue(RoutinesScreen.eventually(timeout: 15) { !bannerText(in: app).exists }, "Step 5: the banner should go away")
        XCTAssertTrue(RoutinesScreen.eventually(timeout: 15) { header(in: app).label.contains("없음") }, "Step 5: the header should read 없음")
        XCTAssertEqual(mode(user), "NONE", "Step 5: 끄기 should store 없음")

        try assertAutoReject("r-\(token)", in: app, user: user, caller: caller, device: device)
        XCTAssertEqual(ApprovalLookup.status(user, pendingID), "PENDING", "P should stay PENDING through both modes")
        app.terminate()
    }

    private func assertAutoReject(_ externalID: String, in app: XCUIApplication, user: BackendAPI, caller: ApprovalCaller, device: String) throws {
        choose("AUTO_REJECT", in: app, user: user, step: "Step 1 (자동 거절)")
        XCTAssertTrue(bannerText(in: app).waitForExistence(timeout: 15), "Step 2 (자동 거절): the banner should be back")
        XCTAssertTrue(bannerText(in: app).label.contains("자동 거절 모드가 켜져 있습니다."), "Step 2 (자동 거절): \(bannerText(in: app).label)")
        let rejected = try create(externalID, caller: caller, expecting: "REJECTED", step: "Step 3 (자동 거절)")
        XCTAssertNil(ApprovalPushReceiver.payload(for: rejected, to: device, timeout: 5), "Step 3 (자동 거절): no push should be sent")
        let history = ApprovalHistoryScreen.open(in: app)
        let row = history.row(externalID)
        XCTAssertTrue(row.waitForExistence(timeout: 15), "Step 4 (자동 거절): the request should be in 처리 이력")
        XCTAssertTrue(row.label.contains("자동 거절"), "Step 4 (자동 거절): \(row.label)")
    }

    private func header(in app: XCUIApplication) -> XCUIElement {
        app.descendants(matching: .any).matching(identifier: "approvals.autoResponse.header").firstMatch
    }

    private func bannerText(in app: XCUIApplication) -> XCUIElement {
        app.staticTexts["approvals.autoResponse.banner.text"]
    }

    private func mode(_ user: BackendAPI) -> String? {
        user.call("GET", "/api/approval-settings").json?["auto_response"] as? String
    }

    private func choose(_ mode: String, in app: XCUIApplication, user: BackendAPI, step: String) {
        let settings = ApprovalKeysScreen.open(in: app)
        let option = app.buttons["approval.settings.mode.\(mode)"]
        XCTAssertTrue(option.waitForExistence(timeout: 15), "\(step): 승인 게이트 설정 should list the \(mode) option")
        XCTAssertTrue(RoutinesScreen.eventually(timeout: 15) { option.isEnabled }, "\(step): the mode options should load")
        option.tap()
        XCTAssertTrue(RoutinesScreen.eventually(timeout: 15) { option.isSelected }, "\(step): \(mode) should be selected")
        XCTAssertEqual(self.mode(user), mode, "\(step): the server should store \(mode)")
        settings.goBack()
    }

    private func create(_ externalID: String, caller: ApprovalCaller, expecting status: String, step: String) throws -> String {
        let reply = BackendAPI(token: caller.key).call(
            "POST", "/api/external/approvals",
            json: ["externalId": externalID, "context": "자동 응답 요청 \(externalID)", "timeoutSeconds": 600]
        )
        XCTAssertEqual(reply.status, 201, "\(step): creating \(externalID) should succeed: \(reply.text)")
        XCTAssertEqual(reply.json?["status"] as? String, status, "\(step): \(externalID) should be \(status) at creation")
        let id = try XCTUnwrap(reply.json?["id"] as? String, "\(step): the create reply should carry an id")
        XCTAssertEqual(ApprovalLookup.callerView(caller, id)?["status"] as? String, status, "\(step): the caller should read \(status)")
        return id
    }
}
