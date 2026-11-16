import XCTest
@testable import PocketAideAPI

final class SignOutReasonTests: XCTestCase {
    func testClientErrorReadsAsRejectionWithOAuthError() {
        let reason = SignOutReason.refreshHTTPFailure(
            status: 400,
            body: #"{"error":"invalid_grant","error_description":"Token is not active"}"#,
            at: Date(timeIntervalSince1970: 0)
        )

        XCTAssertEqual(reason.detail, "토큰 갱신 거절 — HTTP 400: invalid_grant: Token is not active")
    }

    func testServerErrorWithoutBodyReadsAsFailure() {
        let reason = SignOutReason.refreshHTTPFailure(status: 503, body: "", at: Date(timeIntervalSince1970: 0))

        XCTAssertEqual(reason.detail, "토큰 갱신 실패 — HTTP 503")
    }

    func testNonJSONBodyIsFlattenedAndCapped() {
        let body = "<html>\n  bad   gateway\n</html>" + String(repeating: "x", count: 300)

        let reason = SignOutReason.refreshHTTPFailure(status: 502, body: body, at: Date(timeIntervalSince1970: 0))

        XCTAssertTrue(reason.detail.hasPrefix("토큰 갱신 실패 — HTTP 502: <html> bad gateway </html>xxx"))
        XCTAssertTrue(reason.detail.hasSuffix("x…"))
    }

    func testUnauthorizedKeepsEarlierRefreshFailureAndItsTime() {
        let stored = SignOutReason(detail: "토큰 갱신 거절 — HTTP 400: invalid_grant", at: Date(timeIntervalSince1970: 1))

        let reason = SignOutReason.afterUnauthorized(
            stored,
            apiBody: "missing Authorization header\n",
            at: Date(timeIntervalSince1970: 5)
        )

        XCTAssertEqual(reason, SignOutReason(
            detail: "토큰 갱신 거절 — HTTP 400: invalid_grant\nAPI 응답 — HTTP 401: missing Authorization header",
            at: Date(timeIntervalSince1970: 1)
        ))
    }

    func testUnauthorizedWithoutStoredFailureUsesNow() {
        let reason = SignOutReason.afterUnauthorized(
            nil,
            apiBody: "invalid token: oidc: token is expired",
            at: Date(timeIntervalSince1970: 5)
        )

        XCTAssertEqual(reason, SignOutReason(
            detail: "API 응답 — HTTP 401: invalid token: oidc: token is expired",
            at: Date(timeIntervalSince1970: 5)
        ))
    }

    func testMessageLeadsWithLocalTime() throws {
        let reason = SignOutReason(detail: "API 응답 — HTTP 401", at: Date(timeIntervalSince1970: 1_763_253_120))

        XCTAssertEqual(
            reason.message(timeZone: try XCTUnwrap(TimeZone(identifier: "Asia/Seoul"))),
            "11월 16일 오전 9:32에 자동으로 로그아웃됐어요.\n\nAPI 응답 — HTTP 401"
        )
    }

    func testStoreRoundTripsAndClears() throws {
        let suite = "SignOutReasonTests.\(UUID().uuidString)"
        let defaults = try XCTUnwrap(UserDefaults(suiteName: suite))
        defer { defaults.removePersistentDomain(forName: suite) }
        let store = SignOutReasonStore(defaults: defaults)
        let reason = SignOutReason(detail: "토큰 갱신 실패 — HTTP 503", at: Date(timeIntervalSince1970: 7))

        store.save(reason)
        XCTAssertEqual(store.load(), reason)

        store.clear()
        XCTAssertNil(store.load())
    }
}
