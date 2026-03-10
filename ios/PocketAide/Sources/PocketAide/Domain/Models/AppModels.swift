// Domain layer models shared across the application.
// These are pure Swift value types with no framework dependencies.

import Foundation

// MARK: - Health

/// Response from GET /health
public struct HealthResponse: Codable, Equatable {
    public let status: String

    public init(status: String) {
        self.status = status
    }
}

// MARK: - API Error

/// Structured error response returned by the backend.
/// Example: {"message": "Unauthorized", "code": 401}
public struct APIErrorResponse: Codable, Equatable {
    public let message: String
    public let code: Int

    public init(message: String, code: Int) {
        self.message = message
        self.code = code
    }
}

// MARK: - Tab

/// Represents one of the seven top-level tabs in the app.
public enum AppTab: CaseIterable, Equatable {
    case chat
    case routine
    case personalTodo
    case workTodo
    case scratchPad
    case notifications
    case quotes

    /// Human-readable display title used for accessibility and tab labels.
    public var title: String {
        switch self {
        case .chat:          return "Chat"
        case .routine:       return "Routine"
        case .personalTodo:  return "Personal"
        case .workTodo:      return "Work"
        case .scratchPad:    return "Scratch"
        case .notifications: return "Alerts"
        case .quotes:        return "Quotes"
        }
    }

    /// SF Symbol name for the tab bar icon.
    public var symbolName: String {
        switch self {
        case .chat:          return "bubble.left.and.bubble.right"
        case .routine:       return "repeat"
        case .personalTodo:  return "person.crop.circle"
        case .workTodo:      return "briefcase"
        case .scratchPad:    return "pencil.and.scribble"
        case .notifications: return "bell"
        case .quotes:        return "quote.bubble"
        }
    }
}

// MARK: - Speech

/// Represents the result of a single speech recognition pass.
public struct SpeechResult: Equatable {
    public let transcript: String
    public let isFinal: Bool

    public init(transcript: String, isFinal: Bool) {
        self.transcript = transcript
        self.isFinal = isFinal
    }
}

// MARK: - App Group

/// Keys used when reading/writing shared data via App Group UserDefaults.
public enum AppGroupKey: String {
    case lastChatMessage  = "lastChatMessage"
    case pendingNoteCount = "pendingNoteCount"
    case lastSyncDate     = "lastSyncDate"
}
