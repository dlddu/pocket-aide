import XCTest
@testable import PocketAideAPI

final class PRMonitorPushPayloadTests: XCTestCase {
    func testEventIDFromInt64() {
        XCTAssertEqual(
            PRMonitorPushPayload.eventID(from: ["event_id": Int64(42)]),
            42
        )
    }

    func testEventIDFromInt() {
        XCTAssertEqual(
            PRMonitorPushPayload.eventID(from: ["event_id": 99]),
            99
        )
    }

    func testEventIDFromNSNumber() {
        // APNs JSON arrives via NSJSONSerialization, which boxes numbers as
        // NSNumber. The helper has to unbox those before string conversion.
        XCTAssertEqual(
            PRMonitorPushPayload.eventID(from: ["event_id": NSNumber(value: 1234)]),
            1234
        )
    }

    func testEventIDFromString() {
        XCTAssertEqual(
            PRMonitorPushPayload.eventID(from: ["event_id": "7"]),
            7
        )
    }

    func testEventIDReturnsNilWhenMissing() {
        XCTAssertNil(PRMonitorPushPayload.eventID(from: [:]))
        XCTAssertNil(PRMonitorPushPayload.eventID(from: ["other_key": 1]))
    }

    func testEventIDReturnsNilWhenUnparseable() {
        XCTAssertNil(PRMonitorPushPayload.eventID(from: ["event_id": "not-a-number"]))
        XCTAssertNil(PRMonitorPushPayload.eventID(from: ["event_id": ["nested": 1] as [String: Any]]))
    }

    func testDeepLinkURLShape() {
        let url = PRMonitorPushPayload.deepLinkURL(fromUserInfo: ["event_id": 42])
        XCTAssertEqual(url?.absoluteString, "pocketaide://pr-monitor?eventId=42")
    }

    func testDeepLinkURLNilWhenNoEventID() {
        XCTAssertNil(PRMonitorPushPayload.deepLinkURL(fromUserInfo: [:]))
    }
}
