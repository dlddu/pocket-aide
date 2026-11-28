// PocketAideAppTests — the app target's own models, tested as themselves.
//
// Running it (needs macOS + Xcode, which this repo's CI does not currently
// provide — `ci.yml` keeps the iOS jobs off):
//
//     cd ios && xcodegen generate
//     xcodebuild test -project PocketAide.xcodeproj -scheme PocketAide \
//       -destination 'platform=iOS Simulator,name=iPhone 16' \
//       -only-testing:PocketAideAppTests
//
// The target is hosted by the app (TEST_HOST / BUNDLE_LOADER) so
// `@testable import PocketAide` can reach a view model that lives there.
// PocketAideUnitTests is deliberately left alone: it links only the shared
// package and needs no host app, and giving it one would change how every test
// already in it runs.

import Foundation
import PocketAideAPI
import PocketAideStorage
import XCTest

@testable import PocketAide

/// A stand-in for the network that lets a test decide which of two answers
/// lands first.
///
/// The race under test is the server answering a send *after* a poll has
/// already brought that send's message back, so the order is the whole point:
/// leaving it to timing would be leaving it to luck. The POST is therefore
/// parked here — received, but with its answer unwritten — until the test lets
/// it go. Nothing blocks while it waits, so the poll's GET goes through in the
/// meantime, which is exactly the situation being reproduced.
final class FakeTransport: URLProtocol {
    /// Everything the stub shares across the instances URLSession makes, one
    /// per request.
    final class Wire: @unchecked Sendable {
        private let lock = NSLock()
        private var getReplies: [String] = []
        private var seen: [String] = []
        private var parked: (proto: FakeTransport, client: URLProtocolClient)?
        private var waiter: CheckedContinuation<Void, Never>?
        private var postArrived = false

        func reset(getReplies: [String]) {
            lock.lock()
            defer { lock.unlock() }
            self.getReplies = getReplies
            seen = []
            parked = nil
            waiter = nil
            postArrived = false
        }

        /// The paths the stub was asked for, in order — so a test can say that
        /// the poll in the middle of the race was still the incremental read.
        var paths: [String] {
            lock.lock()
            defer { lock.unlock() }
            return seen
        }

        func nextGetReply(for path: String) -> String? {
            lock.lock()
            defer { lock.unlock() }
            seen.append(path)
            return getReplies.isEmpty ? nil : getReplies.removeFirst()
        }

        func park(_ proto: FakeTransport, _ client: URLProtocolClient, path: String) {
            lock.lock()
            seen.append(path)
            parked = (proto, client)
            postArrived = true
            let resume = waiter
            waiter = nil
            lock.unlock()
            resume?.resume()
        }

        /// Resumes as soon as the send's POST has reached the stub — at once if
        /// it already has.
        func awaitPost(_ continuation: CheckedContinuation<Void, Never>) {
            lock.lock()
            if postArrived {
                lock.unlock()
                continuation.resume()
                return
            }
            waiter = continuation
            lock.unlock()
        }

        func takeParked() -> (proto: FakeTransport, client: URLProtocolClient)? {
            lock.lock()
            defer { lock.unlock() }
            let held = parked
            parked = nil
            return held
        }
    }

    static let wire = Wire()

    static func reset(getReplies: [String]) {
        wire.reset(getReplies: getReplies)
    }

    static func session() -> URLSession {
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [FakeTransport.self]
        return URLSession(configuration: config)
    }

    /// Waits until the send's POST is parked, then returns. No sleeping.
    static func waitForParkedSend() async {
        await withCheckedContinuation { continuation in
            wire.awaitPost(continuation)
        }
    }

    /// Lets the parked send answer, with `body`, now.
    static func answerParkedSend(_ body: String) {
        guard let held = wire.takeParked() else {
            XCTFail("보내기의 POST 가 멈춰 있지 않다 — 풀어 줄 응답이 없다")
            return
        }
        held.proto.finish(client: held.client, status: 202, body: body)
    }

    override class func canInit(with request: URLRequest) -> Bool { true }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let client else { return }
        let path = FakeTransport.path(of: request)
        if request.httpMethod == "POST" {
            // Received and held: the answer is written only when the test says.
            FakeTransport.wire.park(self, client, path: path)
            return
        }
        guard let reply = FakeTransport.wire.nextGetReply(for: path) else {
            finish(client: client, status: 500, body: #"{"error":"준비해 둔 GET 응답이 없다"}"#)
            return
        }
        finish(client: client, status: 200, body: reply)
    }

