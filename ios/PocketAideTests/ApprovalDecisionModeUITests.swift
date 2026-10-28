// 검증 시나리오: test-approval-gate.md#시나리오 18
import XCTest

final class ApprovalDecisionModeUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testCallerResponsesCarryDecisionMode() throws {
        let owner = try XCTUnwrap(
            BackendAPI.signIn(subject: "e2e-approval-mode-\(UUID().uuidString.lowercased())"),
            "The test runner should get the user's token"
        )
        let caller = try XCTUnwrap(ApprovalSeed.issueKey(owner, name: "infra-ci"), "The user should issue the infra-ci key")
        addTeardownBlock {
            _ = owner.call("PATCH", "/api/approval-settings", json: ["auto_response": "NONE"])
            _ = owner.call("POST", "/api/approval-keys/\(caller.id)/revoke")
        }

        // Step 1: with no auto mode, the decision mode is null until the user decides.
        let manualID = try create("mode-r1", caller: caller, timeout: 600, step: "Step 1")
        let pending = try XCTUnwrap(ApprovalLookup.callerView(caller, manualID), "Step 1: the caller should read R1")
        XCTAssertEqual(pending["status"] as? String, "PENDING", "Step 1: R1 should wait for the user")
        assertNoDecisionMode(pending, step: "Step 1 before the decision")
        assertNoUserIdentity(pending, step: "Step 1 before the decision")
        let approve = owner.call("POST", "/api/approvals/\(manualID)/approve")
        XCTAssertEqual(approve.status, 200, "Step 1: the user should approve R1: \(approve.text)")
        let approved = try XCTUnwrap(ApprovalLookup.callerView(caller, manualID), "Step 1: the caller should read R1 again")
        XCTAssertEqual(approved["status"] as? String, "APPROVED", "Step 1: R1 should be APPROVED")
        XCTAssertEqual(approved["decisionMode"] as? String, "manual", "Step 1: a user decision should read manual")
        assertNoUserIdentity(approved, step: "Step 1 after the decision")

        // Step 2: with AUTO_APPROVE, the create response is already decided.
        setAutoResponse("AUTO_APPROVE", for: owner, step: "Step 2")
        let autoReply = createReply("mode-r2", caller: caller, timeout: 600, step: "Step 2")
        XCTAssertEqual(autoReply.json?["status"] as? String, "APPROVED", "Step 2: R2 should be APPROVED at creation: \(autoReply.text)")
        XCTAssertEqual(autoReply.json?["decisionMode"] as? String, "auto", "Step 2: the create response should read auto: \(autoReply.text)")
        assertNoUserIdentity(autoReply.json ?? [:], step: "Step 2 create response")
        let autoID = try XCTUnwrap(autoReply.json?["id"] as? String, "Step 2: the create response should carry an id")
        let autoView = try XCTUnwrap(ApprovalLookup.callerView(caller, autoID), "Step 2: the caller should read R2")
        XCTAssertEqual(autoView["decisionMode"] as? String, "auto", "Step 2: the lookup should read auto")

        // Step 3: with the mode off again, an unanswered request expires.
        setAutoResponse("NONE", for: owner, step: "Step 3")
        let expiringID = try create("mode-r3", caller: caller, timeout: 5, step: "Step 3")
        Thread.sleep(forTimeInterval: 6)
        let expired = try XCTUnwrap(ApprovalLookup.callerView(caller, expiringID), "Step 3: the caller should read R3")
        XCTAssertEqual(expired["status"] as? String, "EXPIRED", "Step 3: R3 should be EXPIRED after 6 seconds")
        XCTAssertEqual(expired["decisionMode"] as? String, "expired", "Step 3: an expired request should read expired")
        assertNoUserIdentity(expired, step: "Step 3")
    }

    private func createReply(_ externalID: String, caller: ApprovalCaller, timeout: Int, step: String) -> BackendReply {
        let reply = BackendAPI(token: caller.key).call(
            "POST", "/api/external/approvals",
            json: ["externalId": externalID, "context": "결정 방식 요청 \(externalID)", "timeoutSeconds": timeout]
        )
        XCTAssertEqual(reply.status, 201, "\(step): creating \(externalID) should succeed: \(reply.text)")
        return reply
    }

    private func create(_ externalID: String, caller: ApprovalCaller, timeout: Int, step: String) throws -> String {
        let reply = createReply(externalID, caller: caller, timeout: timeout, step: step)
        return try XCTUnwrap(reply.json?["id"] as? String, "\(step): the create response should carry an id")
    }

    private func setAutoResponse(_ mode: String, for owner: BackendAPI, step: String) {
        let reply = owner.call("PATCH", "/api/approval-settings", json: ["auto_response": mode])
        XCTAssertEqual(reply.status, 200, "\(step): the user should switch the auto-response mode to \(mode): \(reply.text)")
    }

    private func assertNoDecisionMode(_ body: [String: Any], step: String) {
        let value = body["decisionMode"]
        XCTAssertTrue(value == nil || value is NSNull, "\(step): decisionMode should be null, got \(String(describing: value))")
    }

    private func assertNoUserIdentity(_ body: [String: Any], step: String) {
        for field in ["sub", "email", "userId"] {
            XCTAssertNil(body[field], "\(step): the caller response should not carry \(field)")
        }
    }
}
