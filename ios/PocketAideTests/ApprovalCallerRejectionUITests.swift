// 검증 시나리오: test-approval-gate.md#시나리오 3
import XCTest

final class ApprovalCallerRejectionUITests: XCTestCase {
    private let createPath = "/api/external/approvals"

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    func testMalformedOrUnauthenticatedCallerRequestsAreRejected() throws {
        let owner = try XCTUnwrap(
            BackendAPI.signIn(subject: "e2e-approval-\(UUID().uuidString.lowercased())"),
            "The test runner should get user A's token"
        )
        let infra = try issueKey(named: "infra-ci", for: owner)
        let staging = try issueKey(named: "staging-ci", for: owner)
        let deploy: [String: Any] = ["externalId": "deploy-1", "context": "프로덕션 배포 승인"]

        let first = infra.call("POST", createPath, json: deploy)
        XCTAssertEqual(first.status, 201, "infra-ci should create deploy-1 first: \(first.text)")
        XCTAssertEqual(try pendingCount(for: owner), 1, "deploy-1 should be user A's only pending request")

        let missingContext = infra.call("POST", createPath, json: ["externalId": "deploy-2"])
        XCTAssertEqual(missingContext.status, 400, "Step 1: a request without context should be rejected: \(missingContext.text)")
        let legacyUser = infra.call(
            "POST",
            createPath,
            json: ["externalId": "deploy-2", "context": "프로덕션 배포 승인", "userId": "ckabc"]
        )
        XCTAssertEqual(legacyUser.status, 400, "Step 2: the old gatekeeper userId field should be rejected: \(legacyUser.text)")
        XCTAssertEqual(try pendingCount(for: owner), 1, "Steps 1-2 should not create a request")

        let duplicate = infra.call("POST", createPath, json: deploy)
        XCTAssertEqual(duplicate.status, 409, "Step 3: the same key should not reuse deploy-1: \(duplicate.text)")

        let otherKey = staging.call("POST", createPath, json: deploy)
        XCTAssertEqual(otherKey.status, 201, "Step 4: another key may use deploy-1: \(otherKey.text)")
        XCTAssertEqual(try pendingCount(for: owner), 2, "Step 4 should add exactly one request")

        let noHeader = BackendAPI.anonymous("POST", createPath, json: ["externalId": "deploy-3", "context": "c"])
        XCTAssertEqual(noHeader.status, 401, "Step 5: a request without Authorization should be rejected: \(noHeader.text)")
        let wrongKey = BackendAPI(token: "pak_wrong").call("POST", createPath, json: ["externalId": "deploy-3", "context": "c"])
        XCTAssertEqual(wrongKey.status, 401, "Step 5: a wrong key should be rejected: \(wrongKey.text)")

        let userToken = owner.call("POST", createPath, json: ["externalId": "deploy-3", "context": "c"])
        XCTAssertEqual(userToken.status, 401, "Step 6: user A's sign-in token should not open the caller route: \(userToken.text)")

        let appRoute = infra.call("GET", "/api/approvals")
        XCTAssertEqual(appRoute.status, 401, "Step 7: a caller key should not open the app's pending list: \(appRoute.text)")
        XCTAssertEqual(try pendingCount(for: owner), 2, "Steps 5-7 should not create a request")
    }

    private func issueKey(named name: String, for owner: BackendAPI) throws -> BackendAPI {
        let reply = owner.call("POST", "/api/approval-keys", json: ["name": name])
        XCTAssertEqual(reply.status, 201, "User A should issue the \(name) key: \(reply.text)")
        let key = try XCTUnwrap(reply.json?["key"] as? String, "The issue response should carry the raw key once: \(reply.text)")
        XCTAssertTrue(key.hasPrefix("pak_"), "The raw key should start with pak_: \(key.prefix(4))")
        return BackendAPI(token: key)
    }

    private func pendingCount(for owner: BackendAPI) throws -> Int {
        let reply = owner.call("GET", "/api/approvals")
        XCTAssertEqual(reply.status, 200, "User A's pending list should load: \(reply.text)")
        let items = try XCTUnwrap(reply.json?["items"] as? [[String: Any]], "The pending list should carry items: \(reply.text)")
        return items.count
    }
}