    override func stopLoading() {}

    private func finish(client: URLProtocolClient, status: Int, body: String) {
        guard let url = request.url,
              let response = HTTPURLResponse(
                  url: url,
                  statusCode: status,
                  httpVersion: "HTTP/1.1",
                  headerFields: ["Content-Type": "application/json"]
              ) else { return }
        client.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
        client.urlProtocol(self, didLoad: Data(body.utf8))
        client.urlProtocolDidFinishLoading(self)
    }

    private static func path(of request: URLRequest) -> String {
        guard let url = request.url,
              let parts = URLComponents(url: url, resolvingAgainstBaseURL: false) else { return "" }
        return parts.query.map { "\(parts.path)?\($0)" } ?? parts.path
    }
}

/// The token the stub hands out — far from expiring, with nothing to refresh,
/// so the client never leaves the request path under test.
private struct StubTokenStore: TokenStoring {
    func load() throws -> TokenBundle? {
        TokenBundle(accessToken: "access", refreshToken: nil, expiresAt: Date(timeIntervalSinceNow: 3600))
    }

    func save(_ bundle: TokenBundle) throws {}

    func clear() throws {}
}

/// PRD-AI채팅.AC1 · 완료 기준 5, iOS side.
///
/// The server stores a sent message *before* it answers the send, so a poll can
/// overtake that answer and bring the message back first — often with the reply
/// to it, which has a larger id. These drive the real `ChatViewModel` through
/// its real entry points (`load()`, `send(_:)`) over a stubbed transport, and
/// measure the list and the progress state it ends up holding. The merge helper
/// is not called directly: what is under test is what the view model does with
/// an answer that arrives late.
@MainActor
final class ChatSendRaceTests: XCTestCase {
    private static let sentID: Int64 = 2

    private func conversation(
        id: Int64 = 7,
        messages: String,
        pending: String = "null",
        queued: String = "null"
    ) -> String {
        """
        {"conversation":{"id":\(id),"title":"첫 질문"},"messages":[\(messages)],
         "pending":\(pending),"queued":\(queued)}
        """
    }

    private func message(_ id: Int64, _ role: String, _ content: String) -> String {
        #"{"id":\#(id),"role":"\#(role)","content":"\#(content)","created_at":\#(id)}"#
    }

    private var sentAnswer: String {
        #"{"message":\#(message(Self.sentID, "user", "보낸 것"))}"#
    }

    private func viewModel(getReplies: [String]) -> ChatViewModel {
        FakeTransport.reset(getReplies: getReplies)
        let api = APIClient(
            baseURL: URL(string: "https://pocket-aide.test")!,
            tokenStore: StubTokenStore(),
            session: FakeTransport.session()
        )
        return ChatViewModel(api: api)
    }

    /// The reported race: the poll comes back with the sent message *and* the
    /// finished reply to it, and only then does the send answer. That answer
    /// must add nothing — the message once, in id order, the reply still after
    /// it, and no 「입력 중」 bubble put back over an answer already arrived.
    func testALateSendAnswerDoesNotDuplicateReorderOrUndoTheFinishedReply() async {
        let vm = viewModel(getReplies: [
            conversation(messages: message(1, "user", "먼저")),
            conversation(messages: message(Self.sentID, "user", "보낸 것") + "," + message(3, "assistant", "답")),
        ])
        await vm.load()

        let sending = Task { await vm.send("보낸 것") }
        await FakeTransport.waitForParkedSend()

        // The poll overtakes the send's answer.
        await vm.load()
        XCTAssertEqual(vm.messages.map(\.id), [1, 2, 3])
        XCTAssertNil(vm.pending, "폴링이 끝난 답을 가져왔는데 진행 중으로 보인다")

        FakeTransport.answerParkedSend(sentAnswer)
        _ = await sending.value

        XCTAssertNil(vm.pending, "늦은 보내기 응답이 끝난 답 위에 「입력 중」을 다시 세웠다")
        XCTAssertEqual(vm.messages.map(\.id), [1, 2, 3], "보낸 메시지가 한 번, id 오름차순이어야 한다")
        XCTAssertEqual(vm.messages.filter { $0.id == Self.sentID }.count, 1, "보낸 메시지가 두 번 들어갔다")
        XCTAssertEqual(vm.messages.last?.content, "답", "답이 그대로 남아야 한다")
        XCTAssertFalse(vm.isSending)
        XCTAssertNil(vm.error)
        // The race is not fixed by going back to reading the conversation
        // whole: the poll in the middle of it is still the incremental read, and
        // the only page read is the one that opened the screen
        // (PRD-AI채팅.AC1 2026-11-27·11-28 개정).
        let paths = FakeTransport.wire.paths
        XCTAssertTrue(paths.contains { $0.contains("after=") }, "폴링이 증분으로 읽지 않았다 — \(paths)")
        XCTAssertEqual(paths.filter { $0.contains("limit=") }.count, 1, "대화를 다시 통째로 읽었다 — \(paths)")
        XCTAssertEqual(paths.filter { $0.hasSuffix("/api/chat/messages") }.count, 1, "보내기가 한 번만 나가야 한다 — \(paths)")
    }

