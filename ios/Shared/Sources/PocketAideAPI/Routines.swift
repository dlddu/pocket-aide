import Foundation

public enum RoutineCadence: String, Codable, CaseIterable, Sendable, Hashable {
    case daily
    case weekdays
    case weekly
    case monthly

    public var displayName: String {
        switch self {
        case .daily: return "매일"
        case .weekdays: return "특정 요일"
        case .weekly: return "매주"
        case .monthly: return "매월"
        }
    }
}

public enum RoutineWeekdays {
    public static let symbols = ["일", "월", "화", "수", "목", "금", "토"]
    public static let displayOrder = [1, 2, 3, 4, 5, 6, 0]

    public static func bit(_ index: Int) -> Int { 1 << index }

    public static func contains(_ mask: Int, _ index: Int) -> Bool { mask & bit(index) != 0 }

    public static func summary(cadence: RoutineCadence, weekdays: Int, monthDay: Int) -> String {
        switch cadence {
        case .daily:
            return "매일"
        case .weekdays:
            return displayOrder.filter { RoutineWeekdays.contains(weekdays, $0) }.map { symbols[$0] }.joined(separator: "·")
        case .weekly:
            let day = displayOrder.first { RoutineWeekdays.contains(weekdays, $0) } ?? 0
            return "매주 \(symbols[day])요일"
        case .monthly:
            return "매월 \(monthDay)일"
        }
    }
}

public enum RoutineDayFormat {
    public static func string(from date: Date, calendar: Calendar = .current) -> String {
        let parts = calendar.dateComponents([.year, .month, .day], from: date)
        return String(format: "%04d-%02d-%02d", parts.year ?? 0, parts.month ?? 0, parts.day ?? 0)
    }

    public static func title(of day: String) -> String {
        let parts = day.split(separator: "-").compactMap { Int($0) }
        guard parts.count == 3 else { return day }
        return "\(parts[1])월 \(parts[2])일"
    }
}

public struct RoutineStep: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let title: String
    public let position: Int

    public init(id: Int64, title: String, position: Int) {
        self.id = id
        self.title = title
        self.position = position
    }
}

public struct Routine: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let name: String
    public let cadence: RoutineCadence
    public let weekdays: Int
    public let monthDay: Int
    public let startDay: String
    public let steps: [RoutineStep]
    public let createdAt: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case cadence
        case weekdays
        case monthDay = "month_day"
        case startDay = "start_day"
        case steps
        case createdAt = "created_at"
    }

    public var scheduleSummary: String {
        RoutineWeekdays.summary(cadence: cadence, weekdays: weekdays, monthDay: monthDay)
    }
}

public struct RoutineDayStep: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let title: String
    public let position: Int
    public let checked: Bool
}

public struct RoutineDay: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let name: String
    public let cadence: RoutineCadence
    public let weekdays: Int
    public let monthDay: Int
    public let day: String
    public let steps: [RoutineDayStep]
    public let done: Int
    public let total: Int
    public let completed: Bool

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case cadence
        case weekdays
        case monthDay = "month_day"
        case day
        case steps
        case done
        case total
        case completed
    }

    public var scheduleSummary: String {
        RoutineWeekdays.summary(cadence: cadence, weekdays: weekdays, monthDay: monthDay)
    }

    public var progress: Double {
        total > 0 ? Double(done) / Double(total) : 0
    }

    public var progressPercent: Int {
        Int((progress * 100).rounded())
    }
}

public struct RoutineHistoryDay: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let day: String
    public let scheduled: Bool
    public let done: Int
    public let total: Int
    public let completed: Bool

    public var id: String { day }

    public init(day: String, scheduled: Bool, done: Int, total: Int, completed: Bool) {
        self.day = day
        self.scheduled = scheduled
        self.done = done
        self.total = total
        self.completed = completed
    }

    public var statusLabel: String {
        if !scheduled { return "쉬는 날" }
        if completed { return "완료" }
        return "미완료 \(done)/\(total)"
    }
}

public enum RoutineFailureCopy {
    public enum Action: Sendable {
        case load
        case create
        case delete
        case addStep
        case deleteStep
        case check
        case history

