import XCTest

struct WebhookEvent {
    let repo: String
    let pullRequest: Int?
    let commit: String
    let branch: String
    let workflow: String
    let conclusion: String

    var groupKey: String {
        if let pullRequest {
            return "pr:\(repo):\(pullRequest)"
        }
        return "sha:\(repo):\(commit)"
    }

    var envelope: [String: Any] {
        let runID = Int.random(in: 100_000_000...999_999_999)
        var pullRequests: [[String: Any]] = []
        if let pullRequest {
            pullRequests.append([
                "id": runID + 1,
                "number": pullRequest,
                "url": "https://api.github.com/repos/\(repo)/pulls/\(pullRequest)",
                "head": ["ref": branch, "sha": commit],
                "base": ["ref": "main", "sha": String(repeating: "0", count: 40)]
            ])
        }
        let run: [String: Any] = [
            "id": runID,
            "name": workflow,
            "head_branch": branch,
            "head_sha": commit,
            "event": pullRequest == nil ? "push" : "pull_request",
            "status": "completed",
            "conclusion": conclusion,
            "run_number": 1,
            "html_url": "https://github.com/\(repo)/actions/runs/\(runID)",
            "pull_requests": pullRequests
        ]
        let repository: [String: Any] = [
            "id": runID + 2,
            "name": String(repo.split(separator: "/").last ?? ""),
            "full_name": repo,
            "html_url": "https://github.com/\(repo)"
        ]
        let sender: [String: Any] = ["login": "dlddu"]
        return ["action": "completed", "workflow_run": run, "repository": repository, "sender": sender]
    }
}

private final class WebhookReply {
    var status = 0
    var detail = ""
}

enum WebhookHistoryUI {
    static let endpoint = "http://127.0.0.1:4566/"
    static let queueURL = "http://localhost:4566/123456789012/pocket-aide-e2e"
    static let authorization = "AWS4-HMAC-SHA256 Credential=test/20260101/us-east-1/sqs/aws4_request, "
        + "SignedHeaders=host;x-amz-target, Signature=" + String(repeating: "0", count: 64)
    static let groupPrefix = "prmonitor.group."

    static func runToken() -> String {
        String(UUID().uuidString.prefix(8)).lowercased()
    }

    static func commit(_ seed: Int, token: String) -> String {
        String((String(seed) + String(repeating: token, count: 5)).prefix(40))
    }

    static func send(_ events: [WebhookEvent]) {
        for event in events {
            send(event)
            Thread.sleep(forTimeInterval: 1.5)
        }
    }

    static func send(_ event: WebhookEvent) {
        guard let url = URL(string: endpoint),
              let payload = try? JSONSerialization.data(withJSONObject: event.envelope)
        else {
            XCTFail("The workflow_run envelope for \(event.workflow) should encode")
            return
        }
        let message: [String: Any] = [
            "QueueUrl": queueURL,
            "MessageBody": String(bytes: payload, encoding: .utf8) ?? "",
            "MessageAttributes": ["x-github-event": ["DataType": "String", "StringValue": "workflow_run"]]
        ]
        guard let body = try? JSONSerialization.data(withJSONObject: message) else {
            XCTFail("The SendMessage request for \(event.workflow) should encode")
            return
        }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.httpBody = body
        request.setValue("application/x-amz-json-1.0", forHTTPHeaderField: "Content-Type")
        request.setValue("AmazonSQS.SendMessage", forHTTPHeaderField: "X-Amz-Target")
        request.setValue(authorization, forHTTPHeaderField: "Authorization")

        let done = DispatchSemaphore(value: 0)
        let reply = WebhookReply()
        URLSession.shared.dataTask(with: request) { data, response, error in
            reply.status = (response as? HTTPURLResponse)?.statusCode ?? 0
            reply.detail = error.map { String(describing: $0) } ?? (String(bytes: data ?? Data(), encoding: .utf8) ?? "")
            done.signal()
        }.resume()
        XCTAssertTrue(done.wait(timeout: .now() + 20) == .success, "The local SQS endpoint should answer SendMessage")
        XCTAssertEqual(reply.status, 200, "SendMessage for \(event.workflow) should be accepted: \(reply.detail)")
    }

