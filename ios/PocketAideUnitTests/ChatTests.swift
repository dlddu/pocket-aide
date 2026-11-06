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
    }

    func testMapsErrors() {
        XCTAssertEqual(ChatError(APIError.badStatus(409, #"{"error":"a reply is still in progress"}"#)), .busy)
        XCTAssertEqual(ChatError(APIError.badStatus(409, #"{"error":"send a message first"}"#), refreshing: true), .noConversation)
        XCTAssertEqual(ChatError(APIError.badStatus(502, "")), .unreachable)
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
