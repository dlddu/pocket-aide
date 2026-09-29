import Foundation

/// PRD-3 splits todos into two areas backed by separate server collections.
/// The raw value is the path segment of `/api/todos/{area}`.
public enum TodoArea: String, Codable, CaseIterable, Sendable, Hashable {
    case personal
    case work
}

public enum TodoPriority: String, Codable, CaseIterable, Sendable, Hashable {
    case high
    case normal
    case low

    public var displayName: String {
        switch self {
        case .high: return "우선"
        case .normal: return "보통"
        case .low: return "낮음"
        }
    }
}

public struct TodoItem: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let title: String
    public let memo: String
    /// `YYYY-MM-DD`, nil when the todo has no deadline.
    public let dueDate: String?
    public let priority: TodoPriority?
    public let completedAt: Int64?
    public let createdAt: Int64
    public let updatedAt: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case memo
        case dueDate = "due_date"
        case priority
        case completedAt = "completed_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    public init(
        id: Int64,
        title: String,
        memo: String = "",
        dueDate: String? = nil,
        priority: TodoPriority? = nil,
        completedAt: Int64? = nil,
        createdAt: Int64 = 0,
        updatedAt: Int64 = 0
    ) {
        self.id = id
        self.title = title
        self.memo = memo
        self.dueDate = dueDate
        self.priority = priority
        self.completedAt = completedAt
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }

    public var isDone: Bool { completedAt != nil }
}

/// Everything the user can edit on a todo. PATCH overwrites all of it.
public struct TodoDraft: Encodable, Equatable, Sendable {
    public var title: String
    public var memo: String
    public var dueDate: String?
    public var priority: TodoPriority?
    public var done: Bool

    enum CodingKeys: String, CodingKey {
        case title
        case memo
        case dueDate = "due_date"
        case priority
        case done
    }

    public init(title: String, memo: String = "", dueDate: String? = nil, priority: TodoPriority? = nil, done: Bool = false) {
        self.title = title
        self.memo = memo
        self.dueDate = dueDate
        self.priority = priority
        self.done = done
    }

    public init(_ item: TodoItem) {
        self.init(title: item.title, memo: item.memo, dueDate: item.dueDate, priority: item.priority, done: item.isDone)
    }
}

/// Converts between the wire format (`YYYY-MM-DD`) and a calendar `Date`.
public enum TodoDueDate {
    private static func formatter() -> DateFormatter {
        let f = DateFormatter()
        f.calendar = Calendar(identifier: .gregorian)
        f.locale = Locale(identifier: "en_US_POSIX")
        f.timeZone = .current
        f.dateFormat = "yyyy-MM-dd"
        return f
    }

    public static func string(from date: Date) -> String {
        formatter().string(from: date)
    }

    public static func date(from string: String) -> Date? {
        formatter().date(from: string)
    }
}

/// Client-side search over one area's list. The list itself only ever holds
/// that area's items (the server returns one collection per request), so a
/// search can never surface the other area's todos (PRD-3 AC1).
public enum TodoSearch {
    public static func filter(_ items: [TodoItem], query: String) -> [TodoItem] {
        let needle = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !needle.isEmpty else { return items }
        return items.filter {
            $0.title.localizedCaseInsensitiveContains(needle) || $0.memo.localizedCaseInsensitiveContains(needle)
        }
    }
}

struct TodosListResponse: Decodable {
    let items: [TodoItem]
}

public extension APIClient {
    func listTodos(area: TodoArea) async throws -> [TodoItem] {
        let response: TodosListResponse = try await get(
            "/api/todos/\(area.rawValue)",
            authenticated: true
        )
        return response.items
    }

    func createTodo(area: TodoArea, draft: TodoDraft) async throws -> TodoItem {
        try await post("/api/todos/\(area.rawValue)", body: draft, authenticated: true)
    }

    func updateTodo(area: TodoArea, id: Int64, draft: TodoDraft) async throws -> TodoItem {
        try await patch("/api/todos/\(area.rawValue)/\(id)", body: draft, authenticated: true)
    }

    func deleteTodo(area: TodoArea, id: Int64) async throws {
        try await delete("/api/todos/\(area.rawValue)/\(id)", authenticated: true)
    }
}
