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
        let data = Data(#"{"data":null,"errors":[{"type":"RATE_LIMITED","message":"API rate limit exceeded"}]}"#.utf8)
        XCTAssertThrowsError(try OpenPullRequests.decode(data)) { error in
            XCTAssertEqual(error as? GitHubError, .graphQL("API rate limit exceeded"))
        }
    }

    func testSearchQualifiersScopeToLogin() {
        let vars = OpenPullRequests.variables(login: "octo")
        XCTAssertEqual(vars["authored"], "is:pr is:open archived:false author:octo")
        XCTAssertEqual(vars["requested"], "is:pr is:open archived:false review-requested:octo")
        XCTAssertEqual(vars["reviewed"], "is:pr is:open archived:false reviewed-by:octo -author:octo")
    }
}
