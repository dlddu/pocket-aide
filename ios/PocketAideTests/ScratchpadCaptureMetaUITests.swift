// 검증 시나리오: test-scratchpad.md#시나리오 4
import XCTest

final class ScratchpadCaptureMetaUITests: XCTestCase {
    private struct Seed {
        let text: String
        let meta: String
        let section: String
    }

    private struct Planned {
        let text: String
        let source: String
        let day: Date
        let hour: Int
        let minute: Int
        let seed: Seed
    }

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testMetaLinesAndDaySectionsFollowCaptureTimeAndSource() throws {
        ScratchpadCapture.waitPastMidnight()
        let api = try XCTUnwrap(BackendAPI.signIn(), "The test runner should get a backend token from the test IdP")
        api.clearScratchpad()

        let token = TodoUI.uniqueToken()
        let app = TodoUI.launch()
        var scratchpad = ScratchpadScreen.open(in: app)
        let typed = "캡처 메타 텍스트 \(token)"
        scratchpad.add(typed)
        let typedAt = try XCTUnwrap(api.capturedAt(of: typed), "The memo typed in the app should be stored with its capture time")

        let seeds = try seed(api, token: token)
        let expected = [Seed(text: typed, meta: "탭 내 입력 · \(ScratchpadCapture.clock(typedAt))", section: "오늘")] + seeds
        TodoUI.relaunch(app)
        scratchpad = ScratchpadScreen.open(in: app)

        for seed in expected {
            scratchpad.assertMeta(seed.meta, above: seed.text)
        }
        scratchpad.assertSection("오늘", count: 1)
        scratchpad.assertSection("어제", count: 2)
        scratchpad.assertSection(seeds[seeds.count - 1].section, count: 1)

        var order: [XCUIElement] = []
        for (index, seed) in expected.enumerated() {
            if index == 0 || expected[index - 1].section != seed.section {
                order.append(scratchpad.sectionHeader(seed.section))
            }
            order.append(scratchpad.memo(seed.text))
        }
        scratchpad.assertTopToBottom(order, "Sections should run 오늘 → 어제 → older day, newest memo first inside each")

        api.clearScratchpad()
    }

    private func seed(_ api: BackendAPI, token: String) throws -> [Seed] {
        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())
        let yesterday = try XCTUnwrap(calendar.date(byAdding: .day, value: -1, to: today))
        let olderDay = try XCTUnwrap(calendar.date(byAdding: .day, value: -3, to: today))
        let parts = calendar.dateComponents([.month, .day], from: olderDay)
        let olderTitle = "\(parts.month ?? 0)월 \(parts.day ?? 0)일"
        let voice = "캡처 메타 음성 \(token)"
        let shortcut = "캡처 메타 숏컷 \(token)"
        let older = "캡처 메타 지난 메모 \(token)"
        let plan = [
            Planned(text: voice, source: "voice", day: yesterday, hour: 21, minute: 40,
                    seed: Seed(text: voice, meta: "음성 · 21:40", section: "어제")),
            Planned(text: shortcut, source: "shortcut", day: yesterday, hour: 8, minute: 5,
                    seed: Seed(text: shortcut, meta: "숏컷 · 음성 · 08:05", section: "어제")),
            Planned(text: older, source: "text", day: olderDay, hour: 10, minute: 30,
                    seed: Seed(text: older, meta: "탭 내 입력 · 10:30", section: olderTitle))
        ]
        return try plan.map { entry in
            let at = try XCTUnwrap(calendar.date(bySettingHour: entry.hour, minute: entry.minute, second: 0, of: entry.day))
            api.createScratchpad(entry.text, source: entry.source, capturedAt: at)
            return entry.seed
        }
    }
}

private enum ScratchpadCapture {
    static func clock(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.dateFormat = "HH:mm"
        return formatter.string(from: date)
    }

    static func waitPastMidnight() {
        let calendar = Calendar.current
        guard let midnight = calendar.date(byAdding: .day, value: 1, to: calendar.startOfDay(for: Date())) else { return }
        let left = midnight.timeIntervalSinceNow
        if left < 600 {
            Thread.sleep(forTimeInterval: left + 5)
        }
    }
}

