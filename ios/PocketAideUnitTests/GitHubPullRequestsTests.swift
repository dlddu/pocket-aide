import XCTest
@testable import PocketAideAPI

final class GitHubPullRequestsTests: XCTestCase {
    private func node(_ repo: String, _ number: Int, updated: String, author: String = "octo", rollup: String? = nil) -> String {
        let state = rollup.map { #"{"state":"\#($0)"}"# } ?? "null"
        return #"{"number":\#(number),"title":"PR \#(number)","url":"https:\/\/github.com\/\#(repo)/pull/\#(number)","updatedAt":"\#(updated)","repository":{"nameWithOwner":"\#(repo)"},"author":{"login":"\#(author)"},"commits":{"nodes":[{"commit":{"statusCheckRollup":\#(state)}}]}}"#
    }

    private func response(authored: [String], requested: [String] = [], reviewed: [String] = [], errors: String = "") -> Data {
        let body = #"{"data":{"authored":{"nodes":[\#(authored.joined(separator: ","))]},"requested":{"nodes":[\#(requested.joined(separator: ","))]},"reviewed":{"nodes":[\#(reviewed.joined(separator: ","))]}}\#(errors)}"#
        return Data(body.utf8)
    }

    func testMapsRollupStatesToFourStatuses() {
        XCTAssertEqual(PullRequestCIStatus(rollupState: "SUCCESS"), .success)
        XCTAssertEqual(PullRequestCIStatus(rollupState: "FAILURE"), .failure)
        XCTAssertEqual(PullRequestCIStatus(rollupState: "ERROR"), .failure)
        XCTAssertEqual(PullRequestCIStatus(rollupState: "PENDING"), .inProgress)
        XCTAssertEqual(PullRequestCIStatus(rollupState: "EXPECTED"), .inProgress)
        XCTAssertEqual(PullRequestCIStatus(rollupState: nil), .noChecks)
        XCTAssertEqual(
            [PullRequestCIStatus.inProgress, .success, .failure, .noChecks].map(\.label),
            ["진행 중", "성공", "실패", "상태 없음"]
        )
    }

    func testMergesRolesAndSortsByUpdatedDescending() throws {
        let data = response(
            authored: [node("me/app", 1, updated: "2026-10-01T08:00:00Z", author: "me", rollup: "SUCCESS")],
            requested: [node("team/api", 7, updated: "2026-10-01T09:00:00Z", rollup: "PENDING")],
            reviewed: [node("team/web", 3, updated: "2026-09-30T23:00:00Z", rollup: "FAILURE")]
        )
        let result = try OpenPullRequests.decode(data)
        XCTAssertEqual(result.pullRequests.map(\.id), ["team/api#7", "me/app#1", "team/web#3"])
        XCTAssertEqual(result.pullRequests.map(\.role), [.reviewer, .author, .reviewer])
        XCTAssertEqual(result.pullRequests.map(\.ciStatus), [.inProgress, .success, .failure])
        XCTAssertEqual(result.pullRequests[1].author, "me")
        XCTAssertEqual(result.pullRequests[1].url?.absoluteString, "https://github.com/me/app/pull/1")
        XCTAssertEqual(result.inaccessibleCount, 0)
    }

    func testAuthorRoleWinsWhenAlsoReviewRequested() throws {
        let shared = node("me/app", 5, updated: "2026-10-01T08:00:00Z")
        let result = try OpenPullRequests.decode(response(authored: [shared], requested: [shared]))
        XCTAssertEqual(result.pullRequests.count, 1)
        XCTAssertEqual(result.pullRequests.first?.role, .author)
    }

    func testSkipsNullNodesFromPartialErrorsAndCountsThem() throws {
        let errors = #","errors":[{"type":"FORBIDDEN","path":["authored","nodes",1],"message":"Resource protected by organization SAML enforcement."}]"#
        let data = response(
            authored: [node("me/app", 1, updated: "2026-10-01T08:00:00Z"), "null"],
            errors: errors
        )
        let result = try OpenPullRequests.decode(data)
        XCTAssertEqual(result.pullRequests.map(\.id), ["me/app#1"])
        XCTAssertEqual(result.inaccessibleCount, 1)
        XCTAssertEqual(result.pullRequests.first?.ciStatus, .noChecks)
    }

    func testEmptySearchYieldsEmptyList() throws {
        let result = try OpenPullRequests.decode(response(authored: []))
        XCTAssertTrue(result.pullRequests.isEmpty)
        XCTAssertEqual(result.inaccessibleCount, 0)
    }

    func testMissingDataSurfacesGraphQLError() {
        let data = Data(#"{"data":null,"errors":[{"type":"INTERNAL","message":"Something went wrong"}]}"#.utf8)
        XCTAssertThrowsError(try OpenPullRequests.decode(data)) { error in
            XCTAssertEqual(error as? GitHubError, .graphQL("Something went wrong"))
        }
    }

    func testGraphQLRateLimitErrorBecomesRateLimited() {
        let data = Data(#"{"data":null,"errors":[{"type":"RATE_LIMITED","message":"API rate limit exceeded"}]}"#.utf8)
        XCTAssertThrowsError(try OpenPullRequests.decode(data)) { error in
            XCTAssertEqual(error as? GitHubError, .rateLimited(resetAt: nil))
        }
    }

    func testFiltersNarrowListAndKeepUpdatedOrder() throws {
        let data = response(
            authored: [
                node("me/app", 1, updated: "2026-10-01T08:00:00Z", author: "me", rollup: "FAILURE"),
                node("me/lib", 2, updated: "2026-10-01T06:00:00Z", author: "me", rollup: "SUCCESS"),
            ],
            requested: [node("team/api", 7, updated: "2026-10-01T09:00:00Z", rollup: "FAILURE")],
            reviewed: [node("team/web", 3, updated: "2026-10-01T07:00:00Z", rollup: "PENDING")]
        )
        let all = try OpenPullRequests.decode(data).pullRequests
        XCTAssertEqual(OpenPullRequestFilter.all.apply(to: all).map(\.id), ["team/api#7", "me/app#1", "team/web#3", "me/lib#2"])
        XCTAssertEqual(OpenPullRequestFilter.mine.apply(to: all).map(\.id), ["me/app#1", "me/lib#2"])
        XCTAssertEqual(OpenPullRequestFilter.reviewRequested.apply(to: all).map(\.id), ["team/api#7", "team/web#3"])
        XCTAssertEqual(OpenPullRequestFilter.ciFailing.apply(to: all).map(\.id), ["team/api#7", "me/app#1"])
    }

    func testFilterRoundTripsThroughRawValue() {
        XCTAssertEqual(OpenPullRequestFilter.allCases.map(\.label), ["전체", "내 PR만", "리뷰 요청만", "CI 실패만"])
        for filter in OpenPullRequestFilter.allCases {
            XCTAssertEqual(OpenPullRequestFilter(rawValue: filter.rawValue), filter)
        }
    }

    func testClassifiesHTTPFailuresByCause() {
        let now = Date(timeIntervalSince1970: 1_000)
        XCTAssertNil(GitHubError.classify(statusCode: 200))
        XCTAssertEqual(GitHubError.classify(statusCode: 401), .unauthorized)
        XCTAssertEqual(GitHubError.classify(statusCode: 403, remaining: "12"), .forbidden)
        XCTAssertEqual(
            GitHubError.classify(statusCode: 403, remaining: "0", reset: "1791085411"),
            .rateLimited(resetAt: Date(timeIntervalSince1970: 1_791_085_411))
        )
        XCTAssertEqual(
            GitHubError.classify(statusCode: 403, retryAfter: "60", now: now),
            .rateLimited(resetAt: Date(timeIntervalSince1970: 1_060))
        )
        XCTAssertEqual(GitHubError.classify(statusCode: 429), .rateLimited(resetAt: nil))
        XCTAssertEqual(GitHubError.classify(statusCode: 502), .badStatus(502))
    }

    func testAlertNamesCauseAndResolution() {
        XCTAssertEqual(GitHubAlert.make(error: .unauthorized, inaccessibleCount: 0), .tokenRejected)
        XCTAssertEqual(GitHubAlert.make(error: .forbidden, inaccessibleCount: 0), .insufficientAccess(hiddenCount: 0))
        XCTAssertEqual(GitHubAlert.make(error: nil, inaccessibleCount: 2), .insufficientAccess(hiddenCount: 2))
        XCTAssertEqual(GitHubAlert.make(error: .rateLimited(resetAt: nil), inaccessibleCount: 2), .rateLimited(resetAt: nil))
        XCTAssertNil(GitHubAlert.make(error: nil, inaccessibleCount: 0))
        XCTAssertNil(GitHubAlert.make(error: .transport("offline"), inaccessibleCount: 0))
        XCTAssertEqual(
            [GitHubAlert.tokenRejected, .insufficientAccess(hiddenCount: 1), .rateLimited(resetAt: nil)].map(\.actionTitle),
            ["토큰 다시 연결", "토큰 바꾸기", "다시 시도"]
        )
        let reset = Date(timeIntervalSince1970: 0)
        XCTAssertEqual(GitHubAlert.rateLimited(resetAt: reset).message { _ in "09:30" }, "09:30 이후에 다시 시도하세요.")
        XCTAssertTrue(GitHubAlert.insufficientAccess(hiddenCount: 3).message { _ in "" }.hasPrefix("PR 3개는"))
    }

    func testSearchQualifiersScopeToLogin() {
        let vars = OpenPullRequests.variables(login: "octo")
        XCTAssertEqual(vars["authored"], "is:pr is:open archived:false author:octo")
        XCTAssertEqual(vars["requested"], "is:pr is:open archived:false review-requested:octo")
        XCTAssertEqual(vars["reviewed"], "is:pr is:open archived:false reviewed-by:octo -author:octo")
    }
}
