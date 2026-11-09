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
         "queued":{"content":"다음 질문","queued_at":4}}
        """#
        let thread = try JSONDecoder().decode(ChatThread.self, from: Data(json.utf8))
        XCTAssertEqual(thread.queued?.content, "다음 질문")
        XCTAssertEqual(thread.queued?.queuedAt, 4)
    }

    func testSendOutcomeDecodesSentOrQueued() throws {
        let sent = try JSONDecoder().decode(ChatSendResponse.self, from: Data(#"{"message":{"id":1,"role":"user","content":"hi","created_at":1}}"#.utf8))
        XCTAssertEqual(sent.message?.content, "hi")
        XCTAssertNil(sent.queued)

        let queued = try JSONDecoder().decode(ChatSendResponse.self, from: Data(#"{"queued":{"content":"hi","queued_at":5}}"#.utf8))
        XCTAssertNil(queued.message)
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
