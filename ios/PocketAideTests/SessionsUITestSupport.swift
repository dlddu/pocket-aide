import XCTest

struct SessionsScreen {
    let app: XCUIApplication

    static func open(in app: XCUIApplication) -> SessionsScreen {
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 15), "Tab bar should appear after sign-in")
        let tab = tabBar.buttons["세션"]
        XCTAssertTrue(tab.waitForExistence(timeout: 5), "The 세션 tab should sit in the bottom tab bar, not under More")
        tab.tap()
        let screen = SessionsScreen(app: app)
        XCTAssertTrue(screen.title.waitForExistence(timeout: 15), "The 에이전트 세션 screen should open")
        return screen
    }

    var title: XCUIElement { app.staticTexts["sessions.title"] }
    var emptyState: XCUIElement { app.staticTexts["sessions.empty.state"] }
    var createButton: XCUIElement { app.buttons["sessions.create.button"] }
    var saveButton: XCUIElement { app.buttons["sessions.create.save.button"] }
    var catalogRetry: XCUIElement { app.buttons["sessions.create.workload.retry"] }
    var listRetry: XCUIElement { app.buttons["sessions.error.retry"] }
}

/// 세션 플랫폼 스텁(`.github/actions/start-test-backend/session_platform_stub.py`)을 부르는 테스트 러너 쪽 헬퍼.
/// 앱이 부르는 경로는 백엔드 프록시를 거치는 실경로이고, 여기서 하는 일은 시드(세션 만들기)와 스텁 제어뿐이다.
enum SessionPlatformStub {
    private static let base = "http://127.0.0.1:5560"

    static func reset() { _ = send("POST", "/e2e/reset") }
    static func makeUnreachable() { _ = send("POST", "/e2e/unreachable") }
    static func makeReachable() { _ = send("POST", "/e2e/reachable") }
    static func disableArchive() { _ = send("POST", "/e2e/archive-disabled") }

    /// write 마다 응답 조각이 `delay` 초 뒤부터 `interval` 초 간격으로 자라게 한다. `markers` 면 turn 마커를 남기는 플랫폼처럼 감싼다.
    static func streamReplies(_ chunks: [String], delay: Double, interval: Double, markers: Bool = false) {
        let body = try? JSONSerialization.data(withJSONObject: ["chunks": chunks, "delay": delay, "interval": interval, "markers": markers])
        _ = send("POST", "/e2e/reply", body: body)
    }

    static func seedSession(name: String, workloadType: String = "claude-code") -> String? {
        let body = try? JSONSerialization.data(withJSONObject: ["name": name, "workloadType": workloadType])
        guard let data = send("POST", "/api/v1/sessions", body: body),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else { return nil }
        return json["id"] as? String
    }

    static func write(_ id: String, _ payload: String) {
        let body = try? JSONSerialization.data(withJSONObject: ["payload": payload])
        _ = send("POST", "/api/v1/sessions/\(id)/write", body: body)
    }

    static func snapshot(_ id: String) { _ = send("POST", "/api/v1/sessions/\(id)/snapshot") }

    static func sessionState(_ id: String) -> String? {
        guard let data = send("GET", "/api/v1/sessions/\(id)"),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else { return nil }
        return json["state"] as? String
    }

    static func sessionExists(_ id: String) -> Bool {
        guard let url = URL(string: base + "/api/v1/sessions/\(id)") else { return false }
        let done = DispatchSemaphore(value: 0)
        let box = StatusBox()
        URLSession.shared.dataTask(with: url) { _, response, _ in
            box.status = (response as? HTTPURLResponse)?.statusCode ?? 0
            done.signal()
        }.resume()
        _ = done.wait(timeout: .now() + 20)
        return box.status == 200
    }

    private static func send(_ method: String, _ path: String, body: Data? = nil) -> Data? {
        guard let url = URL(string: base + path) else { return nil }
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.httpBody = body
        let done = DispatchSemaphore(value: 0)
        let box = DataBox()
        URLSession.shared.dataTask(with: request) { data, _, _ in
            box.data = data
            done.signal()
        }.resume()
        _ = done.wait(timeout: .now() + 20)
        return box.data
    }
}

private final class DataBox: @unchecked Sendable {
    var data: Data?
}

private final class StatusBox: @unchecked Sendable {
    var status = 0
}
