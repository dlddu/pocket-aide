import Foundation

public enum AgentSessionInteraction: String, Decodable, Sendable, Hashable {
    case agent
    case terminal

    public init(from decoder: Decoder) throws {
        let raw = try decoder.singleValueContainer().decode(String.self)
        self = AgentSessionInteraction(rawValue: raw) ?? .agent
    }
}

public struct AgentSessionModelCatalog: Decodable, Equatable, Sendable, Hashable {
    public let defaultModel: String
    public let models: [String]

    public init(defaultModel: String, models: [String]) {
        self.defaultModel = defaultModel
        self.models = models
    }

    enum CodingKeys: String, CodingKey { case defaultModel, models }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        defaultModel = try container.decodeIfPresent(String.self, forKey: .defaultModel) ?? ""
        models = try container.decodeIfPresent([String].self, forKey: .models) ?? []
    }

    public var initialPick: String? {
        guard !models.isEmpty else { return nil }
        return models.contains(defaultModel) ? defaultModel : models[0]
    }
}

public struct AgentSessionWorkloadType: Decodable, Identifiable, Equatable, Sendable, Hashable {
    public let type: String
    public let label: String?
    public let description: String?
    public let interaction: AgentSessionInteraction
    public let model: AgentSessionModelCatalog?

    public init(
        type: String,
        label: String? = nil,
        description: String? = nil,
        interaction: AgentSessionInteraction = .agent,
        model: AgentSessionModelCatalog? = nil
    ) {
        self.type = type
        self.label = label
        self.description = description
        self.interaction = interaction
        self.model = model
    }

    enum CodingKeys: String, CodingKey { case type, label, description, interaction, model }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        type = try container.decode(String.self, forKey: .type)
        label = try container.decodeIfPresent(String.self, forKey: .label)
        description = try container.decodeIfPresent(String.self, forKey: .description)
        interaction = try container.decodeIfPresent(AgentSessionInteraction.self, forKey: .interaction) ?? .agent
        model = try container.decodeIfPresent(AgentSessionModelCatalog.self, forKey: .model)
    }

    public var id: String { type }
    public var displayName: String { label.flatMap { $0.isEmpty ? nil : $0 } ?? type }
    public var takesModel: Bool { model != nil }
}

public struct AgentSessionConfig: Decodable, Equatable, Sendable {
    public let workloadTypes: [AgentSessionWorkloadType]

    public init(workloadTypes: [AgentSessionWorkloadType]) {
        self.workloadTypes = workloadTypes
    }

    enum CodingKeys: String, CodingKey { case workloadTypes, claudeCode }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let listed = try container.decodeIfPresent([AgentSessionWorkloadType].self, forKey: .workloadTypes) ?? []
        if listed.isEmpty, let legacy = try container.decodeIfPresent(AgentSessionModelCatalog.self, forKey: .claudeCode) {
            workloadTypes = [AgentSessionWorkloadType(type: "claude-code", model: legacy)]
        } else {
            workloadTypes = listed
        }
    }

    public func type(_ id: String) -> AgentSessionWorkloadType? {
        workloadTypes.first { $0.type == id }
    }
}
