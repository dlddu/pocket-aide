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

    func testDecodesWorkloadCatalogInPlatformOrder() throws {
        let json = #"""
        {"claudeCode":{"defaultModel":"sonnet","models":["sonnet"]},
         "workloadTypes":[
          {"type":"claude-code","label":"Claude Code","description":"agent","interaction":"agent","model":{"defaultModel":"sonnet","models":["sonnet","opus"]}},
          {"type":"minimax-code","label":"MiniMax Code","interaction":"agent","model":{"defaultModel":"platform-default","models":[]}},
          {"type":"shell","interaction":"terminal"},
          {"type":"next-kind","interaction":"hologram"}
         ]}
        """#
        let config = try JSONDecoder().decode(AgentSessionConfig.self, from: Data(json.utf8))
        XCTAssertEqual(config.workloadTypes.map(\.type), ["claude-code", "minimax-code", "shell", "next-kind"])
        let claude = try XCTUnwrap(config.type("claude-code"))
        XCTAssertEqual(claude.model?.initialPick, "sonnet")
        let minimax = try XCTUnwrap(config.type("minimax-code"))
        XCTAssertTrue(minimax.takesModel)
        XCTAssertNil(minimax.model?.initialPick, "An empty catalog means free-text input")
        let shell = try XCTUnwrap(config.type("shell"))
        XCTAssertFalse(shell.takesModel)
        XCTAssertEqual(shell.interaction, .terminal)
        XCTAssertEqual(shell.displayName, "shell", "No label falls back to the raw type")
        XCTAssertEqual(config.type("next-kind")?.interaction, .agent)
        XCTAssertNil(config.type("absent"))
    }

    func testLegacyConfigSurfacesClaudeCodeAlone() throws {
        let config = try JSONDecoder().decode(
            AgentSessionConfig.self,
            from: Data(#"{"claudeCode":{"defaultModel":"opus","models":["sonnet","opus"]}}"#.utf8)
        )
        XCTAssertEqual(config.workloadTypes.map(\.type), ["claude-code"])
        XCTAssertEqual(config.workloadTypes.first?.model?.initialPick, "opus")
    }

    func testCreatePayloadSendsTheCatalogType() throws {
        let data = try JSONEncoder().encode(AgentSessionCreatePayload(name: "x", workloadType: "minimax-code", model: nil))
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(object["workloadType"] as? String, "minimax-code")
        XCTAssertNil(object["model"])
    }

    func testKindLabelUsesCatalogLabelOrRawType() throws {
        let json = #"""
        {"sessions":[
         {"id":"s","name":"sh","workloadType":"shell","state":"active","createdAt":"2026-10-22T10:00:00Z","lastAccess":"2026-10-22T10:00:00Z"},
         {"id":"m","name":"mm","workloadType":"minimax-code","state":"idle","model":"MiniMax-M2","createdAt":"2026-10-22T10:00:00Z","lastAccess":"2026-10-22T10:00:00Z"}
        ]}
        """#
        let sessions = try JSONDecoder().decode(AgentSessionListResponse.self, from: Data(json.utf8)).sessions
        XCTAssertEqual(sessions.map(\.workloadType), ["shell", "minimax-code"])
        XCTAssertEqual(sessions[0].kindLabel(AgentSessionWorkloadType(type: "shell", label: "쉘", interaction: .terminal)), "쉘")
        XCTAssertEqual(sessions[1].kindLabel(nil), "minimax-code · MiniMax-M2", "Unknown to the catalog still shows the raw type")
    }

    func testStateDetailShowsArchiveTimeAndSize() throws {
        let now = try XCTUnwrap(AgentSessionFormatting.date("2026-10-22T12:00:00Z"))
        let archived = AgentSession(
            id: "a", name: "n", state: .snapshot,
            createdAt: "2026-10-22T08:00:00Z", lastAccess: "2026-10-22T09:00:00Z",
            checkpoint: AgentSessionCheckpoint(sizeBytes: 3_000_000, createdAt: "2026-10-22T10:00:00Z")
        )
        let detail = AgentSessionFormatting.stateDetail(archived, now: now)
        XCTAssertTrue(detail.hasPrefix("아카이브 "), detail)
        XCTAssertTrue(detail.contains("MB"), detail)
        XCTAssertTrue(detail.hasSuffix("환경 반납됨"), detail)
        XCTAssertEqual(AgentSessionState.snapshot.label, "아카이브됨")
        let running = AgentSession(id: "b", name: "n", state: .active, createdAt: "2026-10-22T11:59:30Z", lastAccess: "2026-10-22T11:59:30Z")
        XCTAssertEqual(AgentSessionFormatting.stateDetail(running, now: now), "환경이 떠 있어요 · 마지막 사용 방금")
    }

    func testTerminalCleanStripsEscapesAndCarriageReturns() {
        let raw = "\u{1B}]0;user@sess:~\u{07}\u{1B}[?2004huser@sess:~$ ls\r\n\u{1B}[01;34mdir\u{1B}[0m  a.txt\r\n\u{1B}=done"
        XCTAssertEqual(AgentSessionTerminal.clean(raw), "user@sess:~$ ls\ndir  a.txt\ndone")
        XCTAssertEqual(AgentSessionTerminal.command("ls -la"), "ls -la\n")
        XCTAssertEqual(AgentSessionTerminal.command("ls\n"), "ls\n")
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
