// Shared data storage backed by UserDefaults with an App Group suite.
// The App Group (group.com.dlddu.pocket-aide) allows the main app, Widget
// Extension, and NotificationServiceExtension to read and write the same values.

import Foundation

// MARK: - AppGroupStorageProtocol

/// Defines the data-sharing contract used by all targets in the App Group.
public protocol AppGroupStorageProtocol: AnyObject {

    /// Persists a `String` value for the given key.
    func set(_ value: String, forKey key: AppGroupKey)

    /// Returns the `String` value stored under `key`, or `nil` if absent.
    func string(forKey key: AppGroupKey) -> String?

    /// Persists an `Int` value for the given key.
    func set(_ value: Int, forKey key: AppGroupKey)

    /// Returns the `Int` value stored under `key`, or `nil` if absent.
    func integer(forKey key: AppGroupKey) -> Int?

    /// Persists a `Date` value for the given key.
    func set(_ value: Date, forKey key: AppGroupKey)

    /// Returns the `Date` value stored under `key`, or `nil` if absent.
    func date(forKey key: AppGroupKey) -> Date?

    /// Removes the value stored under `key`.
    func remove(forKey key: AppGroupKey)
}

// MARK: - AppGroupStorage

/// Live implementation backed by `UserDefaults` with the App Group suite identifier.
///
/// On Linux (CI / unit test hosts without an App Group entitlement) the suite
/// identifier is silently ignored by `UserDefaults`, so this class still compiles
/// and works against the standard user defaults.
///
/// Implementation notes:
/// - `Int` values are wrapped in `NSNumber` before storage so that explicit zero
///   can be distinguished from "key not present" via `object(forKey:)`.
/// - `Date` values are stored as `Double` (timeIntervalSince1970) for portability
///   across Apple and Linux platforms (swift-corelibs-foundation).
public final class AppGroupStorage: AppGroupStorageProtocol {

    // MARK: - Constants

    public static let appGroupIdentifier = "group.com.dlddu.pocket-aide"

    // MARK: - Properties

    private let defaults: UserDefaults

    // MARK: - Init

    /// Creates an `AppGroupStorage` using the shared App Group suite.
    ///
    /// - Parameter suiteName: Defaults to `AppGroupStorage.appGroupIdentifier`.
    ///   Pass a custom suite name in tests to get an isolated domain.
    ///   The initialiser falls back to `UserDefaults.standard` when the suite
    ///   cannot be created (e.g. on Linux without an entitlement).
    public init(suiteName: String? = AppGroupStorage.appGroupIdentifier) {
        self.defaults = UserDefaults(suiteName: suiteName) ?? .standard
    }

    // MARK: - String

    public func set(_ value: String, forKey key: AppGroupKey) {
        defaults.set(value, forKey: key.rawValue)
    }

    public func string(forKey key: AppGroupKey) -> String? {
        defaults.string(forKey: key.rawValue)
    }

    // MARK: - Integer

    /// Stores `value` as an `NSNumber` so that `0` can be distinguished from
    /// "key not present" when reading back with `object(forKey:)`.
    public func set(_ value: Int, forKey key: AppGroupKey) {
        defaults.set(NSNumber(value: value), forKey: key.rawValue)
    }

    public func integer(forKey key: AppGroupKey) -> Int? {
        // `object(forKey:)` returns nil when the key is absent, regardless of
        // the underlying value, so this correctly distinguishes 0 from "absent".
        guard let stored = defaults.object(forKey: key.rawValue) else { return nil }
        if let number = stored as? NSNumber {
            return number.intValue
        }
        // Fallback: handle plain Int stored by older code paths.
        return (stored as? Int)
    }

    // MARK: - Date

    /// Stores `value` as a `Double` (timeIntervalSince1970) to guarantee
    /// round-trip correctness on both Apple platforms and Linux
    /// (swift-corelibs-foundation does not reliably store/retrieve `Date`
    /// via the generic `set(_:forKey:)` / `object(forKey:) as? Date` path).
    public func set(_ value: Date, forKey key: AppGroupKey) {
        defaults.set(value.timeIntervalSince1970, forKey: key.rawValue)
    }

    public func date(forKey key: AppGroupKey) -> Date? {
        guard let stored = defaults.object(forKey: key.rawValue) else { return nil }
        // The value may be stored as Double (our own writes) or as a native Date
        // (written by another target on Apple platforms).
        if let interval = stored as? Double {
            return Date(timeIntervalSince1970: interval)
        }
        if let date = stored as? Date {
            return date
        }
        return nil
    }

    // MARK: - Remove

    public func remove(forKey key: AppGroupKey) {
        defaults.removeObject(forKey: key.rawValue)
    }
}
