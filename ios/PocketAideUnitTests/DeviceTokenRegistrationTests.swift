import XCTest
@testable import PocketAideAPI

final class DeviceTokenRegistrationTests: XCTestCase {
    private func encoded(_ body: DeviceTokenRegistration) throws -> [String: String] {
        let data = try JSONEncoder().encode(body)
        return try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: String])
    }

    func testOmitsNameWhenNotGiven() throws {
        XCTAssertEqual(try encoded(DeviceTokenRegistration(token: "abc")), ["token": "abc"])
    }

    func testSendsNameWhenGiven() throws {
        XCTAssertEqual(
            try encoded(DeviceTokenRegistration(token: "abc", name: "출근용 폰")),
            ["token": "abc", "name": "출근용 폰"]
        )
    }
}
