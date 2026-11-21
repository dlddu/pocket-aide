import Foundation

enum DeviceName {
    static let maxLength = 40

    static func normalize(_ raw: String) -> String {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        return String(trimmed.prefix(maxLength)).trimmingCharacters(in: .whitespacesAndNewlines)
    }

    static func forRegistration(_ stored: String?) -> String? {
        guard let stored, !stored.isEmpty else { return nil }
        return stored
    }
}

struct DeviceNameStore {
    private let defaults: UserDefaults
    private let key = "pocketaide.device.name"

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    func load() -> String? { defaults.string(forKey: key) }

    func save(_ name: String) { defaults.set(name, forKey: key) }
}
