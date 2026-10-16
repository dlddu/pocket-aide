import XCTest
@testable import PocketAideAPI

final class ApprovalKeysTests: XCTestCase {
    func testDecodesKeyListWithRevokedAndUnusedKeys() throws {
        let json = #"""
        {"keys":[
         {"id":3,"name":"old-ci","key_prefix":"pak_b3KdQ1zX","created_at":1759000000,"last_used_at":1759400000,"revoked_at":1759500000},
         {"id":7,"name":"infra-ci","key_prefix":"pak_Xm91aB2c","created_at":1760000000,"last_used_at":null,"revoked_at":null}
        ]}
        """#
        let response = try JSONDecoder().decode(ApprovalKeyListResponse.self, from: Data(json.utf8))
        XCTAssertEqual(response.keys.map(\.name), ["old-ci", "infra-ci"])
        XCTAssertTrue(response.keys[0].isRevoked)
        XCTAssertFalse(response.keys[1].isRevoked)
        XCTAssertNil(response.keys[1].lastUsedAt)
        XCTAssertEqual(response.keys[1].keyPrefix, "pak_Xm91aB2c")
    }

    func testDecodesIssuedKeyWithSecretOnce() throws {
        let json = #"{"id":9,"name":"session-agent","key_prefix":"pak_Rk4vN8sQ","created_at":1760000000,"last_used_at":null,"revoked_at":null,"key":"pak_Rk4vN8sQ2mTzW6yLp0eHc9JdXb7AuF3G"}"#
        let issued = try JSONDecoder().decode(IssuedApprovalKey.self, from: Data(json.utf8))
        XCTAssertEqual(issued.key.id, 9)
        XCTAssertEqual(issued.key.name, "session-agent")
        XCTAssertEqual(issued.secret, "pak_Rk4vN8sQ2mTzW6yLp0eHc9JdXb7AuF3G")
        XCTAssertTrue(issued.secret.hasPrefix(issued.key.keyPrefix))
    }

    func testOrdersActiveKeysNewestFirstThenRevoked() {
        let keys = [
            ApprovalCallerKey(id: 1, name: "a", keyPrefix: "pak_a", createdAt: 100, revokedAt: 400),
            ApprovalCallerKey(id: 2, name: "b", keyPrefix: "pak_b", createdAt: 200),
            ApprovalCallerKey(id: 3, name: "c", keyPrefix: "pak_c", createdAt: 300),
            ApprovalCallerKey(id: 4, name: "d", keyPrefix: "pak_d", createdAt: 350, revokedAt: 500)
        ]
        XCTAssertEqual(ApprovalKeyFormatting.ordered(keys).map(\.name), ["c", "b", "d", "a"])
    }

    func testMetaShowsPrefixIssueDayAndLastUse() {
        let now = Date(timeIntervalSince1970: 1_760_000_000)
        let unused = ApprovalCallerKey(id: 1, name: "infra-ci", keyPrefix: "pak_Xm91aB2c", createdAt: 1_760_000_000)
        let unusedMeta = ApprovalKeyFormatting.meta(unused, now: now)
        XCTAssertTrue(unusedMeta.hasPrefix("pak_Xm91aB2c… · 발급 "), unusedMeta)
        XCTAssertTrue(unusedMeta.hasSuffix(" · 마지막 사용 없음"), unusedMeta)

        let used = ApprovalCallerKey(id: 1, name: "infra-ci", keyPrefix: "pak_Xm91aB2c", createdAt: 1_759_000_000, lastUsedAt: 1_760_000_000)
        let usedMeta = ApprovalKeyFormatting.meta(used, now: now)
        XCTAssertNotNil(usedMeta.range(of: #"마지막 사용 \d{2}:\d{2}$"#, options: .regularExpression), usedMeta)
        XCTAssertFalse(usedMeta.contains("Rk4v"))
    }

    func testRevokedLabelAndTrimmedName() {
        let active = ApprovalCallerKey(id: 1, name: "k", keyPrefix: "pak_k", createdAt: 1)
        XCTAssertNil(ApprovalKeyFormatting.revokedLabel(active))
        let revoked = ApprovalCallerKey(id: 1, name: "k", keyPrefix: "pak_k", createdAt: 1, revokedAt: 1_759_500_000)
        XCTAssertEqual(ApprovalKeyFormatting.revokedLabel(revoked)?.hasPrefix("폐기됨 "), true)
        XCTAssertEqual(ApprovalKeyFormatting.trimmedName("  infra-ci \n"), "infra-ci")
        XCTAssertNil(ApprovalKeyFormatting.trimmedName("   "))
        XCTAssertEqual(ApprovalKeyFormatting.revokeTitle(active), "k 키를 폐기할까요?")
    }

    func testDecodesServiceAccountCallersAlongsideKeysWithOrWithoutTheField() throws {
        let json = #"""
        {"keys":[
         {"id":11,"name":"homelab-k3s-mcp","kind":"k8s_sa","key_prefix":"","service_account":"homelab-k3s-mcp/homelab-k3s-mcp","created_at":1760500000,"last_used_at":null,"revoked_at":null},
         {"id":7,"name":"infra-ci","kind":"key","key_prefix":"pak_Xm91aB2c","service_account":null,"created_at":1760000000,"last_used_at":null,"revoked_at":null},
         {"id":3,"name":"old-ci","key_prefix":"pak_b3KdQ1zX","created_at":1759000000,"last_used_at":null,"revoked_at":1759500000}
        ]}
        """#
        let response = try JSONDecoder().decode(ApprovalKeyListResponse.self, from: Data(json.utf8))
        XCTAssertEqual(response.keys[0].serviceAccount, "homelab-k3s-mcp/homelab-k3s-mcp")
        XCTAssertTrue(response.keys[0].isServiceAccount)
        XCTAssertEqual(response.keys[0].keyPrefix, "")
        XCTAssertNil(response.keys[1].serviceAccount)
        XCTAssertFalse(response.keys[1].isServiceAccount)
        XCTAssertNil(response.keys[2].serviceAccount)
        XCTAssertFalse(response.keys[2].isServiceAccount)
    }

    func testServiceAccountMetaNamesTheAccountInsteadOfAKeyPrefix() {
        let now = Date(timeIntervalSince1970: 1_760_000_000)
        let unused = ApprovalCallerKey(id: 1, name: "mcp", keyPrefix: "", createdAt: 1_760_000_000, serviceAccount: "homelab-k3s-mcp/homelab-k3s-mcp")
        let unusedMeta = ApprovalKeyFormatting.meta(unused, now: now)
        XCTAssertTrue(unusedMeta.hasPrefix("SA homelab-k3s-mcp/homelab-k3s-mcp · 등록 "), unusedMeta)
        XCTAssertTrue(unusedMeta.hasSuffix(" · 마지막 사용 없음"), unusedMeta)
        XCTAssertFalse(unusedMeta.contains("…"), unusedMeta)

        let used = ApprovalCallerKey(id: 2, name: "runner", keyPrefix: "", createdAt: 1_759_000_000, lastUsedAt: 1_760_000_000, serviceAccount: "ci/runner")
        let usedMeta = ApprovalKeyFormatting.meta(used, now: now)
        XCTAssertNotNil(usedMeta.range(of: #"^SA ci/runner · 등록 \d{2}/\d{2} · 마지막 사용 \d{2}:\d{2}$"#, options: .regularExpression), usedMeta)
    }

    func testRevokeMessageFollowsTheCallerKind() {
        let key = ApprovalCallerKey(id: 1, name: "k", keyPrefix: "pak_k", createdAt: 1)
        let account = ApprovalCallerKey(id: 2, name: "sa", keyPrefix: "", createdAt: 1, serviceAccount: "ci/runner")
        XCTAssertEqual(ApprovalKeyFormatting.revokeMessage(for: key), ApprovalKeyFormatting.revokeMessage)
        XCTAssertEqual(ApprovalKeyFormatting.revokeMessage(for: account), ApprovalKeyFormatting.serviceAccountRevokeMessage)
    }

    func testServiceAccountJoinsTwoDNSLabelsLikeTheServer() {
        XCTAssertEqual(ApprovalKeyFormatting.serviceAccount(namespace: " homelab-k3s-mcp ", name: "homelab-k3s-mcp\n"), "homelab-k3s-mcp/homelab-k3s-mcp")
        XCTAssertEqual(ApprovalKeyFormatting.serviceAccount(namespace: "a", name: "b"), "a/b")
        let longest = String(repeating: "n", count: 63)
        XCTAssertEqual(ApprovalKeyFormatting.serviceAccount(namespace: longest, name: longest), "\(longest)/\(longest)")
        let invalid: [(String, String)] = [
            ("", "runner"), ("ci", ""), ("CI", "runner"), ("ci", "Runner"), ("-ci", "runner"), ("ci-", "runner"),
            ("ci", "runner-"), ("ci", "run.ner"), ("c i", "runner"), ("ci/x", "runner"), ("ci", "runner/x"),
            (String(repeating: "n", count: 64), "runner"), ("ci", String(repeating: "s", count: 64))
        ]
        for (namespace, name) in invalid {
            XCTAssertNil(ApprovalKeyFormatting.serviceAccount(namespace: namespace, name: name), "\(namespace)/\(name)")
        }
    }

    func testServiceAccountPayloadAsksForAK8sCaller() throws {
        let data = try JSONEncoder().encode(ApprovalServiceAccountPayload(name: "mcp", serviceAccount: "ci/runner"))
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: String])
        XCTAssertEqual(object, ["name": "mcp", "kind": "k8s_sa", "service_account": "ci/runner"])
    }
}
