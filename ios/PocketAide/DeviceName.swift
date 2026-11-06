import Foundation

/// 푸시 기기 이름. `DeviceNameStore.load()` 가 nil 이면 아직 묻지 않았고, "" 이면 건너뛰었다.
enum DeviceName {
    static let maxLength = 40

    static func normalize(_ raw: String) -> String {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        return String(trimmed.prefix(maxLength)).trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// 서버로 보낼 값 — 건너뛴 경우("")와 묻기 전(nil)은 보내지 않는다.
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
