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

    // PRD-AI채팅.AC1 (2026-11-28 개정): a first read brings back the newest page
    // and says that the conversation goes on before it, with the cursor to ask
    // for that page by.
    func testDecodesTheNewestPageWithItsBackwardCursor() throws {
        let json = #"""
        {"conversation":{"id":7,"title":"첫 질문","created_at":1,"updated_at":2},
         "messages":[{"id":11,"role":"user","content":"열한째","created_at":11}],
         "pending":null,"has_more":true,"prev_cursor":11}
        """#
        let thread = try JSONDecoder().decode(ChatThread.self, from: Data(json.utf8))
        XCTAssertEqual(thread.hasMore, true)
        XCTAssertEqual(thread.prevCursor, 11)
    }

    // An incremental read carries neither field, and a page with nothing before
    // it carries neither either — both read as "no page to offer" rather than
    // as a cursor of zero.
    func testAReadWithoutABackwardCursorSaysSo() throws {
        let incremental = #"""
        {"conversation":{"id":7,"title":"첫 질문","created_at":1,"updated_at":2},
         "messages":[],"pending":null}
        """#
        let thread = try JSONDecoder().decode(ChatThread.self, from: Data(incremental.utf8))
        XCTAssertNil(thread.hasMore)
        XCTAssertNil(thread.prevCursor)
        XCTAssertEqual(thread.hasMore ?? false, false)
        XCTAssertEqual(thread.prevCursor ?? 0, 0)
    }

    // AC2: the previous page is its own answer — messages and a cursor, with no
    // reply in progress and no waiting message to overwrite what is on screen.
    func testDecodesAPageOfThePast() throws {
        let json = #"""
        {"messages":[{"id":9,"role":"user","content":"아홉째","created_at":9},
                     {"id":10,"role":"assistant","content":"열째","created_at":10}],
         "has_more":true,"prev_cursor":9}
        """#
        let page = try JSONDecoder().decode(ChatMessagePage.self, from: Data(json.utf8))
        XCTAssertEqual(page.messages.map(\.id), [9, 10])
        XCTAssertEqual(page.hasMore, true)
        XCTAssertEqual(page.prevCursor, 9)

        let start = try JSONDecoder().decode(
            ChatMessagePage.self,
            from: Data(#"{"messages":[{"id":1,"role":"user","content":"첫째","created_at":1}]}"#.utf8)
        )
        XCTAssertEqual(start.hasMore, nil)
    }

    // AC2 · AC5: a page of the past goes in front of what is held, in order —
    // and applying the same page again (the retry of a read whose answer never
    // arrived) leaves the conversation exactly once.
    func testAPageOfThePastGoesInFrontAndSurvivesARetry() throws {
        func messages(_ json: String) throws -> [ChatMessage] {
            try JSONDecoder().decode([ChatMessage].self, from: Data(json.utf8))
        }
        let held = try messages(#"[{"id":11,"role":"user","content":"열한째","created_at":11}]"#)
        let older = try messages(#"[{"id":9,"role":"user","content":"아홉째","created_at":9},{"id":10,"role":"assistant","content":"열째","created_at":10}]"#)
        XCTAssertEqual(held.merging(older).map(\.id), [9, 10, 11])
        XCTAssertEqual(held.merging(older).merging(older).map(\.id), [9, 10, 11])
        // And a reply arriving meanwhile keeps its place at the end: its id is
        // above everything held, wherever it is merged in.
        let arrived = try messages(#"[{"id":12,"role":"assistant","content":"열둘째","created_at":12}]"#)
        XCTAssertEqual(held.merging(arrived).merging(older).map(\.id), [9, 10, 11, 12])
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
}