private extension BackendAPI {
    func items() -> [[String: Any]] {
        let reply = call("GET", "/api/scratchpad")
        XCTAssertEqual(reply.status, 200, "GET /api/scratchpad should answer: \(reply.text)")
        return reply.json?["items"] as? [[String: Any]] ?? []
    }

    func clearScratchpad() {
        for item in items() {
            guard let id = (item["id"] as? NSNumber)?.int64Value else { continue }
            let reply = call("DELETE", "/api/scratchpad/\(id)")
            XCTAssertTrue((200..<300).contains(reply.status), "DELETE /api/scratchpad/\(id) should succeed: \(reply.text)")
        }
        XCTAssertTrue(items().isEmpty, "The scratchpad should start empty")
    }

    func capturedAt(of text: String) -> Date? {
        let item = items().first { ($0["text"] as? String) == text }
        guard let seconds = (item?["captured_at"] as? NSNumber)?.doubleValue else { return nil }
        return Date(timeIntervalSince1970: seconds)
    }

    func createScratchpad(_ text: String, source: String, capturedAt: Date) {
        let body: [String: Any] = ["text": text, "source": source, "captured_at": Int64(capturedAt.timeIntervalSince1970)]
        let reply = call("POST", "/api/scratchpad", json: body)
        XCTAssertEqual(reply.status, 201, "POST /api/scratchpad should store '\(text)': \(reply.text)")
    }
}

private extension ScratchpadScreen {
    var list: XCUIElement { app.collectionViews.firstMatch }

    @discardableResult
    func reveal(_ element: XCUIElement, attempts: Int = 6) -> Bool {
        for _ in 0..<attempts {
            if element.exists && element.isHittable { return true }
            list.swipeUp()
        }
        return element.exists && element.isHittable
    }

    func sectionHeader(_ title: String) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label == %@ OR label BEGINSWITH %@", title, "\(title), ")).firstMatch
    }

    func assertMeta(_ meta: String, above text: String, file: StaticString = #filePath, line: UInt = #line) {
        let element = memo(text)
        XCTAssertTrue(element.waitForExistence(timeout: 10) && reveal(element), "'\(text)' should be listed", file: file, line: line)
        let floor = element.frame.minY + 1
        let nearest = app.staticTexts.matching(NSPredicate(format: "label == %@", meta)).allElementsBoundByIndex
            .filter { $0.frame.maxY <= floor }
            .max { $0.frame.maxY < $1.frame.maxY }
        XCTAssertNotNil(nearest, "'\(text)' should carry the meta line '\(meta)'", file: file, line: line)
        if let nearest {
            XCTAssertLessThan(element.frame.minY - nearest.frame.maxY, 30, "'\(meta)' should sit right above '\(text)'", file: file, line: line)
        }
    }

    func assertSection(_ title: String, count: Int, file: StaticString = #filePath, line: UInt = #line) {
        let header = sectionHeader(title)
        XCTAssertTrue(header.waitForExistence(timeout: 10) && reveal(header), "The '\(title)' section should be listed", file: file, line: line)
        let wanted = "\(count) ITEMS"
        if header.label.hasSuffix(wanted) { return }
        let beside = app.staticTexts.matching(NSPredicate(format: "label ENDSWITH %@", " ITEMS")).allElementsBoundByIndex
            .filter { abs($0.frame.midY - header.frame.midY) < 8 }
            .map(\.label)
        XCTAssertEqual(beside, [wanted], "The '\(title)' section header should read '\(wanted)'", file: file, line: line)
    }

    func assertTopToBottom(_ elements: [XCUIElement], _ message: String, file: StaticString = #filePath, line: UInt = #line) {
        guard var previous = elements.first else { return }
        list.swipeDown()
        list.swipeDown()
        XCTAssertTrue(reveal(previous), message, file: file, line: line)
        for element in elements.dropFirst() {
            XCTAssertTrue(element.waitForExistence(timeout: 5) && reveal(element), message, file: file, line: line)
            if previous.exists && previous.isHittable {
                XCTAssertLessThan(previous.frame.minY, element.frame.minY, message, file: file, line: line)
            }
            previous = element
        }
    }
}
