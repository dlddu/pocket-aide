import XCTest
@testable import PocketAideAPI

final class ApprovalsTests: XCTestCase {
    func testDecodesServerApprovalResponse() throws {
        let json = #"""
        {"id":"0f1e2d3c4b5a69788796a5b4c3d2e1f0","status":"PENDING","external_id":"deploy-42",
         "context":"prod 배포 v1.42\n변경 3건","requester_name":"Deploy Bot","key_name":"infra-ci",
         "key_revoked":false,"created_at":1760000000,"expires_at":1760000600,"processed_at":null}
        """#
        let item = try JSONDecoder().decode(ApprovalRequest.self, from: Data(json.utf8))
        XCTAssertEqual(item.status, .pending)
        XCTAssertEqual(item.externalId, "deploy-42")
        XCTAssertEqual(item.requesterName, "Deploy Bot")
        XCTAssertEqual(item.keyName, "infra-ci")
        XCTAssertFalse(item.keyRevoked)
        XCTAssertEqual(item.expiresAt, 1_760_000_600)
        XCTAssertNil(item.processedAt)
    }

    func testDecodesSessionAndReasonAndToleratesServersWithoutThem() throws {
        let json = #"""
        {"id":"r1","status":"PENDING","external_id":"x","context":"delete pod","requester_name":"","key_name":"k3s-mcp",
         "key_revoked":false,"created_at":1,"expires_at":null,"processed_at":null,
         "session_id":"0d9c8b7a-6f5e-4d3c","session_name":"배포 파이프라인 정리","reason":"재시작이 필요해서 지운다."}
        """#
        let item = try JSONDecoder().decode(ApprovalRequest.self, from: Data(json.utf8))
        XCTAssertEqual(item.sessionId, "0d9c8b7a-6f5e-4d3c")
        XCTAssertEqual(item.sessionName, "배포 파이프라인 정리")
        XCTAssertEqual(item.reason, "재시작이 필요해서 지운다.")

        let old = #"{"id":"a","status":"PENDING","external_id":"x","context":"c","requester_name":"","key_name":"k","key_revoked":false,"created_at":1,"expires_at":null,"processed_at":null}"#
        let legacy = try JSONDecoder().decode(ApprovalRequest.self, from: Data(old.utf8))
        XCTAssertNil(legacy.sessionId)
        XCTAssertNil(legacy.sessionName)
        XCTAssertNil(legacy.reason)
        XCTAssertNil(ApprovalFormatting.sessionLabel(legacy))
        XCTAssertNil(ApprovalFormatting.agentReason(legacy))
    }

    func testSessionRowLabel() {
        func item(id: String?, name: String?) -> ApprovalRequest {
            ApprovalRequest(id: "a", context: "c", sessionId: id, sessionName: name, createdAt: 1)
        }
        XCTAssertEqual(ApprovalFormatting.sessionLabel(item(id: "0d9c8b7a-6f5e", name: "배포 파이프라인 정리")), "배포 파이프라인 정리")
        XCTAssertEqual(ApprovalFormatting.sessionLabel(item(id: "0d9c8b7a-6f5e", name: "")), "제목 없음 · 0d9c8b7a")
        XCTAssertEqual(ApprovalFormatting.sessionLabel(item(id: "0d9c8b7a-6f5e", name: "  ")), "제목 없음 · 0d9c8b7a")
        XCTAssertEqual(ApprovalFormatting.sessionLabel(item(id: "s-7", name: nil)), "제목 없음 · s-7")
        XCTAssertNil(ApprovalFormatting.sessionLabel(item(id: "", name: "")))
        XCTAssertNil(ApprovalFormatting.sessionLabel(item(id: nil, name: nil)))
        XCTAssertEqual(ApprovalFormatting.sessionLabel(item(id: nil, name: "이름만")), "이름만")
    }

    func testAgentReasonShownOnlyWhenPresent() {
        XCTAssertEqual(ApprovalFormatting.agentReason(ApprovalRequest(id: "a", context: "c", reason: "  이유  ", createdAt: 1)), "이유")
        XCTAssertNil(ApprovalFormatting.agentReason(ApprovalRequest(id: "a", context: "c", reason: " \n ", createdAt: 1)))
        XCTAssertNil(ApprovalFormatting.agentReason(ApprovalRequest(id: "a", context: "c", reason: "", createdAt: 1)))
    }

    func testDecodesListWithoutDeadline() throws {
        let json = #"{"items":[{"id":"a","status":"PENDING","external_id":"x","context":"c","requester_name":"","key_name":"k","key_revoked":true,"created_at":1,"expires_at":null,"processed_at":null}]}"#
        let response = try JSONDecoder().decode(ApprovalListResponse.self, from: Data(json.utf8))
        XCTAssertEqual(response.items.count, 1)
        XCTAssertNil(response.items[0].expiresAt)
        XCTAssertTrue(response.items[0].keyRevoked)
    }

    func testNewestFirstOrdersByCreationTime() {
        let items = [
            ApprovalRequest(id: "b", context: "둘째", createdAt: 200),
            ApprovalRequest(id: "a", context: "첫째", createdAt: 100),
            ApprovalRequest(id: "c", context: "셋째", createdAt: 300),
        ]
        XCTAssertEqual(ApprovalFormatting.newestFirst(items).map(\.id), ["c", "b", "a"])
    }

    func testRemainingTime() {
        let now = Date(timeIntervalSince1970: 1000)
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: nil, now: now), "기한 없음")
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: 1252, now: now), "04:12 남음")
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: 1000 + 3600 + 61, now: now), "1:01:01 남음")
        XCTAssertEqual(ApprovalFormatting.remaining(expiresAt: 900, now: now), "00:00 남음")
    }

    func testSummaryUsesFirstNonEmptyLine() {
        XCTAssertEqual(ApprovalFormatting.summary(of: "\n  deployment/dear-baby-api 삭제 \n상세"), "deployment/dear-baby-api 삭제")
        XCTAssertEqual(ApprovalFormatting.summary(of: "한 줄"), "한 줄")
        XCTAssertEqual(ApprovalFormatting.pendingHeader(3), "대기 · 3건")
    }

    func testDecodesDecisionModeAndConflictBody() throws {
        let json = #"{"error":"already processed","approval":{"id":"r1","status":"APPROVED","external_id":"deploy-1","context":"c","requester_name":"","key_name":"infra-ci","key_revoked":false,"decision_mode":"manual","created_at":1,"expires_at":null,"processed_at":5}}"#
        let conflict = try JSONDecoder().decode(ApprovalConflictResponse.self, from: Data(json.utf8))
        XCTAssertEqual(conflict.approval.status, .approved)
        XCTAssertEqual(conflict.approval.decisionMode, "manual")
        XCTAssertEqual(conflict.approval.processedAt, 5)
        XCTAssertEqual(ApprovalFormatting.conflictNotice(conflict.approval), "이미 처리된 요청입니다 — 승인됨")
    }

    func testExpiryIsJudgedAtTheGivenMoment() {
        let item = ApprovalRequest(id: "a", context: "c", createdAt: 1000, expiresAt: 1030)
        XCTAssertFalse(item.isExpired(at: Date(timeIntervalSince1970: 1029.9)))
        XCTAssertTrue(item.isExpired(at: Date(timeIntervalSince1970: 1030)))
        XCTAssertFalse(ApprovalRequest(id: "b", context: "c", createdAt: 1).isExpired(at: Date(timeIntervalSince1970: 9_999_999)))
        XCTAssertTrue(ApprovalRequest(id: "c", status: .expired, context: "c", createdAt: 1).isExpired(at: Date(timeIntervalSince1970: 0)))
        XCTAssertFalse(ApprovalRequest(id: "d", status: .approved, context: "c", createdAt: 1, expiresAt: 2).isExpired(at: Date(timeIntervalSince1970: 10)))
    }

    func testCountdownAndProgress() {
        let now = Date(timeIntervalSince1970: 1000)
        XCTAssertEqual(ApprovalFormatting.countdown(expiresAt: 1228, now: now), "03:48")
        XCTAssertEqual(ApprovalFormatting.countdown(expiresAt: 900, now: now), "00:00")
        XCTAssertEqual(ApprovalFormatting.countdown(expiresAt: 1000 + 7322, now: now), "2:02:02")
        XCTAssertEqual(ApprovalFormatting.progress(createdAt: 900, expiresAt: 1100, now: now), 0.5, accuracy: 0.0001)
        XCTAssertEqual(ApprovalFormatting.progress(createdAt: 900, expiresAt: 950, now: now), 0)
    }

    func testStatusAndDecisionModeLabels() {
        XCTAssertEqual(ApprovalFormatting.statusLabel(.pending), "대기 중")
        XCTAssertEqual(ApprovalFormatting.statusLabel(.approved), "승인됨")
        XCTAssertEqual(ApprovalFormatting.statusLabel(.rejected), "거절됨")
        XCTAssertEqual(ApprovalFormatting.statusLabel(.expired), "만료")
        XCTAssertEqual(ApprovalFormatting.decisionModeLabel(status: .approved, mode: "manual"), "수동 결정")
        XCTAssertEqual(ApprovalFormatting.decisionModeLabel(status: .expired, mode: "expired"), "만료 — 결정 없음")
        XCTAssertEqual(ApprovalFormatting.decisionModeLabel(status: .expired, mode: nil), "만료 — 결정 없음")
        XCTAssertEqual(ApprovalFormatting.decisionModeLabel(status: .rejected, mode: "auto"), "자동 거절")
        let expired = ApprovalRequest(id: "e", status: .expired, context: "c", createdAt: 1)
        XCTAssertEqual(ApprovalFormatting.conflictNotice(expired), "만료된 요청이라 결정되지 않았습니다")
    }

    func testConfirmCopy() {
        XCTAssertEqual(ApprovalFormatting.confirmTitle(.approve), "승인할까요?")
        XCTAssertEqual(ApprovalFormatting.confirmTitle(.reject), "거절할까요?")
        XCTAssertEqual(
            ApprovalFormatting.confirmMessage(.approve, context: "deployment/dear-baby-api 삭제\n대상: prod"),
            "deployment/dear-baby-api 삭제를 허락합니다. 확정하면 되돌릴 수 없습니다."
        )
        XCTAssertEqual(
            ApprovalFormatting.confirmMessage(.reject, context: "prod 배포 v1.42"),
            "prod 배포 v1.42을(를) 거절합니다. 확정하면 되돌릴 수 없습니다."
        )
        XCTAssertEqual(ApprovalFormatting.confirmMessage(.reject, context: "캐시 비움"), "캐시 비움을 거절합니다. 확정하면 되돌릴 수 없습니다.")
    }

    func testAutoResponseSettingsRoundTripAndCopy() throws {
        let decoded = try JSONDecoder().decode(ApprovalSettings.self, from: Data(#"{"auto_response":"AUTO_REJECT"}"#.utf8))
        XCTAssertEqual(decoded.autoResponse, .autoReject)
        let encoded = try JSONEncoder().encode(ApprovalSettings(autoResponse: .autoApprove))
        XCTAssertEqual(String(decoding: encoded, as: UTF8.self), #"{"auto_response":"AUTO_APPROVE"}"#)

        XCTAssertEqual(ApprovalAutoResponse.allCases.map(\.optionLabel), ["없음 — 알림으로 받기", "자동 승인", "자동 거절"])
        XCTAssertEqual(ApprovalAutoResponse.allCases.map(\.label), ["없음", "자동 승인", "자동 거절"])
        XCTAssertNil(ApprovalAutoResponse.off.bannerTitle)
        XCTAssertEqual(ApprovalAutoResponse.autoApprove.bannerTitle, "자동 승인 모드가 켜져 있습니다.")
        XCTAssertEqual(ApprovalAutoResponse.autoReject.bannerTitle, "자동 거절 모드가 켜져 있습니다.")
        XCTAssertEqual(ApprovalAutoResponse.off.outcomeNotice, "새 요청은 알림으로 받습니다.")
        XCTAssertEqual(ApprovalAutoResponse.autoReject.outcomeNotice, "새 요청은 알림 없이 바로 거절됩니다.")
        XCTAssertEqual(ApprovalAutoResponse.autoApprove.outcomeNotice, "새 요청은 알림 없이 바로 승인됩니다.")
    }
}
