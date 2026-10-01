import XCTest
@testable import PocketAideAPI

final class NotificationSettingsTests: XCTestCase {
    func testDecodesServerShape() throws {
        let json = Data(#"{"enabled":false,"outcomes":"failure"}"#.utf8)
        let decoded = try JSONDecoder().decode(NotificationSettings.self, from: json)
        XCTAssertEqual(decoded, NotificationSettings(enabled: false, outcomes: .failure))
    }

    func testDefaultIsEnabledForBothOutcomes() {
        XCTAssertEqual(NotificationSettings(), NotificationSettings(enabled: true, outcomes: .both))
    }

    func testPatchOmitsUnsetFields() throws {
        let data = try JSONEncoder().encode(NotificationSettingsPatch(outcomes: .success))
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(object.keys.sorted(), ["outcomes"])
        XCTAssertEqual(object["outcomes"] as? String, "success")
    }

    func testPatchEncodesEnabledOnly() throws {
        let data = try JSONEncoder().encode(NotificationSettingsPatch(enabled: false))
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(object.keys.sorted(), ["enabled"])
        XCTAssertEqual(object["enabled"] as? Bool, false)
    }

    func testOutcomeRawValuesMatchServerContract() {
        XCTAssertEqual(NotificationOutcomes.allCases.map(\.rawValue), ["both", "success", "failure"])
        XCTAssertEqual(NotificationOutcomes.allCases.map(\.label), ["둘 다", "성공만", "실패만"])
    }
}
