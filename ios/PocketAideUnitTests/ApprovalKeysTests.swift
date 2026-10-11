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
}
