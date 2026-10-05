import XCTest
@testable import PocketAideAPI

final class RoutinesTests: XCTestCase {
    func testDecodesRoutineWithSteps() throws {
        let json = #"{"id":4,"name":"아침 루틴","cadence":"weekly","weekdays":1,"month_day":0,"start_day":"2026-10-01","steps":[{"id":1,"title":"물 한 컵","position":0},{"id":2,"title":"스트레칭","position":1}],"created_at":1700000000}"#
        let routine = try JSONDecoder().decode(Routine.self, from: Data(json.utf8))
        XCTAssertEqual(routine.cadence, .weekly)
        XCTAssertEqual(routine.steps.map(\.title), ["물 한 컵", "스트레칭"])
        XCTAssertEqual(routine.startDay, "2026-10-01")
        XCTAssertEqual(routine.scheduleSummary, "매주 일요일")
    }

    func testScheduleSummaries() {
        let monWedFri = RoutineWeekdays.bit(1) | RoutineWeekdays.bit(3) | RoutineWeekdays.bit(5)
        XCTAssertEqual(RoutineWeekdays.summary(cadence: .daily, weekdays: 0, monthDay: 0), "매일")
        XCTAssertEqual(RoutineWeekdays.summary(cadence: .weekdays, weekdays: monWedFri, monthDay: 0), "월·수·금")
        XCTAssertEqual(RoutineWeekdays.summary(cadence: .weekdays, weekdays: RoutineWeekdays.bit(0) | RoutineWeekdays.bit(6), monthDay: 0), "토·일")
        XCTAssertEqual(RoutineWeekdays.summary(cadence: .monthly, weekdays: 0, monthDay: 15), "매월 15일")
    }

    func testDayProgressAndCompletion() throws {
        let json = #"{"id":4,"name":"저녁 정리","cadence":"daily","weekdays":0,"month_day":0,"day":"2026-10-01","steps":[{"id":1,"title":"책상","position":0,"checked":true},{"id":2,"title":"옷","position":1,"checked":false}],"done":1,"total":2,"completed":false}"#
        let day = try JSONDecoder().decode(RoutineDay.self, from: Data(json.utf8))
        XCTAssertEqual(day.progressPercent, 50)
        XCTAssertFalse(day.completed)
        XCTAssertEqual(day.steps.filter(\.checked).map(\.id), [1])
    }

    func testHistoryStatusAndSummary() throws {
        let json = #"{"items":[{"day":"2026-09-29","scheduled":true,"done":2,"total":2,"completed":true},{"day":"2026-09-30","scheduled":true,"done":1,"total":2,"completed":false},{"day":"2026-10-01","scheduled":false,"done":0,"total":2,"completed":false}]}"#
        let days = try JSONDecoder().decode(RoutineHistoryResponse.self, from: Data(json.utf8)).items
        XCTAssertEqual(days.map(\.statusLabel), ["완료", "미완료 1/2", "쉬는 날"])
        let summary = RoutineHistorySummary(days)
        XCTAssertEqual(summary.scheduledDays, 2)
        XCTAssertEqual(summary.completedDays, 1)
        XCTAssertEqual(RoutineDayFormat.title(of: "2026-10-01"), "10월 1일")
    }

    func testDayFormatUsesCalendarDay() {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Seoul") ?? .current
        let lateNightUTC = Date(timeIntervalSince1970: 1_790_870_400)
        XCTAssertEqual(RoutineDayFormat.string(from: lateNightUTC, calendar: calendar), "2026-10-02")
    }

    func testDraftEncodesWireKeys() throws {
        let draft = RoutineDraft(name: "월말 정산", cadence: .monthly, weekdays: 0, monthDay: 31, startDay: "2026-10-01", steps: ["카드 내역"])
        let object = try JSONSerialization.jsonObject(with: JSONEncoder().encode(draft)) as? [String: Any]
        XCTAssertEqual(object?["month_day"] as? Int, 31)
        XCTAssertEqual(object?["start_day"] as? String, "2026-10-01")
        XCTAssertEqual(object?["cadence"] as? String, "monthly")
    }

    func testMoveResultCarriesRoutine() throws {
        let json = #"{"target":"routine","routine":{"id":5,"name":"저녁 정리","cadence":"daily","weekdays":0,"month_day":0,"start_day":"2026-10-01","steps":[],"created_at":1}}"#
        let result = try JSONDecoder().decode(ScratchpadMoveResult.self, from: Data(json.utf8))
        XCTAssertEqual(result.target, .routine)
        XCTAssertEqual(result.routine?.name, "저녁 정리")
        XCTAssertEqual(ScratchpadMoveTarget.routine.chipLabel, "→ 루틴")
    }

    func testFailureCopyIsReadableAndHidesRawError() {
        let offline = APIError.transport(URLError(.notConnectedToInternet))
        XCTAssertEqual(
            RoutineFailureCopy.message(for: .check, error: offline),
            "단계 체크를 저장하지 못했습니다. 네트워크 연결을 확인해 주세요."
        )
        let unavailable = APIError.badStatus(503, "service unavailable")
        let server = RoutineFailureCopy.message(for: .check, error: unavailable)
        XCTAssertEqual(server, "단계 체크를 저장하지 못했습니다. 서버가 응답하지 않습니다. 잠시 후 다시 시도해 주세요.")
        XCTAssertFalse(server.contains("503"))
        XCTAssertFalse(server.contains("service unavailable"))
        XCTAssertEqual(
            RoutineFailureCopy.message(for: .load, error: APIError.badStatus(404, "not found")),
            "루틴을 불러오지 못했습니다. 잠시 후 다시 시도해 주세요."
        )
        XCTAssertEqual(
            RoutineFailureCopy.message(for: .history, error: URLError(.timedOut)),
            "이력을 불러오지 못했습니다. 잠시 후 다시 시도해 주세요."
        )
    }
}
