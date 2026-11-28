import XCTest
@testable import PocketAideAPI

final class ChatTests: XCTestCase {
    func testDecodesTheConversationWithAPendingReply() throws {
        let json = #"""
        {"conversation":{"id":7,"title":"첫 질문","created_at":1,"updated_at":2},
         "messages":[{"id":1,"role":"user","content":"안녕","created_at":1},{"id":2,"role":"assistant","content":"네","created_at":2}],
         "pending":{"content":"쓰는 중","started_at":3}}
        """#
        let thread = try JSONDecoder().decode(ChatThread.self, from: Data(json.utf8))
        XCTAssertEqual(thread.conversation.title, "첫 질문")
        XCTAssertEqual(thread.messages.map(\.isMine), [true, false])
        XCTAssertEqual(thread.pending?.content, "쓰는 중")
    }

    // PRD-1 AC1 (2026-11-27 개정): a poll's newer messages join the ones on
    // screen — the one send() already appended is not shown twice, and a reply
    // with a smaller id lands before it.
    func testMergingNewerMessagesKeepsEachIDOnceInOrder() throws {
        func messages(_ json: String) throws -> [ChatMessage] {
            try JSONDecoder().decode([ChatMessage].self, from: Data(json.utf8))
        }
        let onScreen = try messages(#"[{"id":1,"role":"user","content":"a","created_at":1},{"id":4,"role":"user","content":"d","created_at":4}]"#)
        let polled = try messages(#"[{"id":3,"role":"assistant","content":"c","created_at":3},{"id":4,"role":"user","content":"d","created_at":4}]"#)
        XCTAssertEqual(onScreen.merging(polled).map(\.id), [1, 3, 4])
        XCTAssertEqual(onScreen.merging([]).map(\.id), [1, 4])
    }

    func testDecodesAnIdleConversation() throws {
        let json = #"{"conversation":{"id":7,"title":"","created_at":1,"updated_at":1},"messages":[],"pending":null}"#
        let thread = try JSONDecoder().decode(ChatThread.self, from: Data(json.utf8))
        XCTAssertTrue(thread.messages.isEmpty)
        XCTAssertNil(thread.pending)
        XCTAssertNil(thread.queued)
    }

    // PRD-1 AC1 (2026-11-09 개정): a message sent while a reply is in progress
    // is held, surfaced here the same way a pending reply is.
    func testDecodesAQueuedMessage() throws {
        let json = #"""
        {"conversation":{"id":7,"title":"첫 질문","created_at":1,"updated_at":2},
         "messages":[],"pending":{"content":"쓰는 중","started_at":3},
         "queued":{"id":"q1","content":"다음 질문","queued_at":4}}
        """#
        let thread = try JSONDecoder().decode(ChatThread.self, from: Data(json.utf8))
        XCTAssertEqual(thread.queued?.id, "q1")
        XCTAssertEqual(thread.queued?.content, "다음 질문")
        XCTAssertEqual(thread.queued?.queuedAt, 4)
    }

    func testSendOutcomeDecodesSentOrQueued() throws {
        let sent = try JSONDecoder().decode(ChatSendResponse.self, from: Data(#"{"message":{"id":1,"role":"user","content":"hi","created_at":1}}"#.utf8))
        XCTAssertEqual(sent.message?.content, "hi")
        XCTAssertNil(sent.queued)

        let queued = try JSONDecoder().decode(ChatSendResponse.self, from: Data(#"{"queued":{"id":"q2","content":"hi","queued_at":5}}"#.utf8))
        XCTAssertNil(queued.message)
        XCTAssertEqual(queued.queued?.id, "q2")
        XCTAssertEqual(queued.queued?.content, "hi")
    }

    func testMapsErrors() {
        XCTAssertEqual(ChatError(APIError.badStatus(409, #"{"error":"a reply is still in progress"}"#)), .busy)
        XCTAssertEqual(ChatError(APIError.badStatus(409, #"{"error":"send a message first"}"#), refreshing: true), .noConversation)
        XCTAssertEqual(ChatError(APIError.badStatus(502, "")), .unreachable)
        XCTAssertEqual(ChatError(APIError.badStatus(503, "")), .unreachable)
        XCTAssertEqual(ChatError(APIError.badStatus(503, #"{"error":"all login slots are in use","code":"auth_pool_exhausted"}"#)), .slotsBusy)
        XCTAssertEqual(ChatError(APIError.badStatus(413, "")), .tooLarge)
        XCTAssertEqual(ChatError(APIError.transport(URLError(.timedOut))), .unreachable)
    }

    func testPluginRefreshNotice() throws {
        let json = #"{"state":"ready","revision":"abc","added":["p--new"],"updated":[],"removed":["p--old"]}"#
        let result = try JSONDecoder().decode(ChatPluginRefresh.self, from: Data(json.utf8))
        XCTAssertEqual(result.changedCount, 2)
        XCTAssertTrue(result.notice.contains("추가 1"))
        let none = try JSONDecoder().decode(ChatPluginRefresh.self, from: Data(#"{"state":"unconfigured","added":[],"updated":[],"removed":[]}"#.utf8))
        XCTAssertEqual(none.notice, "이 대화에는 새로 받을 스킬이 없어요.")
    }
}

final class ChatLinkTests: XCTestCase {
    private func links(_ text: String) -> [String] {
        chatLinkRanges(in: text).map { String(text[$0]) }
    }

    func testFindsTheCodexSignInAddress() {
        let text = "1. Open this link in your browser\n   https://auth.openai.com/codex/device\n2. Enter this one-time code\n   2V6W-MKPWJ"
        XCTAssertEqual(links(text), ["https://auth.openai.com/codex/device"])
    }

    func testLeavesSentencePunctuationOut() {
        XCTAssertEqual(links("see https://example.com/a?b=1. and (http://x.io/path), done"), ["https://example.com/a?b=1", "http://x.io/path"])
    }

    func testStopsAtAnUnopenedClosingBracketBeforeAParticle() {
        XCTAssertEqual(
            links("[PR #95](https://github.com/dlddu/plugin-marketplace/pull/95)를 머지했고"),
            ["https://github.com/dlddu/plugin-marketplace/pull/95"]
        )
        XCTAssertEqual(links("(https://x.io/pull/96)을 올렸습니다."), ["https://x.io/pull/96"])
    }

    func testKeepsBracketsTheAddressOpened() {
        XCTAssertEqual(links("see https://en.wikipedia.org/wiki/Foo_(bar)."), ["https://en.wikipedia.org/wiki/Foo_(bar)"])
    }

    func testIgnoresTextWithoutAddresses() {
        XCTAssertEqual(links("no links here, just https:// and words"), [])
    }

    // PRD-1 AC11: the conversation says which session answers it, which is what
    // the confirmation names — and an answer from before that field existed
    // decodes as "no session yet" rather than failing.
    func testDecodesTheSessionAnsweringTheConversation() throws {
        let named = #"{"conversation":{"id":7,"title":"q","session_id":"cx1"},"messages":[],"pending":null}"#
        XCTAssertEqual(try JSONDecoder().decode(ChatThread.self, from: Data(named.utf8)).conversation.sessionID, "cx1")

        let absent = #"{"conversation":{"id":7,"title":"q"},"messages":[],"pending":null}"#
        XCTAssertEqual(try JSONDecoder().decode(ChatThread.self, from: Data(absent.utf8)).conversation.sessionID, "")
    }

    // The result is read back from the platform, so the archived case says the
    // environment is handed back and names when and how big — and the accepted
    // case says only that, without claiming the environment is already back.
    func testArchiveResultSaysArchivedOrOnlyAccepted() throws {
        let archived = try JSONDecoder().decode(
            ChatArchiveResult.self,
            from: Data(#"{"session_id":"cx1","state":"snapshot","archived":true,"checkpoint":{"createdAt":"2026-11-28T06:00:00Z","sizeBytes":2048}}"#.utf8)
        )
        XCTAssertEqual(archived.sessionID, "cx1")
        XCTAssertEqual(archived.state, "snapshot")
        XCTAssertTrue(archived.archived)
        XCTAssertEqual(archived.checkpoint?.sizeBytes, 2048)
        XCTAssertTrue(archived.notice.contains("아카이브됐어요"))
        // The conversation is not what was handed back.
        XCTAssertTrue(archived.notice.contains("복원"))

        let accepted = try JSONDecoder().decode(ChatArchiveResult.self, from: Data(#"{"session_id":"cx1","archived":false}"#.utf8))
        XCTAssertFalse(accepted.archived)
        XCTAssertTrue(accepted.notice.contains("요청했어요"))
        XCTAssertFalse(accepted.notice.contains("아카이브됐어요"))
    }

    // Each refusal the app explains differently is read by code, because two
    // share 409 and two share 503.
    func testArchiveRefusalsAreToldApartByCode() {
        let cases: [(Int, String, ChatArchiveError)] = [
            (409, #"{"code":"chat_archive_no_session"}"#, .noSession),
            (409, #"{"code":"chat_archive_session_changed"}"#, .sessionChanged),
            (409, #"{"code":"chat_archive_queued"}"#, .queued),
            (409, #"{"code":"chat_archive_busy"}"#, .busy),
            (404, #"{"code":"chat_archive_session_gone"}"#, .gone),
            (503, #"{"code":"chat_archive_disabled"}"#, .disabled),
            (503, #"{"error":"auth pool exhausted","code":"auth_pool_exhausted"}"#, .slotsBusy),
            (503, #"{"error":"down"}"#, .unreachable),
            (500, #"{"error":"nope"}"#, .failed),
        ]
        for (status, body, want) in cases {
            XCTAssertEqual(ChatArchiveError(APIError.badStatus(status, body)), want, "\(status) \(body)")
        }

        // The gate being off must not be told as a loss (PRD-11 AC8).
        XCTAssertTrue(ChatArchiveError.disabled.message.contains("사라지지 않아요"))
        // Only the two that mean the screen's session is stale re-read.
        for refusal in [ChatArchiveError.sessionChanged, .gone] {
            XCTAssertTrue(refusal.rereadsConversation, "\(refusal)")
        }
        for refusal in [ChatArchiveError.noSession, .queued, .busy, .disabled, .slotsBusy, .unreachable, .failed] {
            XCTAssertFalse(refusal.rereadsConversation, "\(refusal)")
        }
    }

    func testArchivePayloadNamesTheSession() throws {
        let raw = try JSONEncoder().encode(ChatArchivePayload(sessionID: "cx1"))
        XCTAssertEqual(String(data: raw, encoding: .utf8), #"{"session_id":"cx1"}"#)
    }
}