    /// The same race one step earlier: the poll brings back only the sent
    /// message, with no answer yet. The send's answer is then still the only
    /// word that a reply is coming, so the 「입력 중」 bubble stays — the
    /// existing contract, unchanged.
    func testALateSendAnswerStillShowsTheReplyComingWhenNothingNewerWasRead() async {
        let vm = viewModel(getReplies: [
            conversation(messages: message(1, "user", "먼저")),
            conversation(messages: message(Self.sentID, "user", "보낸 것")),
        ])
        await vm.load()

        let sending = Task { await vm.send("보낸 것") }
        await FakeTransport.waitForParkedSend()
        await vm.load()
        FakeTransport.answerParkedSend(sentAnswer)
        _ = await sending.value

        XCTAssertEqual(vm.messages.map(\.id), [1, 2])
        XCTAssertTrue(vm.isReplying, "아직 아무 답도 읽지 않았는데 「입력 중」이 사라졌다")
    }

    /// A conversation replaced under a send: the message that comes back
    /// belongs to the conversation that has left, so it is not shown under the
    /// one now on screen.
    func testALateSendAnswerFromAConversationTheScreenHasLeftIsDropped() async {
        let other = conversation(id: 8, messages: message(5, "user", "새 대화"))
        let vm = viewModel(getReplies: [
            conversation(messages: message(1, "user", "옛 대화")),
            // The poll names another conversation, so load() starts that one's
            // newest page over — two answers for the one load().
            other,
            other,
        ])
        await vm.load()

        let sending = Task { await vm.send("옛 대화에 보낸 것") }
        await FakeTransport.waitForParkedSend()
        await vm.load()
        FakeTransport.answerParkedSend(sentAnswer)
        _ = await sending.value

        XCTAssertEqual(vm.conversation?.id, 8)
        XCTAssertEqual(vm.messages.map(\.id), [5], "옛 대화의 보낸 메시지가 새 대화에 섞였다")
        XCTAssertNil(vm.pending)
    }

    /// A message held for after the reply, answered late: by then the poll has
    /// read past the turn, so the server has already said whether anything is
    /// still waiting. The late answer must not put a waiting bubble back under
    /// the message's own sent bubble.
    func testALateQueuedAnswerDoesNotPutBackAMessageTheServerHasSent() async {
        let vm = viewModel(getReplies: [
            conversation(messages: message(1, "user", "먼저"), pending: #"{"content":"쓰는 중","started_at":1}"#),
            // The turn ended, the waiting message went out, and nothing waits.
            conversation(messages: message(Self.sentID, "assistant", "답") + "," + message(3, "user", "대기한 것")),
        ])
        await vm.load()
        XCTAssertTrue(vm.isReplying)

        let sending = Task { await vm.send("대기한 것") }
        await FakeTransport.waitForParkedSend()
        await vm.load()
        XCTAssertNil(vm.queued)

        FakeTransport.answerParkedSend(#"{"queued":{"id":"q1","content":"대기한 것","queued_at":2}}"#)
        _ = await sending.value

        XCTAssertEqual(vm.messages.map(\.id), [1, 2, 3])
        XCTAssertNil(vm.queued, "서버가 이미 보낸 메시지를 늦은 응답이 대기로 되돌렸다")
        XCTAssertTrue(vm.canSend)
    }
}
