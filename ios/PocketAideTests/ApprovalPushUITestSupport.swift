import XCTest

enum ApprovalPushReceiver {
    static func registerDevice(_ api: BackendAPI, file: StaticString = #filePath, line: UInt = #line) -> String {
        let device = (0..<32).map { _ in String(format: "%02x", UInt8.random(in: 0...255)) }.joined()
        let reply = api.call("POST", "/api/device-tokens", json: ["token": device])
        XCTAssertEqual(reply.status, 201, "The device token should register: \(reply.text)", file: file, line: line)
        return device
    }

    static func payload(for approvalID: String, to device: String, timeout: TimeInterval = 30) -> [String: Any]? {
        let deadline = Date().addingTimeInterval(timeout)
        repeat {
            let match = received(by: device)
                .compactMap { $0["payload"] as? [String: Any] }
                .first { $0["approval_id"] as? String == approvalID }
            if let match {
                return match
            }
            Thread.sleep(forTimeInterval: 1)
        } while Date() < deadline
        return nil
    }

    static func title(_ payload: [String: Any]) -> String? {
        alert(payload)?["title"] as? String
    }

    static func body(_ payload: [String: Any]) -> String? {
        alert(payload)?["body"] as? String
    }

    private static func alert(_ payload: [String: Any]) -> [String: Any]? {
        (payload["aps"] as? [String: Any])?["alert"] as? [String: Any]
    }

    private static func received(by device: String) -> [[String: Any]] {
        guard let url = URL(string: "http://127.0.0.1:5558/e2e/received?device=\(device)") else { return [] }
        let done = DispatchSemaphore(value: 0)
        let box = ReceivedRows()
        URLSession.shared.dataTask(with: url) { data, _, _ in
            box.rows = (try? JSONSerialization.jsonObject(with: data ?? Data())) as? [[String: Any]] ?? []
            done.signal()
        }.resume()
        _ = done.wait(timeout: .now() + 20)
        return box.rows
    }
}

private final class ReceivedRows: @unchecked Sendable {
    var rows: [[String: Any]] = []
}
