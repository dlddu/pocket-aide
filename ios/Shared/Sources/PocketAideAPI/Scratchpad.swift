import Foundation

/// How an item reached the scratchpad (PRD-4 AC3). The raw value is the wire value.
public enum ScratchpadSource: String, Codable, CaseIterable, Sendable, Hashable {
    case text
    case voice
    case shortcut

    public var displayName: String {
        switch self {
        case .text: return "탭 내 입력"
        case .voice: return "음성"
        case .shortcut: return "숏컷 · 음성"
        }
    }
}

public struct ScratchpadItem: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let text: String
    public let source: ScratchpadSource
    public let capturedAt: Int64
    public let createdAt: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case text
        case source
        case capturedAt = "captured_at"
        case createdAt = "created_at"
    }

    public init(id: Int64, text: String, source: ScratchpadSource = .text, capturedAt: Int64, createdAt: Int64 = 0) {
        self.id = id
        self.text = text
        self.source = source
        self.capturedAt = capturedAt
        self.createdAt = createdAt
    }
}

/// Where a scratchpad item can be moved (PRD-4 AC4). The raw value is the wire value.
public enum ScratchpadMoveTarget: String, Codable, CaseIterable, Sendable, Hashable {
    case personal
    case work
    case affirmation
    case routine

    public var chipLabel: String {
        switch self {
        case .personal: return "→ 개인"
        case .work: return "→ 회사"
        case .affirmation: return "→ 다짐"
        case .routine: return "→ 루틴"
        }
    }
}

public struct ScratchpadMoveResult: Decodable, Equatable, Sendable {
    public let target: ScratchpadMoveTarget
    public let todo: TodoItem?
    public let affirmation: Affirmation?
    public let routine: Routine?
}

/// A day's worth of items, newest day first, as the list shows them.
public struct ScratchpadSection: Equatable, Sendable, Identifiable {
    public let title: String
    public let items: [ScratchpadItem]

    public var id: String { title }
}

public enum ScratchpadSections {
    public static func group(
        _ items: [ScratchpadItem],
        now: Date = Date(),
        calendar: Calendar = .current
    ) -> [ScratchpadSection] {
        let sorted = items.sorted { ($0.capturedAt, $0.id) > ($1.capturedAt, $1.id) }
        var sections: [ScratchpadSection] = []
        var currentDay: Date?
        var bucket: [ScratchpadItem] = []
        for item in sorted {
            let day = calendar.startOfDay(for: Date(timeIntervalSince1970: TimeInterval(item.capturedAt)))
            if day != currentDay, let open = currentDay {
                sections.append(ScratchpadSection(title: title(for: open, now: now, calendar: calendar), items: bucket))
                bucket = []
            }
            currentDay = day
            bucket.append(item)
        }
        if let open = currentDay {
            sections.append(ScratchpadSection(title: title(for: open, now: now, calendar: calendar), items: bucket))
        }
        return sections
    }

    static func title(for day: Date, now: Date, calendar: Calendar) -> String {
        let today = calendar.startOfDay(for: now)
        if day == today { return "오늘" }
        if let yesterday = calendar.date(byAdding: .day, value: -1, to: today), day == yesterday { return "어제" }
        let parts = calendar.dateComponents([.month, .day], from: day)
        return "\(parts.month ?? 0)월 \(parts.day ?? 0)일"
    }
}

struct ScratchpadListResponse: Decodable {
    let items: [ScratchpadItem]
}

struct ScratchpadCreatePayload: Encodable {
    let text: String
    let source: ScratchpadSource
}

struct ScratchpadMovePayload: Encodable {
    let target: ScratchpadMoveTarget
}

public extension APIClient {
    func listScratchpad() async throws -> [ScratchpadItem] {
        let response: ScratchpadListResponse = try await get("/api/scratchpad", authenticated: true)
        return response.items
    }

    func createScratchpadItem(text: String, source: ScratchpadSource = .text) async throws -> ScratchpadItem {
        try await post("/api/scratchpad", body: ScratchpadCreatePayload(text: text, source: source), authenticated: true)
    }

    func deleteScratchpadItem(id: Int64) async throws {
        try await delete("/api/scratchpad/\(id)", authenticated: true)
    }

    func moveScratchpadItem(id: Int64, to target: ScratchpadMoveTarget) async throws -> ScratchpadMoveResult {
        try await post("/api/scratchpad/\(id)/move", body: ScratchpadMovePayload(target: target), authenticated: true)
    }
}