        var title: String {
            switch self {
            case .load: return "루틴을 불러오지 못했습니다."
            case .create: return "루틴을 저장하지 못했습니다."
            case .delete: return "루틴을 삭제하지 못했습니다."
            case .addStep: return "단계를 추가하지 못했습니다."
            case .deleteStep: return "단계를 삭제하지 못했습니다."
            case .check: return "단계 체크를 저장하지 못했습니다."
            case .history: return "이력을 불러오지 못했습니다."
            }
        }
    }

    public static func message(for action: Action, error: Error) -> String {
        "\(action.title) \(reason(for: error))"
    }

    private static func reason(for error: Error) -> String {
        guard let apiError = error as? APIError else {
            return "잠시 후 다시 시도해 주세요."
        }
        switch apiError {
        case .transport:
            return "네트워크 연결을 확인해 주세요."
        case .badStatus(let status, _) where status >= 500:
            return "서버가 응답하지 않습니다. 잠시 후 다시 시도해 주세요."
        default:
            return "잠시 후 다시 시도해 주세요."
        }
    }
}

public struct RoutineHistorySummary: Equatable, Sendable {
    public let scheduledDays: Int
    public let completedDays: Int

    public init(_ days: [RoutineHistoryDay]) {
        scheduledDays = days.filter(\.scheduled).count
        completedDays = days.filter(\.completed).count
    }
}

public struct RoutineDraft: Encodable, Equatable, Sendable {
    public let name: String
    public let cadence: RoutineCadence
    public let weekdays: Int
    public let monthDay: Int
    public let startDay: String
    public let steps: [String]

    enum CodingKeys: String, CodingKey {
        case name
        case cadence
        case weekdays
        case monthDay = "month_day"
        case startDay = "start_day"
        case steps
    }

    public init(name: String, cadence: RoutineCadence, weekdays: Int, monthDay: Int, startDay: String, steps: [String]) {
        self.name = name
        self.cadence = cadence
        self.weekdays = weekdays
        self.monthDay = monthDay
        self.startDay = startDay
        self.steps = steps
    }
}

struct RoutineListResponse: Decodable {
    let items: [Routine]
}

struct RoutineDayResponse: Decodable {
    let items: [RoutineDay]
}

struct RoutineHistoryResponse: Decodable {
    let items: [RoutineHistoryDay]
}

struct RoutineStepPayload: Encodable {
    let title: String
}

struct RoutineCheckPayload: Encodable {
    let checked: Bool
}

public extension APIClient {
    func listRoutines() async throws -> [Routine] {
        let response: RoutineListResponse = try await get("/api/routines", authenticated: true)
        return response.items
    }

    func createRoutine(_ draft: RoutineDraft) async throws -> Routine {
        try await post("/api/routines", body: draft, authenticated: true)
    }

    func deleteRoutine(id: Int64) async throws {
        try await delete("/api/routines/\(id)", authenticated: true)
    }

    func addRoutineStep(routineID: Int64, title: String) async throws -> Routine {
        try await post("/api/routines/\(routineID)/steps", body: RoutineStepPayload(title: title), authenticated: true)
    }

    func deleteRoutineStep(routineID: Int64, stepID: Int64) async throws {
        try await delete("/api/routines/\(routineID)/steps/\(stepID)", authenticated: true)
    }

    func listRoutines(on day: String) async throws -> [RoutineDay] {
        let response: RoutineDayResponse = try await get("/api/routines/days/\(day)", authenticated: true)
        return response.items
    }

    func setRoutineStep(routineID: Int64, stepID: Int64, day: String, checked: Bool) async throws -> RoutineDay {
        try await patch(
            "/api/routines/\(routineID)/days/\(day)/steps/\(stepID)",
            body: RoutineCheckPayload(checked: checked),
            authenticated: true
        )
    }

    func routineHistory(routineID: Int64, endingOn day: String) async throws -> [RoutineHistoryDay] {
        let response: RoutineHistoryResponse = try await get("/api/routines/\(routineID)/history/\(day)", authenticated: true)
        return response.items
    }
}
