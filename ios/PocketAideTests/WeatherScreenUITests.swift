// 검증 시나리오: test-widget.md#시나리오 10
import CoreLocation
import XCTest

final class WeatherScreenUITests: XCTestCase {
    private let weatherLink = URL(string: "pocketaide://weather")!
    private let seoul = CLLocation(latitude: 37.5665, longitude: 126.9780)
    private let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
    private let conditions = ["맑음", "대체로 맑음", "구름 조금", "흐림", "안개", "이슬비", "비", "눈", "소나기", "눈 소나기", "뇌우"]

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
        XCUIDevice.shared.location = XCUILocation(location: seoul)
    }

    override func tearDown() {
        XCUIApplication().resetAuthorizationStatus(for: .location)
        super.tearDown()
    }

    func testAllowedLocationShowsTheForecastFromTheWeatherLink() {
        let app = launch(allowingLocation: true)
        let tabBefore = selectedTab(in: app)
        let close = openWeather(in: app)
        XCTAssertFalse(app.tabBars.firstMatch.buttons["날씨"].exists, "Weather should not be a tab")

        let temperature = text(matching: "^-?[0-9]+°$", in: app)
        XCTAssertTrue(temperature.waitForExistence(timeout: 60), "The current temperature should load")
        assertCurrentConditions(in: app)
        assertPlace(in: app)
        assertHourlyAndWeekly(in: app)
        assertPullToRefreshMovesTheUpdateTime(from: temperature, in: app)

        close.tap()
        XCTAssertTrue(close.waitForNonExistence(timeout: 10), "닫기 should dismiss the weather screen")
        XCTAssertEqual(selectedTab(in: app), tabBefore, "The weather screen should sit above the tabs, not replace one")
    }

    func testDeniedLocationAsksForLocationAccess() {
        let app = launch(allowingLocation: false)
        _ = openWeather(in: app)
        let notice = text(matching: "^앱에서 위치 접근을 허용해 주세요\\.$", in: app)
        XCTAssertTrue(notice.waitForExistence(timeout: 30), "A denied location should show the access notice")
        let settings = app.buttons.matching(NSPredicate(format: "label == %@", "설정 열기")).firstMatch
        XCTAssertTrue(settings.exists, "The notice should offer a way to the settings")
        XCTAssertFalse(text(matching: "^-?[0-9]+°$", in: app).exists, "No forecast should show without a location")
    }

    private func launch(allowingLocation allow: Bool) -> XCUIApplication {
        let app = XCUIApplication()
        app.resetAuthorizationStatus(for: .location)
        app.launch()
        let alert = springboard.alerts
            .matching(NSPredicate(format: "label CONTAINS[c] %@ OR label CONTAINS %@", "location", "위치")).firstMatch
        XCTAssertTrue(alert.waitForExistence(timeout: 30), "A signed-in launch should ask for location access")
        let labels = allow
            ? ["Allow While Using App", "앱을 사용하는 동안 허용"]
            : ["Don’t Allow", "Don't Allow", "허용 안 함"]
        let button = alert.buttons.matching(NSPredicate(format: "label IN %@", labels)).firstMatch
        XCTAssertTrue(button.waitForExistence(timeout: 5), "The location prompt should offer \(labels)")
        button.tap()
        XCTAssertTrue(alert.waitForNonExistence(timeout: 10), "The location prompt should close")
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 15), "The app should land on the tab shell")
        return app
    }

    private func openWeather(in app: XCUIApplication) -> XCUIElement {
        app.open(weatherLink)
        let close = app.buttons.matching(NSPredicate(format: "label == %@", "닫기")).firstMatch
        XCTAssertTrue(close.waitForExistence(timeout: 15), "The weather link should open the weather screen")
        return close
    }

    private func selectedTab(in app: XCUIApplication) -> String {
        app.tabBars.firstMatch.buttons.matching(NSPredicate(format: "isSelected == true")).firstMatch.label
    }

    private func text(matching pattern: String, in app: XCUIApplication) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label MATCHES %@", pattern)).firstMatch
    }

    private func labels(matching pattern: String, in app: XCUIApplication) -> [String] {
        app.descendants(matching: .any)
            .matching(NSPredicate(format: "label MATCHES %@", pattern))
            .allElementsBoundByIndex.map(\.label)
    }

    private func assertCurrentConditions(in app: XCUIApplication) {
        let condition = app.staticTexts.matching(NSPredicate(format: "label IN %@", conditions)).firstMatch
        XCTAssertTrue(condition.exists, "The current weather condition should show")
        XCTAssertTrue(text(matching: "^체감 -?[0-9]+°$", in: app).exists, "The apparent temperature should show")
        let detail = text(matching: "^최고 -?[0-9]+ · 최저 -?[0-9]+( · 강수 [0-9]+%)?$", in: app)
        XCTAssertTrue(detail.exists, "Today's high, low and precipitation chance should show")
    }

    private func assertPlace(in app: XCUIApplication) {
        let geocoded = expectation(description: "runner reverse geocode")
        var place: String?
        CLGeocoder().reverseGeocodeLocation(seoul) { marks, _ in
            place = marks?.first?.locality
            geocoded.fulfill()
        }
        wait(for: [geocoded], timeout: 30)
        guard let place, !place.isEmpty else {
            let note = XCTAttachment(string: "Runner reverse geocoding returned no locality; place name not compared")
            note.lifetime = .keepAlways
            add(note)
            return
        }
        let header = app.staticTexts.matching(NSPredicate(format: "label == %@", place)).firstMatch
        XCTAssertTrue(header.waitForExistence(timeout: 30), "The screen should name the neighbourhood \(place)")
    }

    private func assertHourlyAndWeekly(in app: XCUIApplication) {
        let hourPattern = "^[0-9]{1,2}시, .*"
        let hours = Set(labels(matching: hourPattern, in: app).compactMap { $0.components(separatedBy: ",").first })
        XCTAssertEqual(hours.count, 24, "The hourly forecast should hold 24 distinct hours: \(hours.sorted())")

        let hourCells = app.descendants(matching: .any).matching(NSPredicate(format: "label MATCHES %@", hourPattern))
        let cells = hourCells.allElementsBoundByIndex
        if let first = cells.min(by: { $0.frame.minX < $1.frame.minX }),
           let last = cells.max(by: { $0.frame.minX < $1.frame.minX }) {
            let start = app.coordinate(withNormalizedOffset: .zero)
                .withOffset(CGVector(dx: app.frame.width * 0.85, dy: first.frame.midY))
            for _ in 0..<8 where !last.isHittable {
                start.press(forDuration: 0.05, thenDragTo: start.withOffset(CGVector(dx: -app.frame.width * 0.6, dy: 0)))
            }
            XCTAssertTrue(last.isHittable, "Swiping the hourly forecast should reach its last hour")
        }

        let dayPattern = "^(오늘|[0-9]{1,2}/[0-9]{1,2} \\(.\\)), .*"
        let days = Set(labels(matching: dayPattern, in: app).compactMap { $0.components(separatedBy: ", ").first })
        XCTAssertEqual(days.count, 7, "The weekly forecast should hold 7 days: \(days.sorted())")
        XCTAssertTrue(days.contains("오늘"), "The weekly forecast should start with today: \(days.sorted())")
    }

    private func assertPullToRefreshMovesTheUpdateTime(from anchor: XCUIElement, in app: XCUIApplication) {
        let stamp = text(matching: "^[0-9]{2}:[0-9]{2} 갱신$", in: app)
        XCTAssertTrue(stamp.waitForExistence(timeout: 10), "The update time should show")
        let before = stamp.label

        let second = Calendar.current.component(.second, from: Date())
        Thread.sleep(forTimeInterval: TimeInterval(62 - second))
        var after = before
        for _ in 0..<3 where after == before {
            let start = anchor.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
            start.press(forDuration: 0.1, thenDragTo: start.withOffset(CGVector(dx: 0, dy: 420)))
            let moved = XCTNSPredicateExpectation(predicate: NSPredicate(format: "label != %@", before), object: stamp)
            _ = XCTWaiter().wait(for: [moved], timeout: 20)
            after = stamp.label
        }
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm"
        let now = Date()
        let expected = [now, now.addingTimeInterval(-60)].map { "\(formatter.string(from: $0)) 갱신" }
        XCTAssertTrue(app.buttons.matching(NSPredicate(format: "label == %@", "닫기")).firstMatch.exists,
                      "Pulling the forecast should not dismiss the weather screen")
        XCTAssertNotEqual(after, before, "Pull to refresh should move the update time")
        XCTAssertTrue(expected.contains(after), "The update time should be the refresh time: \(after) vs \(expected)")
    }
}
