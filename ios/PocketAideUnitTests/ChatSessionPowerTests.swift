import XCTest
@testable import PocketAideAPI

// PRD-1 AC12: the chat tab shows whether the session answering this
// conversation is on, and turns it back on when it is not. These cover the
// decoding and the words, which is where this feature can quietly lie.
final class ChatSessionPowerTests: XCTestCase {
    private func power(_ json: String) throws -> ChatSessionPower {
        try JSONDecoder().decode(ChatSessionPower.self, from: Data(json.utf8))
    }

    func testDecodesALiveSession() throws {
        let decoded = try power(#"{"power":"on","state":"active","last_access":"2026-11-28T00:00:00Z"}"#)
        XCTAssertEqual(decoded.power, .on)
        XCTAssertEqual(decoded.state, "active")
        XCTAssertFalse(decoded.power.canPowerOn)
        XCTAssertFalse(decoded.restored)
    }

    // PRD-11 AC5: an archived session keeps the archive time and size the
    // sessions tab shows for the same state, and it is the one state that
    // offers the button.
    func testDecodesAnArchivedSessionWithItsCheckpoint() throws {
        let decoded = try power(
            #"{"power":"off","state":"snapshot","checkpoint":{"createdAt":"2026-11-28T00:00:00Z","sizeBytes":2048}}"#
        )
        XCTAssertEqual(decoded.power, .off)
        XCTAssertTrue(decoded.power.canPowerOn)
        XCTAssertEqual(decoded.checkpoint?.sizeBytes, 2048)
        let detail = decoded.detail()
        XCTAssertTrue(detail.contains("아카이브"), detail)
        XCTAssertTrue(detail.contains("이어져요"), detail)
    }

    // The first message is what opens a session (AC4), so before it there is
    // nothing to turn on — and the line says what happens next instead.
    func testNoSessionYetOffersNoButton() throws {
        let decoded = try power(#"{"power":"none"}"#)
        XCTAssertEqual(decoded.power, .absent)
        XCTAssertFalse(decoded.power.canPowerOn)
        XCTAssertEqual(decoded.detail(), "첫 메시지를 보내면 세션이 열려요")
    }

    // A session the platform lost is not a lost conversation: the line says
    // the history stays and the next message opens a new session.
    func testAGoneSessionSaysTheHistoryStays() throws {
        let decoded = try power(#"{"power":"gone"}"#)
        XCTAssertEqual(decoded.power, .gone)
        XCTAssertFalse(decoded.power.canPowerOn)
        XCTAssertTrue(decoded.detail().contains("대화 기록은 그대로"), decoded.detail())
    }

    // A state neither side can place is shown as unknown with the platform's
    // own word kept — not guessed into on or off, and with no button.
    func testAnUnrecognisedStateStaysUnknown() throws {
        let decoded = try power(#"{"power":"hibernating","state":"hibernating"}"#)
        XCTAssertEqual(decoded.power, .unknown)
        XCTAssertFalse(decoded.power.canPowerOn)
        XCTAssertTrue(decoded.detail().contains("hibernating"), decoded.detail())
    }

    // A turn-on says what actually happened: it restored the session, the
    // session was already on, or the platform took the request without a state
    // coming back — which must not read as "turned on".
    func testPowerOnNoticeTellsTheThreeOutcomesApart() throws {
        let restored = try power(#"{"power":"on","state":"active","restored":true}"#)
        XCTAssertEqual(restored.powerOnNotice, "세션을 켰어요. 이어서 보내면 같은 대화로 이어져요.")

        let alreadyOn = try power(#"{"power":"on","state":"active"}"#)
        XCTAssertTrue(alreadyOn.powerOnNotice.contains("이미 켜져"), alreadyOn.powerOnNotice)

        let accepted = try power(#"{"power":"unknown","restored":true}"#)
        XCTAssertTrue(accepted.powerOnNotice.contains("접수"), accepted.powerOnNotice)
        XCTAssertFalse(accepted.powerOnNotice.contains("켰어요"), accepted.powerOnNotice)
    }

    // PRD-1 AC12 · PRD-11 AC11: the refusals the tab explains differently are
    // told apart by the backend's code and status, so a session the platform
    // is busy with never reads as "send a message first".
    func testRefusalsAreToldApart() {
        func mapped(_ status: Int, _ body: String) -> ChatPowerError {
            ChatPowerError(APIError.badStatus(status, body))
        }
        XCTAssertEqual(mapped(409, #"{"code":"chat_power_no_session"}"#), .noSession)
        XCTAssertEqual(mapped(409, #"{"code":"chat_power_busy"}"#), .busy)
        XCTAssertEqual(mapped(404, #"{"code":"chat_power_session_gone"}"#), .gone)
        XCTAssertEqual(mapped(503, #"{"code":"auth_pool_exhausted"}"#), .slotsBusy)
        XCTAssertEqual(mapped(502, #"{"error":"session platform unreachable"}"#), .unreachable)
        XCTAssertEqual(mapped(503, #"{"error":"session platform error"}"#), .unreachable)
        XCTAssertEqual(ChatPowerError(APIError.transport(URLError(.notConnectedToInternet))), .unreachable)
        XCTAssertTrue(ChatPowerError.gone.message.contains("대화 기록은 그대로"))
    }
}
