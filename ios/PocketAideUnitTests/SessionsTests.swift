import XCTest
@testable import PocketAideAPI

final class SessionsTests: XCTestCase {
    func testDecodesSessionListFromProxy() throws {
        let json = #"""
        {"sessions":[
         {"id":"a1","name":"테스트 실패 원인","workloadType":"claude-code","state":"snapshot","model":"sonnet","createdAt":"2026-10-22T13:59:18.483112085Z","lastAccess":"2026-10-22T14:25:15.252261164Z","checkpoint":{"ref":"s3://x","sizeBytes":2048,"createdAt":"2026-10-22T15:30:00Z"}},
         {"id":"b2","name":"리팩터","workloadType":"claude-code","state":"active","createdAt":"2026-10-22T10:00:00Z","lastAccess":"2026-10-22T10:05:00Z"}
        ]}
        """#
        let response = try JSONDecoder().decode(AgentSessionListResponse.self, from: Data(json.utf8))
        XCTAssertEqual(response.sessions.map(\.id), ["a1", "b2"])
        XCTAssertEqual(response.sessions[0].state, .snapshot)
        XCTAssertEqual(response.sessions[0].checkpoint?.sizeBytes, 2048)
        XCTAssertEqual(response.sessions[1].modelLabel, "platform-default")
    }

    func testDecodesConfigWithAndWithoutCatalog() throws {
        let full = try JSONDecoder().decode(
            AgentSessionConfig.self,
            from: Data(#"{"claudeCode":{"defaultModel":"sonnet","models":["sonnet","opus"]}}"#.utf8)
        )
        XCTAssertEqual(full.defaultModel, "sonnet")
        XCTAssertEqual(full.models, ["sonnet", "opus"])
        let empty = try JSONDecoder().decode(AgentSessionConfig.self, from: Data(#"{"claudeCode":{"defaultModel":""}}"#.utf8))
        XCTAssertEqual(empty.models, [])
    }

    func testCreatePayloadDropsBlankModelAndNeverSendsWorkloadType() throws {
        let data = try JSONEncoder().encode(AgentSessionCreatePayload(name: "x", model: nil))
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(object["name"] as? String, "x")
        XCTAssertNil(object["model"])
        XCTAssertNil(object["workloadType"])
    }

    func testOutputAppendsFromServerCursor() {
        var output = AgentSessionOutput()
        XCTAssertEqual(output.apply(read("안녕", 6), requestedOffset: 0), .grew)
        XCTAssertEqual(output.apply(read("하세요", 15), requestedOffset: 6), .grew)
        XCTAssertEqual(output.text, "안녕하세요")
        XCTAssertEqual(output.nextOffset, 15)
        XCTAssertEqual(output.apply(read("", 15), requestedOffset: 15), .unchanged)
        XCTAssertEqual(output.text, "안녕하세요")
    }

    func testOutputAskForFullRereadWhenCursorIsAheadOfRetained() {
        var output = AgentSessionOutput()
        _ = output.apply(read("abcdef", 6), requestedOffset: 0)
        XCTAssertEqual(output.apply(read("", 3), requestedOffset: 6), .needsFullReread)
        XCTAssertEqual(output.text, "abcdef", "A stale cursor must not touch the rendered history before the full re-read")
        XCTAssertEqual(output.apply(read("abc", 3), requestedOffset: 0), .grew)
        XCTAssertEqual(output.text, "abc")
        XCTAssertEqual(output.nextOffset, 3)
    }

    func testOutputKeepsOnlyTheTailPastTheDisplayLimit() {
        var output = AgentSessionOutput()
        let big = String(repeating: "x", count: AgentSessionOutput.displayLimit)
        _ = output.apply(read(big, Int64(big.utf8.count)), requestedOffset: 0)
        XCTAssertFalse(output.isTrimmed)
        _ = output.apply(read("tail", Int64(big.utf8.count + 4)), requestedOffset: Int64(big.utf8.count))
        XCTAssertTrue(output.isTrimmed)
        XCTAssertEqual(output.text.count, AgentSessionOutput.displayLimit)
        XCTAssertTrue(output.text.hasSuffix("tail"))
    }

    func testResponsePhaseNeverClaimsCompletionOnSend() {
        let queued = AgentSessionResponsePhase.queued
        XCTAssertEqual(queued.next(outputGrew: false, secondsSinceLastOutput: 60), .queued)
        let streaming = queued.next(outputGrew: true, secondsSinceLastOutput: 0)
        XCTAssertEqual(streaming, .streaming)
        XCTAssertEqual(streaming.next(outputGrew: false, secondsSinceLastOutput: 3), .streaming)
        let quiet = streaming.next(outputGrew: false, secondsSinceLastOutput: AgentSessionResponsePhase.quietAfter)
        XCTAssertEqual(quiet, .quiet)
        XCTAssertEqual(quiet.next(outputGrew: true, secondsSinceLastOutput: 0), .streaming)
        XCTAssertNil(AgentSessionResponsePhase.none.label)
    }

    func testPromptLimitCountsUTF8Bytes() {
        let underLimit = String(repeating: "a", count: AgentSessionPrompt.maxBytes)
        XCTAssertFalse(AgentSessionPrompt.exceedsLimit(underLimit))
        let hangul = String(repeating: "가", count: AgentSessionPrompt.maxBytes / 3 + 1)
        XCTAssertTrue(AgentSessionPrompt.exceedsLimit(hangul), "3-byte characters must count by bytes, not characters")
    }

    func testErrorCategoriesNeverExposeUpstreamText() {
        let upstream = "dial tcp control-plane.session-platform.svc:80"
        XCTAssertEqual(AgentSessionError(APIError.badStatus(413, upstream)), .tooLarge)
        XCTAssertEqual(AgentSessionError(APIError.badStatus(404, upstream)), .notFound)
        XCTAssertEqual(AgentSessionError(APIError.badStatus(502, upstream)), .unreachable)
        XCTAssertEqual(AgentSessionError(APIError.badStatus(500, upstream)), .unknown)
        XCTAssertEqual(AgentSessionError(APIError.transport(URLError(.notConnectedToInternet))), .unreachable)
        let all: [AgentSessionError] = [.unauthorized, .notFound, .busy, .tooLarge, .snapshotted, .queueFull, .disabled, .quotaFull, .unreachable, .unknown]
        XCTAssertEqual(Set(all.map(\.message)).count, all.count, "Each category needs its own guidance")
        XCTAssertFalse(all.contains { $0.message.contains("session-platform") })
    }

    func testParsesNanosecondTimestamps() {
        let date = AgentSessionFormatting.date("2026-10-22T13:59:18.483112085Z")
        XCTAssertEqual(date, ISO8601DateFormatter().date(from: "2026-10-22T13:59:18Z"))
        XCTAssertNil(AgentSessionFormatting.date(""))
        XCTAssertEqual(AgentSessionFormatting.relative(nil), "—")
    }

    private func read(_ payload: String, _ next: Int64) -> AgentSessionReadResult {
        AgentSessionReadResult(session: nil, payload: payload, nextOffset: next)
    }
}
