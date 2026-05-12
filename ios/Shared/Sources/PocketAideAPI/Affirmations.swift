import Foundation

public enum Priority: String, Codable, CaseIterable, Sendable {
    case high
    case normal
    case low
}

public struct Affirmation: Codable, Identifiable, Equatable, Sendable {
    public let id: Int64
    public let text: String
    public let priority: Priority
    public let createdAt: Int64
    public let updatedAt: Int64

    public init(id: Int64, text: String, priority: Priority, createdAt: Int64, updatedAt: Int64) {
        self.id = id
        self.text = text
        self.priority = priority
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }

    enum CodingKeys: String, CodingKey {
        case id
        case text
        case priority
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

public struct AffirmationInput: Encodable, Sendable {
    public let text: String
    public let priority: Priority

    public init(text: String, priority: Priority) {
        self.text = text
        self.priority = priority
    }
}

public extension APIClient {
    func listAffirmations() async throws -> [Affirmation] {
        try await get("/api/affirmations", authenticated: true, decodeAs: [Affirmation].self)
    }

    func createAffirmation(text: String, priority: Priority) async throws -> Affirmation {
        try await post(
            "/api/affirmations",
            body: AffirmationInput(text: text, priority: priority),
            decodeAs: Affirmation.self
        )
    }

    func updateAffirmation(id: Int64, text: String, priority: Priority) async throws -> Affirmation {
        try await patch(
            "/api/affirmations/\(id)",
            body: AffirmationInput(text: text, priority: priority),
            decodeAs: Affirmation.self
        )
    }

    func deleteAffirmation(id: Int64) async throws {
        try await delete("/api/affirmations/\(id)")
    }
}
