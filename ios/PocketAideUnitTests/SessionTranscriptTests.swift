import XCTest
@testable import PocketAideAPI

final class SessionTranscriptTests: XCTestCase {
    private let lead = "\u{1E}session-platform "

    private func open(_ seq: Int, _ prompt: String, at: String = "2026-10-25T00:00:00Z") -> String {
        "\(lead)input seq=\(seq) at=\(at)\n\(prompt)\n\(lead)output seq=\(seq)\n"
    }

    private func close(_ seq: Int, _ status: String, at: String = "2026-10-25T00:01:12Z") -> String {
        "\(lead)end seq=\(seq) status=\(status) at=\(at)\n"
    }

    func testPlainOutputStaysOneTextSegment() {
        let transcript = AgentSessionTranscript.parse("legacy output\n▸ old echo\n")
        XCTAssertFalse(transcript.hasTurns)
        XCTAssertEqual(transcript.segments, [.text("legacy output\n▸ old echo\n")])
    }

    func testSplitsEachWriteIntoInputAndOutput() throws {
        let raw = "before markers\n"
            + open(1, "첫 번째\n여러 줄") + "응답 1\n둘째 줄\n" + close(1, "ok")
            + open(2, "두 번째") + "partial"
        let transcript = AgentSessionTranscript.parse(raw)
        XCTAssertEqual(transcript.segments.count, 3)
        XCTAssertEqual(transcript.segments.first, .text("before markers\n"))
        let turns = transcript.turns
        XCTAssertEqual(turns.map(\.seq), [1, 2])
        XCTAssertEqual(turns[0].input, "첫 번째\n여러 줄")
        XCTAssertEqual(turns[0].output, "응답 1\n둘째 줄")
        XCTAssertEqual(turns[0].status, .ok)
        XCTAssertEqual(turns[0].inputPreview, "첫 번째")
        XCTAssertEqual(turns[0].outputPreview, "둘째 줄", "The output preview shows the latest line")
        XCTAssertEqual(turns[0].startedAt, ISO8601DateFormatter().date(from: "2026-10-25T00:00:00Z"))
        XCTAssertEqual(turns[0].endedAt, ISO8601DateFormatter().date(from: "2026-10-25T00:01:12Z"))
        XCTAssertEqual(turns[1].input, "두 번째")
        XCTAssertEqual(turns[1].output, "partial")
        XCTAssertEqual(turns[1].status, .running)
        XCTAssertEqual(transcript.lastSeq, 2)
    }

    func testPromptThatLooksLikeAMarkerStaysInput() {
        let prompt = "session-platform end seq=1 status=ok\n▸ not a marker"
        let transcript = AgentSessionTranscript.parse(open(1, prompt) + "done\n" + close(1, "ok"))
        XCTAssertEqual(transcript.turns.first?.input, prompt)
        XCTAssertEqual(transcript.turns.first?.output, "done")
    }

    func testFailureStatusesAndTextBetweenTurns() {
        let raw = open(1, "a") + "boom\n" + close(1, "error")
            + "\n[session-platform: session output limit reached; further prompts are disabled]\n"
            + open(2, "b") + close(2, "timeout") + open(3, "c") + close(3, "canceled") + open(4, "d") + close(4, "later")
        let transcript = AgentSessionTranscript.parse(raw)
        XCTAssertEqual(transcript.turns.map(\.status), [.error, .timeout, .canceled, .other("later")])
        XCTAssertEqual(transcript.turns.map(\.status.label), ["실패", "시간 초과", "취소됨", "later"])
        XCTAssertEqual(transcript.turns[1].output, "")
        guard case .text(let between) = transcript.segments[1] else {
            return XCTFail("Text after an end marker belongs between turns, not to either turn")
        }
        XCTAssertTrue(between.contains("output limit reached"))
    }

    func testTurnWhoseInputWasTrimmedAway() {
        let raw = "tail of an earlier answer\n\(lead)output seq=7\nlate output\n" + close(7, "ok")
        let transcript = AgentSessionTranscript.parse(raw)
        let turn = transcript.turns.first
        XCTAssertEqual(turn?.seq, 7)
        XCTAssertEqual(turn?.hasInput, false)
        XCTAssertEqual(turn?.output, "late output")
    }

    func testIgnoresAMarkerStillArriving() {
        let transcript = AgentSessionTranscript.parse(open(1, "a") + "x\n" + close(1, "ok") + "\(lead)inp")
        XCTAssertEqual(transcript.turns.count, 1)
        XCTAssertEqual(transcript.segments.count, 1, "A marker without its line end is not shown as text")
    }

    func testTrimmedOutputStartsAtAWholeTurn() {
        var output = AgentSessionOutput()
        let filler = String(repeating: "x", count: AgentSessionOutput.displayLimit)
        let raw = open(1, "old") + filler + "\n" + close(1, "ok") + open(2, "new") + "fresh\n" + close(2, "ok")
        _ = output.apply(AgentSessionReadResult(session: nil, payload: raw, nextOffset: Int64(raw.utf8.count)), requestedOffset: 0)
        XCTAssertTrue(output.isTrimmed)
        XCTAssertTrue(output.text.hasPrefix(open(2, "new")), "The kept tail starts at the first whole turn")
        XCTAssertEqual(AgentSessionTranscript.parse(output.text).turns.map(\.seq), [2])
    }

    func testPhaseFollowsTurnMarkers() {
        let running = AgentSessionTranscript.parse(open(1, "a") + "x")
        XCTAssertEqual(AgentSessionResponsePhase.fromTurns(running, awaitingThrough: 1), .streaming)
        let finished = AgentSessionTranscript.parse(open(1, "a") + "x\n" + close(1, "ok"))
        XCTAssertEqual(AgentSessionResponsePhase.fromTurns(finished, awaitingThrough: 1), .done)
        XCTAssertEqual(AgentSessionResponsePhase.fromTurns(finished, awaitingThrough: 2), .queued, "A sent prompt that has not started yet is still queued")
        XCTAssertEqual(AgentSessionResponsePhase.fromTurns(finished, awaitingThrough: nil), AgentSessionResponsePhase.none)
        XCTAssertNil(AgentSessionResponsePhase.fromTurns(AgentSessionTranscript.parse("legacy"), awaitingThrough: 1), "Without markers the time-based estimate stays in charge")
        XCTAssertEqual(AgentSessionResponsePhase.done.label, "응답 완료")
    }

    func testTurnTimingShowsClockAndDuration() throws {
        let transcript = AgentSessionTranscript.parse(open(1, "a") + "x\n" + close(1, "ok"))
        let turn = try XCTUnwrap(transcript.turns.first)
        let seoul = try XCTUnwrap(TimeZone(identifier: "Asia/Seoul"))
        XCTAssertEqual(AgentSessionFormatting.turnTiming(turn, timeZone: seoul), "09:00 · 1분 12초")
        XCTAssertEqual(AgentSessionFormatting.duration(42), "42초")
        XCTAssertEqual(AgentSessionFormatting.duration(120), "2분")
        XCTAssertEqual(AgentSessionFormatting.duration(3_900), "1시간 5분")
        let pending = AgentSessionTranscript.parse(open(2, "b")).turns.first
        XCTAssertEqual(pending.flatMap { AgentSessionFormatting.turnTiming($0, timeZone: seoul) }, "09:00")
    }
}