    static func openMonitor() -> XCUIApplication {
        let app = XCUIApplication()
        app.launch()
        let tab = app.tabBars.firstMatch.buttons["PR 모니터"]
        XCTAssertTrue(tab.waitForExistence(timeout: 15), "PR 모니터 should be a direct tab")
        tab.tap()
        return app
    }

    static func openMonitor(until group: String, reads label: String) -> XCUIApplication {
        var app = openMonitor()
        var relaunches = 0
        while !text(label, of: group, in: app).waitForExistence(timeout: 10) && relaunches < 6 {
            relaunches += 1
            app.terminate()
            app = openMonitor()
        }
        XCTAssertTrue(text(label, of: group, in: app).exists, "Group \(group) should read \(label)")
        return app
    }

    static func texts(of group: String, in app: XCUIApplication) -> XCUIElementQuery {
        app.staticTexts.matching(NSPredicate(format: "identifier BEGINSWITH %@", groupPrefix + group))
    }

    static func text(_ label: String, of group: String, in app: XCUIApplication) -> XCUIElement {
        texts(of: group, in: app).matching(NSPredicate(format: "label == %@", label)).firstMatch
    }

    static func itemAcknowledgeButtons(of group: String, in app: XCUIApplication) -> XCUIElementQuery {
        app.buttons.matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label == %@", groupPrefix + group, "확인")
        )
    }

    static func acknowledgeAllButtons(of group: String, in app: XCUIApplication) -> XCUIElementQuery {
        app.buttons.matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND label CONTAINS %@", groupPrefix + group, "모두 확인")
        )
    }

    static func labels(of group: String, in app: XCUIApplication) -> [String] {
        texts(of: group, in: app).allElementsBoundByIndex
            .map { (top: $0.frame.minY, label: $0.label) }
            .sorted { $0.top < $1.top }
            .map { $0.label }
    }

    static func groups(of repo: String, in app: XCUIApplication) -> Set<String> {
        let leaves = app.staticTexts.matching(
            NSPredicate(format: "identifier BEGINSWITH %@ AND identifier CONTAINS %@", groupPrefix, ":\(repo):")
        )
        return Set(leaves.allElementsBoundByIndex.map { String($0.identifier.dropFirst(groupPrefix.count)) })
    }

    static func scroll(_ app: XCUIApplication, by offset: CGFloat) {
        let start = app.coordinate(withNormalizedOffset: CGVector(dx: 0.03, dy: 0.6))
        start.press(forDuration: 0.1, thenDragTo: start.withOffset(CGVector(dx: 0, dy: offset)))
    }

    @discardableResult
    static func reveal(_ group: String, in app: XCUIApplication, drags: Int = 20) -> Bool {
        let anchor = texts(of: group, in: app).firstMatch
        var left = drags
        while !anchor.exists && left > 0 {
            scroll(app, by: -320)
            left -= 1
        }
        return anchor.exists
    }

    static func tap(_ element: XCUIElement, in app: XCUIApplication, drags: Int = 20) {
        var left = drags
        while !(element.exists && element.isHittable) && left > 0 {
            scroll(app, by: -200)
            left -= 1
        }
        XCTAssertTrue(element.exists && element.isHittable, "The control should be reachable by scrolling the history")
        element.tap()
    }

    static func unreadCount(in app: XCUIApplication) -> Int? {
        let badge = app.descendants(matching: .any).matching(identifier: "prmonitor.unread.badge")
        guard badge.firstMatch.waitForExistence(timeout: 10) else { return nil }
        let joined = badge.allElementsBoundByIndex.map(\.label).joined(separator: " ")
        guard let digits = joined.range(of: "[0-9]+", options: .regularExpression) else { return nil }
        return Int(joined[digits])
    }
}
