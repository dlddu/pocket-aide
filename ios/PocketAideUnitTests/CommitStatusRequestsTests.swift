import XCTest
@testable import PocketAideAPI

final class CommitStatusRequestsTests: XCTestCase {
    func testDecodesServerListResponse() throws {
        let json = #"""
        {"items":[{"id":"c1","status":"PENDING","owner":"dlddu","repo":"pocket-aide",
          "sha":"0123456789abcdef0123456789abcdef01234567","state":"success","context":"review/manual-approval",
          "description":"수동 승인","target_url":"https://github.com/dlddu/pocket-aide/pull/1",
          "requester":"Claude Code 세션","reason":"검토 끝","session_id":"","session_name":"",
          "key_name":"homelab-k3s-mcp","key_revoked":false,"github_status_id":null,"last_error":"",
          "created_at":1760000000,"processed_at":null}],
         "history":[{"id":"c0","status":"APPROVED","owner":"dlddu","repo":"pocket-aide",
          "sha":"0123456789abcdef0123456789abcdef01234567","state":"success","context":"review/manual-approval",
          "description":"","target_url":"","requester":"r","reason":"","session_id":"","session_name":"",
          "key_name":"homelab-k3s-mcp","key_revoked":false,"github_status_id":9001,"last_error":"",
          "created_at":1759999000,"processed_at":1759999100}]}
        """#
        let list = try JSONDecoder().decode(CommitStatusRequestList.self, from: Data(json.utf8))
        XCTAssertEqual(list.items.count, 1)
        let item = list.items[0]
        XCTAssertEqual(item.status, .pending)
        XCTAssertEqual(item.targetUrl, "https://github.com/dlddu/pocket-aide/pull/1")
        XCTAssertEqual(item.keyName, "homelab-k3s-mcp")
        XCTAssertNil(item.githubStatusId)
        XCTAssertEqual(list.history.first?.githubStatusId, 9001)
        XCTAssertEqual(list.history.first?.processedAt, 1_759_999_100)
    }

    func testFormatting() {
        let item = CommitStatusRequest(id: "c1", repo: "pocket-aide", sha: "0123456789abcdef0123456789abcdef01234567",
                                       context: "review/manual-approval", createdAt: 1)
        XCTAssertEqual(CommitStatusFormatting.title(item), "review/manual-approval → success")
        XCTAssertEqual(CommitStatusFormatting.target(item), "dlddu/pocket-aide@0123456")
        XCTAssertEqual(CommitStatusFormatting.statusLabel(.approved), "승인됨 — 기록함")
        XCTAssertEqual(CommitStatusFormatting.pendingHeader(2), "대기 · 2건")
        XCTAssertTrue(CommitStatusFormatting.confirmMessage(.approve, item: item).contains("dlddu/pocket-aide@0123456"))
        XCTAssertTrue(CommitStatusFormatting.confirmMessage(.reject, item: item).contains("아무것도 기록하지 않습니다"))
    }

    func testPushOpensTheRequestAndCarriesLockScreenActions() throws {
        let url = try XCTUnwrap(PushDeepLink.url(fromUserInfo: ["aps": [:], "kind": "commit_status_request", "commit_status_request_id": "c1"]))
        XCTAssertEqual(url.absoluteString, "pocketaide://commit-status?id=c1")
        XCTAssertEqual(CommitStatusPushPayload.requestID(fromDeepLink: url), "c1")
        XCTAssertNil(CommitStatusPushPayload.requestID(from: ["commit_status_request_id": "c1"]))
        XCTAssertNil(CommitStatusPushPayload.requestID(fromDeepLink: try XCTUnwrap(URL(string: "pocketaide://approvals?id=c1"))))

        XCTAssertEqual(CommitStatusNotificationAction.categoryIdentifier, "COMMIT_STATUS_REQUEST")
        XCTAssertEqual(CommitStatusNotificationAction.decision(forActionIdentifier: CommitStatusNotificationAction.approveIdentifier), .approve)
        XCTAssertEqual(CommitStatusNotificationAction.decision(forActionIdentifier: CommitStatusNotificationAction.rejectIdentifier), .reject)
        XCTAssertNil(CommitStatusNotificationAction.decision(forActionIdentifier: ApprovalNotificationAction.approveIdentifier))
    }

    func testResultNoticeCoversEveryOutcome() {
        let approved = CommitStatusRequest(id: "c1", status: .approved, repo: "r", sha: "s", context: "review/x", createdAt: 1)
        XCTAssertEqual(CommitStatusNotificationAction.resultNotice(.success(.decided(approved))), "승인됨 — 기록함 — review/x → success")
        XCTAssertEqual(CommitStatusNotificationAction.resultNotice(.success(.conflict(approved))), "이미 처리된 요청입니다 — 승인됨 — 기록함")
        XCTAssertEqual(CommitStatusNotificationAction.resultNotice(.success(.writeFailed(approved, "GitHub answered 422"))), "GitHub 에 기록하지 못했습니다 — GitHub answered 422")
        XCTAssertEqual(CommitStatusNotificationAction.resultNotice(.success(.unavailable("x"))), CommitStatusFormatting.unavailableNotice)
        XCTAssertEqual(CommitStatusNotificationAction.resultNotice(.failure(APIError.missingBaseURL)), CommitStatusFormatting.failedDecisionNotice)
    }
}
