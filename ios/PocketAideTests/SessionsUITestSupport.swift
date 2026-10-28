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
    private static let base = "http://127.0.0.1:5559"

    static func reset() { _ = send("POST", "/e2e/reset") }
    static func makeUnreachable() { _ = send("POST", "/e2e/unreachable") }
    static func makeReachable() { _ = send("POST", "/e2e/reachable") }

    static func seedSession(name: String, workloadType: String = "claude-code") -> String? {
        let body = try? JSONSerialization.data(withJSONObject: ["name": name, "workloadType": workloadType])
        guard let data = send("POST", "/api/v1/sessions", body: body),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else { return nil }
        return json["id"] as? String
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
