// 검증 시나리오: test-approval-gate.md#시나리오 17
import XCTest

final class ApprovalServiceAccountRegisterUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testRegisteredServiceAccountIsListedWithoutASecretAndCannotBeTakenTwice() throws {
        let token = TodoUI.uniqueToken().lowercased()
        let name = "homelab-k3s-mcp-\(token)"
        let duplicateName = "homelab-k3s-mcp-again-\(token)"
        let malformedName = "homelab-k3s-mcp-bad-\(token)"
        let namespace = "e2e-\(token)"
        let account = "\(namespace)/homelab-k3s-mcp"
        let owner = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token")
        addTeardownBlock {
            ApprovalKeysAPI.revokeAll(owner, named: name)
            ApprovalKeysAPI.revokeAll(owner, named: duplicateName)
            ApprovalKeysAPI.revokeAll(owner, named: malformedName)
        }

        let app = TodoUI.launch()
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let screen = ApprovalKeysScreen.open(in: app)

        screen.registerServiceAccount(name: name, namespace: namespace, account: "homelab-k3s-mcp")
        XCTAssertTrue(TodoUI.waitToDisappear(screen.serviceAccountName, timeout: 15), "A successful 등록 should close the sheet without a secret step")
        XCTAssertFalse(screen.issuedTitle.exists, "Registering a service account should not show a key")
        XCTAssertTrue(screen.row(name).waitForExistence(timeout: 15), "\(name) should be listed")
        let meta = screen.meta(name).label
        XCTAssertTrue(meta.hasPrefix("SA \(account) · 등록 "), "The row should name the service account instead of a key: \(meta)")
        XCTAssertTrue(meta.hasSuffix("마지막 사용 없음"), "A fresh caller should have no last use yet: \(meta)")
        XCTAssertTrue(screen.revoke(name).exists, "The service account caller should be revocable like a key")

        let stored = try XCTUnwrap(ApprovalKeysAPI.key(owner, named: name), "The server should list \(name)")
        XCTAssertEqual(stored["kind"] as? String, "k8s_sa", "The server should store a service account caller")
        XCTAssertEqual(stored["service_account"] as? String, account, "The server should store the service account")
        XCTAssertEqual(stored["key_prefix"] as? String, "", "A service account caller has no key")

        screen.registerServiceAccount(name: duplicateName, namespace: namespace, account: "homelab-k3s-mcp")
        XCTAssertTrue(screen.serviceAccountNotice.waitForExistence(timeout: 15), "Registering the same service account again should be refused")
        XCTAssertEqual(screen.serviceAccountNotice.label, "이미 등록된 서비스 어카운트입니다")
        screen.serviceAccountCancel.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.serviceAccountName, timeout: 10), "취소 should close the sheet")
        XCTAssertFalse(screen.row(duplicateName).exists, "The refused caller should not be listed")

        screen.registerServiceAccount(name: malformedName, namespace: namespace, account: "Homelab_MCP")
        XCTAssertTrue(screen.serviceAccountNotice.waitForExistence(timeout: 15), "A name outside the DNS label rule should be refused")
        screen.serviceAccountCancel.tap()
        XCTAssertTrue(TodoUI.waitToDisappear(screen.serviceAccountName, timeout: 10), "취소 should close the sheet")
        XCTAssertFalse(screen.row(malformedName).exists, "The malformed caller should not be listed")

        let other = try XCTUnwrap(BackendAPI.signIn(subject: "e2e-approval-sa-other-\(token)"), "The test runner should get another user's token")
        let taken = other.call("POST", "/api/approval-keys", json: ["name": "mcp-\(token)", "kind": "k8s_sa", "service_account": account])
        XCTAssertEqual(taken.status, 409, "Another user should not register the same service account: \(taken.text)")
        XCTAssertTrue(taken.text.contains("service account already registered"), taken.text)

        let registered = ApprovalKeysAPI.keys(owner).filter { $0["service_account"] as? String == account }
        XCTAssertEqual(registered.count, 1, "Only one caller should hold \(account)")
        app.terminate()
    }
}
