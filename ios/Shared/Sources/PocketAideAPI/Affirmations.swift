import Foundation

public enum AffirmationPriority: String, Codable, CaseIterable, Sendable, Hashable {
    case high
    case normal
    case low

    public var displayName: String {
        switch self {
        case .high: return "높음"
        case .normal: return "보통"
        case .low: return "가끔"
        }
    }
}

public struct Affirmation: Codable, Identifiable, Equatable, Sendable, Hashable {
    public let id: Int64
    public let text: String
    public let priority: AffirmationPriority
    public let createdAt: Int64
    public let updatedAt: Int64

    enum CodingKeys: String, CodingKey {
        case id
        case text
        case priority
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    public init(id: Int64, text: String, priority: AffirmationPriority, createdAt: Int64, updatedAt: Int64) {
        self.id = id
        self.text = text
        self.priority = priority
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}

struct AffirmationsListResponse: Decodable {
    let items: [Affirmation]
}

struct AffirmationPayload: Encodable {
    let text: String
    let priority: AffirmationPriority
}

public extension APIClient {
    func listAffirmations() async throws -> [Affirmation] {
        let response: AffirmationsListResponse = try await get(
            "/api/affirmations",
            authenticated: true
        )
        return response.items
    }

    func createAffirmation(text: String, priority: AffirmationPriority) async throws -> Affirmation {
        try await post(
            "/api/affirmations",
            body: AffirmationPayload(text: text, priority: priority),
            authenticated: true
        )
    }

    func updateAffirmation(id: Int64, text: String, priority: AffirmationPriority) async throws -> Affirmation {
        try await patch(
            "/api/affirmations/\(id)",
            body: AffirmationPayload(text: text, priority: priority),
            authenticated: true
        )
    }

    func deleteAffirmation(id: Int64) async throws {
        try await delete("/api/affirmations/\(id)", authenticated: true)
    }
}
